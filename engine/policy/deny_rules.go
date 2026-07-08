package policy

// ClaudeDenyRules returns conservative deny rules for Claude Code settings.json.
func ClaudeDenyRules() []string {
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

// ShellCommandDenySubstrings are matched case-sensitively against shell commands.
func ShellCommandDenySubstrings() []string {
	return []string{
		"git push",
		"git commit",
		"git reset",
		"rm -rf",
		"npm publish",
		"pnpm publish",
		"yarn publish",
		"gh release",
		"docker push",
		"kubectl apply",
		"terraform apply",
		"vercel deploy",
		"netlify deploy",
		"firebase deploy",
		"gcloud run deploy",
	}
}

// cursorApprovalGatedShellCommands are governed by SOPs and .cursorrules, not Cursor hook hard-deny.
var cursorApprovalGatedShellCommands = map[string]struct{}{
	"git commit": {},
	"git push":   {},
	"git reset":  {},
	"gh release": {},
}

// CursorShellCommandDenySubstrings returns shell patterns hard-denied by the Cursor hook only.
func CursorShellCommandDenySubstrings() []string {
	all := ShellCommandDenySubstrings()
	out := make([]string, 0, len(all))
	for _, pattern := range all {
		if _, gated := cursorApprovalGatedShellCommands[pattern]; gated {
			continue
		}
		out = append(out, pattern)
	}
	return out
}

// SecretReadPathRules describe file paths that hook enforcement should deny.
type SecretReadPathRule struct {
	Prefix   string
	Contains string
	Suffix   string
}

func SecretReadPathRules() []SecretReadPathRule {
	return []SecretReadPathRule{
		{Suffix: "/.env"},
		{Suffix: ".env"},
		{Contains: "/.env."},
		{Contains: "/secrets/"},
		{Suffix: "/config/credentials.json"},
		{Suffix: "/.npmrc"},
		{Suffix: "/.pypirc"},
		{Suffix: "/.aws/credentials"},
		{Contains: "/.ssh/"},
	}
}
