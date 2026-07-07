package engine

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/universal-governance/ugc/engine/models"
)

func TestBuildPlanBlocksUnmanagedAgentsAndWritesNothing(t *testing.T) {
	root := t.TempDir()
	generated := t.TempDir()
	mustWrite(t, filepath.Join(root, "AGENTS.md"), "human-authored codex guidance")
	mustWrite(t, filepath.Join(generated, "AGENTS.md"), "<!-- UGC-Source-Hash: new -->\n")
	mustWrite(t, filepath.Join(generated, "CLAUDE.md"), "// UGC-Source-Hash: new\n")

	plan, err := PlanGeneratedBuild(root, generated, "sourcehash")
	if err != nil {
		t.Fatalf("PlanGeneratedBuild failed: %v", err)
	}
	if !plan.HasBlockers() {
		t.Fatalf("expected unmanaged AGENTS.md blocker, got %+v", plan.Items)
	}
	if statusForPath(plan, "AGENTS.md") != BuildStatusBlockedUnmanaged {
		t.Fatalf("expected AGENTS.md blocked, got %+v", plan.Items)
	}

	err = ApplyBuildPlan(root, generated, plan)
	if !errors.Is(err, ErrBuildPlanBlocked) {
		t.Fatalf("expected ErrBuildPlanBlocked, got %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md failed: %v", err)
	}
	if string(data) != "human-authored codex guidance" {
		t.Fatal("unmanaged AGENTS.md was overwritten")
	}
	if _, err := os.Stat(filepath.Join(root, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Fatal("blocked build wrote another generated artifact")
	}
	if _, err := os.Stat(filepath.Join(root, BuildManifestPath)); !os.IsNotExist(err) {
		t.Fatal("blocked build wrote build manifest")
	}
}

func TestBuildPlanDryRunWritesNothing(t *testing.T) {
	root := t.TempDir()
	generated := t.TempDir()
	mustWrite(t, filepath.Join(generated, "AGENTS.md"), "<!-- UGC-Source-Hash: new -->\n")
	mustWrite(t, filepath.Join(generated, "CLAUDE.md"), "// UGC-Source-Hash: new\n")

	plan, err := PlanGeneratedBuild(root, generated, "sourcehash")
	if err != nil {
		t.Fatalf("PlanGeneratedBuild failed: %v", err)
	}
	if len(plan.Items) == 0 {
		t.Fatal("expected reportable build plan")
	}

	for _, path := range []string{"AGENTS.md", "CLAUDE.md", BuildManifestPath} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); !os.IsNotExist(err) {
			t.Fatalf("dry-run planning should not write %s", path)
		}
	}
}

func TestBuildPlanOverwritesManagedAgentsWithSourceHashMarker(t *testing.T) {
	root := t.TempDir()
	generated := t.TempDir()
	mustWrite(t, filepath.Join(root, "AGENTS.md"), "<!-- UGC-Source-Hash: old -->\nold content")
	mustWrite(t, filepath.Join(generated, "AGENTS.md"), "<!-- UGC-Source-Hash: new -->\nnew content")

	plan, err := PlanGeneratedBuild(root, generated, "sourcehash")
	if err != nil {
		t.Fatalf("PlanGeneratedBuild failed: %v", err)
	}
	if statusForPath(plan, "AGENTS.md") != BuildStatusManagedOverwrite {
		t.Fatalf("expected managed overwrite, got %+v", plan.Items)
	}
	if err := ApplyBuildPlan(root, generated, plan); err != nil {
		t.Fatalf("ApplyBuildPlan failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md failed: %v", err)
	}
	if string(data) != "<!-- UGC-Source-Hash: new -->\nnew content" {
		t.Fatal("managed AGENTS.md was not overwritten with generated output")
	}
}

func TestBuildPlanTreatsPriorManifestMatchAsManaged(t *testing.T) {
	root := t.TempDir()
	generated := t.TempDir()
	const oldContent = "old generated content without marker"
	mustWrite(t, filepath.Join(root, "AGENTS.md"), oldContent)
	mustWrite(t, filepath.Join(generated, "AGENTS.md"), "<!-- UGC-Source-Hash: new -->\nnew content")

	prior := BuildManifest{
		SchemaVersion:  1,
		SourceHash:     "oldsource",
		EnabledTargets: append([]string(nil), V1Targets...),
		Artifacts: []GeneratedArtifact{{
			Target: "codex",
			Path:   "AGENTS.md",
			SHA256: sha256String(oldContent),
		}},
		CapabilityMatrix: TargetCapabilityMatrix(),
	}
	if err := WriteBuildManifest(root, prior); err != nil {
		t.Fatalf("WriteBuildManifest failed: %v", err)
	}

	plan, err := PlanGeneratedBuild(root, generated, "sourcehash")
	if err != nil {
		t.Fatalf("PlanGeneratedBuild failed: %v", err)
	}
	if statusForPath(plan, "AGENTS.md") != BuildStatusManagedOverwrite {
		t.Fatalf("expected prior-manifest managed overwrite, got %+v", plan.Items)
	}
	if err := ApplyBuildPlan(root, generated, plan); err != nil {
		t.Fatalf("ApplyBuildPlan failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md failed: %v", err)
	}
	if string(data) != "<!-- UGC-Source-Hash: new -->\nnew content" {
		t.Fatal("manifest-managed AGENTS.md was not overwritten with generated output")
	}
}

func TestApplyBuildPlanWritesAllV1ArtifactsAndManifest(t *testing.T) {
	root := t.TempDir()
	generated := t.TempDir()
	gov := &models.Governance{
		BaseRules:  "Approval Gate: ask for approval.\nProtected Surfaces: stay in scope.\nWorklog: update Plans/worklog.md.",
		SourceHash: "sourcehash",
		SOPs: []models.SOP{
			{Name: "UGC_TEST_SOP.md", Content: "Stop Conditions: stop on conflict.\nDestructive action warning: no destructive write without approval.\nWorklog duty: append evidence."},
		},
	}

	for _, emitter := range V1Emitters() {
		if err := emitter.Emit(gov, generated); err != nil {
			t.Fatalf("emit failed: %v", err)
		}
	}

	plan, err := PlanGeneratedBuild(root, generated, gov.SourceHash)
	if err != nil {
		t.Fatalf("PlanGeneratedBuild failed: %v", err)
	}
	if plan.HasBlockers() {
		t.Fatalf("clean build plan should not block: %+v", plan.Items)
	}
	if err := ApplyBuildPlan(root, generated, plan); err != nil {
		t.Fatalf("ApplyBuildPlan failed: %v", err)
	}

	for _, path := range []string{
		"AGENTS.md",
		".agents/AGENTS.md",
		"CLAUDE.md",
		".claude/settings.json",
		".codex/config.toml",
		".codex/rules/ugc.rules",
		".cursorrules",
		".cursor/hooks.json",
		".cursor/hooks/ugc-deny.sh",
		BuildManifestPath,
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}
	manifest, err := ReadBuildManifest(root)
	if err != nil {
		t.Fatalf("ReadBuildManifest failed: %v", err)
	}
	for _, target := range V1Targets {
		if !containsString(manifest.EnabledTargets, target) {
			t.Fatalf("manifest missing target %s: %v", target, manifest.EnabledTargets)
		}
	}
	if !containsArtifact(manifest.Artifacts, "AGENTS.md", "codex") {
		t.Fatalf("manifest missing Codex AGENTS.md artifact: %+v", manifest.Artifacts)
	}
	if !containsArtifact(manifest.Artifacts, ".codex/config.toml", "codex") {
		t.Fatalf("manifest missing Codex config artifact: %+v", manifest.Artifacts)
	}
	if !containsArtifact(manifest.Artifacts, ".claude/settings.json", "claude") {
		t.Fatalf("manifest missing Claude settings artifact: %+v", manifest.Artifacts)
	}
}

func TestApplyBuildPlanFailureLeavesNoNewManifest(t *testing.T) {
	root := t.TempDir()
	generated := t.TempDir()
	plan := BuildPlan{
		Items: []BuildPlanItem{{
			Path:   "CLAUDE.md",
			Status: BuildStatusCreate,
			Reason: "target file does not exist",
		}},
		Manifest: BuildManifest{
			SchemaVersion:    1,
			SourceHash:       "failedsource",
			EnabledTargets:   append([]string(nil), V1Targets...),
			Artifacts:        []GeneratedArtifact{{Target: "claude", Path: "CLAUDE.md", SHA256: "missing"}},
			CapabilityMatrix: TargetCapabilityMatrix(),
		},
	}

	if err := ApplyBuildPlan(root, generated, plan); err == nil {
		t.Fatal("expected apply failure because generated artifact is missing")
	}
	if _, err := os.Stat(filepath.Join(root, BuildManifestPath)); !os.IsNotExist(err) {
		t.Fatal("failed apply wrote a build manifest")
	}
}

func TestApplyBuildPlanRollsBackCreatedArtifactAfterWriteFailure(t *testing.T) {
	root := t.TempDir()
	generated := t.TempDir()
	mustWrite(t, filepath.Join(generated, "A.md"), "new A")
	mustWrite(t, filepath.Join(generated, "B.md"), "new B")

	err := applyBuildPlan(root, generated, testPlanForItems(
		BuildPlanItem{Path: "A.md", Status: BuildStatusCreate},
		BuildPlanItem{Path: "B.md", Status: BuildStatusCreate},
	), buildPlanIO{writeFileAtomic: failAtomicWriteForBase("B.md")})
	if !errors.Is(err, ErrBuildApplyFailed) {
		t.Fatalf("expected ErrBuildApplyFailed, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "A.md")); !os.IsNotExist(err) {
		t.Fatal("rollback left newly-created A.md behind")
	}
	if _, err := os.Stat(filepath.Join(root, "B.md")); !os.IsNotExist(err) {
		t.Fatal("failed write unexpectedly created B.md")
	}
	if _, err := os.Stat(filepath.Join(root, BuildManifestPath)); !os.IsNotExist(err) {
		t.Fatal("failed apply wrote a build manifest")
	}
}

func TestApplyBuildPlanRollsBackManagedOverwriteAfterWriteFailure(t *testing.T) {
	root := t.TempDir()
	generated := t.TempDir()
	const oldA = "<!-- UGC-Source-Hash: old -->\nold A"
	mustWrite(t, filepath.Join(root, "A.md"), oldA)
	mustWrite(t, filepath.Join(generated, "A.md"), "<!-- UGC-Source-Hash: new -->\nnew A")
	mustWrite(t, filepath.Join(generated, "B.md"), "new B")

	err := applyBuildPlan(root, generated, testPlanForItems(
		BuildPlanItem{Path: "A.md", Status: BuildStatusManagedOverwrite},
		BuildPlanItem{Path: "B.md", Status: BuildStatusCreate},
	), buildPlanIO{writeFileAtomic: failAtomicWriteForBase("B.md")})
	if !errors.Is(err, ErrBuildApplyFailed) {
		t.Fatalf("expected ErrBuildApplyFailed, got %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "A.md"))
	if err != nil {
		t.Fatalf("read restored A.md failed: %v", err)
	}
	if string(data) != oldA {
		t.Fatalf("rollback did not restore managed overwrite, got %q", data)
	}
	if _, err := os.Stat(filepath.Join(root, "B.md")); !os.IsNotExist(err) {
		t.Fatal("failed write unexpectedly created B.md")
	}
	if _, err := os.Stat(filepath.Join(root, BuildManifestPath)); !os.IsNotExist(err) {
		t.Fatal("failed apply wrote a build manifest")
	}
}

func TestApplyBuildPlanRollsBackArtifactsWhenManifestWriteFails(t *testing.T) {
	root := t.TempDir()
	generated := t.TempDir()
	mustWrite(t, filepath.Join(generated, "A.md"), "new A")

	err := applyBuildPlan(root, generated, testPlanForItems(
		BuildPlanItem{Path: "A.md", Status: BuildStatusCreate},
	), buildPlanIO{
		writeBuildManifest: func(rootDir string, manifest BuildManifest) error {
			return errors.New("injected manifest failure")
		},
	})
	if !errors.Is(err, ErrBuildApplyFailed) {
		t.Fatalf("expected ErrBuildApplyFailed, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "A.md")); !os.IsNotExist(err) {
		t.Fatal("manifest failure rollback left created artifact behind")
	}
	if _, err := os.Stat(filepath.Join(root, BuildManifestPath)); !os.IsNotExist(err) {
		t.Fatal("failed manifest write left a manifest")
	}
}

func TestApplyBuildPlanReportsRollbackFailure(t *testing.T) {
	root := t.TempDir()
	generated := t.TempDir()
	mustWrite(t, filepath.Join(generated, "A.md"), "new A")
	mustWrite(t, filepath.Join(generated, "B.md"), "new B")

	err := applyBuildPlan(root, generated, testPlanForItems(
		BuildPlanItem{Path: "A.md", Status: BuildStatusCreate},
		BuildPlanItem{Path: "B.md", Status: BuildStatusCreate},
	), buildPlanIO{
		writeFileAtomic: failAtomicWriteForBase("B.md"),
		remove: func(name string) error {
			if filepath.Base(name) == "A.md" {
				return errors.New("injected rollback remove failure")
			}
			return os.Remove(name)
		},
	})
	if !errors.Is(err, ErrBuildRollbackFailed) {
		t.Fatalf("expected ErrBuildRollbackFailed, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "A.md")); err != nil {
		t.Fatalf("rollback failure test expected A.md to remain: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, BuildManifestPath)); !os.IsNotExist(err) {
		t.Fatal("rollback failure wrote a clean build manifest")
	}
}

func TestRestoreGeneratedArtifactRestoresManifestOwnedPath(t *testing.T) {
	root := t.TempDir()
	generated := t.TempDir()
	mustWrite(t, filepath.Join(generated, ".claude", "settings.json"), "expected settings")
	mustWrite(t, filepath.Join(root, ".claude", "settings.json"), "tampered settings")

	manifest, err := BuildManifestForOutputs(generated, "sourcehash")
	if err != nil {
		t.Fatalf("BuildManifestForOutputs failed: %v", err)
	}
	if err := WriteBuildManifest(root, manifest); err != nil {
		t.Fatalf("WriteBuildManifest failed: %v", err)
	}

	if err := RestoreGeneratedArtifact(root, generated, "sourcehash", ".claude/settings.json"); err != nil {
		t.Fatalf("RestoreGeneratedArtifact failed: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read restored file failed: %v", err)
	}
	if string(data) != "expected settings" {
		t.Fatalf("restore wrote %q", data)
	}
}

func TestRestoreGeneratedArtifactRefusesMissingManifest(t *testing.T) {
	root := t.TempDir()
	generated := t.TempDir()
	mustWrite(t, filepath.Join(generated, ".claude", "settings.json"), "expected settings")

	err := RestoreGeneratedArtifact(root, generated, "sourcehash", ".claude/settings.json")
	if !errors.Is(err, ErrRestoreRefused) {
		t.Fatalf("expected ErrRestoreRefused, got %v", err)
	}
}

func TestRestoreGeneratedArtifactRefusesUnknownOrUnexpectedPaths(t *testing.T) {
	root := t.TempDir()
	generated := t.TempDir()
	mustWrite(t, filepath.Join(generated, ".claude", "settings.json"), "expected settings")
	mustWrite(t, filepath.Join(generated, ".codex", "config.toml"), "codex config")

	manifest, err := BuildManifestForOutputs(generated, "sourcehash")
	if err != nil {
		t.Fatalf("BuildManifestForOutputs failed: %v", err)
	}
	manifest.Artifacts = []GeneratedArtifact{{Target: "claude", Path: ".claude/settings.json", SHA256: "old"}}
	if err := WriteBuildManifest(root, manifest); err != nil {
		t.Fatalf("WriteBuildManifest failed: %v", err)
	}

	err = RestoreGeneratedArtifact(root, generated, "sourcehash", ".codex/config.toml")
	if !errors.Is(err, ErrRestoreRefused) {
		t.Fatalf("expected ErrRestoreRefused for non-manifest path, got %v", err)
	}

	err = RestoreGeneratedArtifact(root, generated, "changedsource", ".claude/settings.json")
	if !errors.Is(err, ErrRestoreRefused) {
		t.Fatalf("expected ErrRestoreRefused for source drift, got %v", err)
	}
}

func TestRestoreGeneratedArtifactRejectsUnsafePaths(t *testing.T) {
	for _, path := range []string{"", "/tmp/file", "../CLAUDE.md", "dir/", "*.md"} {
		t.Run(path, func(t *testing.T) {
			if _, err := normalizeRestorePath(path); !errors.Is(err, ErrRestoreRefused) {
				t.Fatalf("expected ErrRestoreRefused for %q, got %v", path, err)
			}
		})
	}
}

func statusForPath(plan BuildPlan, path string) BuildPlanStatus {
	for _, item := range plan.Items {
		if item.Path == path {
			return item.Status
		}
	}
	return ""
}

func sha256String(content string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
}

func testPlanForItems(items ...BuildPlanItem) BuildPlan {
	return BuildPlan{
		Items: items,
		Manifest: BuildManifest{
			SchemaVersion:    1,
			SourceHash:       "testsource",
			EnabledTargets:   append([]string(nil), V1Targets...),
			CapabilityMatrix: TargetCapabilityMatrix(),
		},
	}
}

func failAtomicWriteForBase(base string) func(filename string, data []byte, perm os.FileMode) error {
	return func(filename string, data []byte, perm os.FileMode) error {
		if filepath.Base(filename) == base {
			return errors.New("injected write failure")
		}
		return atomicWriteFile(filename, data, perm)
	}
}
