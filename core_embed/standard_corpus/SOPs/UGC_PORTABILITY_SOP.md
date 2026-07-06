# Backup Checkpoint and Portability SOP

Status: Active
Scope: Material sessions, source-control backup checkpoints, and work-in-progress portability across different environments

## Purpose

This SOP makes valuable development work portable across machines without cluttering the main branch or turning git history into a junk drawer.
It defines a **backup checkpoint** as a way to preserve scoped work and context remotely (e.g. on a checkpoint branch) before it is ready for code review or merging.

## Related SOPs

Read this SOP with:
- [SOPs/README.md](README.md)
- [UGC_WORKLOG_AND_SESSION_SOP.md](UGC_WORKLOG_AND_SESSION_SOP.md)
- [UGC_GOVERNANCE_CHANGE_SOP.md](UGC_GOVERNANCE_CHANGE_SOP.md)

## Canonical Rule

Every material session must perform a Backup Checkpoint Assessment before final handoff.
The agent must ask:
```text
Is valuable work or context at risk of existing only on this machine?
```
If yes, the agent must propose a scoped checkpoint. No commit or push should be executed without explicit approval.

## Assessment States

Use exactly one state:
- `not needed`: no local-only valuable work exists, or it is already safely remote.
- `recommended`: valuable work exists and should be checkpointed.
- `completed`: checkpoint branch/commit/push was created.
- `deferred by user`: user chose not to checkpoint.

## Branch Naming

Preferred format for checkpoint branches:
```text
dev/<area>-<scope>-<yyyymmdd>
```
Do not use vague names like `test`, `fix`, or `backup`.

## Fresh Environment Return Gate

A track is portable only when a new environment can:
1. Clone the repository;
2. Check out the track branch;
3. Find the plans, worklogs, and modifications;
4. Continue the work without needing local-only memory.

## Approval Gates

Explicit user approval is required to:
- Commit and push to a remote repository.
- Merge branches.
- Modify branch protection rules.
