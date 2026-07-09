package antigravity

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/universal-governance/ugc/engine/models"
)

func TestEmitter(t *testing.T) {
	worklogSkill := "---\nname: ugc-worklog-sync\ndescription: Mandates syncing worklog evidence.\n---\n\n# UGC Worklog Sync Skill\n\nWorklog duty: append session evidence."
	gov := &models.Governance{
		BaseRules:  "Approval Gate: ask for approval before destructive actions.\nProtected Surfaces: do not touch neighboring systems.\nWorklog: update Plans/worklog.md.",
		SourceHash: "testhash123",
		SOPs: []models.SOP{
			{Name: "UGC_RELEASE_SOP.md", Content: "# Universal Release and Promotion SOP\n\nStop Conditions: stop before deploy.\nDestructive action warning: no rm without approval."},
			{Name: "UGC_WORKLOG_SYNC_SKILL.md", Content: worklogSkill},
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

	skillFile := filepath.Join(tmpDir, ".agents", "skills", "ugc-release-sop", "SKILL.md")
	skillData, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("Expected SKILL.md to exist: %v", err)
	}
	skillContent := string(skillData)
	for _, want := range []string{
		"---\nname: ugc-release-sop\n",
		"description: \"Use when applying UGC guidance from Universal Release and Promotion SOP.\"",
		"Stop Conditions",
		"Destructive action warning",
	} {
		if !strings.Contains(skillContent, want) {
			t.Errorf("SKILL.md missing governance concept %q", want)
		}
	}
	if !hasSkillFrontmatter(skillContent) {
		t.Fatal("generated Antigravity skill must include Codex-compatible frontmatter")
	}

	worklogSkillFile := filepath.Join(tmpDir, ".agents", "skills", "ugc-worklog-sync", "SKILL.md")
	worklogSkillData, err := os.ReadFile(worklogSkillFile)
	if err != nil {
		t.Fatalf("Expected worklog SKILL.md to exist: %v", err)
	}
	if string(worklogSkillData) != worklogSkill {
		t.Fatal("existing skill frontmatter should be preserved without double wrapping")
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
	skillData2, err := os.ReadFile(filepath.Join(tmpDir2, ".agents", "skills", "ugc-release-sop", "SKILL.md"))
	if err != nil {
		t.Fatalf("Expected second SKILL.md to exist: %v", err)
	}
	if string(skillData) != string(skillData2) {
		t.Fatal("SKILL.md output is not deterministic")
	}
}
