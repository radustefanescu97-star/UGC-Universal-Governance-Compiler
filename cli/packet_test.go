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

func TestPacketVerifyJSONPassesValidApproval(t *testing.T) {
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
		t.Fatalf("packet verify --json failed: %v", err)
	}

	var got packetVerifyJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, buf.String())
	}
	if got.SchemaVersion != packetVerifyJSONSchemaVersion {
		t.Fatalf("schema_version = %d, want %d", got.SchemaVersion, packetVerifyJSONSchemaVersion)
	}
	if !got.OK {
		t.Fatalf("ok = false, want true: %+v", got)
	}
	if got.TaskID != "UGC-TEST-001" {
		t.Fatalf("task_id = %q, want UGC-TEST-001", got.TaskID)
	}
	if got.PacketPath != packetArg {
		t.Fatalf("packet_path = %q, want %q", got.PacketPath, packetArg)
	}
	if got.SHA256 != hash {
		t.Fatalf("sha256 = %q, want %q", got.SHA256, hash)
	}
	if got.Reasons == nil {
		t.Fatal("reasons should be a non-nil array")
	}
	if len(got.Reasons) != 0 {
		t.Fatalf("reasons = %v, want empty", got.Reasons)
	}
}

func TestPacketVerifyJSONFailsInvalidApproval(t *testing.T) {
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

	buf := &bytes.Buffer{}
	oldOut := packetVerifyCmd.OutOrStdout()
	packetVerifyCmd.SetOut(buf)
	defer packetVerifyCmd.SetOut(oldOut)

	packetVerifyPath = packetArg
	packetVerifyApproval = "approval for executing UGC-TEST-001, according to approval packet Plans/Wrong.md SHA256 badhash. Scope, allowed actions, forbidden actions, stop conditions, and Return Gate remain exactly as defined in the packet. No actions outside the packet are authorized."
	packetVerifyJSON = true
	defer func() {
		packetVerifyPath = ""
		packetVerifyApproval = ""
		packetVerifyJSON = false
	}()

	if err := packetVerifyCmd.RunE(packetVerifyCmd, nil); err == nil {
		t.Fatal("expected packet verify --json to fail on invalid approval")
	}

	var got packetVerifyJSONOutput
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, buf.String())
	}
	if got.OK {
		t.Fatal("ok = true, want false")
	}
	if len(got.Reasons) == 0 {
		t.Fatal("expected failure reasons")
	}
	if got.TaskID != "" || got.PacketPath != "" || got.SHA256 != "" {
		t.Fatalf("expected empty success fields on failure, got %+v", got)
	}
}

func TestPacketVerifyHumanOutputRegression(t *testing.T) {
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
	packetVerifyJSON = false
	defer func() {
		packetVerifyPath = ""
		packetVerifyApproval = ""
	}()

	if err := packetVerifyCmd.RunE(packetVerifyCmd, nil); err != nil {
		t.Fatalf("packet verify human failed: %v", err)
	}

	out := buf.String()
	mustContainLine(t, out, "Approval verification passed.")
	mustContainLine(t, out, "Task ID: UGC-TEST-001")
	mustContainLine(t, out, "Packet Path: Plans/UGC_Test_Packet.md")
	mustContainLine(t, out, "Packet SHA256: "+hash)
	if strings.Contains(out, "\"schema_version\"") {
		t.Fatal("human verify output must not contain JSON")
	}
}

func mustWritePacketCLI(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}
