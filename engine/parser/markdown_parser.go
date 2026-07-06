package parser

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/universal-governance/ugc/engine/audit"
	"github.com/universal-governance/ugc/engine/models"
)

// Parse reads the .universal-governance directory and populates the Governance model.
func Parse(targetDir string) (*models.Governance, error) {
	gov := &models.Governance{}

	hash, err := audit.GenerateSourceHash(targetDir)
	if err != nil {
		return nil, err
	}
	gov.SourceHash = hash

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read target directory %s: %w", targetDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
			data, err := os.ReadFile(filepath.Join(targetDir, entry.Name()))
			if err != nil {
				return nil, err
			}
			gov.BaseRules += string(data) + "\n\n"
		}
	}

	sopsDir := filepath.Join(targetDir, "SOPs")
	if _, err := os.Stat(sopsDir); err == nil {
		err = filepath.WalkDir(sopsDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && filepath.Ext(d.Name()) == ".md" {
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				gov.SOPs = append(gov.SOPs, models.SOP{
					Name:    d.Name(),
					Content: string(data),
				})
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk SOPs directory: %w", err)
		}
	}

	return gov, nil
}
