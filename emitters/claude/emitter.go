package claude

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/universal-governance/ugc/engine/models"
)

type Emitter struct{}

func (e *Emitter) Emit(g *models.Governance, targetDir string) error {
	fmt.Println("Emitting Claude configuration...")

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
			Deny: claudeDenyRules(),
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

func claudeDenyRules() []string {
	return []string{
		"Bash(git push)",
		"Bash(git push *)",
		"Bash(git commit)",
		"Bash(git commit *)",
		"Bash(git reset)",
		"Bash(git reset *)",
		"Bash(rm -rf)",
		"Bash(rm -rf *)",
		"Bash(npm publish)",
		"Bash(npm publish *)",
		"Bash(pnpm publish)",
		"Bash(pnpm publish *)",
		"Bash(yarn publish)",
		"Bash(yarn publish *)",
		"Bash(gh release)",
		"Bash(gh release *)",
		"Bash(docker push)",
		"Bash(docker push *)",
		"Bash(kubectl apply)",
		"Bash(kubectl apply *)",
		"Bash(terraform apply)",
		"Bash(terraform apply *)",
		"Bash(vercel deploy)",
		"Bash(vercel deploy *)",
		"Bash(netlify deploy)",
		"Bash(netlify deploy *)",
		"Bash(firebase deploy)",
		"Bash(firebase deploy *)",
		"Bash(gcloud run deploy)",
		"Bash(gcloud run deploy *)",
		"Read(./.env)",
		"Read(./.env.*)",
		"Read(./secrets/**)",
		"Read(./config/credentials.json)",
		"Read(./.npmrc)",
		"Read(./.pypirc)",
		"Read(~/.aws/credentials)",
		"Read(~/.ssh/**)",
	}
}
