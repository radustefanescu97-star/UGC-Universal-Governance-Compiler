package audit

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GenerateSourceHash computes a combined SHA-256 hash of all files in .universal-governance.
func GenerateSourceHash(targetDir string) (string, error) {
	var files []string
	err := filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(d.Name()) == ".md" {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to walk target directory: %w", err)
	}

	sort.Strings(files)

	h := sha256.New()
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		relPath, err := filepath.Rel(targetDir, file)
		if err != nil {
			return "", err
		}
		h.Write([]byte(filepath.ToSlash(relPath))) // Include stable filename in hash
		h.Write(data)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func ValidateSourceStructure(targetDir string) []error {
	var errs []error
	if !fileExists(filepath.Join(targetDir, "AGENTS.md")) {
		errs = append(errs, fmt.Errorf("missing AGENTS.md"))
	}
	if !fileExists(filepath.Join(targetDir, "SOPs", "README.md")) {
		errs = append(errs, fmt.Errorf("missing SOPs/README.md"))
	}

	sopsDir := filepath.Join(targetDir, "SOPs")
	entries, err := os.ReadDir(sopsDir)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to read SOPs directory: %w", err))
		return errs
	}

	ugcSOPCount := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		if strings.HasPrefix(entry.Name(), "UGC_") {
			ugcSOPCount++
		}
	}
	if ugcSOPCount == 0 {
		errs = append(errs, fmt.Errorf("no UGC SOP files found"))
	}
	return errs
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
