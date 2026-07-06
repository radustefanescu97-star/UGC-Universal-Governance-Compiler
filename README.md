# Universal Governance Compiler

Universal Governance Compiler (UGC) is a local CLI for teams that want one repository-level source of governance for multiple AI coding agents.

AI tools do not read the same rule files. Codex, Antigravity, Claude Code, and Cursor each expect different local configuration surfaces. Without a compiler, teams either duplicate instructions by hand or accept that every agent may operate from a slightly different set of rules.

UGC takes the opposite approach: write governance once, compile it into the agent-specific files, and audit the generated output for drift.

## What UGC Does

UGC gives a repository a local governance source under `.universal-governance/`, then generates deterministic target files for the supported V1 agents:

- OpenAI Codex: `AGENTS.md`, `.codex/config.toml`, `.codex/rules/ugc.rules`
- Google Antigravity: `.agents/AGENTS.md`, `.agents/skills/*/SKILL.md`
- Claude Code: `CLAUDE.md`, `.claude/settings.json`
- Cursor: `.cursorrules`

It is designed for teams that care about approval gates, protected surfaces, stop conditions, worklog discipline, and repeatable governance across AI-assisted development tools.

## Why It Matters

AI development is moving faster than the governance layer around it. Teams can already choose between multiple coding agents, IDE assistants, and local automation tools, but the operating rules for those tools are fragmented.

UGC makes governance portable at the repository level:

- one local source of truth;
- deterministic generated files for each supported agent;
- dry-run previews before writes;
- audit checks for generated-file drift and missing artifacts;
- narrow restore for exact UGC-owned generated files;
- approval-packet hashing for scoped human approval workflows.

The goal is not to pretend every third-party agent has identical runtime controls. The goal is to make the governance surface explicit, reproducible, and auditable.

## Current Status

UGC V1 is a local Go CLI. It currently focuses on four official targets: Codex, Antigravity, Claude Code, and Cursor.

The following are intentionally not part of V1:

- Git hooks or CI enforcement;
- GitHub Pages presentation site;
- prebuilt release binaries;
- package-manager installation;
- Copilot or Windsurf targets;
- external services, accounts, or hosted control planes.

## Install From Source

Until tagged releases exist, build UGC from source.

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

Shows the generated write plan without modifying files.

```bash
ugc build
```

Compiles governance into supported target files. UGC refuses to overwrite unmanaged existing target files, applies generated artifacts through a local transaction where possible, and writes the build manifest last.

```bash
ugc audit
```

Checks whether generated artifacts still match the current governance source and reports target capability coverage.

```bash
ugc build --restore <PATH>
```

Restores one exact manifest-owned generated artifact.

```bash
ugc packet new
ugc packet hash
ugc packet verify
```

Creates, hashes, and verifies local approval packet text. Packet verification checks literal approval fields, the declared packet path, and the current SHA-256 hash.

```bash
ugc update --dry-run
```

Previews standard corpus updates without writing files.

## Approval Packets And SHA-256

UGC includes a local approval-packet workflow for scoped changes. A packet can be written, hashed, and checked against an approval sentence before execution.

The hash is SHA-256, part of the Secure Hash Standard family specified by NIST FIPS 180-4. UGC uses it for a practical local purpose: detecting whether the packet text changed after approval.

This is not identity signing, legal attestation, or remote policy enforcement. It is a disciplined local guardrail for teams that want approval text, scope, allowed actions, forbidden actions, stop conditions, and return gates to stay tied to a specific document.

## Capability Labels

UGC reports target capabilities honestly because each agent exposes different control surfaces.

- `constrained`: UGC can emit local configuration that constrains behavior within that tool's supported mechanisms.
- `instructed`: UGC can emit instructions, but the tool's runtime obedience is not treated as machine enforcement.
- `advisory`: UGC can express guidance, but not a hard local constraint.
- `native-skill`: UGC uses a target-native skill-style mechanism for that capability.

V1 examples:

- Claude Code is `constrained` for conservative deny rules emitted in `.claude/settings.json`; UGC does not claim comprehensive hook-based enforcement in V1.
- Codex is `constrained` for project-local approval/sandbox defaults and project rule files when the project layer is trusted; secret-read protection is `advisory` in V1.
- Antigravity uses a mix of `native-skill` and `instructed`.
- Cursor is `instructed` in V1.

## Design Principles

UGC is built around a few practical principles:

- governance should live with the repository;
- generated agent rules should be reproducible;
- dry runs should be available before writes;
- unmanaged local files should not be overwritten silently;
- audits should identify drift instead of relying on trust;
- security and compliance claims should match the real enforcement surface.

## Limitations

UGC verifies generated governance artifacts. It does not guarantee that every third-party AI tool will obey instructions identically at runtime.

V1 does not install git hooks, mutate CI, configure branch protections, deploy a service, or enforce policy outside the supported local generated files.

Build application uses local rollback where the filesystem permits it. If the filesystem itself refuses restoration, UGC reports the incomplete rollback and does not write a clean manifest; run `ugc audit` after any interrupted or failed build.

`ugc packet verify` is a local discipline aid. It is not a cryptographic identity system, not a signing service, and not a substitute for organizational approval policy.

## Roadmap

Future work may include additional agent targets, stronger repository self-defense, CI integration, release packaging, and a separate GitHub Pages presentation site. Those are outside V1 and should be implemented through separate reviewed changes.

## Support And Contact

- Bug reports and feature requests: use GitHub Issues in this repository.
- Maintainer profile: https://github.com/radustefanescu97-star
- X: https://x.com/radu_st1

## License

UGC is licensed under the Apache License 2.0.

SPDX: `Apache-2.0`

Copyright 2026 Radu Stefanescu
