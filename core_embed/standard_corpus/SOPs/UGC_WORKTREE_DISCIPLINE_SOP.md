# Safe Worktree Discipline SOP

Status: Active
Scope: Git workspace, worktree, and repository cleanliness rules

## Purpose

This SOP defines rules for managing safe worktrees and temporary workspaces to prevent:
- Untracked or safe worktrees spreading across arbitrary folders (such as Desktop).
- Accidental filesystem operations that break Git worktree metadata.
- Confusion between active and abandoned workspaces.
- Polluted operator contexts.

## Related SOPs

Read this SOP with:
- [SOPs/README.md](README.md)
- [UGC_WORKLOG_AND_SESSION_SOP.md](UGC_WORKLOG_AND_SESSION_SOP.md)

## Core Rule

Any safe worktrees or temporary workspaces created during development must live inside a structured, ignored container subdirectory inside the repository or a dedicated workspace folder, NOT loose on the desktop or root folders.

Recommended local container:
- `.worktrees/active/`
- `.worktrees/archive/`

Do not move worktree folders using standard filesystem move operations; always use `git worktree move`.

## Creation Rule

When a new safe worktree is needed:
1. Decide scope and target branch.
2. Create it only inside the designated `.worktrees` area.
3. Record its purpose in the worklog.
4. Assign its expected landing posture:
   - `active`
   - `archive`
   - `remove after completion`

## Non-Worktree Snapshot Rule

If a temporary workspace folder is created that is NOT a Git worktree (e.g. source bundles, recovery copies):
- Place it under an ignored path (e.g. `.archives/` or `.worktrees/`).
- Do not leave it loose in the repository root or Desktop.

## Move & Remove Rules

- Relocate worktrees ONLY via `git worktree move <worktree> <new-path>`.
- Remove worktrees cleanly via `git worktree remove <worktree>`.
- Run `git worktree prune` if stale metadata is left behind.

## Mandatory Closeout Posture

At the end of any session that created or used a worktree:
1. Leave it registered in `active` with a recorded reason, or
2. Move it to `archive`, or
3. Remove it cleanly.
Do not leave it unclassified.
