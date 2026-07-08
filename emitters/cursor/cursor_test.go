package cursor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/universal-governance/ugc/engine/models"
	"github.com/universal-governance/ugc/engine/policy"
)

func TestEmitter(t *testing.T) {
	gov := &models.Governance{
		BaseRules:  "Approval Gate: ask for approval before destructive actions.\nProtected Surfaces: do not touch neighboring systems.\nWorklog: update Plans/worklog.md.",
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

	hooksData, err := os.ReadFile(filepath.Join(tmpDir, ".cursor", "hooks.json"))
	if err != nil {
		t.Fatalf("Expected .cursor/hooks.json to exist: %v", err)
	}
	var hooks cursorHooksFile
	if err := json.Unmarshal(hooksData, &hooks); err != nil {
		t.Fatalf("hooks.json is not valid JSON: %v", err)
	}
	if hooks.Version != 1 {
		t.Fatalf("hooks.json version = %d, want 1", hooks.Version)
	}
	for _, event := range []string{"beforeShellExecution", "beforeReadFile"} {
		defs, ok := hooks.Hooks[event]
		if !ok || len(defs) != 1 {
			t.Fatalf("hooks.json missing %s hook definition: %+v", event, hooks.Hooks)
		}
		if defs[0].Command != denyHookRel {
			t.Fatalf("%s hook command = %q, want %q", event, defs[0].Command, denyHookRel)
		}
		if !defs[0].FailClosed {
			t.Fatalf("%s hook must set failClosed=true", event)
		}
	}

	scriptPath := filepath.Join(tmpDir, denyHookRel)
	scriptData, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("Expected %s to exist: %v", denyHookRel, err)
	}
	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("stat hook script failed: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Fatal("hook script is not executable")
	}
	script := string(scriptData)
	if !strings.Contains(script, "# UGC-Source-Hash: testhash123") {
		t.Fatal("hook script missing source hash header")
	}
	for _, pattern := range policy.CursorShellCommandDenySubstrings() {
		if !strings.Contains(script, pattern) {
			t.Errorf("hook script missing shell deny pattern %q", pattern)
		}
	}
	for _, pattern := range []string{"git push", "git commit", "git reset", "gh release"} {
		if strings.Contains(script, fmt.Sprintf(`== *"%s"*`, pattern)) {
			t.Errorf("hook script must not hard-deny approval-gated command %q", pattern)
		}
	}
	if strings.Contains(script, `case " $command " in *git push*)`) {
		t.Fatal("hook script still uses broken case pattern for shell commands")
	}
	if err := exec.Command("bash", "-n", scriptPath).Run(); err != nil {
		t.Fatalf("generated hook script failed bash -n: %v", err)
	}
	if got := runHookScript(t, scriptPath, `{"command":"ls -la"}`); !strings.Contains(got, `"permission":"allow"`) {
		t.Fatalf("benign command: want allow, got %q", got)
	}
	for _, cmd := range []string{
		`{"command":"git push origin main"}`,
		`{"command":"git commit -m test"}`,
		`{"command":"gh release create v1.0.6"}`,
	} {
		if got := runHookScript(t, scriptPath, cmd); !strings.Contains(got, `"permission":"allow"`) {
			t.Fatalf("%s: want allow, got %q", cmd, got)
		}
	}
	if got := runHookScript(t, scriptPath, `{"command":"rm -rf /tmp/x"}`); !strings.Contains(got, `"permission":"deny"`) {
		t.Fatalf("rm -rf: want deny, got %q", got)
	}
	if got := runHookScript(t, scriptPath, `{"path":"/tmp/project/.env"}`); !strings.Contains(got, `"permission":"deny"`) {
		t.Fatalf(".env read: want deny, got %q", got)
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
	hooksData2, err := os.ReadFile(filepath.Join(tmpDir2, ".cursor", "hooks.json"))
	if err != nil {
		t.Fatalf("Expected second hooks.json to exist: %v", err)
	}
	if string(hooksData) != string(hooksData2) {
		t.Fatal("hooks.json output is not deterministic")
	}
}

func runHookScript(t *testing.T, scriptPath, input string) string {
	t.Helper()
	cmd := exec.Command("bash", scriptPath)
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("hook script execution failed: %v", err)
	}
	return strings.TrimSpace(string(out))
}
