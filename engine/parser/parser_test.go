package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create mock markdown files
	rootFile := filepath.Join(tmpDir, "AGENTS.md")
	os.WriteFile(rootFile, []byte("# Root Rules\n\nSome root rules here."), 0644)

	sopsDir := filepath.Join(tmpDir, "SOPs")
	os.MkdirAll(sopsDir, 0755)
	
	sop1File := filepath.Join(sopsDir, "SOP_1.md")
	os.WriteFile(sop1File, []byte("# SOP 1\n\nApproval Gate: architecture_mutation\n"), 0644)

	gov, err := Parse(tmpDir)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if gov.BaseRules == "" {
		t.Error("Expected BaseRules to be populated")
	}

	if len(gov.SOPs) != 1 {
		t.Errorf("Expected 1 SOP, got %d", len(gov.SOPs))
	}

	if gov.SOPs[0].Name != "SOP_1.md" {
		t.Errorf("Expected SOP name SOP_1.md, got %s", gov.SOPs[0].Name)
	}
}
