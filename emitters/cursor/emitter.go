package cursor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/universal-governance/ugc/engine/models"
)

type Emitter struct{}

func (e *Emitter) Emit(g *models.Governance, targetDir string) error {
	fmt.Println("Emitting Cursor configuration...")
	
	var content string
	content += fmt.Sprintf("// UGC-Source-Hash: %s\n\n", g.SourceHash)
	content += g.BaseRules + "\n\n"
	for _, sop := range g.SOPs {
		content += "## " + sop.Name + "\n" + sop.Content + "\n\n"
	}

	return os.WriteFile(filepath.Join(targetDir, ".cursorrules"), []byte(content), 0644)
}
