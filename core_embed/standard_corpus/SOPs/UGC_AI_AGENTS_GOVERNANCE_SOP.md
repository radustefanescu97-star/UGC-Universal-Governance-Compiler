# Universal AI Agents Governance SOP

Status: Active
Created: 2026-07-04
Scope: AI-assisted work on the project, operations, SOPs, plans, release work, and investigations

## Purpose

This SOP defines how AI agents may work on the project without breaking production trust, leaking sensitive data, bypassing approval gates, or silently redefining product truth.

Its job is to ensure every AI-assisted material session has:

- a declared target surface;
- protected neighboring surfaces;
- a bounded scope;
- source-aware implementation or investigation;
- explicit approval gates for risky actions;
- validation before promotion;
- a clear handoff trail.

This SOP does not authorize code changes, deploys, cost-generating tests, provider calls, pricing changes, entitlement changes, or production actions.

## Related SOPs

This SOP must be read with:

- `UGC_WORKLOG_AND_SESSION_SOP.md`
- `UGC_CHECKLIST_SOP.md`
- `UGC_ORCHESTRATION_SOP.md`
- `UGC_COMMIT_DEPLOY_PUSH_GUARDRAILS_SOP.md`
- `UGC_RELEASE_SOP.md`
- `UGC_GOVERNANCE_CHANGE_SOP.md`
- `UGC_ENGINEERING_SIMPLICITY_SOP.md`

If a surface-specific SOP is stricter, the stricter SOP wins.

## Canonical Rule

AI agents may assist with planning, implementation, review, and documentation, but they must not silently redefine production truth.

Every material AI session must preserve:

- existing user-facing behavior unless a change is explicit;
- existing operational reachability unless a navigation change is explicit;
- target separation across environments;
- user data and secret isolation;
- deploy and cost-control discipline.

If parallel agents are used, all delegated agents inherit these obligations. The orchestrator remains accountable for reviewing their reports, rejecting unsafe or unsupported findings, and following `UGC_ORCHESTRATION_SOP.md`.

Delegated agents must not create, request, spawn, schedule, coordinate, simulate, or rely on child agents. Only the top-level orchestrator may assign delegated agents.

## Operating Requirements

Before implementation or runtime action, the agent must:

- inspect relevant source files and SOPs;
- identify the target surface;
- identify protected neighboring surfaces;
- distinguish investigation from mutation;
- avoid assumptions that can be answered from repo or read-only cloud state;
- name any remaining product, legal, billing, cost, provider, or approval ambiguity.

During implementation, the agent must:

- keep changes tightly scoped;
- prefer existing patterns and dependencies;
- avoid direct production deploys unless explicitly requested and SOP-approved;
- avoid live provider calls, load tests, or cost-generating tests unless explicitly approved and costed;
- avoid secrets in browser code or repo-tracked files;
- preserve unrelated user changes in the worktree;
- avoid broad rewrites when additive changes are enough.
- avoid clever code and unneeded abstractions unless `UGC_ENGINEERING_SIMPLICITY_SOP.md` is satisfied.

Before handoff, the agent must:

- run the highest-signal feasible checks that do not violate cost or approval SOPs;
- report checks that could not run;
- list residual production risks;
- identify preview, production, provider, billing, or approval gates still required;
- update the relevant worklog for material sessions;
- close with the post-session SOP compliance checklist.

## Governance Change Preflight Rule

When a session changes governance artifacts, the agent must follow `UGC_GOVERNANCE_CHANGE_SOP.md` before implementation and before final closeout.

Governance artifacts include:

- `AGENTS.md`;
- `SOPs/**`;
- local installed or repo-tracked governance skills;
- governance scripts, validation workflows, branch protection, rulesets, required checks, deploy wrappers, live-surface contracts, approval semantics.

Every dependent artifact must finish as `synced`, `not applicable`, `blocked`, or `requires separate approval`. Unknown or unchecked dependencies block closeout.

## Data And Access Rules

AI agents must not request, expose, or store:

- passwords;
- payment card numbers;
- private keys;
- service account secrets;
- provider API secrets;
- unnecessary customer personal data;
- raw billing/payment data;
- local-only credential values.

Secret validation must be done without printing the secret.

Internal operational tooling must use server-side authorization for sensitive data. Browser-only allowlists are not enough for sensitive billing, customer, support, provider, or operational data.

## Production Truth Rule

Agents must distinguish:

- source truth;
- built artifact truth;
- deployed hosting truth;
- live revision truth;
- project/target truth;
- runtime secret/config truth;
- provider quota/cost truth.

If these disagree, record the drift and stop before promotion unless the session is explicitly a restore or reconciliation session.

## Approval Gate Rule

The agent must ask for explicit approval before:

- deploys;
- traffic shifts;
- service updates;
- hosting deploys;
- database rules/index deploys;
- secret value changes;
- IAM/billing mutations;
- provider migrations;
- load tests or cost-generating tests;
- automated job execution;
- pricing, entitlement, or billing behavior changes.

An approval is valid only when the user's sentence contains the literal word `aprobare` or `aproval`. If the sentence does not contain one of those words, the agent must treat it as discussion, direction, preference, or planning context, not approval, even if the wording sounds permissive.

Generic approval is not enough when a stricter SOP requires exact wording, cost estimate, target, or rollback posture.

Agent delegation, agent consensus, or an agent report is never approval for an approval-gated action.

## Documentation Rule

SOPs remain canonical in `SOPs/`.

Plans and worklogs remain canonical in `Plans/`.

Delegated agent reports are evidence inputs. They become operationally useful only after orchestrator review and, for material sessions, worklog/checklist recording.

Every material AI-assisted session must update the relevant worklog according to `UGC_WORKLOG_AND_SESSION_SOP.md`.

Every material AI-assisted session must close with the checklist from `UGC_CHECKLIST_SOP.md`.

## Final Rule

AI assistance should make the project safer, faster, and clearer.

If the agent cannot prove a change preserves the correct surface, protected neighboring behavior, cost posture, and approval boundaries, the change is not ready to ship.

## Separation of Plans Rule

Do not confuse the agent's native planning artifact (e.g., `implementation_plan.md`) with the repository's Product Masterplan. The repository Masterplan must be a distinct file (e.g., `Plans/Product_Masterplan.md`). The agent's native planning artifact must only be used as a short-lived, disposable Approval Packet to propose changes to the repository. It is never the single source of truth.
