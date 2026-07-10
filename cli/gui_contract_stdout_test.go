package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/universal-governance/ugc/engine"
)

func TestJSONSurfacesBuiltBinaryStdout(t *testing.T) {
	binary := buildUGCTestBinary(t)

	t.Run("audit", func(t *testing.T) {
		root := buildAuditCLIFixture(t)
		stdout := execUGC(t, binary, root, "audit", "--json")
		assertSingleJSONObjectStdout(t, stdout)
	})

	t.Run("build_dry_run", func(t *testing.T) {
		root := buildAuditCLIFixture(t)
		stdout := execUGC(t, binary, root, "build", "--dry-run", "--json")
		assertSingleJSONObjectStdout(t, stdout)
	})

	t.Run("build_apply", func(t *testing.T) {
		root := buildAuditCLIFixture(t)
		sopPath := filepath.Join(root, engine.GovernanceDir, "SOPs", "UGC_TEST_SOP.md")
		original, err := os.ReadFile(sopPath)
		if err != nil {
			t.Fatalf("read sop: %v", err)
		}
		if err := os.WriteFile(sopPath, append(original, []byte("\nBinary stdout apply fixture.\n")...), 0644); err != nil {
			t.Fatalf("write sop: %v", err)
		}
		stdout := execUGC(t, binary, root, "build", "--json")
		assertSingleJSONObjectStdout(t, stdout)
		var got buildDryRunJSONOutput
		if err := json.Unmarshal(stdout, &got); err != nil {
			t.Fatalf("unmarshal apply json: %v", err)
		}
		if got.DryRun {
			t.Fatal("dry_run must be false")
		}
	})

	t.Run("update_dry_run", func(t *testing.T) {
		root := t.TempDir()
		stdout := execUGC(t, binary, root, "update", "--dry-run", "--json")
		assertSingleJSONObjectStdout(t, stdout)
	})
}

func TestPacketVerifyMissingFlagsNoUsage(t *testing.T) {
	binary := buildUGCTestBinary(t)
	cmd := exec.Command(binary, "packet", "verify", "--json")
	cmd.Dir = t.TempDir()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected packet verify to fail without required flags")
	}
	errText := stderr.String()
	if strings.Contains(errText, "Usage:") {
		t.Fatalf("stderr must not contain usage block:\n%s", errText)
	}
	if strings.Count(errText, "packet verify requires") != 1 {
		t.Fatalf("expected one clean error line, got:\n%s", errText)
	}
}

func TestRegenerateGUIContractGoldens(t *testing.T) {
	if os.Getenv("UGC_REGENERATE_GOLDENS") != "1" {
		t.Skip("set UGC_REGENERATE_GOLDENS=1 to regenerate golden files")
	}

	binary := buildUGCTestBinary(t)
	root := buildAuditCLIFixture(t)

	writeGoldenFile(t, "audit_clean.json", execUGC(t, binary, root, "audit", "--json"))
	writeGoldenFile(t, "build_dry_run.json", execUGC(t, binary, root, "build", "--dry-run", "--json"))

	sopPath := filepath.Join(root, engine.GovernanceDir, "SOPs", "UGC_TEST_SOP.md")
	original, err := os.ReadFile(sopPath)
	if err != nil {
		t.Fatalf("read sop: %v", err)
	}
	if err := os.WriteFile(sopPath, append(original, []byte("\nGolden regenerate apply edit.\n")...), 0644); err != nil {
		t.Fatalf("write sop: %v", err)
	}
	writeGoldenFile(t, "build_apply.json", execUGC(t, binary, root, "build", "--json"))

	packetRoot := t.TempDir()
	packetPath := filepath.Join(packetRoot, "Plans", "UGC_Test_Packet.md")
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
		t.Fatalf("PacketSHA256: %v", err)
	}
	approval := "approval for executing UGC-TEST-001, according to approval packet Plans/UGC_Test_Packet.md SHA256 " + hash + ". Scope, allowed actions, forbidden actions, stop conditions, and Return Gate remain exactly as defined in the packet. No actions outside the packet are authorized."
	writeGoldenFile(t, "packet_verify_pass.json", execUGC(t, binary, packetRoot, "packet", "verify", "--json", "--packet", packetArg, "--approval", approval))

	writeGoldenFile(t, "update_dry_run.json", execUGC(t, binary, t.TempDir(), "update", "--dry-run", "--json"))
	writeGoldenFile(t, "version_no_check.json", execUGC(t, binary, t.TempDir(), "version", "--json", "--no-check"))
}

func buildUGCTestBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "ugc")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = ugcRepoRoot(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
	return bin
}

func ugcRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if filepath.Base(wd) == "cli" {
		return filepath.Dir(wd)
	}
	if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
		return wd
	}
	t.Fatal("cannot locate ugc repo root from test working directory")
	return ""
}

func execUGC(t *testing.T, binary, dir string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Dir = dir
	stdout, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("%s failed: %v\nstderr: %s", strings.Join(args, " "), err, exitErr.Stderr)
		}
		t.Fatalf("%s failed: %v", strings.Join(args, " "), err)
	}
	return stdout
}

func assertSingleJSONObjectStdout(t *testing.T, stdout []byte) {
	t.Helper()
	trimmed := bytes.TrimSpace(stdout)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		t.Fatalf("stdout must begin with a JSON object, got:\n%s", stdout)
	}
	if bytes.Contains(stdout, []byte("Emitting ")) {
		t.Fatalf("stdout must not contain emitter progress lines:\n%s", stdout)
	}
	var raw map[string]any
	if err := json.Unmarshal(trimmed, &raw); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout)
	}
}

func writeGoldenFile(t *testing.T, name string, data []byte) {
	t.Helper()
	path := filepath.Join(guiContractTestdata, name)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write golden %s: %v", name, err)
	}
}

func captureRealStdout(t *testing.T, fn func()) []byte {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return buf.Bytes()
}
