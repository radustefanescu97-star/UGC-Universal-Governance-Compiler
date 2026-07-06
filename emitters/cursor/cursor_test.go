package cursor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/universal-governance/ugc/engine/models"
)

func TestEmitter(t *testing.T) {
	gov := &models.Governance{
		BaseRules:  "Approval Gate: ask for aprobare before destructive actions.\nProtected Surfaces: do not touch neighboring systems.\nWorklog: update Plans/worklog.md.",
		SourceHash: "testhash123",
		SOPs: []models.SOP{
			{Name: "UGC_TEST_SOP.md", Content: "Stop Conditions: stop before deploy.\nDestructive action warning: no rm without approval.\nWorklog duty: append session evidence."},
		},
	}

	tmpDir := t.TempDir()
	e := &Emitter{}
	if err := e.Emit(gov, tmpDir); err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	outputFile := filepath.Join(tmpDir, ".cursorrules")
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Expected .cursorrules to exist: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "// UGC-Source-Hash: testhash123") {
		t.Error(".cursorrules missing source hash header")
	}
	for _, want := range []string{"Approval Gate", "Protected Surfaces", "Worklog", "Stop Conditions", "Destructive action warning", "Worklog duty"} {
		if !strings.Contains(content, want) {
			t.Errorf(".cursorrules missing governance concept %q", want)
		}
	}

	tmpDir2 := t.TempDir()
	if err := e.Emit(gov, tmpDir2); err != nil {
		t.Fatalf("second Emit failed: %v", err)
	}
	data2, err := os.ReadFile(filepath.Join(tmpDir2, ".cursorrules"))
	if err != nil {
		t.Fatalf("Expected second .cursorrules to exist: %v", err)
	}
	if string(data) != string(data2) {
		t.Fatal(".cursorrules output is not deterministic")
	}
}
