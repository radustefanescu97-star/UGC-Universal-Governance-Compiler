# UGC Clarify Intent SOP

Status: Active
Created: 2026-07-05
Scope: AI agent conversation, intent interpretation, and ambiguity resolution

## Purpose

This SOP prevents AI agents from executing destructive or unapproved mutations based on conversational assumptions.
Its job is to force the agent to stop and ask the user for clarification when the verbal command is ambiguous between a read-only request (e.g. "check this", "investigate") and a write request (e.g. "fix this", "amend this").

## Canonical Rule

When a user's instruction is ambiguous, underspecified, or could be interpreted as either an investigation or a mutation, the agent MUST treat it as a Hard Stop.

The agent must NOT proactively generate artifacts, mutate files, or propose hash-bound packets for implementation without explicitly clarifying if the user intended a write action.

## Clarification Gate

Before writing any new plan artifact or mutating any workspace file on an ambiguous command, the agent must:
1. Halt execution of any file generation.
2. Output a conversational response asking the user to clarify their intent. Example: "Did you want me to just audit this file, or do you want me to write a plan to fix the errors?"
3. Wait for the user's explicit response.

## Stop Conditions

- **STOP** and clarify if the command uses generic verbs like "check", "verify", "look into", but the agent detects errors that would normally require a fix.
- **STOP** if the user provides feedback on a plan but does not explicitly say "update the plan" or "fix it".
