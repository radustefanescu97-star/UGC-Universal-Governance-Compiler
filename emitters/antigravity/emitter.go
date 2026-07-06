package antigravity

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/universal-governance/ugc/engine/models"
)

type Emitter struct{}

func (e *Emitter) Emit(g *models.Governance, targetDir string) error {
	fmt.Println("Emitting Antigravity configuration...")
	agentsDir := filepath.Join(targetDir, ".agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return err
	}

	// Write AGENTS.md
	baseContent := fmt.Sprintf("<!-- UGC-Source-Hash: %s -->\n\n%s", g.SourceHash, g.BaseRules)
	if err := os.WriteFile(filepath.Join(agentsDir, "AGENTS.md"), []byte(baseContent), 0644); err != nil {
		return err
	}

	// Write skills
	skillsDir := filepath.Join(agentsDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return err
	}

	for _, sop := range g.SOPs {
		skillName := strings.TrimSuffix(sop.Name, ".md")
		skillName = strings.ToLower(strings.ReplaceAll(skillName, "_", "-"))
		if sop.Name == "UGC_WORKLOG_SYNC_SKILL.md" {
			skillName = "ugc-worklog-sync"
		}

		skillPath := filepath.Join(skillsDir, skillName)
		if err := os.MkdirAll(skillPath, 0755); err != nil {
			return err
		}

		if err := os.WriteFile(filepath.Join(skillPath, "SKILL.md"), []byte(sop.Content), 0644); err != nil {
			return err
		}
	}

	return nil
}
