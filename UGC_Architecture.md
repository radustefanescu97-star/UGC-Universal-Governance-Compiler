# Universal Governance Compiler (UGC)

## The Problem
AI agent environments (Google Antigravity, Cursor, GitHub Copilot, Windsurf) currently suffer from massive fragmentation. Each agent requires its own proprietary configuration format (e.g., `.agents/AGENTS.md`, `.cursorrules`, `.github/copilot-instructions.md`). This forces teams to maintain redundant rules, increasing the risk of compliance gaps when developers use different agents on the same repository.

## The Solution
The **Universal Governance Compiler (UGC)** is an open-source CLI tool and architectural framework. It establishes a **Single Source of Truth** for AI governance within a repository and compiles it down into the specific, vendor-locked formats required by individual agents.

### The Form of the Software

The UGC should be a lightweight, zero-dependency CLI tool. 
**Recommended Tech Stack:** **Go** (Golang) is ideal because it compiles into a single executable binary that works across Mac, Windows, and Linux without requiring a runtime (like Node or Python). Alternatively, **TypeScript/Node** is excellent for rapid prototyping and npm distribution.

The software is structured into three distinct layers:

#### 1. The Ingestion Layer (The Source)
The tool reads from a standard, agnostic directory (e.g., `.universal-governance/`).
*   **Auto-Discovery (`ugc analyze`):** Optionally, UGC can scan the project structure (`package.json`, `go.mod`, CI workflows) to infer existing tech-stack constraints and pre-populate the ingestion directory.
*   **`governance.md`**: Defines core constraints, truth order, intent classification, and stop conditions.
*   **`worklog-schema.md`**: Defines the audit trail format.
*   **`SOPs/`**: A folder containing standard operating procedures (e.g., "Deployment Protocol", "Code Review").

#### 2. The Translation Engine 
This layer parses the agnostic markdown/JSON rules into an internal mapping, matching universal concepts (like "Gated Action") to the specific capabilities of target agents.

#### 3. The Emitter Layer (The Compilers)
The engine outputs the proprietary files into the repository. To guarantee compliance and prevent governance dilution, UGC intentionally restricts the community plugin system for V1 and ships with a closed, auditable set of official emitters:
*   **Antigravity Emitter:** Generates `.agents/AGENTS.md` and compiles SOPs into distinct `.agents/skills/SKILL.md` directories.
*   **Cursor Emitter:** Flattens and concatenates rules into a single `.cursorrules` file.
*   **Copilot Emitter:** Formats guidelines into `.github/copilot-instructions.md`.
*   **Claude Code Emitter:** Formats guidelines into `CLAUDE.md`.

#### 4. The Audit Layer (Drift Detection)
*   **`ugc audit`:** A mandatory verification step that checks if the generated vendor-specific files (e.g., `.cursorrules`) match the exact hash of the compiled `.universal-governance/` source. This detects and rejects unauthorized manual edits to agent configs, protecting the Single Source of Truth.

### Runtime Compliance: The Worklog Problem
To ensure all agents respect a unified audit trail (e.g., `worklog.md`), the UGC injects **Runtime Injection Policies**. 

For highly capable agents like Antigravity, the UGC automatically compiles a specialized skill (e.g., `universal-worklog-sync/SKILL.md`). This skill acts as a mandatory runtime policy dictating: *"You may use your internal `task.md` or `walkthrough.md` for thinking, but before your execution ends, you MUST append your actions to the repository's main `worklog.md`."* 

## Open-Source Strategy & Security
Because governance involves strict compliance, security, and access control, the UGC must be open-source.
*   **Trust:** Organizations must be able to audit the compiler to ensure it doesn't introduce prompt-injection vulnerabilities or leak rules.
*   **Community Adapters:** As new AI IDEs and agents emerge, the community can easily contribute new "Emitters" for them without changing the core engine.
