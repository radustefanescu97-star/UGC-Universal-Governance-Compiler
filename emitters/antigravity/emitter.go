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

		skillContent := renderSkillContent(skillName, sop.Name, sop.Content)
		if err := os.WriteFile(filepath.Join(skillPath, "SKILL.md"), []byte(skillContent), 0644); err != nil {
			return err
		}
	}

	return nil
}

func renderSkillContent(skillName, sourceName, content string) string {
	if hasSkillFrontmatter(content) {
		return content
	}

	descriptionSubject := firstMarkdownHeading(content)
	if descriptionSubject == "" {
		descriptionSubject = strings.TrimSuffix(sourceName, ".md")
		descriptionSubject = strings.ToLower(strings.ReplaceAll(descriptionSubject, "_", " "))
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("name: %s\n", skillName))
	b.WriteString(fmt.Sprintf("description: %q\n", "Use when applying UGC guidance from "+descriptionSubject+"."))
	b.WriteString("---\n\n")
	b.WriteString(strings.TrimLeft(content, "\r\n"))
	return b.String()
}

func hasSkillFrontmatter(content string) bool {
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		return false
	}
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			header := strings.Join(lines[1:i], "\n")
			return strings.Contains(header, "name:") && strings.Contains(header, "description:")
		}
	}
	return false
}

func firstMarkdownHeading(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") {
			continue
		}
		heading := strings.TrimSpace(strings.TrimLeft(line, "#"))
		if heading != "" {
			return heading
		}
	}
	return ""
}
