package engine

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestApprovalPacketTemplateIncludesRequiredFields(t *testing.T) {
	template := ApprovalPacketTemplate(ApprovalPacketOptions{
		TaskID:         "UGC-TEST-001",
		Path:           "Plans/UGC_Test_Packet.md",
		SourcePath:     "Plans/source.md",
		MasterplanPath: "Plans/Product_Masterplan.md",
	})

	for _, want := range []string{
		"**Task ID:** `UGC-TEST-001`",
		"**Packet Path:** `Plans/UGC_Test_Packet.md`",
		"**Packet SHA256:**",
		"## 6. Allowed Actions",
		"## 7. Forbidden Actions",
		"## 9. Stop Conditions",
		"## 10. Return Gate",
		"Fara actiuni in afara packetului.",
	} {
		if !strings.Contains(template, want) {
			t.Fatalf("packet template missing %q", want)
		}
	}
}

func TestVerifyApprovalPacketPassesValidApproval(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	packetPath := filepath.Join(root, "Plans", "UGC_Test_Packet.md")
	packetArg := filepath.Join("Plans", "UGC_Test_Packet.md")
	template := ApprovalPacketTemplate(ApprovalPacketOptions{
		TaskID:         "UGC-TEST-001",
		Path:           "Plans/UGC_Test_Packet.md",
		SourcePath:     "Plans/source.md",
		MasterplanPath: "Plans/Product_Masterplan.md",
	})
	mustWrite(t, packetPath, template)
	hash, err := PacketSHA256(packetPath)
	if err != nil {
		t.Fatalf("PacketSHA256 failed: %v", err)
	}

	approval := "aprobare pentru executarea UGC-TEST-001, conform approval packetului Plans/UGC_Test_Packet.md SHA256 " + hash + ". Scope-ul, allowed actions, forbidden actions, stop conditions si Return Gate raman exact cele din packet. Fara actiuni in afara packetului."
	result := VerifyApprovalPacket(packetArg, approval)
	if !result.OK {
		t.Fatalf("expected valid approval, got reasons %v", result.Reasons)
	}
}

func TestVerifyApprovalPacketFailsClosed(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	packetPath := filepath.Join(root, "Plans", "UGC_Test_Packet.md")
	packetArg := filepath.Join("Plans", "UGC_Test_Packet.md")
	template := ApprovalPacketTemplate(ApprovalPacketOptions{
		TaskID:         "UGC-TEST-001",
		Path:           "Plans/UGC_Test_Packet.md",
		SourcePath:     "Plans/source.md",
		MasterplanPath: "Plans/Product_Masterplan.md",
	})
	mustWrite(t, packetPath, template)
	hash, err := PacketSHA256(packetPath)
	if err != nil {
		t.Fatalf("PacketSHA256 failed: %v", err)
	}

	cases := map[string]string{
		"missing literal":     "execute UGC-TEST-001 Plans/UGC_Test_Packet.md SHA256 " + hash + ". Scope-ul, allowed actions, forbidden actions, stop conditions si Return Gate raman exact cele din packet. Fara actiuni in afara packetului.",
		"wrong path":          "aprobare pentru executarea UGC-TEST-001, conform approval packetului Plans/Wrong.md SHA256 " + hash + ". Scope-ul, allowed actions, forbidden actions, stop conditions si Return Gate raman exact cele din packet. Fara actiuni in afara packetului.",
		"stale hash":          "aprobare pentru executarea UGC-TEST-001, conform approval packetului Plans/UGC_Test_Packet.md SHA256 badhash. Scope-ul, allowed actions, forbidden actions, stop conditions si Return Gate raman exact cele din packet. Fara actiuni in afara packetului.",
		"missing no outside":  "aprobare pentru executarea UGC-TEST-001, conform approval packetului Plans/UGC_Test_Packet.md SHA256 " + hash + ". Scope-ul, allowed actions, forbidden actions, stop conditions si Return Gate raman exact cele din packet.",
		"missing exact scope": "aprobare pentru executarea UGC-TEST-001, conform approval packetului Plans/UGC_Test_Packet.md SHA256 " + hash + ". Fara actiuni in afara packetului.",
	}

	for name, approval := range cases {
		t.Run(name, func(t *testing.T) {
			result := VerifyApprovalPacket(packetArg, approval)
			if result.OK {
				t.Fatalf("expected invalid approval")
			}
			if len(result.Reasons) == 0 {
				t.Fatalf("expected concise failure reasons")
			}
		})
	}
}

func TestVerifyApprovalPacketFailsWhenArgumentPathDiffersFromDeclaredPath(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	template := ApprovalPacketTemplate(ApprovalPacketOptions{
		TaskID:         "UGC-TEST-001",
		Path:           "Plans/UGC_Test_Packet.md",
		SourcePath:     "Plans/source.md",
		MasterplanPath: "Plans/Product_Masterplan.md",
	})
	copyPath := filepath.Join(root, "Plans", "Copy.md")
	mustWrite(t, copyPath, template)
	hash, err := PacketSHA256(copyPath)
	if err != nil {
		t.Fatalf("PacketSHA256 failed: %v", err)
	}

	approval := "aprobare pentru executarea UGC-TEST-001, conform approval packetului Plans/UGC_Test_Packet.md SHA256 " + hash + ". Scope-ul, allowed actions, forbidden actions, stop conditions si Return Gate raman exact cele din packet. Fara actiuni in afara packetului."
	result := VerifyApprovalPacket(filepath.Join("Plans", "Copy.md"), approval)
	if result.OK {
		t.Fatalf("expected copied packet verification to fail")
	}
	if !containsString(result.Reasons, "packet argument path mismatch") {
		t.Fatalf("expected packet argument path mismatch, got %v", result.Reasons)
	}
}
