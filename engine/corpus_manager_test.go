package engine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateCorpusDryRunWritesNothing(t *testing.T) {
	tmpDir := t.TempDir()
	withWorkingDir(t, tmpDir)

	if err := UpdateCorpus(true); err != nil {
		t.Fatalf("dry-run update failed: %v", err)
	}
	if _, err := os.Stat(GovernanceDir); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create %s", GovernanceDir)
	}
}

func TestUpdateCorpusFailsOnCorruptState(t *testing.T) {
	tmpDir := t.TempDir()
	withWorkingDir(t, tmpDir)
	mustWrite(t, StateFile, "{bad json")

	if err := UpdateCorpus(false); err == nil {
		t.Fatal("expected corrupt state to fail")
	}
}

func TestUpdateCorpusPreservesUnverifiedLegacyEdit(t *testing.T) {
	tmpDir := t.TempDir()
	withWorkingDir(t, tmpDir)
	localPath := filepath.Join(GovernanceDir, "AGENTS.md")
	mustWrite(t, localPath, "local custom governance")

	if err := UpdateCorpus(false); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	data, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("failed reading local file: %v", err)
	}
	if string(data) != "local custom governance" {
		t.Fatal("unverified legacy local edit was overwritten")
	}
}

func TestUpdateCorpusPreservesTrustedLocalEdit(t *testing.T) {
	tmpDir := t.TempDir()
	withWorkingDir(t, tmpDir)

	if err := UpdateCorpus(false); err != nil {
		t.Fatalf("initial update failed: %v", err)
	}

	localPath := filepath.Join(GovernanceDir, "AGENTS.md")
	const localEdit = "trusted local edit"
	mustWrite(t, localPath, localEdit)

	if err := UpdateCorpus(false); err != nil {
		t.Fatalf("second update failed: %v", err)
	}

	data, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("failed reading local file: %v", err)
	}
	if string(data) != localEdit {
		t.Fatal("trusted local edit was overwritten")
	}
}

func TestUpdateCorpusUnchangedPathDoesNotRewriteContent(t *testing.T) {
	tmpDir := t.TempDir()
	withWorkingDir(t, tmpDir)

	if err := UpdateCorpus(false); err != nil {
		t.Fatalf("initial update failed: %v", err)
	}

	localPath := filepath.Join(GovernanceDir, "AGENTS.md")
	before, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("failed reading local file: %v", err)
	}

	writes := 0
	_, err = updateCorpus(false, false, corpusUpdateIO{
		writeFileAtomic: func(filename string, data []byte, perm os.FileMode) error {
			writes++
			return atomicWriteFile(filename, data, perm)
		},
		saveState: func(CorpusState) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("second update failed: %v", err)
	}
	if writes != 0 {
		t.Fatalf("unchanged corpus should not rewrite files, got %d writes", writes)
	}

	after, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("failed reading local file after update: %v", err)
	}
	if string(after) != string(before) {
		t.Fatal("unchanged file content mutated")
	}
}

func TestUpdateCorpusWriteFailureDoesNotSaveState(t *testing.T) {
	tmpDir := t.TempDir()
	withWorkingDir(t, tmpDir)

	saveCalled := false
	writeErr := errors.New("write failed")
	_, err := updateCorpus(false, false, corpusUpdateIO{
		writeFileAtomic: func(filename string, data []byte, perm os.FileMode) error {
			return writeErr
		},
		saveState: func(CorpusState) error {
			saveCalled = true
			return nil
		},
	})
	if !errors.Is(err, writeErr) {
		t.Fatalf("expected write failure, got %v", err)
	}
	if saveCalled {
		t.Fatal("state was saved after corpus write failure")
	}
}

func withWorkingDir(t *testing.T, dir string) {
	t.Helper()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Fatalf("restore chdir failed: %v", err)
		}
	})
}
