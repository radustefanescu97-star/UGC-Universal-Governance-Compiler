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
		"No actions outside the packet are authorized.",
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

	approval := "approval for executing UGC-TEST-001, according to approval packet Plans/UGC_Test_Packet.md SHA256 " + hash + ". Scope, allowed actions, forbidden actions, stop conditions, and Return Gate remain exactly as defined in the packet. No actions outside the packet are authorized."
	result := VerifyApprovalPacket(packetArg, approval)
	if !result.OK {
		t.Fatalf("expected valid approval, got reasons %v", result.Reasons)
	}
}

func TestVerifyApprovalPacketAcceptsAprovalCompatibilityAlias(t *testing.T) {
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

	approval := "aproval for executing UGC-TEST-001, according to approval packet Plans/UGC_Test_Packet.md SHA256 " + hash + ". Scope, allowed actions, forbidden actions, stop conditions, and Return Gate remain exactly as defined in the packet. No actions outside the packet are authorized."
	result := VerifyApprovalPacket(packetArg, approval)
	if !result.OK {
		t.Fatalf("expected compatibility alias to pass, got reasons %v", result.Reasons)
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
		"missing literal":                 "execute UGC-TEST-001 Plans/UGC_Test_Packet.md SHA256 " + hash + ". Scope, allowed actions, forbidden actions, stop conditions, and Return Gate remain exactly as defined in the packet. No actions outside the packet are authorized.",
		"packet reference is not literal": "execute UGC-TEST-001 according to approval packet Plans/UGC_Test_Packet.md SHA256 " + hash + ". Scope, allowed actions, forbidden actions, stop conditions, and Return Gate remain exactly as defined in the packet. No actions outside the packet are authorized.",
		"wrong path":                      "approval for executing UGC-TEST-001, according to approval packet Plans/Wrong.md SHA256 " + hash + ". Scope, allowed actions, forbidden actions, stop conditions, and Return Gate remain exactly as defined in the packet. No actions outside the packet are authorized.",
		"stale hash":                      "approval for executing UGC-TEST-001, according to approval packet Plans/UGC_Test_Packet.md SHA256 badhash. Scope, allowed actions, forbidden actions, stop conditions, and Return Gate remain exactly as defined in the packet. No actions outside the packet are authorized.",
		"missing no outside":              "approval for executing UGC-TEST-001, according to approval packet Plans/UGC_Test_Packet.md SHA256 " + hash + ". Scope, allowed actions, forbidden actions, stop conditions, and Return Gate remain exactly as defined in the packet.",
		"missing exact scope":             "approval for executing UGC-TEST-001, according to approval packet Plans/UGC_Test_Packet.md SHA256 " + hash + ". No actions outside the packet are authorized.",
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

	approval := "approval for executing UGC-TEST-001, according to approval packet Plans/UGC_Test_Packet.md SHA256 " + hash + ". Scope, allowed actions, forbidden actions, stop conditions, and Return Gate remain exactly as defined in the packet. No actions outside the packet are authorized."
	result := VerifyApprovalPacket(filepath.Join("Plans", "Copy.md"), approval)
	if result.OK {
		t.Fatalf("expected copied packet verification to fail")
	}
	if !containsString(result.Reasons, "packet argument path mismatch") {
		t.Fatalf("expected packet argument path mismatch, got %v", result.Reasons)
	}
}
