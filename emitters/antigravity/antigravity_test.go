package antigravity

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/universal-governance/ugc/engine/models"
)

func TestEmitter(t *testing.T) {
	gov := &models.Governance{
		BaseRules:  "Approval Gate: ask for approval before destructive actions.\nProtected Surfaces: do not touch neighboring systems.\nWorklog: update Plans/worklog.md.",
		SourceHash: "testhash123",
		SOPs: []models.SOP{
			{Name: "UGC_WORKLOG_SYNC_SKILL.md", Content: "Stop Conditions: stop before deploy.\nDestructive action warning: no rm without approval.\nWorklog duty: append session evidence."},
		},
	}

	tmpDir := t.TempDir()
	e := &Emitter{}
	if err := e.Emit(gov, tmpDir); err != nil {
		t.Fatalf("Emit failed: %v", err)
	}

	agentsFile := filepath.Join(tmpDir, ".agents", "AGENTS.md")
	data, err := os.ReadFile(agentsFile)
	if err != nil {
		t.Fatalf("Expected AGENTS.md to exist: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "<!-- UGC-Source-Hash: testhash123 -->") {
		t.Error("AGENTS.md missing source hash header")
	}
	for _, want := range []string{"Approval Gate", "Protected Surfaces", "Worklog"} {
		if !strings.Contains(content, want) {
			t.Errorf("AGENTS.md missing governance concept %q", want)
		}
	}

	skillFile := filepath.Join(tmpDir, ".agents", "skills", "ugc-worklog-sync", "SKILL.md")
	skillData, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("Expected SKILL.md to exist: %v", err)
	}
	skillContent := string(skillData)
	for _, want := range []string{"Stop Conditions", "Destructive action warning", "Worklog duty"} {
		if !strings.Contains(skillContent, want) {
			t.Errorf("SKILL.md missing governance concept %q", want)
		}
	}

	tmpDir2 := t.TempDir()
	if err := e.Emit(gov, tmpDir2); err != nil {
		t.Fatalf("second Emit failed: %v", err)
	}
	data2, err := os.ReadFile(filepath.Join(tmpDir2, ".agents", "AGENTS.md"))
	if err != nil {
		t.Fatalf("Expected second AGENTS.md to exist: %v", err)
	}
	if string(data) != string(data2) {
		t.Fatal("AGENTS.md output is not deterministic")
	}
}
