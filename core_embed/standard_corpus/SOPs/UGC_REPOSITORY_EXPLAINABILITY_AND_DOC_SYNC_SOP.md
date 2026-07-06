# Universal Repository Explainability And Doc Sync SOP

Status: Active
Created: 2026-07-04
Scope: Material changes, bug fixes, architecture decisions, debugging lessons, and documentation governance

## Purpose

This SOP prevents documentation from drifting away from the code and operational memory.

Its job is to make future debugging and refactoring easier by ensuring material sessions record:

- what changed;
- why it changed;
- what failure mode or decision it relates to;
- how it was validated;
- what remains risky;
- whether canonical explainability/debugging docs need an update.

This SOP is meant to improve auditability, not to slow ordinary work.

It does not authorize code changes, deploys, provider calls, billing changes, runtime checks, cloud mutations, cost-generating tests, or SOP loosening.

## Related SOPs

Read this SOP with:

- `UGC_REPOSITORY_EXPLAINABILITY_SOP.md`
- `UGC_ENGINEERING_SIMPLICITY_SOP.md`
- `UGC_WORKLOG_AND_SESSION_SOP.md`
- `UGC_CHECKLIST_SOP.md`
- `UGC_AI_AGENTS_GOVERNANCE_SOP.md`
- `UGC_GOVERNANCE_CHANGE_SOP.md`

If another SOP is stricter, the stricter SOP wins.

## Canonical Rule

For every material change, either:

1. update the relevant explainability/debugging/SOP documentation; or
2. record `Doc Impact: none` with a short reason in the worklog.

This rule applies only to material work. It is not a requirement to rewrite documentation for tiny mechanical changes that do not affect behavior, architecture, debugging, risk, protected surfaces, or operational assumptions.

## Material Doc-Impact Triggers

Check doc impact when work touches:

- architecture or request flow;
- bug fixes or incident response;
- public behavior;
- provider/model/backend behavior;
- billing, pricing, entitlement, public API, or CRM behavior;
- environment boundaries (development vs staging vs production);
- deployment, runtime config, or cost posture;
- tests or guardrails that future agents rely on;
- new abstractions, shared gates, or simplification decisions that future agents must understand;
- a known failure mode;
- a newly discovered recurring failure mode;
- a stale or superseded plan/worklog assumption.

## Lightweight Rule

Do not make this SOP bureaucratic.

Allowed for small material changes:

```md
Doc Impact:
- None. This change is a narrow copy/test-only edit and does not affect architecture, debugging, risk, or protected surfaces.
```

Required for meaningful material changes:

```md
Doc Impact:
- Updated:
- Not updated:
- Reason:
```

## Bug Fix Closeout Rule

For material bug fixes, record:

- symptom;
- affected surface;
- root cause;
- why the chosen fix is correct;
- files touched;
- validation performed;
- validation not performed and why;
- residual risk;
- regression prevention;
- docs updated or `Doc Impact: none` with reason.

The worklog entry can be concise. It must not be only "fixed" or "patched".

## Architecture Decision Rule

For architecture or flow changes, record:

- decision;
- alternatives considered;
- why the chosen path fits governance;
- blast radius;
- protected neighboring surfaces;
- rollback or stop condition;
- what future maintainers must not infer from the change.

## Canonical Documentation Targets

Use the smallest useful update:

| If The Lesson Is About | Update Or Check |
| --- | --- |
| Repository shape, high-risk files, request flow, or why a major choice exists | `UGC_REPOSITORY_EXPLAINABILITY_SOP.md` |
| Worklog/doc sync behavior | this SOP and `UGC_WORKLOG_AND_SESSION_SOP.md` if needed |
| One-session evidence, investigation notes, or unresolved debt | a plan/worklog/evidence file in `Plans/` |

Historical detail should stay in `Plans/`. SOPs should stay concise and operational.

## Doc-Sync Debt Rule

If a documentation update is needed but unsafe or too large for the current session, create or update a doc-sync debt entry in `Plans/` with:

- target doc;
- missing update;
- why it was deferred;
- risk if it remains deferred;
- recommended next action.

Do not silently defer material documentation.

## Governance Doc-Sync Rule

If the documentation update creates or changes an SOP, governance skill, governance script, approval rule, validation rule, or publication/sync path, also apply `UGC_GOVERNANCE_CHANGE_SOP.md`.

Every related governance dependency must be synced, marked not applicable with a reason, blocked with debt in `Plans/`, or deferred behind a named separate approval gate.

## Prohibited Practices

Do not:

- copy raw secrets into documentation or worklogs;
- claim live production truth from static/local evidence;
- loosen an existing SOP by calling the change "just docs";
- delete historical evidence only because it is stale;
- turn stale docs into current truth without reconciliation;
- force long narrative documentation for tiny/mechanical changes;
- promote draft plans into SOPs without explicit approval.

## Approval Gates

This SOP does not reduce any existing approval gate.

Explicit approval is still required for:

- canonical SOP creation or update;
- deploys, previews, and traffic shifts;
- infrastructure, IAM, or billing mutations;
- pricing, entitlement, public API, or CRM behavior changes;
- live provider/model/data calls;
- load tests or cost-generating tests;
- automated job execution;
- staging, commit, push, PR, merge, or release actions when those are requested.

## Stop Conditions

Stop when:

- a needed doc update would conflict with an existing active SOP;
- a doc claim requires live/runtime/cloud/provider/billing evidence that is not approved;
- a secret-bearing file would need to be printed or copied;
- a proposed documentation change would broaden scope into technical remediation;
- the worktree dirty state makes ownership of documentation changes unclear.
