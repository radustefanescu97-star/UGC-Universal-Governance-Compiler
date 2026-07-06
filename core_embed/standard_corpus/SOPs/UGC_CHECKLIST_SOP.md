# Post-Session SOP Compliance Checklist SOP

Status: Active
Scope: Every material AI-assisted session

## Purpose

This SOP defines the post-session SOP compliance checklist.
Its job is to prevent:
- silent SOP drift;
- incomplete handoffs;
- missing worklog entries;
- unclear target surfaces;
- hidden residual risks;
- accidental leakage of secrets.

This SOP is documentation discipline. It does not authorize runtime changes.

## Related SOPs

This SOP must be read with:
- [UGC_WORKLOG_AND_SESSION_SOP.md](UGC_WORKLOG_AND_SESSION_SOP.md)
- [UGC_REPOSCAN_AND_BOOTSTRAP_SOP.md](UGC_REPOSCAN_AND_BOOTSTRAP_SOP.md)
- [UGC_APPROVAL_PACKET_SOP.md](UGC_APPROVAL_PACKET_SOP.md)
- [UGC_GOVERNANCE_CHANGE_SOP.md](UGC_GOVERNANCE_CHANGE_SOP.md)

## Canonical Rule

Every material session must close with an SOP compliance checklist.
Record the checklist in the relevant `Plans/` worklog.
For trivial answer-only sessions, the final response may state that no material repo/runtime action occurred.

## Mandatory Checklist Fields

The checklist must answer:
- Which SOPs were consulted or applied?
- Which SOPs were not applicable, and why?
- Was the target surface explicit?
- Were protected neighboring surfaces named?
- Was the correct workspace/worktree used?
- Were dirty or unrelated files avoided?
- Was the engineering simplicity check completed, with any new abstraction justified in writing?
- Were secrets, credentials, and customer data protected?
- Were parallel agents used, and was the no-child-agent rule enforced?
- Were delegated reports reviewed by the orchestrator?
- Was validation run, and what passed or failed?
- If governance artifacts changed, was the governance preflight and closure matrix completed and recorded?
- Was the relevant worklog updated?
- What residual risk remains?
- Why did the session stop?

## Minimum Checklist Template

Use this template in the relevant worklog:

```md
### SOP Compliance Checklist

- SOPs consulted:
- SOPs applied:
- SOPs not applicable:
- Target surface:
- Protected neighboring surfaces:
- Workspace/worktree:
- Dirty state handled:
- Engineering simplicity check:
- Secrets/data handling:
- Parallel agents:
- Reports reviewed:
- Recursive delegation blocked:
- Validation:
- Governance change preflight:
- Governance closure matrix:
- Worklog updated:
- Residual risk:
- Stop reason:
```

## Final Response Rule

For material sessions, the final response must include a brief SOP compliance summary.
It does not need to reproduce the full checklist if the full checklist is in the worklog, but it must say:
- whether the checklist was completed;
- where it was recorded;
- any SOP exceptions or unresolved items;
- whether parallel agents were used.

## Exception Rule

If the user explicitly asks for investigation only, the checklist should state:
- no code/runtime changes made;
- evidence sources used.

## Final Rule

No material session is complete until SOP compliance has been checked, recorded, and handed off clearly.
