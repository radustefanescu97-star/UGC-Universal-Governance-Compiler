package engine

import "github.com/universal-governance/ugc/engine/models"

type Emitter interface {
	Emit(g *models.Governance, targetDir string) error
}
