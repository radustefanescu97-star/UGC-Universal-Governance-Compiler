package claude

import (
	"encoding/json"
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

	outputFile := filepath.Join(tmpDir, "CLAUDE.md")
	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Expected CLAUDE.md to exist: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "// UGC-Source-Hash: testhash123") {
		t.Error("CLAUDE.md missing source hash header")
	}
	for _, want := range []string{"Approval Gate", "Protected Surfaces", "Worklog", "Stop Conditions", "Destructive action warning", "Worklog duty"} {
		if !strings.Contains(content, want) {
			t.Errorf("CLAUDE.md missing governance concept %q", want)
		}
	}

	tmpDir2 := t.TempDir()
	if err := e.Emit(gov, tmpDir2); err != nil {
		t.Fatalf("second Emit failed: %v", err)
	}
	data2, err := os.ReadFile(filepath.Join(tmpDir2, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("Expected second CLAUDE.md to exist: %v", err)
	}
	if string(data) != string(data2) {
		t.Fatal("CLAUDE.md output is not deterministic")
	}

	settingsData, err := os.ReadFile(filepath.Join(tmpDir, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("Expected .claude/settings.json to exist: %v", err)
	}
	var rawSettings map[string]any
	if err := json.Unmarshal(settingsData, &rawSettings); err != nil {
		t.Fatalf("settings.json is not valid JSON: %v", err)
	}
	if _, ok := rawSettings["ugc"]; ok {
		t.Fatal("settings.json must not contain non-standard top-level ugc metadata")
	}
	var settings claudeSettings
	if err := json.Unmarshal(settingsData, &settings); err != nil {
		t.Fatalf("settings.json is not valid JSON: %v", err)
	}
	for _, want := range []string{"Bash(git push)", "Bash(npm publish)", "Bash(rm -rf *)", "Read(./.env)", "Read(~/.ssh/**)"} {
		if !containsString(settings.Permissions.Deny, want) {
			t.Errorf("settings.json missing deny rule %q", want)
		}
	}

	settingsData2, err := os.ReadFile(filepath.Join(tmpDir2, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("Expected second .claude/settings.json to exist: %v", err)
	}
	if string(settingsData) != string(settingsData2) {
		t.Fatal("settings.json output is not deterministic")
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
