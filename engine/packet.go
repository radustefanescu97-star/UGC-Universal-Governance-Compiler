package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ApprovalPacketOptions struct {
	TaskID         string
	Path           string
	SourcePath     string
	MasterplanPath string
}

type ParsedApprovalPacket struct {
	TaskID     string
	PacketPath string
}

type ApprovalVerification struct {
	OK         bool
	TaskID     string
	PacketPath string
	SHA256     string
	Reasons    []string
}

func ApprovalPacketTemplate(opts ApprovalPacketOptions) string {
	taskID := valueOrPlaceholder(opts.TaskID, "<TASK_ID>")
	packetPath := filepath.ToSlash(valueOrPlaceholder(opts.Path, "<PACKET_PATH>"))
	sourcePath := filepath.ToSlash(valueOrPlaceholder(opts.SourcePath, "<SOURCE_PATH>"))
	masterplanPath := filepath.ToSlash(valueOrPlaceholder(opts.MasterplanPath, "<MASTERPLAN_PATH>"))

	var b strings.Builder
	fmt.Fprintf(&b, "# Approval Packet: %s\n\n", taskID)
	fmt.Fprintf(&b, "**Task ID:** `%s`\n", taskID)
	fmt.Fprintf(&b, "**Packet Path:** `%s`\n", packetPath)
	b.WriteString("**Packet SHA256:** computed from this saved file and supplied in the approval sentence; if this file changes after hash calculation, approval is stale.\n")
	fmt.Fprintf(&b, "**Connected source truth:** `%s`\n", sourcePath)
	fmt.Fprintf(&b, "**Connected masterplan:** `%s`\n\n", masterplanPath)
	b.WriteString("## 1. Objective\n\n")
	b.WriteString("State the smallest approved objective.\n\n")
	b.WriteString("## 2. Governance Basis\n\n")
	b.WriteString("- UGC_APPROVAL_PACKET_SOP.md\n")
	b.WriteString("- UGC_GOVERNANCE_CHANGE_SOP.md when governance behavior changes\n")
	b.WriteString("- UGC_ENGINEERING_SIMPLICITY_SOP.md\n")
	b.WriteString("- UGC_WORKLOG_AND_SESSION_SOP.md\n")
	b.WriteString("- UGC_CHECKLIST_SOP.md\n")
	b.WriteString("- UGC_WORKTREE_DISCIPLINE_SOP.md\n")
	b.WriteString("- UGC_COMMIT_DEPLOY_PUSH_GUARDRAILS_SOP.md\n\n")
	b.WriteString("If another active SOP is stricter, the stricter SOP wins.\n\n")
	b.WriteString("## 3. Source Truth / Plan Truth / Local Truth\n\n")
	fmt.Fprintf(&b, "- Source truth: %s\n", sourcePath)
	b.WriteString("- Plan truth: this packet\n")
	fmt.Fprintf(&b, "- Product truth: %s\n", masterplanPath)
	b.WriteString("- Local truth: current repository worktree at execution time\n\n")
	b.WriteString("## 4. Target Surface\n\n")
	b.WriteString("List every file or directory path that may be changed.\n\n")
	b.WriteString("## 5. Protected Neighboring Surfaces\n\n")
	b.WriteString("List every neighboring surface that must not be touched.\n\n")
	b.WriteString("## 6. Allowed Actions\n\n")
	b.WriteString("List only the actions authorized by this packet.\n\n")
	b.WriteString("## 7. Forbidden Actions\n\n")
	b.WriteString("List actions that remain forbidden, including git mutation, release, deploy, cloud, secrets, credentials, and actions outside this packet unless separately approved.\n\n")
	b.WriteString("## 8. Validation Plan\n\n")
	b.WriteString("List required tests and functional checks.\n\n")
	b.WriteString("## 9. Stop Conditions\n\n")
	b.WriteString("Stop and report before continuing if source truth, approval, scope, validation, protected surfaces, or SOP requirements conflict.\n\n")
	b.WriteString("## 10. Return Gate\n\n")
	b.WriteString("The implementation is complete only after validation passes, worklog/checklist closure is recorded, residual risks are stated, and no action outside this packet has occurred.\n\n")
	b.WriteString("## 11. Worklog/Checklist Path\n\n")
	b.WriteString("`Plans/worklog.md`\n\n")
	b.WriteString("## 12. Approval Formula\n\n")
	b.WriteString("Use the hash printed by:\n\n")
	b.WriteString("```bash\n")
	fmt.Fprintf(&b, "ugc packet hash %s\n", packetPath)
	b.WriteString("```\n\n")
	b.WriteString("Approval sentence:\n\n")
	b.WriteString("```text\n")
	fmt.Fprintf(&b, "aprobare pentru executarea %s, conform approval packetului %s SHA256 <HASH>. Scope-ul, allowed actions, forbidden actions, stop conditions si Return Gate raman exact cele din packet. Fara actiuni in afara packetului.\n", taskID, packetPath)
	b.WriteString("```\n")
	return b.String()
}

func PacketSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func VerifyApprovalPacket(packetPath, approval string) ApprovalVerification {
	actualPacketPath := cleanPacketPathForCompare(packetPath)
	result := ApprovalVerification{PacketPath: actualPacketPath}

	data, err := os.ReadFile(packetPath)
	if err != nil {
		result.Reasons = append(result.Reasons, fmt.Sprintf("packet read failed: %v", err))
		return result
	}

	hash, err := PacketSHA256(packetPath)
	if err != nil {
		result.Reasons = append(result.Reasons, fmt.Sprintf("packet hash failed: %v", err))
		return result
	}
	result.SHA256 = hash

	parsed := ParseApprovalPacket(string(data))
	result.TaskID = parsed.TaskID
	if parsed.PacketPath != "" {
		result.PacketPath = parsed.PacketPath
	}

	if parsed.TaskID == "" {
		result.Reasons = append(result.Reasons, "packet task id missing")
	}
	if parsed.PacketPath == "" {
		result.Reasons = append(result.Reasons, "packet path missing")
	}
	if parsed.PacketPath != "" && actualPacketPath != cleanPacketPathForCompare(parsed.PacketPath) {
		result.Reasons = append(result.Reasons, "packet argument path mismatch")
	}

	normalizedApproval := normalizeApproval(approval)
	if !strings.Contains(normalizedApproval, "aprobare") && !strings.Contains(normalizedApproval, "aproval") {
		result.Reasons = append(result.Reasons, "approval literal missing")
	}
	if parsed.TaskID != "" && !strings.Contains(normalizedApproval, strings.ToLower(parsed.TaskID)) {
		result.Reasons = append(result.Reasons, "task id mismatch")
	}
	if parsed.PacketPath != "" && !strings.Contains(normalizedApproval, normalizeApproval(parsed.PacketPath)) {
		result.Reasons = append(result.Reasons, "packet path mismatch")
	}
	if !strings.Contains(normalizedApproval, strings.ToLower(hash)) {
		result.Reasons = append(result.Reasons, "packet sha256 mismatch")
	}
	if !hasExactBoundaryStatement(normalizedApproval) {
		result.Reasons = append(result.Reasons, "exact scope/actions/stop/return-gate statement missing")
	}
	if !hasNoOutsideActionsStatement(normalizedApproval) {
		result.Reasons = append(result.Reasons, "no-outside-actions statement missing")
	}

	result.OK = len(result.Reasons) == 0
	return result
}

func ParseApprovalPacket(content string) ParsedApprovalPacket {
	return ParsedApprovalPacket{
		TaskID:     markdownField(content, "Task ID"),
		PacketPath: filepath.ToSlash(markdownField(content, "Packet Path")),
	}
}

func markdownField(content, field string) string {
	prefix := strings.ToLower("**" + field + ":**")
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(strings.ToLower(line), prefix) {
			continue
		}
		value := strings.TrimSpace(line[len(prefix):])
		value = strings.Trim(value, "` \t\r\n")
		return value
	}
	return ""
}

func hasExactBoundaryStatement(approval string) bool {
	required := []string{"allowed actions", "forbidden actions", "stop conditions", "return gate", "exact"}
	for _, phrase := range required {
		if !strings.Contains(approval, phrase) {
			return false
		}
	}
	return strings.Contains(approval, "scope-ul") || strings.Contains(approval, "scope")
}

func hasNoOutsideActionsStatement(approval string) bool {
	return strings.Contains(approval, "fara actiuni in afara packetului") ||
		strings.Contains(approval, "no actions outside the packet") ||
		strings.Contains(approval, "no action outside the packet")
}

func normalizeApproval(value string) string {
	value = strings.ToLower(strings.Join(strings.Fields(value), " "))
	value = strings.ReplaceAll(value, "`", "")
	value = strings.ReplaceAll(value, " /", "/")
	value = strings.ReplaceAll(value, "/ ", "/")
	return value
}

func cleanPacketPathForCompare(path string) string {
	clean := filepath.ToSlash(filepath.Clean(strings.TrimSpace(path)))
	return strings.TrimPrefix(clean, "./")
}

func valueOrPlaceholder(value, placeholder string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return placeholder
	}
	return value
}
