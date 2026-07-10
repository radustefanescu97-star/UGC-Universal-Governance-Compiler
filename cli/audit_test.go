package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/universal-governance/ugc/engine"
	"github.com/universal-governance/ugc/engine/parser"
)

func TestAuditJSONCleanFixture(t *testing.T) {
	root := buildAuditCLIFixture(t)
	t.Chdir(root)

	buf := &bytes.Buffer{}
	oldOut := auditCmd.OutOrStdout()
	auditCmd.SetOut(buf)
	defer auditCmd.SetOut(oldOut)

	auditJSON = true
	defer func() { auditJSON = false }()

	if err := auditCmd.RunE(auditCmd, nil); err != nil {
		t.Fatalf("audit --json failed: %v", err)
	}

	var got auditJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, buf.String())
	}

	if got.SchemaVersion != auditJSONSchemaVersion {
		t.Fatalf("schema_version = %d, want %d", got.SchemaVersion, auditJSONSchemaVersion)
	}
	if !got.AuditPassed {
		t.Fatalf("audit_passed = false, want true: %+v", got)
	}
	if !got.SourceValid {
		t.Fatalf("source_valid = false, want true")
	}
	if got.SourceHash == "" {
		t.Fatal("expected source_hash to be set")
	}
	if got.CorpusState != "missing" && got.CorpusState != "ok" {
		t.Fatalf("unexpected corpus_state %q", got.CorpusState)
	}
	if len(got.CapabilityCoverage) == 0 {
		t.Fatal("expected capability_coverage map")
	}
	if len(got.ExpectedArtifacts) == 0 {
		t.Fatal("expected expected_artifacts to be populated")
	}
	if !containsString(got.ExpectedArtifacts, "AGENTS.md") {
		t.Fatalf("expected AGENTS.md in expected_artifacts: %v", got.ExpectedArtifacts)
	}
}

func TestAuditJSONDriftFixture(t *testing.T) {
	root := buildAuditCLIFixture(t)
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("tampered"), 0644); err != nil {
		t.Fatalf("write tampered CLAUDE.md: %v", err)
	}
	t.Chdir(root)

	buf := &bytes.Buffer{}
	oldOut := auditCmd.OutOrStdout()
	auditCmd.SetOut(buf)
	defer auditCmd.SetOut(oldOut)

	auditJSON = true
	defer func() { auditJSON = false }()

	err := auditCmd.RunE(auditCmd, nil)
	if err == nil {
		t.Fatal("expected audit --json to fail on drift")
	}

	buf.Reset()
	result, auditErr := engine.AuditProject(".")
	if auditErr != nil {
		t.Fatalf("AuditProject failed: %v", auditErr)
	}
	buf.Reset()
	if err := printAuditJSON(buf, result, false); err != nil {
		t.Fatalf("printAuditJSON failed: %v", err)
	}

	var got auditJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, buf.String())
	}
	if got.AuditPassed {
		t.Fatal("audit_passed = true, want false")
	}
	if len(got.Drift) == 0 {
		t.Fatalf("expected drift entries, got %+v", got)
	}
	if !driftIncludesPath(got.Drift, "CLAUDE.md") {
		t.Fatalf("expected CLAUDE.md drift, got %+v", got.Drift)
	}
}

func TestAuditHumanOutputRegression(t *testing.T) {
	root := buildAuditCLIFixture(t)
	t.Chdir(root)

	buf := &bytes.Buffer{}
	oldOut := auditCmd.OutOrStdout()
	auditCmd.SetOut(buf)
	defer auditCmd.SetOut(oldOut)

	auditJSON = false
	if err := auditCmd.RunE(auditCmd, nil); err != nil {
		t.Fatalf("audit human failed: %v", err)
	}

	out := buf.String()
	mustContainLine(t, out, "Auditing governance configurations...")
	mustContainLine(t, out, "Source validity: ok")
	mustContainLine(t, out, "Source Hash:")
	mustContainLine(t, out, "Corpus state:")
	mustContainLine(t, out, "Target capability coverage:")
	mustContainLine(t, out, "approval_gates:")
	mustContainLine(t, out, "Audit complete. No drift detected.")
}

func TestAuditJSONHardErrorPath(t *testing.T) {
	buf := &bytes.Buffer{}
	result := engine.AuditResult{CorpusState: "unknown"}
	if err := printAuditJSON(buf, result, true); err != nil {
		t.Fatalf("printAuditJSON failed: %v", err)
	}
	var got auditJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if got.AuditPassed {
		t.Fatal("audit_passed must be false on hard error path")
	}
	if got.SourceValid {
		t.Fatal("source_valid must be false on hard error path")
	}
	if got.CorpusState != "unknown" {
		t.Fatalf("corpus_state = %q, want unknown", got.CorpusState)
	}
}

func buildAuditCLIFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mustWriteAuditCLI(t, filepath.Join(root, engine.GovernanceDir, "AGENTS.md"), "# Root Rules\n\nApproval Gate: ask for approval.\nProtected Surfaces: respect scope.\nRead `SOPs/README.md`.\n")
	mustWriteAuditCLI(t, filepath.Join(root, engine.GovernanceDir, "SOPs", "README.md"), "# SOP Index\n\n- UGC_TEST_SOP.md\n")
	mustWriteAuditCLI(t, filepath.Join(root, engine.GovernanceDir, "SOPs", "UGC_TEST_SOP.md"), "# Test SOP\n\nStop Conditions: stop on conflict.\nDestructive action warning: no destructive action without approval.\nWorklog duty: append evidence.\n")

	gov, err := parser.Parse(filepath.Join(root, engine.GovernanceDir))
	if err != nil {
		t.Fatalf("parse fixture failed: %v", err)
	}
	for _, e := range engine.V1Emitters() {
		if err := e.Emit(gov, root); err != nil {
			t.Fatalf("emit fixture failed: %v", err)
		}
	}
	manifest, err := engine.BuildManifestForOutputs(root, gov.SourceHash)
	if err != nil {
		t.Fatalf("BuildManifestForOutputs failed: %v", err)
	}
	if err := engine.WriteBuildManifest(root, manifest); err != nil {
		t.Fatalf("WriteBuildManifest failed: %v", err)
	}
	return root
}

func mustWriteAuditCLI(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}

func mustContainLine(t *testing.T, output, want string) {
	t.Helper()
	if !strings.Contains(output, want) {
		t.Fatalf("output missing %q\n%s", want, output)
	}
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func driftIncludesPath(drifts []auditDriftJSON, path string) bool {
	for _, drift := range drifts {
		if drift.Path == path {
			return true
		}
	}
	return false
}
