package codex

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/universal-governance/ugc/engine/models"
)

func TestEmitter(t *testing.T) {
	gov := &models.Governance{
		BaseRules:  "Approval Gate: ask for aprobare before destructive actions.\nRead `SOPs/README.md` before material work.\nProtected Surfaces: do not touch neighboring systems.\nWorklog: update Plans/worklog.md.",
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

	outputFile := filepath.Join(tmpDir, "AGENTS.md")
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Expected AGENTS.md to exist: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "<!-- UGC-Source-Hash: testhash123 -->") {
		t.Error("AGENTS.md missing source hash header")
	}
	for _, want := range []string{"OpenAI Codex", "Approval Gate", "Stop Conditions", "Protected Surfaces", "Worklog", "Destructive action warning"} {
		if !strings.Contains(content, want) {
			t.Errorf("AGENTS.md missing governance concept %q", want)
		}
	}
	if strings.Contains(content, "`SOPs/README.md`") {
		t.Error("AGENTS.md kept source-relative SOP path instead of generated root path")
	}
	if !strings.Contains(content, "`.universal-governance/SOPs/README.md`") {
		t.Error("AGENTS.md missing generated root SOP path")
	}

	tmpDir2 := t.TempDir()
	if err := e.Emit(gov, tmpDir2); err != nil {
		t.Fatalf("second Emit failed: %v", err)
	}
	data2, err := os.ReadFile(filepath.Join(tmpDir2, "AGENTS.md"))
	if err != nil {
		t.Fatalf("Expected second AGENTS.md to exist: %v", err)
	}
	if string(data) != string(data2) {
		t.Fatal("AGENTS.md output is not deterministic")
	}

	configData, err := os.ReadFile(filepath.Join(tmpDir, ".codex", "config.toml"))
	if err != nil {
		t.Fatalf("Expected .codex/config.toml to exist: %v", err)
	}
	configContent := string(configData)
	for _, want := range []string{
		"# UGC-Source-Hash: testhash123",
		"approval_policy = \"on-request\"",
		"approvals_reviewer = \"user\"",
		"sandbox_mode = \"workspace-write\"",
	} {
		if !strings.Contains(configContent, want) {
			t.Errorf("config.toml missing %q", want)
		}
	}

	rulesData, err := os.ReadFile(filepath.Join(tmpDir, ".codex", "rules", "ugc.rules"))
	if err != nil {
		t.Fatalf("Expected .codex/rules/ugc.rules to exist: %v", err)
	}
	rulesContent := string(rulesData)
	for _, want := range []string{
		"# UGC-Source-Hash: testhash123",
		"pattern = [\"git\", \"push\"]",
		"pattern = [\"rm\", \"-rf\"]",
		"decision = \"forbidden\"",
	} {
		if !strings.Contains(rulesContent, want) {
			t.Errorf("ugc.rules missing %q", want)
		}
	}

	configData2, err := os.ReadFile(filepath.Join(tmpDir2, ".codex", "config.toml"))
	if err != nil {
		t.Fatalf("Expected second .codex/config.toml to exist: %v", err)
	}
	if string(configData) != string(configData2) {
		t.Fatal("config.toml output is not deterministic")
	}
	rulesData2, err := os.ReadFile(filepath.Join(tmpDir2, ".codex", "rules", "ugc.rules"))
	if err != nil {
		t.Fatalf("Expected second .codex/rules/ugc.rules to exist: %v", err)
	}
	if string(rulesData) != string(rulesData2) {
		t.Fatal("ugc.rules output is not deterministic")
	}
}
