package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/universal-governance/ugc/engine"
)

func TestGUISOPEditRoundTripJSON(t *testing.T) {
	root := buildAuditCLIFixture(t)
	sopPath := filepath.Join(root, engine.GovernanceDir, "SOPs", "UGC_TEST_SOP.md")
	original, err := os.ReadFile(sopPath)
	if err != nil {
		t.Fatalf("read sop: %v", err)
	}
	if err := os.WriteFile(sopPath, append(original, []byte("\nRound-trip governed edit.\n")...), 0644); err != nil {
		t.Fatalf("write sop edit: %v", err)
	}
	t.Chdir(root)

	previewBuf := &bytes.Buffer{}
	oldBuildOut := buildCmd.OutOrStdout()
	buildCmd.SetOut(previewBuf)
	defer buildCmd.SetOut(oldBuildOut)

	buildDryRun = true
	buildJSON = true
	if err := buildCmd.RunE(buildCmd, nil); err != nil {
		t.Fatalf("preview build --dry-run --json failed: %v", err)
	}

	var preview buildDryRunJSONOutput
	if err := json.Unmarshal(previewBuf.Bytes(), &preview); err != nil {
		t.Fatalf("invalid preview json: %v\n%s", err, previewBuf.String())
	}
	if !preview.DryRun {
		t.Fatal("preview dry_run = false, want true")
	}
	if preview.HasBlockers {
		t.Fatal("preview has_blockers = true, want false")
	}
	if preview.Summary[string(engine.BuildStatusManagedOverwrite)] == 0 {
		t.Fatalf("expected managed-overwrite in preview, got %+v", preview.Summary)
	}

	applyBuf := &bytes.Buffer{}
	buildCmd.SetOut(applyBuf)
	buildDryRun = false
	buildJSON = true
	defer func() {
		buildDryRun = false
		buildJSON = false
	}()

	if err := buildCmd.RunE(buildCmd, nil); err != nil {
		t.Fatalf("apply build --json failed: %v", err)
	}

	var applied buildDryRunJSONOutput
	if err := json.Unmarshal(applyBuf.Bytes(), &applied); err != nil {
		t.Fatalf("invalid apply json: %v\n%s", err, applyBuf.String())
	}
	if applied.DryRun {
		t.Fatal("apply dry_run = true, want false")
	}
	if applied.HasBlockers {
		t.Fatal("apply has_blockers = true, want false")
	}

	auditBuf := &bytes.Buffer{}
	oldAuditOut := auditCmd.OutOrStdout()
	auditCmd.SetOut(auditBuf)
	defer auditCmd.SetOut(oldAuditOut)

	auditJSON = true
	defer func() { auditJSON = false }()

	if err := auditCmd.RunE(auditCmd, nil); err != nil {
		t.Fatalf("audit --json failed: %v", err)
	}

	var auditGot auditJSONOutput
	if err := json.Unmarshal(auditBuf.Bytes(), &auditGot); err != nil {
		t.Fatalf("invalid audit json: %v\n%s", err, auditBuf.String())
	}
	if !auditGot.AuditPassed {
		t.Fatalf("audit_passed = false after round-trip, got %+v", auditGot)
	}
	if !auditGot.SourceValid {
		t.Fatal("source_valid = false after round-trip")
	}
}
