# Universal AI Governance Operating Instructions

This file is the mandatory entrypoint for AI-assisted work in this repository. 
We use this governance standard to ensure safe, predictable, and auditable AI coding behavior.

## Mandatory Bootstrap
Before any repository scan, implementation, architecture change, or file mutation, you MUST read the following SOPs in order:

1. `SOPs/README.md`
2. `SOPs/UGC_REPOSCAN_AND_BOOTSTRAP_SOP.md`
3. `SOPs/UGC_WORKLOG_AND_SESSION_SOP.md`
4. `SOPs/UGC_ENGINEERING_SIMPLICITY_SOP.md`

## Additional Active SOPs
The following SOPs apply to specific operational actions:
- **`SOPs/UGC_CLARIFY_INTENT_SOP.md`**: Rules for stopping and clarifying ambiguous user intent before mutating files.
- **`SOPs/UGC_APPROVAL_PACKET_SOP.md`**: Rules for creating and approving hash-bound implementation plans.
- **`SOPs/UGC_GOVERNANCE_CHANGE_SOP.md`**: Preflight and closure gates for any governance/ruleset modification.
- **`SOPs/UGC_ORCHESTRATION_SOP.md`**: Coordination rules for parallel/delegated sub-agents.
- **`SOPs/UGC_CHECKLIST_SOP.md`**: Post-session compliance checklist format.
- **`SOPs/UGC_WORKTREE_DISCIPLINE_SOP.md`**: Git workspace, worktree, and repository cleanliness rules.
- **`SOPs/UGC_PORTABILITY_SOP.md`**: Rules for backup checkpoints and portability across development environments.
- **`SOPs/UGC_AI_AGENTS_GOVERNANCE_SOP.md`**: Rules for AI agents operating on the project.
- **`SOPs/UGC_COMMIT_DEPLOY_PUSH_GUARDRAILS_SOP.md`**: Mandatory repository guardrails for commit and deploy.
- **`SOPs/UGC_RELEASE_SOP.md`**: Mandatory release workflow and environment promotion.
- **`SOPs/UGC_REPOSITORY_EXPLAINABILITY_SOP.md`**: Repository orientation and high-risk file mapping.
- **`SOPs/UGC_REPOSITORY_EXPLAINABILITY_AND_DOC_SYNC_SOP.md`**: Doc-sync debt and documentation governance.

## The Golden Rules
- **Clarify Intent:** Do not make assumptions on ambiguous commands. Stop and clarify if the user wants an investigation or a mutation.
- **No changes without a plan:** Do not mutate files without first writing a plan.
- **Approval for Architecture:** Any change to the core engine architecture or the structure of the project requires explicit approval.
- **No secrets/costs:** Do not add cloud dependencies, API keys, or cost-generating loops without explicit authorization.

## Approval Rule
Explicit approval is required for:
- Releasing a new version.
- Changing the architecture of the parsing engine or project emitters.
- Modifying the underlying governance rules.

Approval is valid ONLY when the current thread contains the exact word `aprobare` or `aproval` from the project owner. Generic direction like "ok" or "go" is NOT approval for these gated actions.
