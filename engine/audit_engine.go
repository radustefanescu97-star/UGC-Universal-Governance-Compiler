package engine

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	sourceaudit "github.com/universal-governance/ugc/engine/audit"
	"github.com/universal-governance/ugc/engine/parser"
)

type AuditDrift struct {
	Path    string
	Message string
}

type AuditResult struct {
	SourceHash          string
	SourceErrors        []string
	Drift               []AuditDrift
	UnexpectedArtifacts []string
	ManifestFindings    []string
	CorpusState         string
	CorpusStateMessage  string
	CapabilitySummary   string
	ExpectedArtifacts   []string
}

func (r AuditResult) Failed() bool {
	return len(r.SourceErrors) > 0 ||
		len(r.Drift) > 0 ||
		len(r.UnexpectedArtifacts) > 0 ||
		len(r.ManifestFindings) > 0 ||
		r.CorpusState == "failed"
}

func AuditProject(rootDir string) (AuditResult, error) {
	if rootDir == "" {
		rootDir = "."
	}

	result := AuditResult{
		CorpusState:       "unknown",
		CapabilitySummary: CapabilitySummary(),
	}
	sourceDir := filepath.Join(rootDir, GovernanceDir)

	sourceErrs := sourceaudit.ValidateSourceStructure(sourceDir)
	if len(sourceErrs) > 0 {
		for _, err := range sourceErrs {
			result.SourceErrors = append(result.SourceErrors, err.Error())
		}
		return result, nil
	}

	hash, err := sourceaudit.GenerateSourceHash(sourceDir)
	if err != nil {
		return result, fmt.Errorf("generate source hash: %w", err)
	}
	result.SourceHash = hash

	gov, err := parser.Parse(sourceDir)
	if err != nil {
		return result, fmt.Errorf("parse corpus: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "ugc-audit-*")
	if err != nil {
		return result, fmt.Errorf("create audit temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	for _, e := range V1Emitters() {
		if err := e.Emit(gov, tmpDir); err != nil {
			return result, fmt.Errorf("emit expected target: %w", err)
		}
	}

	expectedPaths, err := CollectV1ArtifactPaths(tmpDir)
	if err != nil {
		return result, fmt.Errorf("collect expected generated artifacts: %w", err)
	}
	result.ExpectedArtifacts = expectedPaths

	if err := compareGeneratedArtifacts(rootDir, tmpDir, &result); err != nil {
		return result, err
	}

	unexpected, err := FindUnexpectedGeneratedArtifacts(rootDir, expectedPaths)
	if err != nil {
		return result, fmt.Errorf("check unexpected generated artifacts: %w", err)
	}
	result.UnexpectedArtifacts = unexpected

	expectedManifest, err := BuildManifestForOutputs(tmpDir, gov.SourceHash)
	if err != nil {
		return result, fmt.Errorf("build expected manifest: %w", err)
	}
	result.ManifestFindings = BuildManifestFindings(rootDir, expectedManifest)

	if _, err := LoadStateForRoot(rootDir); err != nil {
		switch {
		case errors.Is(err, ErrStateMissing):
			result.CorpusState = "missing"
			result.CorpusStateMessage = "missing .state.json (legacy/unverified)"
		case errors.Is(err, ErrStateLegacy):
			result.CorpusState = "legacy"
			result.CorpusStateMessage = "legacy .state.json schema (legacy/unverified)"
		default:
			result.CorpusState = "failed"
			result.CorpusStateMessage = err.Error()
		}
	} else {
		result.CorpusState = "ok"
		result.CorpusStateMessage = "ok"
	}

	return result, nil
}

func compareGeneratedArtifacts(rootDir, expectedDir string, result *AuditResult) error {
	return filepath.WalkDir(expectedDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(expectedDir, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		expectedData, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		actualPath := filepath.Join(rootDir, filepath.FromSlash(relPath))
		actualData, err := os.ReadFile(actualPath)
		if err != nil {
			if os.IsNotExist(err) {
				result.Drift = append(result.Drift, AuditDrift{
					Path:    relPath,
					Message: fmt.Sprintf("Missing file %s", relPath),
				})
				return nil
			}
			return err
		}

		if !bytes.Equal(expectedData, actualData) {
			result.Drift = append(result.Drift, AuditDrift{
				Path:    relPath,
				Message: fmt.Sprintf("File %s has been modified (mismatch against expected generation).", relPath),
			})
		}
		return nil
	})
}
