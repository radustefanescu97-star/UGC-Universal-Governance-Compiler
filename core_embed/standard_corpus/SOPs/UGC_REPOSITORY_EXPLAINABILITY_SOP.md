# Universal Repository Explainability SOP

Status: Active
Created: 2026-07-04
Scope: Repository orientation, debugging, refactoring, extension planning, and documentation governance

## Purpose

This SOP makes the project easier to understand without changing runtime behavior.

Its job is to give humans and AI agents a canonical map of:

- which surfaces exist;
- which files are high-risk;
- how a typical request moves through the system;
- why major architecture and governance choices exist;
- where to start before debugging, refactoring, or extending the project.

This SOP does not authorize code changes, deploys, provider calls, runtime checks, billing changes, cloud mutations, cost-generating tests, or canonical live-state claims.

## Related SOPs

Read this SOP with:

- `UGC_AI_AGENTS_GOVERNANCE_SOP.md`
- `UGC_WORKLOG_AND_SESSION_SOP.md`
- `UGC_ENGINEERING_SIMPLICITY_SOP.md`
- `UGC_REPOSITORY_EXPLAINABILITY_AND_DOC_SYNC_SOP.md`
- `UGC_COMMIT_DEPLOY_PUSH_GUARDRAILS_SOP.md`
- `UGC_RELEASE_SOP.md`

If another surface-specific SOP is stricter, the stricter SOP wins.

## Canonical Rule

Do not treat the project as a file inventory.

Before material debugging, refactoring, extension, or documentation work, identify:

- target surface;
- protected neighboring surfaces;
- governing SOPs;
- approval gates;
- whether live/runtime/provider/cloud/billing/cost risk exists;
- which current plan or worklog explains recent context.

Repository explainability is an operating safety layer. It must improve auditability without loosening existing guardrails.

Engineering simplicity is a companion rule: a solution that hides control flow, provider choice, safety gates, fallback behavior, thresholds, or validation evidence behind clever or unnecessary abstraction is not explainable enough for material work.

## Truth Boundaries

Use the right truth for the question:

| Question | Primary Truth |
| --- | --- |
| What rules govern the work? | `AGENTS.md`, `SOPs/README.md`, and active SOPs |
| What did a prior session decide or discover? | Approved plans, worklogs, incident reports, and evidence ledgers in `Plans/` |
| What does the local repo currently contain? | Current source files in the active worktree |
| What is currently live? | Separately approved live/runtime evidence |

Historical worklogs are evidence, not automatic current live truth.
Local source is evidence, not automatic deployed truth.
Live truth checks may require approval if they can touch production, providers, cloud, billing, runtime, or cost-generating paths.

## Generic Surface Map

| Surface | Role | Explainability Rule |
| --- | --- | --- |
| `Core Application` | Main API, frontend artifacts, orchestrations | Treat as the main product surface. Avoid broad rewrites. |
| `Backend Services` | Isolated data backends, contracts, jobs | Treat as strict contract surface. |
| `Internal Portal` | Operational/admin visibility | Treat as sensitive. Browser-only gating is not enough for sensitive data. |
| `Plans/` | Plans, worklogs, incident memory, evidence ledgers | Historical and operational memory. Keep material work traceable here. |
| `SOPs/` | Standing operating rules | Canonical operating rules. SOP changes require explicit approval. |
| `.agents/` | Active project governance | Controls AI behavior. Governed by change control. |

## High-Risk Local Files

Treat these as high-risk edit zones:

- Main entry points (`index.html`, `main.py`, `app.js`, etc.)
- Provider routing or integration files
- Deployment configuration files (e.g., `firebase.json`, `docker-compose.yml`, `Dockerfile`)
- `scripts/` directory
- `tests/` directory
- Any local deploy environment or credential-bearing file

Do not print, quote, summarize, or copy secret values into docs, worklogs, terminal summaries, plans, or tickets.

## Why Core Choices Exist

| Choice | Why It Exists | What Not To Infer |
| --- | --- | --- |
| Worklogs as memory | The project has repeated failure classes; future maintainers need symptom, cause, fix, validation, and residual risk. | Do not close material bugs with only "fixed". |
| Local fixtures before live checks | Provider/model/runtime checks can create cost or side effects. | Do not run live checks without approval. |
| Server-side billing decisions | Browser-provided identity/prices are not billing authority. | Do not treat UI pricing as entitlement truth. |

## Required Use

Use this SOP when:

- starting debugging or refactoring;
- explaining repository structure;
- planning a material extension;
- updating documentation or SOPs;
- reviewing a change that touches known high-risk surfaces.

For tiny answer-only or mechanical sessions, do not expand scope just to satisfy this SOP. Record no material action when appropriate.

## Forbidden Practices

Do not:

- claim live production truth from local/static evidence;
- use this SOP to loosen a stricter surface SOP;
- turn historical worklogs into current truth without reconciliation;
- read or print secret-bearing files for orientation;
- touch runtime/product/deploy/provider/billing/cloud files under an explainability-only scope.

## Stop Conditions

Stop and request explicit approval or a narrower plan when:

- the next step requires a deploy, live runtime check, cloud mutation, provider call, billing/payment action, or cost-generating test;
- local source conflicts with active SOPs;
- a proposed documentation change would loosen an existing guardrail;
- a secret-bearing file would need to be inspected;
- the workspace dirty state makes file ownership unclear.
