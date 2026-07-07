package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindUnexpectedGeneratedArtifacts(t *testing.T) {
	tmpDir := t.TempDir()
	mustWrite(t, filepath.Join(tmpDir, "AGENTS.md"), "codex")
	mustWrite(t, filepath.Join(tmpDir, ".agents", "AGENTS.md"), "agents")
	mustWrite(t, filepath.Join(tmpDir, ".agents", "skills", "expected", "SKILL.md"), "expected")
	mustWrite(t, filepath.Join(tmpDir, ".agents", "skills", "stale", "SKILL.md"), "stale")
	mustWrite(t, filepath.Join(tmpDir, "CLAUDE.md"), "claude")
	mustWrite(t, filepath.Join(tmpDir, ".claude", "settings.json"), "claude settings")
	mustWrite(t, filepath.Join(tmpDir, ".claude", "settings.local.json"), "claude local settings")
	mustWrite(t, filepath.Join(tmpDir, ".claude", "commands", "local.md"), "claude command")
	mustWrite(t, filepath.Join(tmpDir, ".claude", "agents", "local.md"), "claude agent")
	mustWrite(t, filepath.Join(tmpDir, ".claude", "skills", "local", "SKILL.md"), "claude skill")
	mustWrite(t, filepath.Join(tmpDir, ".claude", "hooks", "local.sh"), "claude hook")
	mustWrite(t, filepath.Join(tmpDir, ".codex", "config.toml"), "codex config")
	mustWrite(t, filepath.Join(tmpDir, ".codex", "rules", "ugc.rules"), "codex rules")
	mustWrite(t, filepath.Join(tmpDir, ".codex", "rules", "local.rules"), "local codex rules")
	mustWrite(t, filepath.Join(tmpDir, ".cursorrules"), "cursor")
	mustWrite(t, filepath.Join(tmpDir, ".cursor", "hooks.json"), "cursor hooks")
	mustWrite(t, filepath.Join(tmpDir, ".cursor", "hooks", "ugc-deny.sh"), "cursor deny hook")
	mustWrite(t, filepath.Join(tmpDir, ".cursor", "hooks", "local.sh"), "cursor local hook")
	mustWrite(t, filepath.Join(tmpDir, ".codexrules"), "codex")

	expected := []string{
		"AGENTS.md",
		".agents/AGENTS.md",
		".agents/skills/expected/SKILL.md",
		"CLAUDE.md",
		".claude/settings.json",
		".codex/config.toml",
		".codex/rules/ugc.rules",
		".cursorrules",
		".cursor/hooks.json",
		".cursor/hooks/ugc-deny.sh",
	}

	unexpected, err := FindUnexpectedGeneratedArtifacts(tmpDir, expected)
	if err != nil {
		t.Fatalf("FindUnexpectedGeneratedArtifacts failed: %v", err)
	}

	want := map[string]bool{
		".agents/skills/stale/SKILL.md": false,
		".codexrules":                   false,
	}
	for _, path := range unexpected {
		if _, ok := want[path]; ok {
			want[path] = true
		}
	}
	for path, found := range want {
		if !found {
			t.Errorf("expected unexpected artifact %s, got %v", path, unexpected)
		}
	}
	for _, path := range []string{
		".claude/settings.local.json",
		".claude/commands/local.md",
		".claude/agents/local.md",
		".claude/skills/local/SKILL.md",
		".claude/hooks/local.sh",
		".codex/rules/local.rules",
		".cursor/hooks/local.sh",
	} {
		if containsString(unexpected, path) {
			t.Errorf("vendor/user artifact %s should not be unexpected: %v", path, unexpected)
		}
	}
}

func TestBuildManifestForOutputs(t *testing.T) {
	tmpDir := t.TempDir()
	mustWrite(t, filepath.Join(tmpDir, "AGENTS.md"), "codex")
	mustWrite(t, filepath.Join(tmpDir, ".agents", "AGENTS.md"), "agents")
	mustWrite(t, filepath.Join(tmpDir, ".agents", "skills", "ugc-worklog-sync", "SKILL.md"), "worklog")
	mustWrite(t, filepath.Join(tmpDir, ".agents", "skills", "ugc-governance", "SKILL.md"), "codex governance")
	mustWrite(t, filepath.Join(tmpDir, "CLAUDE.md"), "claude")
	mustWrite(t, filepath.Join(tmpDir, ".claude", "settings.json"), "claude settings")
	mustWrite(t, filepath.Join(tmpDir, ".codex", "config.toml"), "codex config")
	mustWrite(t, filepath.Join(tmpDir, ".codex", "rules", "ugc.rules"), "codex rules")
	mustWrite(t, filepath.Join(tmpDir, ".cursorrules"), "cursor")
	mustWrite(t, filepath.Join(tmpDir, ".cursor", "hooks.json"), "cursor hooks")
	mustWrite(t, filepath.Join(tmpDir, ".cursor", "hooks", "ugc-deny.sh"), "cursor deny hook")

	manifest, err := BuildManifestForOutputs(tmpDir, "sourcehash")
	if err != nil {
		t.Fatalf("BuildManifestForOutputs failed: %v", err)
	}
	if manifest.SourceHash != "sourcehash" {
		t.Fatalf("unexpected source hash %q", manifest.SourceHash)
	}
	if len(manifest.Artifacts) != 11 {
		t.Fatalf("expected 11 artifacts, got %d", len(manifest.Artifacts))
	}
	if !containsArtifact(manifest.Artifacts, "AGENTS.md", "codex") {
		t.Fatalf("expected Codex AGENTS.md artifact, got %+v", manifest.Artifacts)
	}
	if !containsArtifact(manifest.Artifacts, ".agents/skills/ugc-governance/SKILL.md", "codex") {
		t.Fatalf("expected Codex governance skill artifact, got %+v", manifest.Artifacts)
	}
	if !containsArtifact(manifest.Artifacts, ".codex/rules/ugc.rules", "codex") {
		t.Fatalf("expected Codex rules artifact, got %+v", manifest.Artifacts)
	}
	if !containsArtifact(manifest.Artifacts, ".agents/skills/ugc-worklog-sync/SKILL.md", "antigravity") {
		t.Fatalf("expected Antigravity skill artifact, got %+v", manifest.Artifacts)
	}
	if !containsArtifact(manifest.Artifacts, ".claude/settings.json", "claude") {
		t.Fatalf("expected Claude settings artifact, got %+v", manifest.Artifacts)
	}
	if !containsArtifact(manifest.Artifacts, ".cursor/hooks.json", "cursor") {
		t.Fatalf("expected Cursor hooks.json artifact, got %+v", manifest.Artifacts)
	}
	if !containsString(manifest.EnabledTargets, "codex") {
		t.Fatalf("expected codex enabled target, got %v", manifest.EnabledTargets)
	}
	if manifest.CapabilityMatrix["approval_gates"]["codex"] == "" {
		t.Fatal("expected Codex capability matrix entry")
	}
	if len(manifest.CapabilityMatrix) == 0 {
		t.Fatal("expected capability matrix")
	}
}

func TestCapabilityMatrixHonestV1Labels(t *testing.T) {
	matrix := TargetCapabilityMatrix()
	for _, concept := range []string{"approval_gates", "stop_conditions", "protected_surfaces", "destructive_action_warnings"} {
		if matrix[concept]["claude"] != "constrained" {
			t.Fatalf("expected Claude %s to be constrained, got %q", concept, matrix[concept]["claude"])
		}
		if matrix[concept]["codex"] != "constrained" {
			t.Fatalf("expected Codex %s to be constrained, got %q", concept, matrix[concept]["codex"])
		}
		if matrix[concept]["antigravity"] == "enforced" {
			t.Fatalf("Antigravity must not be labeled machine-enforced for %s", concept)
		}
		if matrix[concept]["cursor"] != "constrained" {
			t.Fatalf("expected Cursor %s to be constrained, got %q", concept, matrix[concept]["cursor"])
		}
	}
	if matrix["secret_read_protection"]["claude"] != "constrained" {
		t.Fatalf("expected Claude secret read protection to be constrained, got %q", matrix["secret_read_protection"]["claude"])
	}
	if matrix["secret_read_protection"]["cursor"] != "constrained" {
		t.Fatalf("expected Cursor secret read protection to be constrained, got %q", matrix["secret_read_protection"]["cursor"])
	}
	if matrix["worklog_duty"]["codex"] != "native-skill" {
		t.Fatalf("expected Codex worklog duty to be native-skill, got %q", matrix["worklog_duty"]["codex"])
	}
}

func TestBuildManifestFindingsDetectsMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	mustWrite(t, filepath.Join(tmpDir, "AGENTS.md"), "codex")
	expected, err := BuildManifestForOutputs(tmpDir, "sourcehash")
	if err != nil {
		t.Fatalf("BuildManifestForOutputs failed: %v", err)
	}

	bad := expected
	bad.Artifacts = append([]GeneratedArtifact(nil), expected.Artifacts...)
	bad.Artifacts[0].SHA256 = "bad"
	if err := WriteBuildManifest(tmpDir, bad); err != nil {
		t.Fatalf("WriteBuildManifest failed: %v", err)
	}

	findings := BuildManifestFindings(tmpDir, expected)
	if len(findings) == 0 {
		t.Fatal("expected manifest mismatch finding")
	}
}

func containsArtifact(artifacts []GeneratedArtifact, path, target string) bool {
	for _, artifact := range artifacts {
		if artifact.Path == path && artifact.Target == target {
			return true
		}
	}
	return false
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}
