package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/universal-governance/ugc/engine"
)

func TestBuildDryRunJSONCleanFixture(t *testing.T) {
	root := buildAuditCLIFixture(t)
	t.Chdir(root)

	buf := &bytes.Buffer{}
	oldOut := buildCmd.OutOrStdout()
	buildCmd.SetOut(buf)
	defer buildCmd.SetOut(oldOut)

	buildDryRun = true
	buildJSON = true
	defer func() {
		buildDryRun = false
		buildJSON = false
	}()

	if err := buildCmd.RunE(buildCmd, nil); err != nil {
		t.Fatalf("build --dry-run --json failed: %v", err)
	}

	var got buildDryRunJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, buf.String())
	}

	if got.SchemaVersion != buildJSONSchemaVersion {
		t.Fatalf("schema_version = %d, want %d", got.SchemaVersion, buildJSONSchemaVersion)
	}
	if !got.DryRun {
		t.Fatal("dry_run = false, want true")
	}
	if got.HasBlockers {
		t.Fatal("has_blockers = true, want false")
	}
	if len(got.Items) == 0 {
		t.Fatal("expected build plan items")
	}
	if got.Summary[string(engine.BuildStatusCreate)] == 0 && got.Summary[string(engine.BuildStatusUnchanged)] == 0 {
		t.Fatalf("expected summary counts, got %+v", got.Summary)
	}
	for _, key := range []string{
		string(engine.BuildStatusCreate),
		string(engine.BuildStatusUnchanged),
		string(engine.BuildStatusManagedOverwrite),
		string(engine.BuildStatusBlockedUnmanaged),
	} {
		if _, ok := got.Summary[key]; !ok {
			t.Fatalf("summary missing stable key %q: %+v", key, got.Summary)
		}
	}
}

func TestBuildDryRunJSONBlockedFixture(t *testing.T) {
	root := buildAuditCLIFixture(t)
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("human-authored codex guidance"), 0644); err != nil {
		t.Fatalf("write unmanaged AGENTS.md: %v", err)
	}
	t.Chdir(root)

	buf := &bytes.Buffer{}
	oldOut := buildCmd.OutOrStdout()
	buildCmd.SetOut(buf)
	defer buildCmd.SetOut(oldOut)

	buildDryRun = true
	buildJSON = true
	defer func() {
		buildDryRun = false
		buildJSON = false
	}()

	if err := buildCmd.RunE(buildCmd, nil); err == nil {
		t.Fatal("expected build --dry-run --json to fail when blockers exist")
	}

	var got buildDryRunJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, buf.String())
	}
	if !got.HasBlockers {
		t.Fatal("has_blockers = false, want true")
	}
	if got.Summary[string(engine.BuildStatusBlockedUnmanaged)] == 0 {
		t.Fatalf("expected blocked-unmanaged in summary, got %+v", got.Summary)
	}
}

func TestBuildApplyJSONAfterSOPEdit(t *testing.T) {
	root := buildAuditCLIFixture(t)
	sopPath := filepath.Join(root, engine.GovernanceDir, "SOPs", "UGC_TEST_SOP.md")
	original, err := os.ReadFile(sopPath)
	if err != nil {
		t.Fatalf("read sop: %v", err)
	}
	edited := string(original) + "\n\nGUI round-trip edit marker.\n"
	if err := os.WriteFile(sopPath, []byte(edited), 0644); err != nil {
		t.Fatalf("write sop edit: %v", err)
	}
	t.Chdir(root)

	buf := &bytes.Buffer{}
	oldOut := buildCmd.OutOrStdout()
	buildCmd.SetOut(buf)
	defer buildCmd.SetOut(oldOut)

	buildDryRun = false
	buildJSON = true
	defer func() {
		buildDryRun = false
		buildJSON = false
	}()

	if err := buildCmd.RunE(buildCmd, nil); err != nil {
		t.Fatalf("build --json failed: %v", err)
	}

	var got buildDryRunJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, buf.String())
	}
	if got.DryRun {
		t.Fatal("dry_run = true, want false")
	}
	if got.HasBlockers {
		t.Fatal("has_blockers = true, want false")
	}
	if got.Summary[string(engine.BuildStatusManagedOverwrite)] == 0 && got.Summary[string(engine.BuildStatusCreate)] == 0 {
		t.Fatalf("expected apply plan changes, got summary %+v", got.Summary)
	}
}

func TestBuildJSONRejectsRestore(t *testing.T) {
	buildJSON = true
	buildRestorePath = "AGENTS.md"
	defer func() {
		buildJSON = false
		buildRestorePath = ""
	}()

	if err := buildCmd.RunE(buildCmd, nil); err == nil {
		t.Fatal("expected error when --json is combined with --restore")
	}
}

func TestBuildApplyHumanOutputRegression(t *testing.T) {
	root := buildAuditCLIFixture(t)
	sopPath := filepath.Join(root, engine.GovernanceDir, "SOPs", "UGC_TEST_SOP.md")
	original, err := os.ReadFile(sopPath)
	if err != nil {
		t.Fatalf("read sop: %v", err)
	}
	if err := os.WriteFile(sopPath, append(original, []byte("\nHuman apply regression marker.\n")...), 0644); err != nil {
		t.Fatalf("write sop edit: %v", err)
	}
	t.Chdir(root)

	buf := &bytes.Buffer{}
	oldOut := buildCmd.OutOrStdout()
	buildCmd.SetOut(buf)
	defer buildCmd.SetOut(oldOut)

	buildDryRun = false
	buildJSON = false
	defer func() { buildDryRun = false }()

	if err := buildCmd.RunE(buildCmd, nil); err != nil {
		t.Fatalf("build apply failed: %v", err)
	}

	out := buf.String()
	mustContainLine(t, out, "Building governance targets...")
	mustContainLine(t, out, "Build plan:")
	mustContainLine(t, out, "Build complete.")
	if strings.Contains(out, "\"schema_version\"") {
		t.Fatal("human apply output must not contain JSON")
	}
}

func TestBuildDryRunHumanOutputRegression(t *testing.T) {
	root := buildAuditCLIFixture(t)
	t.Chdir(root)

	buf := &bytes.Buffer{}
	oldOut := buildCmd.OutOrStdout()
	buildCmd.SetOut(buf)
	defer buildCmd.SetOut(oldOut)

	buildDryRun = true
	buildJSON = false
	defer func() { buildDryRun = false }()

	if err := buildCmd.RunE(buildCmd, nil); err != nil {
		t.Fatalf("build --dry-run failed: %v", err)
	}

	out := buf.String()
	mustContainLine(t, out, "Building governance targets...")
	mustContainLine(t, out, "Build plan:")
	mustContainLine(t, out, "Dry run complete. No files written.")
	if strings.Contains(out, "\"schema_version\"") {
		t.Fatal("human dry-run output must not contain JSON")
	}
}
