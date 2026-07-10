# UGC GUI Contract (Phase 0 JSON Surfaces)

Status: Phase 0 (`UGC-GUI-P0-5` + `UGC-GUI-P0-FIX`)  
Audience: UGC Cockpit (GUI) implementers and CLI maintainers  
Schema policy: **additive-only** until GUI v1.0

## Purpose

The UGC Cockpit is a native GUI over the `ugc` CLI. It must not reimplement parsing, emission, drift detection, hashing, or audit logic. It consumes **machine-readable CLI output** documented here.

Human-readable CLI output is unchanged and is not part of this contract.

## Stdout and stderr rules

| Stream | `--json` mode | Human mode |
| --- | --- | --- |
| **stdout** | Exactly **one** JSON object (first byte `{`). No other lines. | Human command output (`Build plan:`, `Audit complete…`, etc.) |
| **stderr** | One clean error line on failure (no cobra `Usage:` dump). Emitter progress (`Emitting … configuration...`) during audit/build. | Same emitter progress on stderr; errors on stderr |

The GUI must parse **stdout only** for JSON. Progress and errors belong on stderr.

## Schema evolution

Every JSON surface includes `schema_version` (integer, currently `1`).

Until **GUI v1.0**:

- **Allowed:** new optional fields within the same `schema_version`
- **Forbidden without escalation:** removing fields, changing field types, or changing exit-code semantics
- **Breaking changes:** require a new `schema_version` and a separate approved packet

Golden examples live in `cli/testdata/gui_contract/`. Regenerate with:

```bash
UGC_REGENERATE_GOLDENS=1 go test ./cli -run TestRegenerateGUIContractGoldens -count=1
```

Tests in `cli/gui_contract_test.go` and `cli/gui_contract_stdout_test.go` validate golden shapes and assert the **built binary** emits a single JSON object on real stdout (no emitter leakage).

### Redaction rules for golden examples

| Field | Rule |
| --- | --- |
| `source_hash`, `sha256` | Fixture-specific hex; live values differ per repo state |
| `binary_version`, `go_version`, `platform` | Live values follow the running binary |
| `expected_artifacts`, `items`, list lengths | Vary by fixture; types and keys are stable |
| `corpus_state_message` | Wording may vary slightly; `corpus_state` enum is stable |
| `capability_coverage` | Full matrix in golden; same shape as live output |

---

## 1. `ugc audit --json`

**Invocation:** `ugc audit --json`  
**Stdout:** single JSON object  
**Stderr:** emitter progress during drift check; parse/runtime errors  
**Exit code:** `0` when `audit_passed` is true; non-zero when audit fails or hard-errors

On a **hard audit error** (e.g. emit/temp-dir failure), JSON still prints with `audit_passed: false` and `source_valid: false`.

### Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `schema_version` | int | yes | Currently `1` |
| `audit_passed` | bool | yes | `true` when no drift, manifest, or source failures |
| `source_valid` | bool | yes | `true` when governance source parses; `false` on hard-error path |
| `source_errors` | string[] | yes | Parse/validation errors (empty when valid) |
| `source_hash` | string | when valid | SHA-256 of connected governance source |
| `drift` | object[] | yes | `{ "path", "message" }` per drift item |
| `unexpected_artifacts` | string[] | yes | Generated files not in manifest |
| `manifest_findings` | string[] | yes | Build manifest problems |
| `corpus_state` | string | yes | `ok`, `missing`, `legacy`, `failed`, or `unknown` |
| `corpus_state_message` | string | yes | Human-readable corpus state detail |
| `capability_coverage` | object | yes | **Capability → target → status** map (from `engine.TargetCapabilityMatrix()`) |
| `expected_artifacts` | string[] | yes | Manifest-owned generated paths |

**`capability_coverage` status values:** `constrained`, `instructed`, `advisory`, `native-skill`

Example path: `coverage["approval_gates"]["codex"]` → `"constrained"`

### Golden example

Path: `cli/testdata/gui_contract/audit_clean.json` (full matrix captured from real output)

```json
{
  "schema_version": 1,
  "audit_passed": true,
  "source_valid": true,
  "corpus_state": "missing",
  "capability_coverage": {
    "approval_gates": {
      "codex": "constrained",
      "cursor": "constrained",
      "claude": "constrained",
      "antigravity": "instructed"
    },
    "worklog_duty": {
      "codex": "native-skill",
      "cursor": "instructed"
    }
  }
}
```

---

## 2. `ugc build --dry-run --json` (preview)

**Invocation:** `ugc build --dry-run --json`  
**Stdout:** single JSON object  
**Stderr:** emitter progress  
**Exit code:** `0` when `has_blockers` is false; non-zero when blockers exist (JSON still printed)

### Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `schema_version` | int | yes | Currently `1` |
| `dry_run` | bool | yes | Always `true` for this invocation |
| `has_blockers` | bool | yes | `true` if unmanaged files would be overwritten |
| `items` | object[] | yes | `{ "path", "status", "reason?" }` per planned write |
| `summary` | object | yes | Stable counts for all statuses (zero-filled when absent) |

**Item `status` values:** `create`, `unchanged`, `managed-overwrite`, `blocked-unmanaged`

**`summary` keys (always present):** `create`, `unchanged`, `managed-overwrite`, `blocked-unmanaged` — each an integer count.

### Golden example

Path: `cli/testdata/gui_contract/build_dry_run.json`

```json
{
  "schema_version": 1,
  "dry_run": true,
  "has_blockers": false,
  "summary": {
    "create": 0,
    "unchanged": 12,
    "managed-overwrite": 0,
    "blocked-unmanaged": 0
  }
}
```

---

## 3. `ugc build --json` (apply result)

**Invocation:** `ugc build --json` (no `--dry-run`)  
**Stdout:** single JSON object on **successful** apply  
**Stderr:** emitter progress  
**Exit code:** `0` on success; non-zero on blockers or apply errors (no JSON guaranteed on failure)

### Fields

Same schema as §2. Differences:

| Field | Value on apply |
| --- | --- |
| `dry_run` | `false` |
| `has_blockers` | `false` on success |

Represents the plan that was applied, not a second dry-run.

### Golden example

Path: `cli/testdata/gui_contract/build_apply.json`

### GUI round-trip (governed SOP edit)

1. User edits `.universal-governance/SOPs/...`
2. GUI runs `ugc build --dry-run --json` → show preview
3. User confirms → GUI runs `ugc build --json`
4. GUI runs `ugc audit --json` → show pass/fail

Automated coverage: `cli/gui_roundtrip_test.go`, `cli/gui_contract_stdout_test.go`

---

## 4. `ugc packet verify --json`

**Invocation:** `ugc packet verify --packet <path> --approval "<sentence>" --json`  
**Stdout:** single JSON object  
**Stderr:** one error line on failure (no `Usage:` block)  
**Exit code:** `0` when `ok` is true; non-zero when verification fails (JSON still printed)

### Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `schema_version` | int | yes | Currently `1` |
| `ok` | bool | yes | Verification result |
| `task_id` | string | on success | Task id from approval sentence |
| `packet_path` | string | on success | Packet path from approval sentence |
| `sha256` | string | on success | Packet hash from approval sentence |
| `reasons` | string[] | yes | Failure reasons (empty on success) |

### Golden example

Path: `cli/testdata/gui_contract/packet_verify_pass.json`

---

## 5. `ugc update --dry-run --json` (corpus preview)

**Invocation:** `ugc update --dry-run --json`  
**Stdout:** single JSON object (no emitter calls; stdout stays clean)  
**Exit code:** `0` on success  
**Note:** `--json` without `--dry-run` is rejected with one stderr error line

### Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `schema_version` | int | yes | Currently `1` |
| `dry_run` | bool | yes | Always `true` |
| `state_warning` | string | optional | Present when `.state.json` is missing or legacy |
| `summary` | object | yes | Counts: `created`, `updated`, `unchanged`, `skipped_local_edits`, `skipped_unverified_legacy`, `failed` |
| `created` | string[] | yes | Relative paths under `.universal-governance/` |
| `updated` | string[] | yes | Same |
| `unchanged` | string[] | yes | Same |
| `skipped_local_edits` | string[] | yes | Same |
| `skipped_unverified_legacy` | string[] | yes | Same |
| `failed` | string[] | yes | Same |

### Golden example

Path: `cli/testdata/gui_contract/update_dry_run.json`

---

## Optional handshake: `ugc version --json`

Not part of the governed edit round-trip, but the GUI may call this at startup to confirm binary/corpus compatibility.

**Invocation:** `ugc version --json` (use `--no-check` to skip GitHub release lookup)

### Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `schema_version` | int | yes | Currently `1` |
| `binary_version` | string | yes | CLI version |
| `corpus_version` | string | yes | Embedded standard corpus version |
| `go_version` | string | yes | Toolchain used to build the binary |
| `platform` | string | yes | `GOOS/GOARCH` |
| `latest_release` | string | optional | Present when update check succeeds |
| `update_available` | bool | optional | Present when update check succeeds |

Golden example (no network): `cli/testdata/gui_contract/version_no_check.json`

---

## Validation

| Check | Command |
| --- | --- |
| Contract + golden shape | `go test ./cli/... -run GUIContract` |
| Real stdout JSON guard | `go test ./cli/... -run JSONSurfacesBuiltBinaryStdout` |
| Full regression | `go test ./...` and `go vet ./...` |
| Regenerate goldens | `UGC_REGENERATE_GOLDENS=1 go test ./cli -run TestRegenerateGUIContractGoldens` |

## Related documents

- `Plans/UGC_GUI_master_plan.md` — Phase 0 packet definitions
- `Plans/UGC_GUI_Phase0_redteam_review.md` — findings addressed by `UGC-GUI-P0-FIX`
- `cli/gui_roundtrip_test.go` — SOP edit round-trip integration test
