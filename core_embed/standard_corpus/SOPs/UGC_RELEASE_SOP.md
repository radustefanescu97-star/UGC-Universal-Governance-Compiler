# Universal Release and Promotion SOP

## Purpose

This document defines the mandatory release workflow for the project so that:

- `production` stays stable for users
- product improvements do not get tested directly on production
- every deploy has a clear validation path and a clear rollback path
- frontend, API, and operational changes are promoted in a controlled way

This SOP is launch-critical and must be followed for all releases, hotfixes, and production promotions.

## Related SOPs

This release SOP is complemented by:

- `UGC_COMMIT_DEPLOY_PUSH_GUARDRAILS_SOP.md`
- `UGC_ORCHESTRATION_SOP.md`

## Production Rule

Production is not a development surface.
No direct experimentation, debugging, trial fixes, or partial validation happens on production.

If a change has not already been validated in a non-production surface, it does not go to production.

If parallel agents are used during release or incident work, they may investigate bounded questions, but there is still one orchestrator-owned production track. Delegated reports do not authorize deploys, promotions, rollback actions, or traffic shifts.

## Environment Roles

### 1. `development` (Local / Dev)

Purpose:

- engineering sandbox
- risky iteration
- debugging
- exploratory fixes
- experimental UI or runtime changes

Allowed:

- incomplete work
- debugging instrumentation
- temporary logs
- behavior experiments

Not allowed:

- user-facing production promotion

### 2. `staging` (Pre-Production)

Purpose:

- release candidate surface
- near-production validation
- final product verification before production

Allowed:

- only changes already stabilized in `development`
- integrated validation of frontend + API behavior
- smoke testing with real product flows

Not allowed:

- half-finished experiments
- direct debugging hacks
- knowingly broken or partial states

### 3. `production` (Live)

Purpose:

- live user traffic only

Allowed:

- only controlled promotions from `staging`
- only approved hotfixes with rollback plan

Not allowed:

- direct development
- speculative fixes
- “let’s try it live”

## Non-Negotiable Release Policy

Every change follows this path:

1. `development`
2. `staging`
3. `production`

There is no direct `development -> production` path.

## Scope Control

Before any release work starts, define:

- exact files in scope
- exact services in scope
- exact deploy target in scope
- exact things explicitly out of scope

If scope changes mid-session, stop and re-approve scope before continuing.

## Branch and Worktree Rules

Every release session uses:

- a clean dedicated worktree
- a dedicated branch
- one release objective per branch

Never mix unrelated changes in the same branch or commit.
Never use a dirty worktree for release promotion.

## Commit / Push / Deploy Guardrails

Follow repository guardrails from `UGC_COMMIT_DEPLOY_PUSH_GUARDRAILS_SOP.md`.

Mandatory rules:

- no direct production tinkering without a scoped branch/worktree
- no deploy before diff review
- no push of unrelated changes
- no API traffic shift before validation on a no-traffic revision
- no frontend production promotion before validation in `staging`

## Release Types

### A. Frontend-only release

Examples:

- UI rendering
- layout fixes
- copy and messaging

Required flow:

1. validate in `development`
2. validate in `staging`
3. confirm browser-new-session behavior
4. promote to production hosting

### B. API / backend release

Examples:

- API services
- backend tool dispatch
- database migrations

Required flow:

1. validate in `development` or isolated environment
2. deploy `no-traffic` to staging/production
3. smoke the tagged revision directly
4. shift production traffic only after successful smoke

### C. Integrated release

Examples:

- frontend + API behavior changes
- data contract changes

Required flow:

1. validate in `development`
2. validate in `staging`
3. deploy API `no-traffic`
4. validate API revision
5. promote frontend and API in controlled order
6. confirm end-to-end on production immediately after promotion

## Mandatory Release Gate

No change goes to production until all applicable checks below are green.

### Browser Sessions
Must pass in:
- fresh browser session
- incognito session
- authenticated session

### Console and Logs
Must have:
- no uncaught frontend exceptions for release-critical flows
- no unexpected 4xx/5xx on release-critical endpoints
- no new production errors introduced by the candidate

## Production Freeze Rule

When the project is in a stable launch state:

- production enters release freeze
- only P0 hotfixes or approved validated promotions may enter production

No opportunistic improvements during freeze.
No “small harmless tweak” directly on production.

## Hotfix SOP

If production is broken:

1. stop all non-essential release work
2. identify whether the issue is frontend, API, or config
3. choose the smallest safe fix or rollback
4. if possible, validate in `staging` first
5. verify production immediately after release
6. sync GitHub to live state after production is stable

## Rollback Rule

Every release must have a known rollback target before promotion.

That means we must know, before deploy:

- current live frontend version
- current live API revision
- current config envelope if relevant

If production degrades:

- rollback immediately to last known good state
- investigate after service is restored

## Emergency Incident Response

This section applies when production is degraded, unstable, or behaving unexpectedly for real users.

### Incident Priorities

#### `P0`
Production is unusable or materially degraded for users.

#### `P1`
Production works, but trust or critical UX is visibly degraded.

#### `P2`
Non-critical defect, workaround exists, no direct launch-risk impact.

### Immediate Response Rules

When a `P0` or `P1` incident is active:

1. stop all unrelated release work immediately
2. no parallel “while we are here” improvements
3. no new features
4. no scope expansion
5. no architecture work mixed with the hotfix

The only goal is service restoration.

### Incident Commander

Every incident must have one temporary owner for the response.
That owner is responsible for:

- freezing scope
- deciding whether the fix is frontend, API, or config
- enforcing rollback when needed
- keeping a single source of truth for what changed

There must not be multiple simultaneous release tracks into production during an incident.

### First 10 Minutes Rule

In the first 10 minutes of a production incident, do only this:

1. reproduce the issue
2. confirm whether the failure is frontend, API, or config
3. identify the last known good revision or artifact
4. decide: rollback now, or smallest safe hotfix

Do not spend the first 10 minutes speculating or broadening scope.

### Rollback-First Criteria

Rollback is mandatory if any of the following is true:

- users cannot complete standard flows
- production was healthy before the last promotion
- the cause is not yet isolated
- the hotfix would touch multiple systems at once
- production is worse than the last known good state

In those cases:

- restore service first
- investigate second

### Hotfix Criteria

A production hotfix is allowed only if:

- the fault domain is isolated
- the patch is smaller and safer than rollback
- the scope is tightly bounded
- rollback target is still known

### Production Incident Logging

For every `P0` or `P1`, capture:

- timestamp first observed
- exact user-visible symptom
- affected environment
- last known good revision/artifact
- actual remediation applied
- final verification result

This should be written into an incident note or worklog before the incident is considered closed.

### GitHub Sync Rule After Incident

After production is restored:

- GitHub must be brought back in sync with live state
- but only after service is stable

Never leave production ahead of source control for long.
Never push a dirty or mixed branch just to “catch up”.

### Resume Rule

After an incident:

- no return to feature work until the incident is closed
- no return to the original larger plan until:
  - production is green
  - rollback target is refreshed
  - GitHub is aligned
  - next scope is re-approved

## Definition of Done for a Release

A release is done only if:

- the change passed `development`
- the change passed `staging`
- production promotion followed this SOP
- production smoke is green
- no new console/runtime errors appear for release-critical flows
- rollback target is known
- GitHub state is brought back in sync with live state
