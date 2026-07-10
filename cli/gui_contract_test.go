package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/universal-governance/ugc/engine"
)

const guiContractTestdata = "testdata/gui_contract"

func TestGUIContractGoldenFilesUnmarshal(t *testing.T) {
	t.Run("audit_clean", func(t *testing.T) {
		var got auditJSONOutput
		loadGUIContractGolden(t, "audit_clean.json", &got)
		assertIntField(t, got.SchemaVersion, "schema_version", auditJSONSchemaVersion)
		if got.CapabilityCoverage == nil {
			t.Fatal("capability_coverage must be an object")
		}
		if got.Drift == nil {
			t.Fatal("drift must be a non-nil array")
		}
	})

	t.Run("build_dry_run", func(t *testing.T) {
		var got buildDryRunJSONOutput
		loadGUIContractGolden(t, "build_dry_run.json", &got)
		assertIntField(t, got.SchemaVersion, "schema_version", buildJSONSchemaVersion)
		if !got.DryRun {
			t.Fatal("dry_run must be true")
		}
		if len(got.Items) == 0 {
			t.Fatal("items must be a non-empty array in golden example")
		}
	})

	t.Run("build_apply", func(t *testing.T) {
		var got buildDryRunJSONOutput
		loadGUIContractGolden(t, "build_apply.json", &got)
		if got.DryRun {
			t.Fatal("dry_run must be false")
		}
		if got.HasBlockers {
			t.Fatal("has_blockers must be false on successful apply golden")
		}
	})

	t.Run("packet_verify_pass", func(t *testing.T) {
		var got packetVerifyJSONOutput
		loadGUIContractGolden(t, "packet_verify_pass.json", &got)
		assertIntField(t, got.SchemaVersion, "schema_version", packetVerifyJSONSchemaVersion)
		if !got.OK {
			t.Fatal("ok must be true")
		}
		if got.Reasons == nil {
			t.Fatal("reasons must be a non-nil array")
		}
	})

	t.Run("update_dry_run", func(t *testing.T) {
		var got updateDryRunJSONOutput
		loadGUIContractGolden(t, "update_dry_run.json", &got)
		assertIntField(t, got.SchemaVersion, "schema_version", updateJSONSchemaVersion)
		if !got.DryRun {
			t.Fatal("dry_run must be true")
		}
		for _, field := range []struct {
			name string
			got  []string
		}{
			{"created", got.Created},
			{"updated", got.Updated},
			{"unchanged", got.Unchanged},
			{"skipped_local_edits", got.SkippedLocalEdits},
			{"skipped_unverified_legacy", got.SkippedUnverifiedLegacy},
			{"failed", got.Failed},
		} {
			if field.got == nil {
				t.Fatalf("%s must be a non-nil array", field.name)
			}
		}
	})

	t.Run("version_no_check", func(t *testing.T) {
		var got versionOutput
		loadGUIContractGolden(t, "version_no_check.json", &got)
		assertIntField(t, got.SchemaVersion, "schema_version", versionJSONSchemaVersion)
		if got.BinaryVersion == "" || got.CorpusVersion == "" {
			t.Fatal("binary_version and corpus_version must be strings")
		}
	})
}

func TestGUIContractLiveShapeMatchesGolden(t *testing.T) {
	t.Run("audit", func(t *testing.T) {
		golden := readGUIContractGoldenBytes(t, "audit_clean.json")
		live := runAuditJSONFixture(t)
		assertJSONTopLevelShapeMatches(t, golden, live)
	})

	t.Run("build_dry_run", func(t *testing.T) {
		golden := readGUIContractGoldenBytes(t, "build_dry_run.json")
		live := runBuildDryRunJSONFixture(t)
		assertJSONTopLevelShapeMatches(t, golden, live)
	})

	t.Run("build_apply", func(t *testing.T) {
		golden := readGUIContractGoldenBytes(t, "build_apply.json")
		live := runBuildApplyJSONFixture(t)
		assertJSONTopLevelShapeMatches(t, golden, live)
	})

	t.Run("packet_verify", func(t *testing.T) {
		golden := readGUIContractGoldenBytes(t, "packet_verify_pass.json")
		live := runPacketVerifyJSONFixture(t)
		assertJSONTopLevelShapeMatches(t, golden, live)
	})

	t.Run("update_dry_run", func(t *testing.T) {
		golden := readGUIContractGoldenBytes(t, "update_dry_run.json")
		live := runUpdateDryRunJSONFixture(t)
		assertJSONTopLevelShapeMatches(t, golden, live)
	})
}

func loadGUIContractGolden(t *testing.T, name string, target any) {
	t.Helper()
	data := readGUIContractGoldenBytes(t, name)
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("golden %s invalid json: %v", name, err)
	}
}

func readGUIContractGoldenBytes(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join(guiContractTestdata, name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", name, err)
	}
	return data
}

func assertIntField(t *testing.T, got int, name string, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %d, want %d", name, got, want)
	}
}

func assertJSONTopLevelShapeMatches(t *testing.T, golden, live []byte) {
	t.Helper()
	goldenTypes := jsonTopLevelTypes(t, golden)
	liveTypes := jsonTopLevelTypes(t, live)
	if !reflect.DeepEqual(goldenTypes, liveTypes) {
		t.Fatalf("top-level shape mismatch\ngolden: %#v\nlive:   %#v", goldenTypes, liveTypes)
	}
}

func jsonTopLevelTypes(t *testing.T, data []byte) map[string]string {
	t.Helper()
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	out := make(map[string]string, len(raw))
	for key, value := range raw {
		out[key] = jsonValueType(value)
	}
	return out
}

func jsonValueType(raw json.RawMessage) string {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return "invalid"
	}
	switch v.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case float64:
		return "number"
	case string:
		return "string"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return "unknown"
	}
}

func runAuditJSONFixture(t *testing.T) []byte {
	t.Helper()
	root := buildAuditCLIFixture(t)
	t.Chdir(root)
	auditJSON = true
	defer func() { auditJSON = false }()

	oldOut := auditCmd.OutOrStdout()
	auditCmd.SetOut(nil)
	defer auditCmd.SetOut(oldOut)

	var runErr error
	stdout := captureRealStdout(t, func() {
		runErr = auditCmd.RunE(auditCmd, nil)
	})
	if runErr != nil {
		t.Fatalf("audit --json fixture failed: %v", runErr)
	}
	assertSingleJSONObjectStdout(t, stdout)
	return stdout
}

func runBuildDryRunJSONFixture(t *testing.T) []byte {
	t.Helper()
	root := buildAuditCLIFixture(t)
	t.Chdir(root)
	buildDryRun = true
	buildJSON = true
	defer func() {
		buildDryRun = false
		buildJSON = false
	}()

	oldOut := buildCmd.OutOrStdout()
	buildCmd.SetOut(nil)
	defer buildCmd.SetOut(oldOut)

	var runErr error
	stdout := captureRealStdout(t, func() {
		runErr = buildCmd.RunE(buildCmd, nil)
	})
	if runErr != nil {
		t.Fatalf("build --dry-run --json fixture failed: %v", runErr)
	}
	assertSingleJSONObjectStdout(t, stdout)
	return stdout
}

func runBuildApplyJSONFixture(t *testing.T) []byte {
	t.Helper()
	root := buildAuditCLIFixture(t)
	sopPath := filepath.Join(root, engine.GovernanceDir, "SOPs", "UGC_TEST_SOP.md")
	original, err := os.ReadFile(sopPath)
	if err != nil {
		t.Fatalf("read sop: %v", err)
	}
	if err := os.WriteFile(sopPath, append(original, []byte("\nContract apply fixture edit.\n")...), 0644); err != nil {
		t.Fatalf("write sop edit: %v", err)
	}
	t.Chdir(root)
	buildDryRun = false
	buildJSON = true
	defer func() {
		buildDryRun = false
		buildJSON = false
	}()

	oldOut := buildCmd.OutOrStdout()
	buildCmd.SetOut(nil)
	defer buildCmd.SetOut(oldOut)

	var runErr error
	stdout := captureRealStdout(t, func() {
		runErr = buildCmd.RunE(buildCmd, nil)
	})
	if runErr != nil {
		t.Fatalf("build --json fixture failed: %v", runErr)
	}
	assertSingleJSONObjectStdout(t, stdout)
	return stdout
}

func runPacketVerifyJSONFixture(t *testing.T) []byte {
	t.Helper()
	root := t.TempDir()
	t.Chdir(root)
	packetPath := filepath.Join(root, "Plans", "UGC_Test_Packet.md")
	packetArg := filepath.Join("Plans", "UGC_Test_Packet.md")
	template := engine.ApprovalPacketTemplate(engine.ApprovalPacketOptions{
		TaskID:         "UGC-TEST-001",
		Path:           "Plans/UGC_Test_Packet.md",
		SourcePath:     "Plans/source.md",
		MasterplanPath: "Plans/Product_Masterplan.md",
	})
	mustWritePacketCLI(t, packetPath, template)
	hash, err := engine.PacketSHA256(packetPath)
	if err != nil {
		t.Fatalf("PacketSHA256 failed: %v", err)
	}
	approval := "approval for executing UGC-TEST-001, according to approval packet Plans/UGC_Test_Packet.md SHA256 " + hash + ". Scope, allowed actions, forbidden actions, stop conditions, and Return Gate remain exactly as defined in the packet. No actions outside the packet are authorized."

	buf := &bytes.Buffer{}
	oldOut := packetVerifyCmd.OutOrStdout()
	packetVerifyCmd.SetOut(buf)
	defer packetVerifyCmd.SetOut(oldOut)
	packetVerifyPath = packetArg
	packetVerifyApproval = approval
	packetVerifyJSON = true
	defer func() {
		packetVerifyPath = ""
		packetVerifyApproval = ""
		packetVerifyJSON = false
	}()
	if err := packetVerifyCmd.RunE(packetVerifyCmd, nil); err != nil {
		t.Fatalf("packet verify --json fixture failed: %v", err)
	}
	return buf.Bytes()
}

func runUpdateDryRunJSONFixture(t *testing.T) []byte {
	t.Helper()
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
		t.Fatalf("update --dry-run --json fixture failed: %v", err)
	}
	return buf.Bytes()
}
