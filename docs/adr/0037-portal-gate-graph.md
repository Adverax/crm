# ADR-0037: Portal as Gate Graph — Unified Navigation and Execution Model

**Date:** 2026-03-01

**Status:** Proposed

**Participants:** @roman_myakotin

**Supersedes:** ADR-0022 (Portal as bounded context adapter), ADR-0036 (Portal Action Model)

**Amends:** ADR-0027 (Layout embedded in Gate, `metadata.layouts` deprecated)

**Extends:** ADR-0035 (data binding model — queries move from Portal-level to Gate-level)

## Context

### Problem: flat Portal model cannot express multi-step workflows

After ADR-0035 (data binding) and ADR-0036 (action model), Portal evolved into a
capable page configuration with explicit queries and transactional actions. However,
the model remains **flat** — a single configuration that describes one screen:

| Aspect | Current state | Problem |
|--------|--------------|---------|
| Navigation | Actions return results but don't lead anywhere | Frontend manually hardcodes `router.push()` after action success |
| CRUD flow | List → Detail → Edit → Delete | Four separate Portals or one Portal with implicit URL-based routing |
| Layout | Separate `metadata.layouts` table, FK to Portal | Two entities to manage; complex merge logic in Describe handler |
| Actions | Transactional DML/Scenario with no output structure | Action completes, but what does the user see next? |
| Multi-step | Impossible | A wizard (step 1 → step 2 → confirmation) requires hardcoded frontend |

The core issue: **Portal describes a single page, but user interactions are
inherently a graph of transitions.** The current model forces the frontend to
implement navigation logic that should be declarative.

### Insight: Portal as a higher-order function

A Portal can be reconceptualized as a **graph of functions**. Each function
(called a **Gate**) receives input, executes a chain of operations, and returns
its result along with a list of available next steps:

```
Gate(args) → { layout?, datasets, outcomes, error? }
```

This maps naturally to CRUD:
- **List gate**: executes a query, returns records, offers "Create" and "Detail" outcomes
- **Detail gate**: executes a row query, returns record, offers "Edit", "Delete", "Back" outcomes
- **Create gate**: on GET returns empty form, on POST validates and inserts, offers "Detail" outcome
- **Delete gate**: executes DML, has no layout — auto-collapses back to List

The model unifies navigation and execution into a single declarative structure.

### Relationship with prior ADRs

| ADR | Relationship |
|-----|-------------|
| ADR-0022 | Superseded — Portal config changes from flat `read` to `gates` map |
| ADR-0027 | Amended — Layout is embedded in Gate; `metadata.layouts` table deprecated |
| ADR-0032 | Extended — Portal is still unbound from object_id; navigation still profile-based |
| ADR-0035 | Extended — PortalViewField and query type inference preserved, moved to gate level |
| ADR-0036 | Superseded — Actions become gates with DML body steps; no separate PortalAction model |

## Options

### Option A: Keep flat Portal + add navigation links

Add an `outcomes` field to existing PortalAction to declare where to navigate after
action completion. Portal remains a single-page config.

**Pros:** Minimal change.
**Cons:** Navigation is bolted on — an afterthought, not a first-class concept.
Cannot express multi-step workflows, wizards, or graphs where a "page" leads to
different outcomes depending on result. Actions and views remain separate concepts.

### Option B: Portal as Gate Graph (chosen)

Portal becomes a **graph of gates**. Each gate is a stateless function with:
body (SOQL/DML chain), optional layout, and declared outcomes (edges to other gates).

**Pros:**
- Unified model: navigation, data display, and actions are all gates
- CRUD is a graph pattern, not a special case
- Multi-step workflows are natural (wizard = linear gate chain)
- HTTP method semantics inferred from gate body (no explicit type declaration)
- Single entity in DB — entire graph is one Portal record
- Admin works with a coherent graph, not disconnected Portals

**Cons:**
- Significant refactoring of Portal types, handlers, and admin UI
- Existing Portal configs require migration
- Gate graph adds cognitive complexity for simple single-page portals

### Option C: Full state machine with persistent state

Each gate is a state, transitions are guarded, and state is persisted server-side
between transitions.

**Pros:** Enables long-running workflows with recovery.
**Cons:** Over-engineered for navigation. Stateful = session management, recovery,
conflict resolution. This is what the Scenario Engine (ADR-0025) is designed for.
Portal should remain stateless.

## Decision

**Option B: Portal as Gate Graph.**

### Core concepts

#### Portal = the graph

A Portal is a named graph of gates stored as a single record in `metadata.portals`.
The Portal-level config declares the entry gate and optional global args:

```go
type PortalConfig struct {
    Args      []PortalArg            `json:"args,omitempty"`
    EntryGate string                 `json:"entry_gate"`
    Gates     map[string]PortalGate  `json:"gates"`
}
```

#### Gate = a vertex in the graph

Each gate is a **stateless function**. It receives args, executes a body (chain
of SOQL/DML operations), and returns layout + datasets + outcomes:

```go
type PortalGate struct {
    Label    string         `json:"label"`
    Args     []PortalArg    `json:"args,omitempty"`
    Body     []GateBodyStep `json:"body"`
    Layout   *GateLayout    `json:"layout,omitempty"`
    Outcomes []GateOutcome  `json:"outcomes,omitempty"`
}
```

- **Args**: gate-specific typed parameters (merged with portal-level args at runtime)
- **Body**: ordered chain of SOQL/DML operations
- **Layout**: optional presentation config (nil = action gate)
- **Outcomes**: declared transitions to other gates

#### GateBodyStep = a single operation in the chain

```go
type GateBodyStep struct {
    Name string `json:"name"`
    Type string `json:"type"`            // "soql" | "dml"
    SOQL string `json:"soql,omitempty"`  // for type="soql"
    DML  string `json:"dml,omitempty"`   // for type="dml"
    When string `json:"when,omitempty"`  // CEL condition (skip if false)
}
```

Steps execute in order. Each step has a unique `name` used to reference its
result in subsequent steps and in the `datasets` response.

The optional `PageSize` field enables pagination and dynamic query capabilities
for this step (see "Dynamic query capabilities" below).

Updated type with pagination and filtering support:

```go
type GateBodyStep struct {
    Name            string   `json:"name"`
    Type            string   `json:"type"`                       // "soql" | "dml"
    SOQL            string   `json:"soql,omitempty"`              // for type="soql"
    DML             string   `json:"dml,omitempty"`               // for type="dml"
    When            string   `json:"when,omitempty"`              // CEL condition (skip if false)
    PageSize        int      `json:"page_size,omitempty"`         // enables pagination + dynamic filtering
    RestrictFilters []string `json:"restrict_filters,omitempty"`  // opt-out: exclude fields from filtering
}
```

#### GateLayout = presentation config

GateLayout merges what was previously split between Portal fields (ADR-0022)
and Layout config (ADR-0027):

```go
type GateLayout struct {
    Fields        []PortalViewField            `json:"fields,omitempty"`
    Root          *LayoutComponent             `json:"root,omitempty"`
    SectionConfig map[string]SectionConfig     `json:"section_config,omitempty"`
    FieldConfig   map[string]LayoutFieldConfig `json:"field_config,omitempty"`
    ListConfig    *ListConfig                  `json:"list_config,omitempty"`
}
```

- `Fields`: unified field list (from ADR-0035 PortalViewField)
- `Root`, `SectionConfig`, `FieldConfig`, `ListConfig`: presentation from ADR-0027

`metadata.shared_layouts` remains available — `FieldConfig` entries can reference
shared layouts via `layout_ref` (same shallow merge semantics as ADR-0027).

#### GateOutcome = an edge in the graph

```go
type GateOutcome struct {
    Name         string            `json:"name"`
    Gate         string            `json:"gate"`
    Label        string            `json:"label,omitempty"`
    Icon         string            `json:"icon,omitempty"`
    Type         string            `json:"type,omitempty"`  // "primary"|"secondary"|"danger"
    ArgsTemplate map[string]string `json:"args_template,omitempty"`
}
```

Outcomes are **pure declarations** — they tell the client what transitions are
available, not when to trigger them. The client decides based on user interaction.

`ArgsTemplate` maps gate arg names to CEL expressions evaluated against the current
gate's context (`args`, `datasets`). Example: `{"id": "datasets.account.Id"}`.

### HTTP method semantics — inferred from body

The gate's supported HTTP methods are determined by analyzing its body steps:

| Body content | GET | POST | Gate behavior |
|-------------|-----|------|---------------|
| Only SOQL steps | Yes | No (405) | View gate — returns layout + datasets |
| SOQL + DML steps, has layout | Yes | Yes | Form gate — GET shows form, POST validates and executes |
| Only DML steps, no layout | No (405) | Yes | Action gate — executes and auto-collapses |
| SOQL + DML steps, no layout | No (405) | Yes | Action gate — executes and auto-collapses |

**No explicit type declaration needed.** The gate's behavior emerges from its
composition — the admin adds body steps and layout, and the system derives what
HTTP methods are supported.

#### GET execution (read path)

On GET, only SOQL body steps execute. DML steps are skipped.
Steps with `page_size` support dynamic filtering, sorting, and pagination
via query parameters (see "Dynamic query capabilities"):

```
GET /api/v1/portal/account_mgmt/gate/list?filter.Industry=Tech&sort=Name:asc&page=2

1. Resolve portal "account_mgmt" from cache
2. Find gate "list"
3. Resolve args (portal-level + gate-level, from query params)
4. Execute body steps in order, SOQL only:
   - step "accounts" (page_size=20):
     a. Clone cached AST
     b. Inject filter: WHERE Industry = $1 (parameterized)
     c. Replace sort: ORDER BY Name ASC
     d. Apply pagination: LIMIT 20 OFFSET 20
     e. Execute → list result with total_count
   - step "save" (type=dml): SKIPPED
5. Resolve layout for current user's profile
6. Build query_config for paginated datasets (from cached field metadata)
7. Return: { layout, datasets, query_config, outcomes }
```

For steps without `page_size` (e.g., scalar queries), the SOQL executes as-is
with only arg substitution — no filter/sort/pagination injection:

```
GET /api/v1/portal/account_mgmt/gate/detail?id=uuid-123

1. Resolve portal and gate
2. Resolve args
3. Execute body steps in order, SOQL only:
   - step "account": SELECT ROW ... WHERE Id = :id → scalar result
   - step "contacts": SELECT ... WHERE AccountId = :id → list result
4. Resolve layout, return: { layout, datasets, outcomes }
```

#### POST execution (write path)

On POST, all body steps execute in order within a transaction:

```
POST /api/v1/portal/account_mgmt/gate/create
Body: { "args": { "name": "Acme", "industry": "Tech" } }

1. Resolve portal and gate
2. Resolve args (from POST body)
3. Execute ALL body steps in a transaction:
   - step "defaults": SELECT ROW ... → default values (SOQL)
   - step "insert": INSERT INTO Account ... → created record (DML)
   - step "result": SELECT ROW ... WHERE Id = :new_id → read back (SOQL)
4. If any step fails: rollback, return error + layout (for form re-render)
5. If success: return { datasets: {result: {...}}, outcomes: [...] }
```

#### Validation errors

When a DML step fails validation (via DML pipeline — ADR-0020), the gate returns
an error response with the layout intact, so the frontend can re-render the form
with error messages:

```json
{
  "gate": "create",
  "layout": { "fields": [...], "root": {...} },
  "datasets": { "defaults": {...} },
  "outcomes": [...],
  "errors": [
    { "field": "name", "message": "Name is required", "code": "required" }
  ]
}
```

### Chain collapse

When a gate has **no layout** and executes **successfully**, the frontend does not
render anything — it automatically navigates to the first outcome:

```
detail ──[Delete]──→ delete (no layout, DML, success) ──→ collapse ──→ list
                     ↑ user does not see this gate
```

If the action gate fails, the error is displayed in the context of the calling
gate (the one that declared the outcome).

### Stateless execution

Every gate call is independent. No server-side session or state is maintained
between calls. All necessary context is passed via args:

- Portal-level args: global parameters (e.g., `object_api_name`)
- Gate-level args: specific to this gate (e.g., `id` for detail view)
- Args are resolved from URL query params (GET) or request body (POST)

### Layout resolution per profile

The gate returns its layout as part of the response. Layout resolution is
profile-aware:

1. Gate has `layout` defined → use it
2. Gate has no `layout` → action gate (no presentation)

Future extension: per-profile layout overrides within a gate (e.g., mobile vs
desktop). For now, a single layout per gate is sufficient — the `form_factor`
concept from ADR-0027 can be reintroduced as a gate-level concern if needed.

### CRUD as a graph pattern

Standard CRUD for any object follows a common graph structure:

```
┌─────────┐     create      ┌─────────┐
│         │ ──────────────→ │         │
│  list   │                 │ create  │
│ (entry) │ ←────────────── │  (form) │
│         │     cancel       │         │
└────┬────┘                 └────┬────┘
     │                           │
     │ detail                    │ detail (on success)
     ↓                           ↓
┌─────────┐                 ┌─────────┐
│         │ ──────────────→ │         │
│ detail  │      edit       │  edit   │
│ (view)  │ ←────────────── │  (form) │
│         │     cancel       │         │
└────┬────┘                 └─────────┘
     │
     │ delete
     ↓
┌─────────┐
│ delete  │ ──→ collapse ──→ list
│(action) │
└─────────┘
```

### Full CRUD example: Account Management

```json
{
  "args": [],
  "entry_gate": "list",
  "gates": {
    "list": {
      "label": "Accounts",
      "body": [
        {
          "name": "accounts",
          "type": "soql",
          "soql": "SELECT Id, Name, Industry, Phone FROM Account ORDER BY Name",
          "page_size": 20
        }
      ],
      "layout": {
        "fields": [
          {"name": "Name"},
          {"name": "Industry"},
          {"name": "Phone"}
        ],
        "list_config": {
          "view": "table",
          "columns": [
            {"field": "Name"},
            {"field": "Industry"},
            {"field": "Phone"}
          ]
        }
      },
      "outcomes": [
        {"name": "create", "gate": "create", "label": "New Account", "icon": "plus", "type": "primary"},
        {"name": "detail", "gate": "detail", "label": "View", "args_template": {"id": "datasets.accounts[index].Id"}}
      ]
    },

    "detail": {
      "label": "Account Detail",
      "args": [{"name": "id", "type": "string"}],
      "body": [
        {
          "name": "account",
          "type": "soql",
          "soql": "SELECT ROW Id, Name, Industry, Phone, CreatedAt FROM Account WHERE Id = :id"
        },
        {
          "name": "contacts",
          "type": "soql",
          "soql": "SELECT Id, FirstName, LastName, Email FROM Contact WHERE AccountId = :id ORDER BY LastName"
        }
      ],
      "layout": {
        "fields": [
          {"name": "Name"},
          {"name": "Industry"},
          {"name": "Phone"}
        ],
        "root": {
          "type": "grid",
          "columns": 2,
          "children": [
            {"type": "field_section", "key": "details"},
            {"type": "related_list", "key": "contacts"}
          ]
        }
      },
      "outcomes": [
        {"name": "edit", "gate": "edit", "label": "Edit", "icon": "pencil", "type": "primary", "args_template": {"id": "args.id"}},
        {"name": "delete", "gate": "delete", "label": "Delete", "icon": "trash-2", "type": "danger", "args_template": {"id": "args.id"}},
        {"name": "list", "gate": "list", "label": "Back to List", "icon": "arrow-left"}
      ]
    },

    "create": {
      "label": "New Account",
      "body": [
        {
          "name": "insert",
          "type": "dml",
          "dml": "INSERT INTO Account (Name, Industry, Phone, OwnerId) VALUES (:name, :industry, :phone, :current_user_id)"
        }
      ],
      "layout": {
        "fields": [
          {"name": "Name"},
          {"name": "Industry"},
          {"name": "Phone"}
        ]
      },
      "outcomes": [
        {"name": "detail", "gate": "detail", "label": "View Created", "args_template": {"id": "datasets.insert.Id"}},
        {"name": "list", "gate": "list", "label": "Cancel"}
      ]
    },

    "edit": {
      "label": "Edit Account",
      "args": [{"name": "id", "type": "string"}],
      "body": [
        {
          "name": "account",
          "type": "soql",
          "soql": "SELECT ROW Id, Name, Industry, Phone FROM Account WHERE Id = :id"
        },
        {
          "name": "save",
          "type": "dml",
          "dml": "UPDATE Account SET Name = :name, Industry = :industry, Phone = :phone WHERE Id = :id"
        }
      ],
      "layout": {
        "fields": [
          {"name": "Name"},
          {"name": "Industry"},
          {"name": "Phone"}
        ]
      },
      "outcomes": [
        {"name": "detail", "gate": "detail", "label": "View", "args_template": {"id": "args.id"}},
        {"name": "list", "gate": "list", "label": "Cancel"}
      ]
    },

    "delete": {
      "label": "Delete Account",
      "args": [{"name": "id", "type": "string"}],
      "body": [
        {
          "name": "delete",
          "type": "dml",
          "dml": "DELETE FROM Account WHERE Id = :id"
        }
      ],
      "outcomes": [
        {"name": "list", "gate": "list", "label": "Back to List"}
      ]
    }
  }
}
```

Note: the `delete` gate has no `layout` — it auto-collapses on success, and the
frontend navigates to the `list` outcome. On failure, the error is shown in the
context of the `detail` gate (which declared the "delete" outcome).

### API

#### Gate execution endpoint (replaces query + action endpoints)

```
GET  /api/v1/portal/:portalApiName/gate/:gateName?arg1=val1&arg2=val2
POST /api/v1/portal/:portalApiName/gate/:gateName
     Body: { "args": { "arg1": "val1" }, "data": { "Name": "Acme" } }
```

The `data` field in POST contains form field values (for DML parameter substitution).

#### Response structure

For detail/form gates (no pagination):

```json
{
  "data": {
    "gate": "detail",
    "layout": {
      "fields": ["..."],
      "root": {"...": "..."},
      "section_config": {"...": "..."},
      "field_config": {"...": "..."}
    },
    "datasets": {
      "account": {"Id": "uuid-123", "Name": "Acme"},
      "contacts": [{"Id": "uuid-456", "FirstName": "John"}]
    },
    "outcomes": [
      {"name": "edit", "gate": "edit", "label": "Edit", "icon": "pencil", "type": "primary", "args_template": {"id": "args.id"}},
      {"name": "delete", "gate": "delete", "label": "Delete", "icon": "trash-2", "type": "danger"}
    ],
    "errors": null
  }
}
```

For list gates (with pagination and query_config):

```json
{
  "data": {
    "gate": "list",
    "layout": {
      "fields": ["..."],
      "list_config": {"...": "..."}
    },
    "datasets": {
      "accounts": {
        "records": [{"Id": "uuid-123", "Name": "Acme", "Industry": "Tech"}],
        "total_count": 150,
        "page": 1,
        "per_page": 20
      }
    },
    "query_config": {
      "accounts": {
        "filterable": [
          {"field": "Name", "type": "text", "operators": ["eq", "contains", "starts_with"]},
          {"field": "Industry", "type": "picklist", "operators": ["eq", "in"], "values": ["Tech", "Finance", "Healthcare"]},
          {"field": "Phone", "type": "phone", "operators": ["eq", "contains"]}
        ],
        "sortable": ["Name", "Industry", "Phone"],
        "searchable": ["Name", "Phone"],
        "default_sort": [{"field": "Name", "direction": "asc"}]
      }
    },
    "outcomes": [
      {"name": "create", "gate": "create", "label": "New Account", "icon": "plus", "type": "primary"},
      {"name": "detail", "gate": "detail", "label": "View"}
    ],
    "errors": null
  }
}
```

`query_config` is only present when at least one dataset has `page_size` configured.
It provides the frontend with all metadata needed to render filter controls,
sort headers, and search boxes — without hardcoding any field information.
```

#### Portal metadata endpoint (for navigation, admin)

```
GET /api/v1/portal/:portalApiName
```

Returns portal metadata: api_name, label, entry_gate, gate names (without body/SOQL
for security). Used by navigation system and admin UI.

#### Admin CRUD (unchanged path, new config shape)

```
POST   /api/v1/admin/portals
GET    /api/v1/admin/portals
GET    /api/v1/admin/portals/:id
PUT    /api/v1/admin/portals/:id
DELETE /api/v1/admin/portals/:id
```

Config shape changes from `{args, read: {fields, actions, queries}}` to
`{args, entry_gate, gates: {...}}`.

### Describe handler compatibility

The Describe handler (`GET /api/v1/describe/:objectName`) remains as a
**compatibility shim**. It resolves the Portal by object api_name, finds the
entry gate (or a gate named `detail`/`list`), and extracts its layout to produce
the same `formDescribe` response the frontend currently expects.

This allows existing RecordDetailView and RecordListView to work unchanged during
the migration period. New views can call the gate endpoint directly.

### Impact on `metadata.layouts`

The `metadata.layouts` table is **deprecated**. Layout config (root, section_config,
field_config, list_config) moves into `GateLayout` within each gate definition.

- Existing layout data is migrated into the corresponding gate's layout field
- `metadata.shared_layouts` **remains** — reusable layout fragments via `layout_ref`
- The `metadata.layouts` table is dropped after data migration

### Validation rules (at Portal save time)

1. **Entry gate exists** — `entry_gate` must reference a key in `gates`
2. **Outcome targets exist** — every `outcome.gate` must reference a key in `gates`
3. **No orphan gates** — every gate (except entry) must be reachable from entry
4. **Body step names unique** — within each gate
5. **Body step validity** — SOQL/DML syntax check, `when` CEL compiles
6. **Gate arg names unique** — no conflict with portal-level args
7. **Field names unique** — within each gate's layout
8. **DAG validation** — computed fields within layout (from ADR-0035)

### Platform limits

| Parameter | Limit | Rationale |
|-----------|-------|-----------|
| Max gates per Portal | 20 | Manageable graph complexity |
| Max body steps per gate | 10 | Transaction scope |
| Max outcomes per gate | 10 | UI usability |
| Max args per gate | 20 | Query param limit |
| DML transaction timeout | 5s | Prevent long-running transactions |

### CEL context for expressions

All CEL expressions within a gate have access to:

| Variable | Type | Available in | Description |
|----------|------|-------------|-------------|
| `args` | map | Body steps, outcomes, layout | Resolved gate arguments |
| `datasets` | map | Subsequent body steps, outcomes | Results of prior body steps |
| `data` | map | POST body steps | Form field values from POST request |
| `user` | object | All | Current user (id, profile_id, role_id) |

### Dynamic query capabilities — auto-derived from SOQL

#### Problem: how does the frontend know what's filterable?

A list gate returns a dataset, but the frontend needs metadata to render filter
controls, sort headers, and search boxes. Manually configuring `filterable`,
`sortable`, and `searchable` arrays per body step is verbose and error-prone —
the information already exists in the SOQL query and field definitions.

#### Solution: auto-derive query_config from SOQL AST + field metadata

At **portal load time** (when the portal config is cached), the system parses each
SOQL body step's query string into an AST. Combined with field definitions from
the metadata cache, it auto-generates a `query_config` per dataset:

```
SOQL: "SELECT Id, Name, Industry, Phone FROM Account ORDER BY Name"
                                          ↓
                                    AST parsing
                                          ↓
SELECT fields: [Id, Name, Industry, Phone]  +  field_definitions(Account)
                                          ↓
query_config: {
  filterable: [
    {field: "Name",     type: "text",     operators: ["eq","contains","starts_with"]},
    {field: "Industry", type: "picklist", operators: ["eq","in"], values: ["Tech","Finance",...]},
    {field: "Phone",    type: "phone",    operators: ["eq","contains"]}
  ],
  sortable: ["Name", "Industry", "Phone"],
  searchable: ["Name", "Phone"]
}
```

**No manual configuration needed.** The admin writes SOQL, and the system infers
everything the frontend needs to render filter/sort/search UI.

#### Field type → operator mapping

The system maps field types (from `metadata.field_definitions`) to available
filter operators:

| Field type | Subtype | Operators | UI control |
|-----------|---------|-----------|------------|
| text | — | eq, ne, contains, starts_with | Text input |
| text | email | eq, ne, contains | Text input |
| text | phone | eq, contains | Text input |
| text | url | eq, contains | Text input |
| text | textarea | contains | Text input |
| number | integer | eq, ne, gt, gte, lt, lte | Number input |
| number | decimal | eq, ne, gt, gte, lt, lte | Number input |
| number | currency | eq, ne, gt, gte, lt, lte | Number input |
| number | percent | eq, ne, gt, gte, lt, lte | Number input |
| datetime | date | eq, ne, gt, gte, lt, lte | Date picker |
| datetime | datetime | eq, ne, gt, gte, lt, lte | Datetime picker |
| boolean | — | eq | Checkbox |
| picklist | — | eq, ne, in, not_in | Dropdown (values from field config) |
| reference | — | eq, ne | Reference picker |

Picklist values are extracted from `field_definitions.config.values` — no separate
configuration needed.

Fields **not** in the SELECT clause are not filterable (they're not projected, so
filtering on them would be misleading to the user).

#### Searchable fields

Fields of type `text` (any subtype) are automatically marked as searchable.
Search generates a compound `OR` condition: `WHERE (Name LIKE '%term%' OR
Phone LIKE '%term%' OR ...)`.

#### Sortable fields

All projected fields except `textarea` and `reference` are sortable by default.
The default sort order comes from the SOQL `ORDER BY` clause (if present).

#### Client request format

The frontend sends filter/sort/search/pagination as query parameters:

```
GET /api/v1/portal/account_mgmt/gate/list
    ?filter.Industry=Tech
    &filter.Name.contains=Acme
    &sort=Name:asc
    &search=hello
    &page=2
    &per_page=25
```

Multiple filters are combined with AND. The backend injects them into the SOQL
AST — never via string concatenation.

#### Server-side filter injection

The gate handler processes dynamic query parameters by **modifying a clone of the
parsed AST**, not by string concatenation:

```
1. Load cached AST for the SOQL body step
2. Clone the AST
3. Parse filter params → append WHERE conditions to the cloned AST
4. Parse sort param → replace ORDER BY in the cloned AST (if provided)
5. Parse search param → append OR conditions for searchable fields
6. Apply pagination (LIMIT/OFFSET)
7. Compile modified AST → parameterized SQL
8. Execute with bound parameters
```

This approach:
- Prevents SQL injection (parameterized queries only)
- Preserves PostgreSQL query plan caching (parameterized form)
- Allows the original SOQL WHERE clause to coexist with user filters (AND)

#### Opt-out mechanism

By default, all fields derived from SOQL are filterable/sortable/searchable based
on their type. If an admin wants to **restrict** filtering on certain fields, the
gate body step supports an optional `restrict_filters` field:

```json
{
  "name": "accounts",
  "type": "soql",
  "soql": "SELECT Id, Name, Industry, Phone FROM Account",
  "page_size": 20,
  "restrict_filters": ["Id"]
}
```

This is opt-out, not opt-in — zero configuration for the common case.

#### query_config in gate response

The gate response includes `query_config` per dataset, giving the frontend all
metadata needed to render filter/sort/search UI:

```json
{
  "data": {
    "gate": "list",
    "layout": { "..." : "..." },
    "datasets": {
      "accounts": {
        "records": [...],
        "total_count": 150,
        "page": 1,
        "per_page": 20
      }
    },
    "query_config": {
      "accounts": {
        "filterable": [
          {"field": "Name", "type": "text", "operators": ["eq", "contains", "starts_with"]},
          {"field": "Industry", "type": "picklist", "operators": ["eq", "in"], "values": ["Tech", "Finance", "Healthcare"]},
          {"field": "Phone", "type": "phone", "operators": ["eq", "contains"]}
        ],
        "sortable": ["Name", "Industry", "Phone"],
        "searchable": ["Name", "Phone"],
        "default_sort": [{"field": "Name", "direction": "asc"}]
      }
    },
    "outcomes": [...],
    "errors": null
  }
}
```

`query_config` is **only included for datasets with `page_size` set** in their
body step. Datasets without `page_size` (e.g., scalar queries, small related lists)
return raw data without filter metadata.

#### Caching strategy

| What | Cached? | When invalidated |
|------|---------|-----------------|
| SOQL AST (parsed query) | Yes — at portal load | Portal config change |
| query_config (field metadata) | Yes — at portal load | Portal config change or field_definitions change |
| GateLayout | Yes — at portal load | Portal config change |
| Dataset results (query data) | Never cached | Always fresh (CRM data is dynamic) |

The AST and query_config are computed once when the portal is loaded into the
metadata cache. Only the actual query execution (with injected filters) happens
per request.

## Consequences

### Positive

- **Unified model.** Navigation, data display, and mutations are all gates in a graph.
  No artificial separation between "page" and "action".
- **Declarative navigation.** Frontend does not hardcode `router.push()` — it
  follows outcomes declared by the server.
- **Multi-step workflows.** Wizards, approval flows, and complex CRUD are natural
  graph patterns — no custom frontend code.
- **Single entity.** The entire graph is one Portal record — admin works with one
  coherent configuration, not scattered entities.
- **HTTP method inference.** No explicit gate type — behavior emerges from composition
  (SOQL-only = GET, DML = POST). Less configuration, fewer mistakes.
- **Chain collapse.** Action gates (no layout) auto-complete — the user sees a
  seamless experience without intermediate screens.
- **Stateless.** No server-side session. Every request is self-contained. Scales
  horizontally without session affinity.
- **Backward compatible.** Describe handler serves as a shim. Existing views work
  during migration.
- **Zero-config filtering.** The admin writes SOQL and the system auto-derives
  filterable/sortable/searchable fields from the AST and field metadata. No manual
  `filterable: [...]` arrays to maintain.

### Negative

- **Significant refactoring.** Portal types, handlers, validation, admin UI, and
  tests all change. This is a large surface area.
- **Graph complexity.** For simple single-page portals (e.g., a dashboard), the
  graph model adds unnecessary indirection (one gate is sufficient, but the concept
  of "graph" and "entry gate" still applies).
- **Migration effort.** Existing portal configs must be transformed from flat
  structure to gate graph. Automated via PL/pgSQL JSONB migration.
- **Layout loss of independence.** Currently, admins can configure layouts separately
  from portals. In the new model, layout is embedded in the gate — changing layout
  means editing the portal config. Mitigated by shared layouts (`layout_ref`).

### Risks

- **Admin UI complexity.** A graph editor is harder to build than a flat form editor.
  Mitigated: start with a tab-based gate list (no visual graph), add graph
  visualization later.
- **Config size.** A Portal with 5+ gates and full layouts may have a large JSONB
  config. Mitigated: JSONB compression in PostgreSQL, shared layouts reduce
  duplication.
- **Describe handler drift.** The compatibility shim may diverge from the gate model
  over time. Mitigated: plan to deprecate Describe in favor of direct gate calls.

## Migration strategy (high-level)

### Existing Portal configs → Gate model

Each existing Portal becomes a single-gate Portal:

```
Before: { args, read: { fields, actions, queries } }
After:  { args, entry_gate: "main", gates: { "main": { body: [...], layout: {...}, outcomes: [...] } } }
```

Transformation:
1. `read.queries` → gate body steps with `type: "soql"`
2. `read.fields` → gate layout fields
3. `read.actions` with DML `apply` → separate action gates with DML body steps
4. `read.actions` without `apply` → outcomes from main gate (UI-only links)
5. Layout data from `metadata.layouts` → merged into gate's `layout` field

### Down migration

Reverse transformation is possible for single-gate portals. Multi-gate portals
(created after migration) cannot be losslessly converted back.

## Related ADRs

- **ADR-0022** — Portal as bounded context (superseded by this ADR)
- **ADR-0027** — Layout + Form (amended: layouts embedded in gates)
- **ADR-0032** — Profile Navigation (unchanged: portal api_name still used for navigation items)
- **ADR-0035** — Data Binding Model (extended: queries and fields move to gate level)
- **ADR-0036** — Portal Action Model (superseded: actions become gates with DML body)
- **ADR-0009..0012** — Security layers (unchanged: SOQL/DML enforcement applies within gates)
- **ADR-0020** — DML Pipeline (unchanged: DML body steps go through full pipeline)
- **ADR-0024** — Procedure Engine (complementary: procedures for non-transactional side effects)
- **ADR-0025** — Scenario Engine (complementary: stateful long-running workflows)

---

## Amendment: ArgRules (2026-03-02)

### Context

Portal args originally embedded validation inline (`validation` + `error_message` on `PortalArg`).
This was limiting:

- Cross-arg rules (e.g., `args.min < args.max`) had to be arbitrarily placed on one arg
- Violated the project pattern where Validation Rules are separate from field definitions
- A single validation rule per arg is restrictive

### Decision

Extract validation into a separate `[]PortalArgRule` on `PortalConfig` and `PortalGate`.

```
PortalArgRule:
  name:          string   # unique within scope, matches ^[a-z][a-z0-9_]*$
  condition:     string   # CEL expression (must return bool), env: args only
  error_message: string   # returned to client when condition evaluates to false
```

**Placement:**
- `PortalConfig.arg_rules` — portal-level rules (validated before gate-level)
- `PortalGate.arg_rules` — gate-level rules (validated after portal-level)

**Limits:** max 10 rules per scope.

**CEL environment:** `args` variable only (same as `newPortalCELEnv()`).

**Runtime evaluation order:**
1. Resolve and type-convert args
2. Evaluate `config.arg_rules` (portal-level)
3. Evaluate `gate.arg_rules` (gate-level)
4. Execute body steps

### Consequences

- `PortalArg.validation` and `PortalArg.error_message` fields are removed
- `PortalArg` retains only: `name`, `type`, `default`
- Cross-arg validation is now straightforward (a single rule can reference multiple args)
- Multiple validation rules per arg scope allow composable error messages
