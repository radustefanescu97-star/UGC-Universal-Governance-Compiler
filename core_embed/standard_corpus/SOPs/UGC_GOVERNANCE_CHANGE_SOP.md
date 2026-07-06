# Governance Change Preflight and Closure Gate SOP

Status: Active
Scope: Project governance changes, SOP updates, governance skill updates, ruleset changes, and documentation sync rules

## Purpose

This SOP creates a mandatory preflight and closure gate for governance changes.
Its job is to prevent governance work from closing as complete while dependent artifacts are unknown, unchecked, silently skipped, or left to memory.

This SOP adds an audit gate. It does not authorize code changes or git stage/commit/push.

## Related SOPs

Read this SOP with:
- [SOPs/README.md](README.md)
- [UGC_REPOSCAN_AND_BOOTSTRAP_SOP.md](UGC_REPOSCAN_AND_BOOTSTRAP_SOP.md)
- [UGC_WORKLOG_AND_SESSION_SOP.md](UGC_WORKLOG_AND_SESSION_SOP.md)
- [UGC_CHECKLIST_SOP.md](UGC_CHECKLIST_SOP.md)

If another SOP is stricter, the stricter SOP wins.

## Canonical Rule

A governance change is not complete until every dependent governance artifact is assigned one of these closure states:
- `synced`: updated or verified in the current session;
- `not applicable`: not relevant, with a short reason;
- `blocked`: cannot be completed safely now, with a blocker and debt file in `Plans/`;
- `requires separate approval`: intentionally deferred because it needs an approval gate outside the current scope.

`unknown`, `unchecked`, `assumed`, or "later" without a debt file is a hard stop.

## What Counts As A Governance Change

Use this SOP when the work creates or changes:
- `AGENTS.md` (active governance instructions);
- `SOPs/**` (SOP documents);
- Installed local agent skills or skill definitions;
- Governance or validation scripts;
- Git and repository workflows (e.g. CI/CD actions, rulesets, git hooks).

This SOP is not required for ordinary product code, copy, or test edits unless those edits change governance behavior or a governance dependency.

## Preflight Gate

Before implementing an approved governance change, record a preflight matrix in the plan or worklog.
Use the smallest useful version of this matrix:

| Dependency | Required Preflight State |
| --- | --- |
| Target governance file(s) | Named explicitly |
| Related SOPs and indexes | Listed and read |
| `SOPs/README.md` | Update needed / not needed with reason |
| `AGENTS.md` | Update needed / not needed with reason |
| Local installed governance skill | Update needed / not needed with reason |
| Validation scripts or workflows | Update needed / absent / debt needed / not applicable |
| Relevant plans/worklogs | Update required / not applicable |
| Approval gates | Explicitly named |

Do not begin implementation if any required dependency has no state.

## Closure Gate

Before final response, record a closure matrix in the worklog for every dependency considered in preflight.
The closure matrix must show:
- dependency;
- final state: `synced`, `not applicable`, `blocked`, or `requires separate approval`;
- evidence or file path;
- residual risk, if any.

A governance session may not be described as `done`, `complete`, or `closed` if any dependency remains unknown or unchecked.

## Sync And Debt Rules

When a governance change affects an SOP, skill, index, or validation rule, update the dependent artifact in the same session unless:
- the dependency is outside the approved scope;
- the dependency is missing and cannot be safely recreated;
- the dependency ownership is unclear.

For each blocked dependency, create or update a debt record in `Plans/` that names:
- missing artifact or update;
- why it was deferred;
- risk if it remains deferred;
- required approval or next plan.

## Validation

Allowed validation is local/static unless a stricter approved plan says otherwise:
- file existence checks;
- static searches for required references;
- static checklist review;
- git status for touched files only.

## Stop Conditions

Stop and report a blocker when:
- a dependency cannot be assigned an allowed closure state;
- the current thread lacks explicit approval for the governance file being changed;
- a proposed governance change conflicts with a stricter active SOP.
