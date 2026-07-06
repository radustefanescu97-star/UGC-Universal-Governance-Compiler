---
name: ugc-worklog-sync
description: Mandates syncing internal agent task and walkthrough data to the main project Plans/worklog.md file before ending a session.
---

# UGC Worklog Sync Skill

This skill enforces the UGC Runtime Compliance for worklogs. 

## Instructions

Whenever you are about to finish a material session (where you have created plans, touched files, or executed commands on the UGC project):

1. DO NOT end the conversation yet.
2. Read your internal progress from `task.md` and `walkthrough.md` (or your internal memory if artifacts aren't used).
3. Open or create the file `Plans/worklog.md` in the root of the repository.
4. Append a new entry to `Plans/worklog.md` detailing:
    - **Date/Time**: The current session date.
    - **Objective**: What you were asked to do.
    - **Actions Taken**: Files modified, commands run.
    - **Approvals**: Any approvals requested or bypassed.
    - **Stop Reason**: Why you are stopping (e.g., "Task complete", "Waiting for approval").
5. Only AFTER appending to `Plans/worklog.md` may you end your turn and notify the user.

This ensures the Single Source of Truth for project history is maintained in the repository, not trapped in the agent's ephemeral conversation history.
