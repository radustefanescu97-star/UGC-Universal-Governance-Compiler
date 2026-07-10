# Universal Governance Compiler

Universal Governance Compiler (UGC) is a local CLI for teams and individuals who want one repository-level source of governance for multiple AI coding agents.

AI tools do not read the same rule files. OpenAI Codex, Google Antigravity, Anthropic Claude Code, and Cursor by Anysphere each expect different local configuration surfaces. Without a compiler, users either duplicate instructions by hand or accept that every agent may operate from a slightly different set of rules.

UGC takes the opposite approach: write governance once, compile it into agent-specific files, and audit the generated output for drift.

## What UGC Does

UGC gives a repository a local governance source under `.universal-governance/`, then generates deterministic target files for the supported V1 agents:

- OpenAI Codex: `AGENTS.md`, `.codex/config.toml`, `.codex/rules/ugc.rules`, `.agents/skills/ugc-governance/SKILL.md`
- Google Antigravity: `.agents/AGENTS.md`, `.agents/skills/*/SKILL.md`
- Anthropic Claude Code: `CLAUDE.md`, `.claude/settings.json`
- Cursor by Anysphere: `.cursorrules`, `.cursor/hooks.json`, `.cursor/hooks/ugc-deny.sh`

It is designed for people who care about approval gates, protected surfaces, stop conditions, worklog discipline, and repeatable governance across AI-assisted development tools.

## Why It Matters

AI development is moving faster than the governance layer around it. Teams and individuals can already choose between multiple coding agents, IDE assistants, and local automation tools, but the operating rules for those tools are fragmented.

UGC makes governance portable at the repository level:

- one local source of truth;
- deterministic generated files for each supported agent;
- dry-run previews before writes;
- audit checks for generated-file drift and missing artifacts;
- narrow restore for exact UGC-owned generated files;
- approval-packet hashing for scoped human approval workflows;
- a standard SOP corpus that gives agents concrete operating discipline from the first `ugc init`.

The goal is not to pretend every third-party agent has identical runtime controls. The goal is to make the governance surface explicit, reproducible, and auditable.

## The Standard SOP Corpus

The standard SOP corpus is the main differentiator in UGC. `ugc init` installs it into `.universal-governance/`; `ugc build` then compiles that corpus into the supported agents' local rule surfaces.

This means UGC does not only generate empty config files. It gives a project a starting operating system for AI-assisted work: approval discipline, worklog memory, safe mutation boundaries, repo hygiene, release guardrails, and documentation accountability.

The embedded corpus currently includes:

- `UGC_AI_AGENTS_GOVERNANCE_SOP.md`: sets safe boundaries for AI-assisted work so agents do not silently redefine product truth, bypass approvals, or touch protected surfaces.
- `UGC_APPROVAL_PACKET_SOP.md`: defines hash-bound approval packets so a short approval in chat can reference a complete implementation boundary by path and SHA-256 hash.
- `UGC_CHECKLIST_SOP.md`: requires closure checklists so sessions do not end with hidden residual risks, unclear validation, or missing handoff details.
- `UGC_CLARIFY_INTENT_SOP.md`: forces clarification before ambiguous requests are treated as permission to mutate files.
- `UGC_COMMIT_DEPLOY_PUSH_GUARDRAILS_SOP.md`: defines guardrails before commit, push, deploy, or promotion work so unrelated changes and unsafe release actions do not get mixed together.
- `UGC_ENGINEERING_SIMPLICITY_SOP.md`: keeps engineering explicit, boring, and auditable by discouraging unnecessary abstractions, hidden control flow, and clever but hard-to-review implementation.
- `UGC_GOVERNANCE_CHANGE_SOP.md`: adds preflight and closure gates for governance changes so dependent artifacts are checked instead of assumed.
- `UGC_ORCHESTRATION_SOP.md`: governs bounded parallel or delegated agent work and makes the orchestrator responsible for final synthesis and correctness.
- `UGC_PORTABILITY_SOP.md`: helps preserve valuable local work through backup checkpoint assessment and fresh-environment readiness.
- `UGC_RELEASE_SOP.md`: defines controlled release and promotion discipline, including separation between development and production surfaces.
- `UGC_REPOSCAN_AND_BOOTSTRAP_SOP.md`: requires agents to identify the target surface, protected neighboring surfaces, applicable SOPs, and approval gates before material work.
- `UGC_REPOSITORY_EXPLAINABILITY_SOP.md`: gives humans and agents a framework for understanding high-risk files, request flow, and extension points before debugging or refactoring.
- `UGC_REPOSITORY_EXPLAINABILITY_AND_DOC_SYNC_SOP.md`: keeps documentation aligned with behavior, decisions, debugging lessons, and material changes.
- `UGC_WORKLOG_AND_SESSION_SOP.md`: defines session history requirements so the project keeps a durable record of objective, files touched, approvals, validations, stop reasons, and residual risks.
- `UGC_WORKLOG_SYNC_SKILL.md`: gives generated agent surfaces a runtime reminder to sync internal task notes into the project worklog before ending a material session.
- `UGC_WORKTREE_DISCIPLINE_SOP.md`: keeps git worktrees and temporary workspaces clean, scoped, and recoverable.

These SOPs are intentionally practical. They are not meant to slow down ordinary work; they are meant to make important work reviewable after the conversation window is gone.

## Worklogs As Persistent Project Memory

Worklogs act as persistent project memory for AI-assisted development. They record what happened across sessions: objectives, target surfaces, files touched, approvals, commands, validation results, stop reasons, and residual risks.

That matters because AI agent conversations are temporary and fragmented. A repository worklog gives the next human or agent a durable trail of why a change happened, what was validated, what was deliberately deferred, and what must not be assumed.

For teams, this improves handoff and accountability. For individuals, it reduces context loss when returning to a project days or weeks later.

## Current Status

UGC V1 is a local Go CLI. It currently focuses on four official targets: OpenAI Codex, Google Antigravity, Anthropic Claude Code, and Cursor by Anysphere.

The following are intentionally not part of V1:

- Git hooks or CI enforcement;
- package-manager installation;
- Copilot or Windsurf targets;
- external services, accounts, or hosted control planes.

## Install

UGC v1.0.9 provides prebuilt GitHub Release archives for Linux, macOS, and Windows. Choose the archive for your OS and CPU architecture from the v1.0.9 release, then verify it with `ugc_1.0.9_checksums.txt`.

v1.0.9 adds machine-readable `--json` output for `ugc audit`, `ugc build` (including `--dry-run`), `ugc packet verify`, and `ugc update --dry-run` so GUI tools can consume CLI results without reimplementing engine logic; emitter progress goes to stderr so stdout stays a single JSON object. It publishes `docs/GUI_CONTRACT.md` as the canonical schema reference. It retains `ugc version` / `ugc version --json` from v1.0.8, Antigravity skill frontmatter fixes from v1.0.7, Cursor hook fixes from v1.0.6, Codex three-layer output from v1.0.5, English-first public wording from v1.0.4, Cursor deny hooks from v1.0.3, and minimal public GitHub Actions CI from v1.0.2.

Available archives:

- `ugc_1.0.9_linux_amd64.tar.gz`
- `ugc_1.0.9_linux_arm64.tar.gz`
- `ugc_1.0.9_darwin_amd64.tar.gz`
- `ugc_1.0.9_darwin_arm64.tar.gz`
- `ugc_1.0.9_windows_amd64.zip`

macOS binaries are cross-compiled and are not signed or notarized.

Linux/macOS example:

```bash
tar -xzf ugc_1.0.9_linux_amd64.tar.gz
./ugc_1.0.9_linux_amd64/ugc --help
```

Windows PowerShell example:

```powershell
Expand-Archive .\ugc_1.0.9_windows_amd64.zip
.\ugc_1.0.9_windows_amd64\ugc.exe --help
```

Checksum verification on systems with `sha256sum`:

```bash
sha256sum -c ugc_1.0.9_checksums.txt
```

Source build remains available.

Requirements:

- Go installed locally;
- Git installed locally.

```bash
git clone https://github.com/radustefanescu97-star/UGC-Universal-Governance-Compiler.git
cd UGC-Universal-Governance-Compiler
go build -o ugc .
```

Then run:

```bash
./ugc --help
```

## Quickstart

In a repository where you want UGC governance:

```bash
ugc init
ugc build --dry-run
ugc build
ugc audit
```

Typical workflow:

1. `ugc init` creates the local `.universal-governance/` corpus.
2. `ugc build --dry-run` previews generated files without writing them.
3. `ugc build` writes the supported agent targets.
4. `ugc audit` verifies source structure, generated-file drift, missing files, unexpected stale files, manifest consistency, target capability coverage, and corpus/update state.

If a generated artifact drifts and is still owned by the build manifest, restore that exact path:

```bash
ugc build --restore .claude/settings.json
ugc audit
```

`--restore` is intentionally narrow. It accepts exact manifest-owned generated artifact paths; it is not a broad force mode.

## CLI Commands

```bash
ugc init
```

Bootstraps `.universal-governance/` with the embedded standard corpus.

```bash
ugc build --dry-run
```

Shows the generated write plan without modifying files. Use `--json` for a machine-readable dry-run plan on stdout.

```bash
ugc build
```

Compiles governance into supported target files. UGC refuses to overwrite unmanaged existing target files, applies generated artifacts through a local transaction where possible, and writes the build manifest last. Use `--json` for machine-readable apply results on stdout.

```bash
ugc audit
```

Checks whether generated artifacts still match the current governance source and reports target capability coverage. Use `--json` for a single machine-readable audit report on stdout (see `docs/GUI_CONTRACT.md`).

```bash
ugc version
```

Prints binary version, embedded corpus version, and Go runtime/platform information. Use `--json` for machine-readable output. By default, checks GitHub for a newer release; use `--no-check` to skip network access.

```bash
ugc build --restore <PATH>
```

Restores one exact manifest-owned generated artifact.

```bash
ugc packet new
ugc packet hash
ugc packet verify
```

Creates, hashes, and verifies local approval packet text. Packet verification checks literal approval fields, the declared packet path, and the current SHA-256 hash. Use `--json` with `ugc packet verify` for machine-readable verification output on stdout.

```bash
ugc update --dry-run
```

Previews standard corpus updates without writing files. Use `--json` for a machine-readable update preview on stdout.

Machine-readable JSON shapes for GUI consumers are documented in `docs/GUI_CONTRACT.md`.

## Approval Packets And SHA-256

UGC includes a local approval-packet workflow for scoped changes. A packet can be written, hashed, and checked against an approval sentence before execution.

The hash is SHA-256, part of the Secure Hash Standard family specified by NIST FIPS 180-4. UGC uses it for a practical local purpose: detecting whether the packet text changed after approval.

This is not identity signing, legal attestation, or remote policy enforcement. It is a disciplined local guardrail for teams and individuals who want approval text, scope, allowed actions, forbidden actions, stop conditions, and return gates to stay tied to a specific document.

## Capability Labels

UGC reports target capabilities honestly because each agent exposes different control surfaces.

- `constrained`: UGC can emit local configuration that constrains behavior within that tool's supported mechanisms.
- `instructed`: UGC can emit instructions, but the tool's runtime obedience is not treated as machine enforcement.
- `advisory`: UGC can express guidance, but not a hard local constraint.
- `native-skill`: UGC uses a target-native skill-style mechanism for that capability.

V1 examples:

- Anthropic Claude Code is `constrained` for conservative deny rules emitted in `.claude/settings.json`; UGC does not claim comprehensive hook-based enforcement in V1.
- OpenAI Codex is `constrained` for project-local approval/sandbox defaults and project rule files when the project layer is trusted. Its worklog workflow is represented through a repo-local governance skill, while secret-read protection remains `advisory` in V1.
- Google Antigravity uses a mix of `native-skill` and `instructed`.
- Cursor by Anysphere is `constrained` for conservative deny hooks emitted in `.cursor/hooks.json` and `.cursor/hooks/ugc-deny.sh`; UGC relies on hook `deny` semantics only and does not claim comprehensive enforcement across every Cursor runtime surface.

## Design Principles

UGC is built around a few practical principles:

- governance should live with the repository;
- generated agent rules should be reproducible;
- dry runs should be available before writes;
- unmanaged local files should not be overwritten silently;
- audits should identify drift instead of relying on trust;
- worklogs should preserve operational memory outside transient agent conversations;
- security and compliance claims should match the real enforcement surface.

## Limitations

UGC verifies generated governance artifacts. It does not guarantee that every third-party AI tool will obey instructions identically at runtime.

V1 does not install git hooks, configure branch protections, deploy a service, or enforce policy outside the supported local generated files.

Build application uses local rollback where the filesystem permits it. If the filesystem itself refuses restoration, UGC reports the incomplete rollback and does not write a clean manifest; run `ugc audit` after any interrupted or failed build.

`ugc packet verify` is a local discipline aid. It is not a cryptographic identity system, not a signing service, and not a substitute for organizational approval policy.

Additional targets, release packaging, and stronger repository self-defense remain separate future work.

## Support And Contact

- Bug reports and feature requests: use GitHub Issues in this repository.
- Maintainer profile: https://github.com/radustefanescu97-star
- X: https://x.com/radu_st1

## License

UGC is licensed under the Apache License 2.0.

SPDX: `Apache-2.0`

Copyright 2026 Radu Stefanescu
