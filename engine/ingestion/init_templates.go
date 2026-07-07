package ingestion

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/universal-governance/ugc/core_embed"
	"github.com/universal-governance/ugc/engine"
)

type Manifest struct {
	SchemaVersion int               `json:"schema_version"`
	CorpusVersion string            `json:"corpus_version"`
	TargetList    []string          `json:"target_list"`
	SOPIDs        []string          `json:"sop_ids"`
	CriticalRules []string          `json:"critical_rules"`
	ApprovalGates []string          `json:"approval_gates"`
	Hashes        map[string]string `json:"hashes"`
}

func InitTemplates() error {
	targetDir := ".universal-governance"

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	manifest := Manifest{
		SchemaVersion: 1,
		CorpusVersion: "1.0.0", // Initial version
		TargetList:    append([]string(nil), engine.V1Targets...),
		Hashes:        make(map[string]string),
	}
	sopIDs := map[string]bool{}
	criticalRules := map[string]bool{}
	approvalGates := map[string]bool{}

	err := fs.WalkDir(core_embed.StandardCorpus, "standard_corpus", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel("standard_corpus", path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		destPath := filepath.Join(targetDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		data, err := fs.ReadFile(core_embed.StandardCorpus, path)
		if err != nil {
			return err
		}

		// Calculate hash
		hash := fmt.Sprintf("%x", sha256.Sum256(data))
		manifest.Hashes[relPath] = hash
		deriveManifestConcepts(relPath, data, sopIDs, criticalRules, approvalGates)

		if _, err := os.Stat(destPath); err == nil {
			fmt.Printf("Skipping %s (already exists)\n", destPath)
			return nil
		}

		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return err
		}

		fmt.Printf("Created %s\n", destPath)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to copy templates: %w", err)
	}

	manifest.SOPIDs = sortedKeys(sopIDs)
	manifest.CriticalRules = sortedKeys(criticalRules)
	manifest.ApprovalGates = sortedKeys(approvalGates)

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(filepath.Join(targetDir, "manifest.json"), manifestData, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	if err := engine.SaveState(engine.CorpusState{
		SchemaVersion: engine.StateSchemaVersion,
		CorpusVersion: manifest.CorpusVersion,
		Hashes:        manifest.Hashes,
	}); err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}

	fmt.Println("Initialization complete. Manifest generated.")
	return nil
}

func deriveManifestConcepts(relPath string, data []byte, sopIDs, criticalRules, approvalGates map[string]bool) {
	if strings.HasPrefix(filepath.ToSlash(relPath), "SOPs/") && filepath.Ext(relPath) == ".md" && filepath.Base(relPath) != "README.md" {
		sopIDs[strings.TrimSuffix(filepath.Base(relPath), ".md")] = true
	}

	lower := strings.ToLower(string(data))
	if strings.Contains(lower, "approval") || strings.Contains(lower, "aproval") {
		criticalRules["approval_gates"] = true
		approvalGates["human_approval_literal"] = true
	}
	if strings.Contains(lower, "stop condition") || strings.Contains(lower, "stop and report") || strings.Contains(lower, "stop reason") {
		criticalRules["stop_conditions"] = true
	}
	if strings.Contains(lower, "protected neighboring") || strings.Contains(lower, "protected surface") {
		criticalRules["protected_surfaces"] = true
	}
	if strings.Contains(lower, "worklog") {
		criticalRules["worklog_duty"] = true
	}
	if strings.Contains(lower, "destructive") || strings.Contains(lower, "cost-generating") {
		criticalRules["destructive_action_warnings"] = true
	}
	if strings.Contains(lower, "governance change") {
		approvalGates["governance_change"] = true
	}
	if strings.Contains(lower, "architecture") {
		approvalGates["architecture_mutation"] = true
	}
}

func sortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
