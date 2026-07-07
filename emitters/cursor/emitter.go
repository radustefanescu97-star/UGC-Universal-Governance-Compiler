package cursor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/universal-governance/ugc/engine/models"
	"github.com/universal-governance/ugc/engine/policy"
)

const (
	hooksDirRel     = ".cursor/hooks"
	hooksJSONRel    = ".cursor/hooks.json"
	denyHookRel     = ".cursor/hooks/ugc-deny.sh"
	denyHookName    = "ugc-deny.sh"
	sourceHashMarker = "UGC-Source-Hash"
)

type Emitter struct{}

func (e *Emitter) Emit(g *models.Governance, targetDir string) error {
	fmt.Println("Emitting Cursor configuration...")

	var content string
	content += fmt.Sprintf("// %s: %s\n\n", sourceHashMarker, g.SourceHash)
	content += g.BaseRules + "\n\n"
	for _, sop := range g.SOPs {
		content += "## " + sop.Name + "\n" + sop.Content + "\n\n"
	}

	if err := os.WriteFile(filepath.Join(targetDir, ".cursorrules"), []byte(content), 0644); err != nil {
		return err
	}
	return writeCursorHooks(g, targetDir)
}

type cursorHooksFile struct {
	Version int                       `json:"version"`
	Hooks   map[string][]cursorHookDef `json:"hooks"`
}

type cursorHookDef struct {
	Command    string `json:"command"`
	FailClosed bool   `json:"failClosed"`
}

func writeCursorHooks(g *models.Governance, targetDir string) error {
	hooksPath := filepath.Join(targetDir, hooksJSONRel)
	hookScriptPath := filepath.Join(targetDir, denyHookRel)

	if err := os.MkdirAll(filepath.Join(targetDir, hooksDirRel), 0755); err != nil {
		return err
	}

	script := denyHookScript(g.SourceHash)
	if err := os.WriteFile(hookScriptPath, []byte(script), 0755); err != nil {
		return err
	}

	hooks := cursorHooksFile{
		Version: 1,
		Hooks: map[string][]cursorHookDef{
			"beforeShellExecution": {
				{Command: denyHookRel, FailClosed: true},
			},
			"beforeReadFile": {
				{Command: denyHookRel, FailClosed: true},
			},
		},
	}

	data, err := json.MarshalIndent(hooks, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(hooksPath, data, 0644)
}

func denyHookScript(sourceHash string) string {
	var b strings.Builder
	b.WriteString("#!/usr/bin/env bash\n")
	b.WriteString("# " + sourceHashMarker + ": " + sourceHash + "\n")
	b.WriteString("# UGC-generated Cursor deny hook. Enforcement uses permission=deny only.\n")
	b.WriteString("set -euo pipefail\n\n")
	b.WriteString("input=$(cat)\n\n")
	b.WriteString("deny() {\n")
	b.WriteString(`  printf '%s\n' '{"permission":"deny","user_message":"UGC policy blocked this action. Explicit owner approval (aprobare/aproval) is required before gated actions.","agent_message":"Blocked by UGC-generated Cursor deny hook."}'` + "\n")
	b.WriteString("  exit 0\n")
	b.WriteString("}\n\n")
	b.WriteString("allow() {\n")
	b.WriteString(`  printf '%s\n' '{"permission":"allow"}'` + "\n")
	b.WriteString("  exit 0\n")
	b.WriteString("}\n\n")
	b.WriteString(`command=$(printf '%s' "$input" | sed -n 's/.*"command"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -1)` + "\n")
	b.WriteString(`path=$(printf '%s' "$input" | sed -n 's/.*"path"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -1)` + "\n")
	b.WriteString(`if [ -z "$path" ]; then` + "\n")
	b.WriteString(`  path=$(printf '%s' "$input" | sed -n 's/.*"filePath"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -1)` + "\n")
	b.WriteString("fi\n\n")
	b.WriteString("if [ -n \"$command\" ]; then\n")
	for _, pattern := range policy.ShellCommandDenySubstrings() {
		fmt.Fprintf(&b, "  case \" $command \" in *%s*) deny ;; esac\n", pattern)
	}
	b.WriteString("fi\n\n")
	b.WriteString("if [ -n \"$path\" ]; then\n")
	for _, rule := range policy.SecretReadPathRules() {
		switch {
		case rule.Suffix != "":
			fmt.Fprintf(&b, "  case \"$path\" in *%s) deny ;; esac\n", rule.Suffix)
		case rule.Contains != "":
			fmt.Fprintf(&b, "  case \"$path\" in *%s*) deny ;; esac\n", rule.Contains)
		case rule.Prefix != "":
			fmt.Fprintf(&b, "  case \"$path\" in %s*) deny ;; esac\n", rule.Prefix)
		}
	}
	b.WriteString("fi\n\n")
	b.WriteString("allow\n")
	return b.String()
}
