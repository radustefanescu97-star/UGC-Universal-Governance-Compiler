package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/universal-governance/ugc/engine"
)

func TestUpdateDryRunJSONFixture(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	buf := &bytes.Buffer{}
	oldOut := updateCmd.OutOrStdout()
	updateCmd.SetOut(buf)
	defer updateCmd.SetOut(oldOut)

	dryRun = true
	updateJSON = true
	defer func() {
		dryRun = false
		updateJSON = false
	}()

	if err := updateCmd.RunE(updateCmd, nil); err != nil {
		t.Fatalf("update --dry-run --json failed: %v", err)
	}

	var got updateDryRunJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, buf.String())
	}

	if got.SchemaVersion != updateJSONSchemaVersion {
		t.Fatalf("schema_version = %d, want %d", got.SchemaVersion, updateJSONSchemaVersion)
	}
	if !got.DryRun {
		t.Fatal("dry_run = false, want true")
	}
	if got.StateWarning == "" {
		t.Fatal("expected state_warning for missing state file")
	}
	if got.Summary.Created == 0 {
		t.Fatalf("expected created count > 0, got %+v", got.Summary)
	}
	if got.Created == nil || len(got.Created) == 0 {
		t.Fatal("expected created item list")
	}
	if got.Updated == nil || got.Unchanged == nil || got.SkippedLocalEdits == nil || got.SkippedUnverifiedLegacy == nil || got.Failed == nil {
		t.Fatalf("expected non-nil category lists, got %+v", got)
	}
	if _, err := os.Stat(engine.GovernanceDir); !os.IsNotExist(err) {
		t.Fatalf("dry-run json should not create %s", engine.GovernanceDir)
	}
}

func TestUpdateJSONRequiresDryRunFlag(t *testing.T) {
	updateJSON = true
	dryRun = false
	defer func() {
		updateJSON = false
		dryRun = false
	}()

	if err := updateCmd.RunE(updateCmd, nil); err == nil {
		t.Fatal("expected error when --json is used without --dry-run")
	}
}

func TestUpdateDryRunHumanOutputRegression(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe failed: %v", err)
	}
	os.Stdout = w

	dryRun = true
	updateJSON = false
	defer func() {
		dryRun = false
		os.Stdout = oldStdout
	}()

	runErr := updateCmd.RunE(updateCmd, nil)
	w.Close()
	var outBuf bytes.Buffer
	if _, err := io.Copy(&outBuf, r); err != nil {
		t.Fatalf("read stdout failed: %v", err)
	}
	os.Stdout = oldStdout

	if runErr != nil {
		t.Fatalf("update --dry-run failed: %v", runErr)
	}

	out := outBuf.String()
	mustContainLine(t, out, "State: missing .state.json")
	mustContainLine(t, out, "Dry run complete:")
	if strings.Contains(out, "\"schema_version\"") {
		t.Fatal("human dry-run output must not contain JSON")
	}
}
