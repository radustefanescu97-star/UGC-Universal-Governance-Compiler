package claude

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/universal-governance/ugc/engine/models"
	"github.com/universal-governance/ugc/engine/policy"
)

type Emitter struct{}

func (e *Emitter) Emit(g *models.Governance, targetDir string) error {
	fmt.Fprintln(os.Stderr, "Emitting Claude configuration...")

	var content string
	content += fmt.Sprintf("// UGC-Source-Hash: %s\n\n", g.SourceHash)
	content += g.BaseRules + "\n\n"
	for _, sop := range g.SOPs {
		content += "## " + sop.Name + "\n" + sop.Content + "\n\n"
	}

	if err := os.WriteFile(filepath.Join(targetDir, "CLAUDE.md"), []byte(content), 0644); err != nil {
		return err
	}
	return writeClaudeSettings(g, targetDir)
}

type claudeSettings struct {
	Permissions claudePermissions `json:"permissions"`
}

type claudePermissions struct {
	Deny []string `json:"deny"`
}

func writeClaudeSettings(g *models.Governance, targetDir string) error {
	settings := claudeSettings{
		Permissions: claudePermissions{
			Deny: policy.ClaudeDenyRules(),
		},
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	settingsPath := filepath.Join(targetDir, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(settingsPath, data, 0644)
}

