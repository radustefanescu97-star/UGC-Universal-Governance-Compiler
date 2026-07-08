package policy

import (
	"slices"
	"testing"
)

func TestCursorShellCommandDenySubstringsExcludesApprovalGatedGit(t *testing.T) {
	cursor := CursorShellCommandDenySubstrings()
	for _, gated := range []string{"git commit", "git push", "git reset", "gh release"} {
		if slices.Contains(cursor, gated) {
			t.Fatalf("cursor deny list must not include approval-gated %q", gated)
		}
	}
	if !slices.Contains(cursor, "rm -rf") {
		t.Fatal("cursor deny list must still include rm -rf")
	}
}
