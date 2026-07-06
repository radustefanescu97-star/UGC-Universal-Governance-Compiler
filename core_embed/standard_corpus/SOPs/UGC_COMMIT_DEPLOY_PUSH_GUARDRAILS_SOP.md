# Universal Commit, Deploy, and Push Guardrails SOP

## Purpose

This SOP defines the mandatory repository guardrails before any `commit`, `push`, or `deploy` touching the project surfaces.

Its job is to stop:

- mixed branches
- accidental production regressions
- cross-thread contamination
- deploys from repo states that are not proven safe

This SOP applies to:

- application core components
- backend runtime services
- portal or internal admin surfaces
- any document or configuration change that can influence release behavior

It is complemented by:

- `UGC_WORKTREE_DISCIPLINE_SOP.md`
- `UGC_ORCHESTRATION_SOP.md`

## Core Rule

No `commit`, `push`, or `deploy` is allowed until the acting branch, worktree, scope, and target surface have been explicitly checked and found clean enough for the intended action.

If there is uncertainty, the correct action is to stop and isolate.

## Rule 1: One Scope Per Session

Every session must have:

- one clear objective
- one clear surface
- one clean worktree
- one dedicated branch

The worktree itself must follow `UGC_WORKTREE_DISCIPLINE_SOP.md`.

Forbidden:

- mixing API and hosting work casually
- mixing internal portal work into public runtime work
- bundling unrelated docs or experiments into the same push

If parallel agents are used, only the orchestrator controls final staging, commit, push, deploy, and handoff decisions. Delegated agents may edit only explicitly assigned, non-overlapping files.

## Rule 2: Mandatory Preflight Before Risky Actions

Before any `commit`, `push`, or `deploy`, run a preflight scan and record the answers:

1. What is dirty locally?
2. What is ahead of `origin/master`?
3. What is live on the exact surface being touched?
4. What files or branches belong to other threads and must be avoided?

Minimum checks:

- `git status --short --branch`
- `git diff --name-only`
- `git diff --name-only --cached`
- `git log --oneline origin/master..HEAD`
- `git log --oneline HEAD..origin/master`

If the action touches a live service or public hosting, also verify:

- current live revision or artifact
- target deploy surface
- rollback target

## Rule 3: Live-Aligned Source First

For production fixes on backend/runtime surfaces:

- start from a source state proven to match live
- do not patch blindly from `master`
- do not assume `master` equals live

Acceptable starting points:

- the exact live-aligned worktree or source bundle
- a branch explicitly rebased to the live revision state

Forbidden:

- cutting emergency runtime fixes from an unverified repo base

## Rule 4: Never Push From A Contaminated Branch

Do not push a branch to `master` if it contains work from another thread or another surface.

Examples of contamination:

- internal UI edits mixed into an API-only hotfix
- hosting edits mixed into a backend-only patch
- product copy changes mixed into runtime recovery work

Safe recovery paths:

- isolate the intended changes on a clean branch
- cherry-pick only the intended commit(s)
- stop and realign if the branch cannot be trusted

## Rule 5: Stage Only Owned Files

Before commit:

- stage only intended files
- inspect staged diff or staged stat
- confirm nothing unrelated is included

Never stage:

- cache noise
- unrelated dirty files
- files touched by another live thread
- opportunistic cleanup outside the current scope

## Rule 6: Commit Scope Must Be Narrow and Honest

A commit message must describe only what is truly inside the commit.

Forbidden:

- broad messages hiding mixed work
- "cleanup" commits that also contain behavior changes
- runtime fixes bundled with docs or hosting unless both are explicitly in scope

## Rule 7: Deploy Only the Intended Surface

Every deploy must target only the surface in scope.

Examples:

- API task: deploy API only
- Backend task: deploy backend only
- Hosting task: deploy hosting only
- Portal task: deploy portal only

If unrelated changes exist on the same surface, stop and isolate first.

## Rule 8: Push Checks Must Be Re-Run Right Before Push

Immediately before `push`, refresh branch delta checks against remote.

Do not rely on an earlier scan.

Minimum refresh:

- `git log --oneline origin/master..HEAD`
- `git diff --name-only origin/master..HEAD`

If anything unrelated appears, stop.

## Rule 9: Deploy Requires Verification

After any deploy, run targeted verification for the deployed surface.

Examples:

- health check
- one or more scoped smoke checks
- relevant logs
- portal snapshot if the change affects control-plane behavior

Verification must answer:

- did the intended surface change?
- did the main flow recover?
- did we introduce an unrelated regression?

## Rule 10: High-Risk Surfaces Need Extra Isolation

These areas are high-risk and must be treated conservatively:

- main `index.html` or entry point files
- rollout policy and serve-mode logic
- billing, pricing, and dashboard surfaces
- anything changing both product behavior and operator visibility

If the task is not explicitly about one of these surfaces, avoid it.

## Rule 11: Worktree Cleanliness Matters

The worktree should be as clean as possible before and after the task.

If unrelated dirty files exist:

- do not revert them
- do not stage them
- do not pretend they do not matter
- explicitly note the collision risk

Workspace location matters too. Follow standard practices for isolating safe worktrees.

## Rule 12: Block Risk, Not Reality

If a preflight finds:

- contamination
- live/local mismatch
- uncertain ownership
- wider deploy surface than intended

then the correct behavior is:

1. stop
2. explain the exact risk
3. isolate the work safely

The wrong behavior is to "push through" uncertainty.

## Required Checklist

Before `commit`:

- [ ] scope is explicit
- [ ] worktree and branch are identified
- [ ] preflight scan is complete
- [ ] staged files are owned by the session
- [ ] staged diff matches the real objective

Before `push`:

- [ ] branch delta vs `origin/master` was refreshed
- [ ] no unrelated commit is included
- [ ] push target is correct

Before `deploy`:

- [ ] target surface is explicit
- [ ] target surface is named canonically
- [ ] source surface is explicit if this is a promotion
- [ ] source is live-aligned if runtime hotfix
- [ ] unrelated changes on that surface are excluded
- [ ] rollback target is known
- [ ] no change in scope quietly worsens secret/config posture

After `deploy`:

- [ ] scoped verification ran
- [ ] revision or artifact identifier is recorded
- [ ] no unrelated regression signal appeared

After any session that created a safe workspace:

- [ ] workspace closeout is recorded
- [ ] any non-worktree safe snapshot does not remain loose on Desktop or repo root

## Final Rule

Speed matters.
Progress matters.
But branch truth, scope isolation, and production safety matter more.

The project must never be committed, pushed, or deployed from a state that we do not understand.
