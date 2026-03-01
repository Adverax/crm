# ADR-0036: OV Action Model — Read + Transactional Actions

**Date:** 2026-02-27

**Status:** Proposed

**Participants:** @roman_myakotin

## Context

### Problem: implicit CRUD operations with no execution model

After ADR-0032 (OV unbinding from object_id) and ADR-0035 (data binding model),
Object View became a general-purpose page configuration with explicit data sources
(queries). However, the **write side** of the model has fundamental issues:

| Aspect | Current state | Problem |
|--------|--------------|---------|
| `edit` config | Optional section with fields/validation/defaults/mutations | Assumes implicit CRUD on a single object |
| CRUD operations | Determined by URL context (has recordId → update, no recordId → create) | No explicit definition of what happens on "Save" |
| Actions | UI buttons with visibility_expr, no execution model | Cannot trigger DML, scenarios, or any server-side logic |
| Mutations | Raw DML strings in `edit.mutations` | Not integrated with DML engine, no transactional guarantees |
| Operation scope | `edit` mode = both Create and Update | No way to show different forms or apply different rules per operation |

The core issue: **OV knows how to display data (Read) but does not define what
operations can be performed on it (Write).** The `edit` section was a half-measure —
it assumed a single implicit object and fixed CRUD semantics.

### Relationship with prior ADRs

| ADR | Relationship |
|-----|-------------|
| ADR-0022 | Introduced OV with `view`/`edit` split; actions were UI-only decorations |
| ADR-0024 | Procedure Engine — JSON DSL for multi-step logic; contains non-transactional commands |
| ADR-0027 | Layout + Form — presentation layer with `mode: edit\|view` |
| ADR-0031 | Automation Rules — reactive triggers on DML events (post-transaction) |
| ADR-0032 | OV unbound from object_id — OV may not have an implicit target object |
| ADR-0035 | Data binding model — queries as first-class data sources for Read |

### Why Procedures don't fit as OV actions

ADR-0024 defined the Procedure Engine with 6 command types: record, compute,
flow, integration, notification, wait. Procedures can contain **non-transactional**
operations (HTTP calls, email, delays). This makes them unsuitable as the
execution model for OV actions, which must be **strictly transactional** — either
all changes commit together, or none do.

OV actions need a simpler, safer primitive: a set of DML operations within a
single database transaction, or a scenario start (which is itself just an INSERT
into `scenario_runs` — transactional).

### Why actions don't have their own fields

An earlier iteration of this ADR included per-action field definitions
(`OVActionField` with name, type, label, required, default) that would render
as a modal dialog when the user clicks an action button. This was removed for
three reasons:

1. **Duplication.** For CRUD actions (95% of cases), action fields duplicate
   information already present in `field_definitions` (type, label, required,
   default). The admin re-enters what metadata already knows.

2. **Bad UX.** The user sees a form on the page, clicks a button, and gets
   *another* form in a modal with different fields. This is confusing.

3. **Separation of concerns.** Actions operate on data **already present on
   the page** (form fields from OV read config). If an action needs its own
   input UI (e.g., "Send Email" with subject/body), it belongs to a different
   mechanism — a Procedure (ADR-0024) or a separate OV page.

## Options

### Option A: Read + Actions unified model (chosen)

Every OV page is fundamentally a **Read** (display data via queries + form) combined
with a list of **available Actions**. CRUD operations are not special — they are
predefined actions with the same structure as any custom action.

Each action is a complete operation unit:
1. **Validation** — auto-extracted from `metadata.validation_rules` for DML target fields
2. **Apply** (required) — transactional execution: DML set OR scenario start

Actions do not define their own input fields. They operate on data from the
current page: the form field values (`data`) and the current record (`record`).

**Pros:**
- Unified model — no artificial split between "CRUD" and "custom actions"
- Explicit — no action configured = operation not supported
- Transactional safety — actions are strictly within one DB transaction
- No field duplication — actions use page data, field metadata stays in one place
- Works with unbound OVs — no implicit target object assumed

**Cons:**
- More configuration for simple CRUD cases (mitigated by Constructor UI templates)
- Breaking change to OVConfig structure (migration required)
- Actions that need custom input require a separate mechanism (Procedure or OV page)

### Option B: Keep `edit` section, add action execution

Extend the existing `edit` section with an execution model. Keep `view`/`edit` split.

**Pros:** Minimal structural change. **Cons:** Still assumes implicit CRUD.
Cannot support custom actions beyond edit/create. Does not work with unbound OVs.

### Option C: Extend Layout `mode` to CRUD enum

Change Layout `mode` from `edit|view` to `create|read|update|delete`. Each mode
gets its own Layout.

**Pros:** Differentiates create from update at the form level.
**Cons:** Still hardcoded to 4 operations. Cannot add custom actions. Layout
proliferation (4 modes × 3 form factors = 12 layouts per OV).

## Decision

**Option A: Read + Actions unified model.**

### Conceptual model

The interaction flow is always the same:

```
1. User opens page → Read (queries execute, form renders data)
2. User sees available actions (buttons, based on visibility_expr)
3. User clicks an action
4. Frontend submits current page data + action key
5. Server: validate → execute transactional action
```

There is no modal dialog for actions. Actions operate on the data already
visible on the page. If an action needs only the record context (e.g., "Delete",
"Mark Hot"), no form data is required — just `record.id`.

For operations that require dedicated input (e.g., "Send Email" with subject
and body fields), the correct approach is:
- **Procedure** (ADR-0024) — for side-effect operations with their own UI
- **Separate OV page** — a dedicated page with its own fields and queries

### Data model changes

#### OVConfig restructured

The `view`/`edit` split is replaced by a single `read` section with embedded actions:

```go
type OVConfig struct {
    Read OVReadConfig `json:"read"`
}

type OVReadConfig struct {
    Queries []OVQuery     `json:"queries,omitempty"` // data sources (SOQL)
    Fields  []OVViewField `json:"fields,omitempty"`  // display fields
    Actions []OVAction    `json:"actions,omitempty"` // available operations
}
```

- `read.queries` — SOQL data sources (from ADR-0035).
- `read.fields` — display fields (from ADR-0035, unified OVViewField).
- `read.actions` — all available operations, including CRUD.

There is no `primary_object` or auto-generation. All actions are explicitly
configured by the admin. For convenience, Constructor UI may offer **templates**
(e.g., "Standard CRUD for Account") that pre-fill the action config, but this
is a UX feature — not a model concept. The generated config is identical to
manually written config.

#### OVAction — no own fields

```go
type OVAction struct {
    // Identity + UI
    Key            string `json:"key"`              // unique within OV
    Label          string `json:"label"`            // UI button text
    Type           string `json:"type"`             // "primary"|"secondary"|"danger"
    Icon           string `json:"icon"`             // lucide icon name
    VisibilityExpr string `json:"visibility_expr"`  // CEL: show/hide button

    // Execution model
    Apply *OVActionApply `json:"apply,omitempty"` // transactional action
}
```

Actions do not have a `Fields` property. They receive data from the page form
(`data` in CEL context) and the current record (`record`).

When `apply` is nil, the action is UI-only (e.g., a link or a client-side toggle).
When `apply` is set, the action is executable server-side.

#### Transactional apply

```go
type OVActionApply struct {
    Type     string         `json:"type"`               // "dml" | "scenario"
    DML      []string       `json:"dml,omitempty"`      // DML query texts
    Scenario *OVScenarioRef `json:"scenario,omitempty"` // for type="scenario"
}
```

**Type "dml"** — a set of DML queries executed within a single transaction.
Each string is a DML query text that the DML engine parses and executes:

```
INSERT INTO Account (Name, Industry, OwnerId) VALUES (data.Name, data.Industry, user.id)
UPDATE Account SET Status = 'hot' WHERE id = record.id
DELETE FROM Task WHERE AccountId = record.id AND Status = 'cancelled'
```

DML queries go through the full DML pipeline (parse → resolve → validate →
execute) with OLS/FLS enforcement. CEL expressions within values are evaluated
against the action's context variables (`data`, `user`, `record`, `result`).

**Type "scenario"** — starts a scenario (INSERT into scenario_runs):

```go
type OVScenarioRef struct {
    APIName string            `json:"api_name"`           // scenario api_name
    Params  map[string]string `json:"params,omitempty"`   // param_name → CEL expr
}
```

Both types are strictly transactional. DML operations go through the DML engine
(with OLS/FLS enforcement). Scenario start is an INSERT into `scenario_runs`.

#### Custom action replaces standard DML

When an action has `apply` defined, it **completely defines** the transactional
behavior. There is no implicit "also do the standard INSERT/UPDATE". If the admin
wants "Create Deal + Create Task", they must include both DML queries in `apply.dml[]`.

#### No action = operation not supported

If no action with a given key is configured on the OV, the operation is not
available. The button is not rendered, and the server returns 404 if the action
key is submitted. This is explicit by design — no implicit CRUD.

### CEL context for expressions

All CEL expressions within an action have access to:

| Variable | Type | Description |
|----------|------|-------------|
| `data` | map | Current form data from the page (OV field values) |
| `user` | object | Current user (id, profile_id, role_id) |
| `record` | map | Current record data (from default query, if available) |
| `result` | list | Results of previous DML operations in the same transaction (for chaining) |

The `data` variable contains the current values of the page form fields —
the same fields defined in OV `read.fields` and rendered by the Layout.
This is the key difference from the earlier design: there is no separate
action-specific form. Actions always work with page data.

The `result` variable enables chaining: a second DML can reference the ID of a
record created by the first DML (e.g., `result[0].id`).

### Example: standard CRUD actions

For an Account object, the admin (or Constructor UI template) configures:

**Action "create":**
```json
{
  "key": "create",
  "label": "Create",
  "type": "primary",
  "icon": "plus",
  "apply": {
    "type": "dml",
    "dml": [
      "INSERT INTO Account (Name, Industry, Status, OwnerId) VALUES (data.Name, data.Industry, data.Status ?? 'new', user.id)"
    ]
  }
}
```

**Action "edit":**
```json
{
  "key": "edit",
  "label": "Save",
  "type": "secondary",
  "icon": "check",
  "visibility_expr": "has(record)",
  "apply": {
    "type": "dml",
    "dml": [
      "UPDATE Account SET Name = data.Name, Industry = data.Industry WHERE Id = record.id"
    ]
  }
}
```

**Action "delete":**
```json
{
  "key": "delete",
  "label": "Delete",
  "type": "danger",
  "icon": "trash-2",
  "visibility_expr": "has(record)",
  "apply": {
    "type": "dml",
    "dml": [
      "DELETE FROM Account WHERE Id = record.id"
    ]
  }
}
```

**Custom action "mark_hot":**
```json
{
  "key": "mark_hot",
  "label": "Mark Hot",
  "type": "secondary",
  "icon": "flame",
  "visibility_expr": "has(record) && record.Status != 'hot'",
  "apply": {
    "type": "dml",
    "dml": [
      "UPDATE Account SET Status = 'hot' WHERE Id = record.id"
    ]
  }
}
```

These are all identical in structure — CRUD and custom actions are indistinguishable
at the model level. None of them define their own fields — they all use `data.*`
(page form values) and `record.*` (current record).

### Validation rules auto-extraction

Actions do **not** define their own validation rules. Instead, validation rules
are automatically extracted from `metadata.validation_rules` at Describe API time,
filtered by the fields each DML statement modifies.

The extraction algorithm:

1. Parse DML statements → extract target objects and modified fields
   (INSERT columns, UPDATE SET fields, UPSERT columns; DELETE is skipped)
2. For each target object: load validation rules from `metadata.validation_rules`
3. For each active rule: parse the CEL expression AST, extract `record.*` field
   references via `cel-go` AST walker (`PreOrderVisit` + `SelectKind`/`IdentKind`)
4. Include the rule only if the intersection of rule fields and DML fields is non-empty
5. Deduplicate by rule ID (same rule from multiple DMLs targeting the same object)

The extracted rules are included in the Describe API response as
`FormAction.validation_rules[]`:

```json
{
  "key": "create",
  "label": "Create",
  "type": "primary",
  "icon": "plus",
  "validation_rules": [
    {
      "expression": "size(record.Name) > 0",
      "error_message": "Name is required",
      "error_code": "name_required"
    }
  ]
}
```

The frontend uses these rules for client-side pre-validation (fail fast with
user-friendly messages). The DML pipeline validation remains the **enforcement
layer** (cannot be bypassed).

For **scenario** actions, validation is the scenario's responsibility — no
automatic extraction is performed.

For **DELETE** operations, no frontend validation is extracted (no fields modified).

### Impact on Layout model

The Layout table currently uses `mode: edit|view`. With the new action model:

- **Layout `mode: view`** → becomes the **Read layout** (how to display data)
- **Layout `mode: edit`** → **deprecated** (actions use page form, not a separate form)

Migration path: existing `mode: edit` layouts are preserved for backward
compatibility but are no longer used for form resolution. The `mode` column
constraint changes from `('edit', 'view')` to `('read', 'view')` in a future
migration, or `edit` is treated as an alias for `read`.

### Actions needing custom input

Some operations require user input that is not part of the page form (e.g.,
"Send Email" needs subject and body). These are **not** OV actions — they
belong to a different mechanism:

| Mechanism | When to use | Example |
|-----------|-------------|---------|
| **Procedure** (ADR-0024) | Side-effect operations with their own Constructor UI | Send Email, HTTP integration, Create + notify |
| **Separate OV page** | Operations needing a dedicated form with its own queries | Email compose page, Bulk update wizard |

The OV action can **navigate** to a Procedure or another OV page via a UI-only
action (no `apply`, just client-side navigation). Or the Procedure can be
triggered by an Automation Rule (ADR-0031) as a post-DML hook.

### API changes

#### Execute action endpoint

```
POST /api/v1/view/:ovApiName/action/:actionKey
Content-Type: application/json

{
  "data": { "Name": "Acme Corp", "Industry": "Technology" },
  "record_id": "uuid-of-existing-record"  // optional, for actions on existing records
}
```

The `data` field contains the current page form values. The frontend collects
field values from the rendered form and submits them with the action key.

Response (success):
```json
{
  "success": true,
  "results": [
    {"operation": "insert", "object": "Account", "id": "new-uuid"}
  ]
}
```

Response (validation error):
```json
{
  "success": false,
  "errors": [
    {"code": "invalid_amount", "message": "Amount must be positive", "field": "amount"}
  ]
}
```

#### Describe API extension

The Describe response includes action definitions (without `apply`
details — security). The frontend uses this to render action buttons:

```json
{
  "api_name": "Account",
  "fields": [...],
  "form": {
    "queries": [...],
    "actions": [
      {
        "key": "create",
        "label": "Create",
        "type": "primary",
        "icon": "plus",
        "validation_rules": [
          {
            "expression": "size(record.Name) > 0",
            "error_message": "Name is required",
            "error_code": "name_required"
          }
        ]
      },
      {
        "key": "edit",
        "label": "Save",
        "type": "secondary",
        "icon": "check",
        "visibility_expr": "has(record)",
        "validation_rules": [
          {
            "expression": "size(record.Name) > 0",
            "error_message": "Name is required",
            "error_code": "name_required"
          }
        ]
      },
      {
        "key": "delete",
        "label": "Delete",
        "type": "danger",
        "icon": "trash-2",
        "visibility_expr": "has(record)"
      }
    ]
  }
}
```

Note: `apply` is **not included** in the Describe response — it contains
server-side logic (DML targets, CEL expressions) that should not be exposed
to the client. `validation_rules` are auto-extracted from metadata at Describe
time — they are not stored in OV config. Actions no longer include field
definitions either — the page form fields serve as the input source.

### Platform limits

| Parameter | Limit | Rationale |
|-----------|-------|-----------|
| Max actions per OV | 20 | UI usability |
| Max DML operations per action | 10 | Transaction scope |
| DML transaction timeout | 5s | Prevent long-running transactions |

## Consequences

### Positive

- **Unified model.** CRUD and custom actions share the same structure. No special
  cases, no implicit behavior.
- **Explicit operations.** Each OV declares exactly what can be done — no guessing
  from URL context.
- **Transactional safety.** All operations within an action execute in a single DB
  transaction, or the entire action rolls back.
- **No field duplication.** Actions use page form data. Field definitions (type,
  label, required, default) stay in metadata — single source of truth.
- **No confusing modals.** Users interact with the page form directly. No
  surprise modal with different fields when clicking an action button.
- **Works with unbound OVs.** No assumption about a target object — actions
  explicitly declare their DML targets.

### Negative

- **More verbose config.** Simple CRUD requires explicit action definitions
  (mitigated by Constructor UI templates).
- **Breaking change.** OVConfig structure changes significantly. Existing OV
  configs must be migrated to the new format.
- **Custom input requires separate mechanism.** Operations needing their own
  input fields (e.g., "Send Email") cannot be simple OV actions — they must
  use Procedures or separate OV pages. This is intentional separation of concerns.
- **DML chaining via `result`.** Referencing previous DML results in subsequent
  queries requires index-based access (`result[0].id`), which is fragile if
  query order changes.

### Risks

- **Transaction scope.** Multiple DML operations in one transaction increase lock
  contention. Mitigated by the 10-operation limit and 5s timeout.
- **CEL expression complexity.** Expressions in DML values and `where` can become
  hard to debug. Mitigated by a dry-run endpoint (from Procedure Engine pattern).
- **Configuration burden.** Every OV requires manual action setup. Mitigated
  by Constructor UI templates and the ability to copy/clone OV configs.

### Related ADRs

- **ADR-0022** — Object View (supersedes the `view`/`edit` config split)
- **ADR-0024** — Procedure Engine (for operations needing custom input or side effects)
- **ADR-0027** — Layout + Form (Layout `mode` impact)
- **ADR-0031** — Automation Rules (post-transaction hooks — complementary)
- **ADR-0032** — OV unbinding (enables unbound OVs with explicit action targets)
- **ADR-0035** — Data binding model (queries as Read data sources)
