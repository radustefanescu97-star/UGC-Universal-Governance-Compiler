# Approval Packet Standard SOP

Status: Active
Scope: Project approval-gated planning, implementation, validation, and git actions

## Purpose

This SOP keeps approval requests short enough to review while preserving auditability.
Its job is to allow a concise chat approval to reference a complete hash-bound approval packet artifact (implementation plan) instead of repeating every allowed action, forbidden action, stop condition, artifact hash, command, and Return Gate in chat.

This SOP does not authorize any action by itself. It does not reduce any stricter SOP requirement.

## Related SOPs

Read this SOP with:
- [SOPs/README.md](README.md)
- [UGC_REPOSCAN_AND_BOOTSTRAP_SOP.md](UGC_REPOSCAN_AND_BOOTSTRAP_SOP.md)
- [UGC_WORKLOG_AND_SESSION_SOP.md](UGC_WORKLOG_AND_SESSION_SOP.md)
- [UGC_ENGINEERING_SIMPLICITY_SOP.md](UGC_ENGINEERING_SIMPLICITY_SOP.md)
- [UGC_GOVERNANCE_CHANGE_SOP.md](UGC_GOVERNANCE_CHANGE_SOP.md)

If another SOP is stricter, the stricter SOP wins.

## Canonical Rule

An approval may be short only when it references a complete approval packet artifact by path and SHA256 hash.
The approval packet artifact carries the full operational boundary.
The chat approval carries the operator decision.
The short approval must still contain the literal word `aprobare` or `aproval`.

## Approval Packet Requirements

A hash-bound approval packet must include:
- task id;
- packet path;
- packet SHA256;
- target surface;
- protected neighboring surfaces;
- source truth, plan truth, and local/worktree truth if relevant;
- allowed actions;
- forbidden actions;
- exact commands or command-rendering rules when applicable;
- validation plan;
- stop conditions;
- worklog/checklist path.

If any required field is missing, the packet is not approval-ready.

## Short Approval Form

Use this form when a complete packet exists:

```text
aprobare pentru executarea <TASK_ID>, conform approval packetului <PACKET_PATH> SHA256 <HASH>. Scope-ul, allowed actions, forbidden actions, stop conditions si Return Gate raman exact cele din packet. Fara actiuni in afara packetului.
```

Equivalent wording is acceptable only if it includes:
- `aprobare` or `aproval`;
- exact task id;
- exact packet path;
- exact packet SHA256;
- explicit statement that scope, allowed actions, forbidden actions, stop conditions, and Return Gate are exactly those in the packet;
- explicit statement that no action outside the packet is authorized.

## Drift Rule

If the packet changes after approval, the approval is stale.
The operator must approve the updated packet hash before execution continues.
Do not treat an old hash, old path, or old packet title as approval for a changed packet.

## Worklog Rule

Material sessions using a hash-bound approval packet must record in the worklog:
- packet path;
- packet SHA256;
- approval sentence or a concise reference to it;
- whether packet hash drift was checked before execution;
- what action was performed;
- whether any requested action was outside the packet and therefore blocked.

## Red Team And Review Rule

Reviewers should reject a packet when:
- it lacks a path or SHA256;
- it relies on prose instead of exact allowed/forbidden actions;
- it omits stop conditions;
- it hides cost, deploy, provider, cloud, secret, env, traffic, or git actions in broad wording;
- the packet hash does not match the artifact being executed.
