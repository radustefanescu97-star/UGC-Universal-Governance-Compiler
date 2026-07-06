# Orchestrator and Parallel Agents SOP

Status: Active
Scope: AI-assisted sessions that use an orchestrator and delegated parallel agents

## Purpose

This SOP defines how the AI may operate as an orchestrator when decomposing work.
Its job is to make parallel sub-agents useful without allowing them to:
- bypass canonical SOPs;
- create recursive child-agent chains;
- treat delegated reports as final truth;
- make critical decisions without explicit human approval.

## Related SOPs

This SOP must be read with:
- [UGC_WORKLOG_AND_SESSION_SOP.md](UGC_WORKLOG_AND_SESSION_SOP.md)
- [UGC_CHECKLIST_SOP.md](UGC_CHECKLIST_SOP.md)
- [UGC_APPROVAL_PACKET_SOP.md](UGC_APPROVAL_PACKET_SOP.md)

If another SOP is stricter, the stricter SOP wins.

## Canonical Rule

The orchestrator may decompose work into bounded parallel tasks, but the orchestrator remains accountable for synthesis, correctness, SOP compliance, final recommendations, and final edits.
Delegated agent output is evidence, not authority. No delegated report is self-executing.

## When Orchestrator Mode Is Appropriate

Use orchestrator mode only when parallel work materially improves safety, speed, or review quality.
Good uses:
- read-only research across separate repository areas;
- independent SOP review;
- separate risk analysis tracks;
- codebase mapping;
- implementation across disjoint files or modules with explicit ownership.

Poor uses:
- tiny tasks where delegation adds overhead;
- tasks where the next step depends on one immediate answer;
- work where parallelism would create overlapping edits.

## Orchestrator Accountability Rule

The orchestrator owns the final result.
The orchestrator must:
- write specific tasks for each delegated agent;
- assign clear ownership and boundaries;
- name applicable SOPs;
- make read-only versus edit permission explicit;
- forbid child agents in every delegated task;
- review every report before acting;
- compare reports against source files, SOPs, and current user instructions;
- personally decide what to edit, run, recommend, or ask approval for;
- document material orchestration in the relevant `Plans/` worklog.

The orchestrator must not say "the agents approved it" as a substitute for review.

## SOP Canonicality Rule For All Agents

All agents must treat `SOPs/` as canonical.
Every delegated task must name the relevant SOPs when known.
If a delegated agent discovers another applicable SOP, it must follow the stricter rule and report that discovery.
No delegated agent may loosen, reinterpret, bypass, or silently ignore an active SOP.

## No Recursive Delegation Rule

Only the top-level orchestrator may assign delegated agents.
Delegated agents must not spawn, request, instruct, simulate, schedule, coordinate, or rely on any sub-agent or child agent.
Every delegated task must include this exact rule:

```text
Do not spawn, request, instruct, simulate, schedule, coordinate, or rely on any sub-agent or child agent. You must complete this task yourself.
```

Every delegated report must confirm whether any sub-agent or child-agent behavior was used or requested.

## Default Permission Rule

Delegated agents are read-only by default.
Edit permission must be explicit and must include:
- file or directory ownership;
- protected neighboring files or surfaces;
- validation expectations;
- forbidden actions.

Two agents must not receive overlapping edit ownership.

## Mandatory Delegated Task Template

Every delegated task must use this template:

```md
You are Agent <ID> for an orchestrated task.

Coordination:
- You are one of multiple parallel agents.
- You are not alone in the codebase.
- Do not edit files unless this task explicitly says edits are allowed.
- Do not spawn, request, instruct, simulate, schedule, coordinate, or rely on any sub-agent or child agent. You must complete this task yourself.

Task:
- <Specific objective>

Scope:
- Target surface:
- Protected neighboring surfaces:
- Files/directories you may inspect:
- Files/directories you may edit, if any:
- Explicitly out of scope:

Canonical SOPs:
- <List applicable SOPs>
- If another SOP applies, follow the stricter rule and report it.

Allowed actions:
- <Read-only commands, specific checks, or specific edits>

Forbidden actions:
- No unauthorized edits.
- No child agents.
- No secrets/keys modification.
- No cost-generating tests.

Required output:
- Scope completed:
- Files/SOPs inspected:
- Evidence:
- Findings:
- Risks:
- Recommended next action:
- What was not checked:
- SOP conflicts or gaps:
- Confirmation of no forbidden actions:
- Confirmation that no sub-agents or child agents were used or requested:
```

## Orchestrator Review And Iteration Rule

After delegated reports return, the orchestrator must:
1. verify report claims against available evidence where material;
2. compare reports for contradictions;
3. classify findings as accepted, rejected, unresolved, or follow-up required;
4. stop and ask the user when the next step requires approval;
5. record material delegated work in the relevant worklog.

## Worklog Rule

For material orchestrated sessions, the worklog must record:
- orchestrator identity;
- delegated agent roster;
- task given to each agent;
- read-only or edit permission for each agent;
- SOPs named for each agent;
- whether no-child-agent instructions were included;
- report summaries;
- accepted/rejected findings;
- follow-up tasks sent;
- stop reason.
