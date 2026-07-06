package engine

import (
	"github.com/universal-governance/ugc/emitters/antigravity"
	"github.com/universal-governance/ugc/emitters/claude"
	"github.com/universal-governance/ugc/emitters/codex"
	"github.com/universal-governance/ugc/emitters/cursor"
)

func V1Emitters() []Emitter {
	return []Emitter{
		&codex.Emitter{},
		&antigravity.Emitter{},
		&claude.Emitter{},
		&cursor.Emitter{},
	}
}
