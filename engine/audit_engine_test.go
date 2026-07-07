package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/universal-governance/ugc/engine/parser"
)

func TestAuditProjectCleanGeneratedArtifacts(t *testing.T) {
	root := t.TempDir()
	writeAuditSource(t, root)
	buildAuditOutputs(t, root)

	result, err := AuditProject(root)
	if err != nil {
		t.Fatalf("AuditProject failed: %v", err)
	}
	if result.Failed() {
		t.Fatalf("expected clean audit, got %+v", result)
	}
	if !strings.Contains(result.CapabilitySummary, "codex=") {
		t.Fatalf("capability summary missing Codex: %s", result.CapabilitySummary)
	}
	if !containsString(result.ExpectedArtifacts, "AGENTS.md") {
		t.Fatalf("expected Codex artifact in audit paths: %v", result.ExpectedArtifacts)
	}
}

func TestAuditProjectDetectsModifiedGeneratedFile(t *testing.T) {
	root := t.TempDir()
	writeAuditSource(t, root)
	buildAuditOutputs(t, root)
	mustWrite(t, filepath.Join(root, "CLAUDE.md"), "tampered")

	result, err := AuditProject(root)
	if err != nil {
		t.Fatalf("AuditProject failed: %v", err)
	}
	if !hasDrift(result, "CLAUDE.md") {
		t.Fatalf("expected CLAUDE.md drift, got %+v", result.Drift)
	}
}

func TestAuditProjectDetectsMissingGeneratedFile(t *testing.T) {
	root := t.TempDir()
	writeAuditSource(t, root)
	buildAuditOutputs(t, root)
	if err := os.Remove(filepath.Join(root, ".cursorrules")); err != nil {
		t.Fatalf("remove .cursorrules failed: %v", err)
	}

	result, err := AuditProject(root)
	if err != nil {
		t.Fatalf("AuditProject failed: %v", err)
	}
	if !hasDrift(result, ".cursorrules") {
		t.Fatalf("expected .cursorrules drift, got %+v", result.Drift)
	}
}

func TestAuditProjectDetectsModifiedCodexArtifact(t *testing.T) {
	root := t.TempDir()
	writeAuditSource(t, root)
	buildAuditOutputs(t, root)
	mustWrite(t, filepath.Join(root, "AGENTS.md"), "tampered codex")

	result, err := AuditProject(root)
	if err != nil {
		t.Fatalf("AuditProject failed: %v", err)
	}
	if !hasDrift(result, "AGENTS.md") {
		t.Fatalf("expected Codex AGENTS.md drift, got %+v", result.Drift)
	}
}

func TestAuditProjectDetectsModifiedEnforcementArtifact(t *testing.T) {
	root := t.TempDir()
	writeAuditSource(t, root)
	buildAuditOutputs(t, root)
	mustWrite(t, filepath.Join(root, ".claude", "settings.json"), "tampered claude settings")

	result, err := AuditProject(root)
	if err != nil {
		t.Fatalf("AuditProject failed: %v", err)
	}
	if !hasDrift(result, ".claude/settings.json") {
		t.Fatalf("expected Claude settings drift, got %+v", result.Drift)
	}
}

func TestAuditProjectDetectsMissingCodexArtifact(t *testing.T) {
	root := t.TempDir()
	writeAuditSource(t, root)
	buildAuditOutputs(t, root)
	if err := os.Remove(filepath.Join(root, "AGENTS.md")); err != nil {
		t.Fatalf("remove AGENTS.md failed: %v", err)
	}

	result, err := AuditProject(root)
	if err != nil {
		t.Fatalf("AuditProject failed: %v", err)
	}
	if !hasDrift(result, "AGENTS.md") {
		t.Fatalf("expected missing Codex AGENTS.md drift, got %+v", result.Drift)
	}
}

func TestAuditProjectDetectsMissingEnforcementArtifact(t *testing.T) {
	root := t.TempDir()
	writeAuditSource(t, root)
	buildAuditOutputs(t, root)
	if err := os.Remove(filepath.Join(root, ".codex", "rules", "ugc.rules")); err != nil {
		t.Fatalf("remove Codex rules failed: %v", err)
	}

	result, err := AuditProject(root)
	if err != nil {
		t.Fatalf("AuditProject failed: %v", err)
	}
	if !hasDrift(result, ".codex/rules/ugc.rules") {
		t.Fatalf("expected missing Codex rules drift, got %+v", result.Drift)
	}
}

func TestAuditProjectDetectsUnexpectedGeneratedArtifacts(t *testing.T) {
	root := t.TempDir()
	writeAuditSource(t, root)
	buildAuditOutputs(t, root)
	mustWrite(t, filepath.Join(root, ".codexrules"), "stale codex")
	mustWrite(t, filepath.Join(root, ".agents", "skills", "stale", "SKILL.md"), "stale skill")

	result, err := AuditProject(root)
	if err != nil {
		t.Fatalf("AuditProject failed: %v", err)
	}
	for _, want := range []string{".codexrules", ".agents/skills/stale/SKILL.md"} {
		if !hasUnexpected(result, want) {
			t.Fatalf("expected unexpected artifact %s, got %+v", want, result.UnexpectedArtifacts)
		}
	}
}

func TestAuditProjectIgnoresVendorUserFilesInClaudeAndCodexDirs(t *testing.T) {
	root := t.TempDir()
	writeAuditSource(t, root)
	buildAuditOutputs(t, root)
	mustWrite(t, filepath.Join(root, ".claude", "settings.local.json"), "{}")
	mustWrite(t, filepath.Join(root, ".claude", "commands", "local.md"), "local command")
	mustWrite(t, filepath.Join(root, ".claude", "agents", "local.md"), "local agent")
	mustWrite(t, filepath.Join(root, ".claude", "skills", "local", "SKILL.md"), "local skill")
	mustWrite(t, filepath.Join(root, ".claude", "hooks", "local.sh"), "local hook")
	mustWrite(t, filepath.Join(root, ".cursor", "hooks", "local.sh"), "local cursor hook")
	mustWrite(t, filepath.Join(root, ".codex", "rules", "local.rules"), "prefix_rule(pattern = [\"gh\"], decision = \"prompt\")")

	result, err := AuditProject(root)
	if err != nil {
		t.Fatalf("AuditProject failed: %v", err)
	}
	if result.Failed() {
		t.Fatalf("expected vendor/user files to be ignored, got %+v", result)
	}
}

func TestAuditProjectDetectsMissingBuildManifest(t *testing.T) {
	root := t.TempDir()
	writeAuditSource(t, root)
	buildAuditOutputs(t, root)
	if err := os.Remove(filepath.Join(root, BuildManifestPath)); err != nil {
		t.Fatalf("remove build manifest failed: %v", err)
	}

	result, err := AuditProject(root)
	if err != nil {
		t.Fatalf("AuditProject failed: %v", err)
	}
	if len(result.ManifestFindings) == 0 {
		t.Fatalf("expected build manifest finding, got %+v", result)
	}
}

func writeAuditSource(t *testing.T, root string) {
	t.Helper()
	mustWrite(t, filepath.Join(root, GovernanceDir, "AGENTS.md"), "# Root Rules\n\nApproval Gate: ask for approval.\nProtected Surfaces: respect scope.\nRead `SOPs/README.md`.\n")
	mustWrite(t, filepath.Join(root, GovernanceDir, "SOPs", "README.md"), "# SOP Index\n\n- UGC_TEST_SOP.md\n")
	mustWrite(t, filepath.Join(root, GovernanceDir, "SOPs", "UGC_TEST_SOP.md"), "# Test SOP\n\nStop Conditions: stop on conflict.\nDestructive action warning: no destructive action without approval.\nWorklog duty: append evidence.\n")
}

func buildAuditOutputs(t *testing.T, root string) {
	t.Helper()
	gov, err := parser.Parse(filepath.Join(root, GovernanceDir))
	if err != nil {
		t.Fatalf("parse audit source failed: %v", err)
	}
	for _, e := range V1Emitters() {
		if err := e.Emit(gov, root); err != nil {
			t.Fatalf("emit failed: %v", err)
		}
	}
	manifest, err := BuildManifestForOutputs(root, gov.SourceHash)
	if err != nil {
		t.Fatalf("BuildManifestForOutputs failed: %v", err)
	}
	if err := WriteBuildManifest(root, manifest); err != nil {
		t.Fatalf("WriteBuildManifest failed: %v", err)
	}
}

func hasDrift(result AuditResult, path string) bool {
	for _, drift := range result.Drift {
		if drift.Path == path {
			return true
		}
	}
	return false
}

func hasUnexpected(result AuditResult, path string) bool {
	for _, artifact := range result.UnexpectedArtifacts {
		if artifact == path {
			return true
		}
	}
	return false
}
