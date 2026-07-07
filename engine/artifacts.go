package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const BuildManifestPath = ".universal-governance/build_manifest.json"

var V1Targets = []string{"codex", "antigravity", "claude", "cursor"}

type GeneratedArtifact struct {
	Target string `json:"target"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type BuildManifest struct {
	SchemaVersion    int                          `json:"schema_version"`
	SourceHash       string                       `json:"source_hash"`
	EnabledTargets   []string                     `json:"enabled_targets"`
	Artifacts        []GeneratedArtifact          `json:"artifacts"`
	CapabilityMatrix map[string]map[string]string `json:"capability_matrix"`
}

func TargetCapabilityMatrix() map[string]map[string]string {
	return map[string]map[string]string{
		"approval_gates": {
			"codex":       "constrained",
			"antigravity": "instructed",
			"claude":      "constrained",
			"cursor":      "constrained",
		},
		"stop_conditions": {
			"codex":       "constrained",
			"antigravity": "instructed",
			"claude":      "constrained",
			"cursor":      "constrained",
		},
		"protected_surfaces": {
			"codex":       "constrained",
			"antigravity": "instructed",
			"claude":      "constrained",
			"cursor":      "constrained",
		},
		"destructive_action_warnings": {
			"codex":       "constrained",
			"antigravity": "instructed",
			"claude":      "constrained",
			"cursor":      "constrained",
		},
		"secret_read_protection": {
			"codex":       "advisory",
			"antigravity": "advisory",
			"claude":      "constrained",
			"cursor":      "constrained",
		},
		"worklog_duty": {
			"codex":       "instructed",
			"antigravity": "native-skill",
			"claude":      "instructed",
			"cursor":      "instructed",
		},
	}
}

func BuildManifestForOutputs(rootDir, sourceHash string) (BuildManifest, error) {
	paths, err := CollectV1ArtifactPaths(rootDir)
	if err != nil {
		return BuildManifest{}, err
	}

	artifacts := make([]GeneratedArtifact, 0, len(paths))
	for _, relPath := range paths {
		hash, err := hashPath(filepath.Join(rootDir, filepath.FromSlash(relPath)))
		if err != nil {
			return BuildManifest{}, err
		}
		artifacts = append(artifacts, GeneratedArtifact{
			Target: targetForArtifact(relPath),
			Path:   relPath,
			SHA256: hash,
		})
	}

	return BuildManifest{
		SchemaVersion:    1,
		SourceHash:       sourceHash,
		EnabledTargets:   append([]string(nil), V1Targets...),
		Artifacts:        artifacts,
		CapabilityMatrix: TargetCapabilityMatrix(),
	}, nil
}

func WriteBuildManifest(rootDir string, manifest BuildManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(rootDir, BuildManifestPath)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return atomicWriteFile(path, data, 0644)
}

func ReadBuildManifest(rootDir string) (BuildManifest, error) {
	data, err := os.ReadFile(filepath.Join(rootDir, BuildManifestPath))
	if err != nil {
		return BuildManifest{}, err
	}
	var manifest BuildManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return BuildManifest{}, err
	}
	return manifest, nil
}

func BuildManifestFindings(rootDir string, expected BuildManifest) []string {
	actual, err := ReadBuildManifest(rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{fmt.Sprintf("missing %s", BuildManifestPath)}
		}
		return []string{fmt.Sprintf("failed to read %s: %v", BuildManifestPath, err)}
	}

	var findings []string
	if actual.SchemaVersion != expected.SchemaVersion {
		findings = append(findings, fmt.Sprintf("schema_version mismatch: got %d want %d", actual.SchemaVersion, expected.SchemaVersion))
	}
	if actual.SourceHash != expected.SourceHash {
		findings = append(findings, "source_hash mismatch")
	}
	if !equalStrings(actual.EnabledTargets, expected.EnabledTargets) {
		findings = append(findings, fmt.Sprintf("enabled_targets mismatch: got %v want %v", actual.EnabledTargets, expected.EnabledTargets))
	}
	findings = append(findings, compareArtifacts(actual.Artifacts, expected.Artifacts)...)
	findings = append(findings, compareCapabilityMatrix(actual.CapabilityMatrix, expected.CapabilityMatrix)...)
	return findings
}

func CollectV1ArtifactPaths(rootDir string) ([]string, error) {
	var paths []string

	for _, relPath := range []string{
		"AGENTS.md",
		".agents/AGENTS.md",
		"CLAUDE.md",
		".claude/settings.json",
		".codex/config.toml",
		".codex/rules/ugc.rules",
		".cursorrules",
		".cursor/hooks.json",
		".cursor/hooks/ugc-deny.sh",
	} {
		if fileExists(filepath.Join(rootDir, filepath.FromSlash(relPath))) {
			paths = append(paths, relPath)
		}
	}

	skillsDir := filepath.Join(rootDir, ".agents", "skills")
	if dirExists(skillsDir) {
		err := filepath.WalkDir(skillsDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || d.Name() != "SKILL.md" {
				return nil
			}
			relPath, err := filepath.Rel(rootDir, path)
			if err != nil {
				return err
			}
			paths = append(paths, filepath.ToSlash(relPath))
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	sort.Strings(paths)
	return paths, nil
}

func FindUnexpectedGeneratedArtifacts(rootDir string, expected []string) ([]string, error) {
	expectedSet := make(map[string]bool, len(expected))
	for _, path := range expected {
		expectedSet[path] = true
	}

	actual, err := CollectV1ArtifactPaths(rootDir)
	if err != nil {
		return nil, err
	}

	var unexpected []string
	for _, path := range actual {
		if !expectedSet[path] {
			unexpected = append(unexpected, path)
		}
	}

	for _, path := range []string{".codexrules"} {
		if fileExists(filepath.Join(rootDir, filepath.FromSlash(path))) {
			unexpected = append(unexpected, path)
		}
	}

	sort.Strings(unexpected)
	return unexpected, nil
}

func ArtifactPaths(manifest BuildManifest) []string {
	paths := make([]string, 0, len(manifest.Artifacts))
	for _, artifact := range manifest.Artifacts {
		paths = append(paths, artifact.Path)
	}
	sort.Strings(paths)
	return paths
}

func compareArtifacts(actual, expected []GeneratedArtifact) []string {
	actualByPath := map[string]GeneratedArtifact{}
	for _, artifact := range actual {
		actualByPath[artifact.Path] = artifact
	}
	expectedByPath := map[string]GeneratedArtifact{}
	for _, artifact := range expected {
		expectedByPath[artifact.Path] = artifact
	}

	var findings []string
	for path, want := range expectedByPath {
		got, ok := actualByPath[path]
		if !ok {
			findings = append(findings, fmt.Sprintf("manifest missing artifact %s", path))
			continue
		}
		if got.Target != want.Target || got.SHA256 != want.SHA256 {
			findings = append(findings, fmt.Sprintf("manifest artifact mismatch for %s", path))
		}
	}
	for path := range actualByPath {
		if _, ok := expectedByPath[path]; !ok {
			findings = append(findings, fmt.Sprintf("manifest has unexpected artifact %s", path))
		}
	}
	sort.Strings(findings)
	return findings
}

func compareCapabilityMatrix(actual, expected map[string]map[string]string) []string {
	var findings []string
	for concept, expectedTargets := range expected {
		actualTargets, ok := actual[concept]
		if !ok {
			findings = append(findings, fmt.Sprintf("manifest missing capability concept %s", concept))
			continue
		}
		for target, want := range expectedTargets {
			if got := actualTargets[target]; got != want {
				findings = append(findings, fmt.Sprintf("manifest capability mismatch for %s/%s", concept, target))
			}
		}
	}
	for concept := range actual {
		if _, ok := expected[concept]; !ok {
			findings = append(findings, fmt.Sprintf("manifest has unexpected capability concept %s", concept))
		}
	}
	sort.Strings(findings)
	return findings
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func targetForArtifact(path string) string {
	switch {
	case path == "AGENTS.md":
		return "codex"
	case path == ".agents/AGENTS.md" || strings.HasPrefix(path, ".agents/skills/"):
		return "antigravity"
	case path == "CLAUDE.md" || strings.HasPrefix(path, ".claude/"):
		return "claude"
	case strings.HasPrefix(path, ".codex/"):
		return "codex"
	case path == ".cursorrules" || strings.HasPrefix(path, ".cursor/"):
		return "cursor"
	default:
		return "unknown"
	}
}

func hashPath(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func CapabilitySummary() string {
	matrix := TargetCapabilityMatrix()
	concepts := make([]string, 0, len(matrix))
	for concept := range matrix {
		concepts = append(concepts, concept)
	}
	sort.Strings(concepts)

	var b strings.Builder
	for _, concept := range concepts {
		b.WriteString(fmt.Sprintf("- %s:", concept))
		for _, target := range V1Targets {
			b.WriteString(fmt.Sprintf(" %s=%s", target, matrix[concept][target]))
			if target != V1Targets[len(V1Targets)-1] {
				b.WriteString(",")
			}
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
