# ADR-0035: OV Data Binding Model

**Date:** 2026-02-27

**Status:** Accepted

**Participants:** @roman_myakotin

## Context

### Problem: disconnected Fields, Queries, and Computed

After ADR-0032 (OV unbinding from object_id), Object View became a general-purpose
screen configuration that can represent any page — record detail, dashboard, or
multi-object view. However, the data model still has fundamental limitations:

| Component | Current state | Problem |
|-----------|--------------|---------|
| `OVViewConfig.Fields` | `[]string` — bare field names | No data source declaration; assumes "current record" |
| `OVViewConfig.Queries` | `[]OVQuery` with `name`+`soql`+`when` | Defined but nothing references them; no `type` or `default` flag |
| `OVViewConfig.Computed` | `[]OVViewComputed` — separate array | Disconnected from Fields; cannot be placed in sections alongside regular fields |

This creates several issues:

1. **No explicit data binding.** Fields implicitly read from "the record", but
   there is no formal concept of "the record" — it is just whatever the URL params
   resolve to. Dashboards and multi-object views cannot exist in this model.

2. **Queries are orphaned.** They are declared but never connected to fields or
   the rendering pipeline. The frontend has no way to know which query provides
   data for which field.

3. **Computed fields are second-class.** They live in a separate `computed` array
   and cannot participate in sections, highlights, or list columns alongside
   regular fields.

4. **No cycle detection.** Computed fields can reference each other, but there is
   no validation to prevent circular dependencies.

### Relationship with prior ADRs

| ADR | Relationship |
|-----|-------------|
| ADR-0022 | Introduced OV as bounded context adapter with sections/actions/queries |
| ADR-0027 | Layout + Form — presentation layer that consumes OV fields |
| ADR-0032 | Unbound OV from object_id, making OV a general-purpose page config |

## Options

### Option A: Query-first explicit binding (chosen)

Queries become first-class data sources. Every field can optionally reference a
query via a CEL expression in `expr`. Fields without `expr` are simple field
references resolved against the default query.

```json
{
  "queries": [
    {"name": "main", "soql": "SELECT ROW Id, Name FROM Account WHERE Id = :id"},
    {"name": "contacts", "soql": "SELECT Id, Name FROM Contact WHERE AccountId = :id"}
  ],
  "fields": [
    {"name": "Name"},
    {"name": "contact_count", "type": "int", "expr": "size(contacts)"},
    {"name": "display_name", "type": "string", "expr": "main.Name + ' (' + main.Industry + ')'"}
  ]
}
```

**Pros:**
- Explicit data flow — every field's source is traceable
- Enables dashboard pages (multiple queries, no implicit record)
- Unifies regular fields and computed fields into one array
- Cycle detection via DAG validation at save time
- Per-query data endpoint for frontend consumption

**Cons:**
- Breaking change to `OVViewConfig.Fields` (mitigation: backward-compat UnmarshalJSON)
- Slightly more verbose for simple record views

### Option B: Source enum per field

Add a `source` field to each entry: `"record"`, `"query:<name>"`, `"computed"`.

**Pros:** Explicit source identification.
**Cons:** Verbose enum, still needs special handling for computed. Does not
naturally support expressions that combine multiple queries.

### Option C: Keep implicit record + extend

Keep `Fields []string`, add optional `field_sources` map.

**Pros:** Minimal change to existing configs.
**Cons:** Two parallel structures to maintain, doesn't solve the computed field
problem, implicit assumptions remain.

## Decision

**Option A: Query-first explicit binding.**

### Data model changes

#### OVViewField replaces bare string

```go
type OVViewField struct {
    Name string `json:"name"`           // field API name
    Type string `json:"type,omitempty"` // "string"|"int"|"float"|"bool"|"timestamp" (for computed)
    Expr string `json:"expr,omitempty"` // CEL expression referencing queries
    When string `json:"when,omitempty"` // visibility condition
}
```

Fields without `expr` are simple field references. Fields with `expr` are computed.
This **unifies** the old `Fields` and `Computed` arrays into a single `Fields` array.

#### OVQuery — type inferred from SOQL syntax

```go
type OVQuery struct {
    Name string `json:"name"`
    SOQL string `json:"soql"`
    When string `json:"when,omitempty"`
}
```

The query type (scalar vs list) is determined by the SOQL syntax — one source of truth:
- `SELECT ROW Id, Name FROM Account WHERE Id = :id` → scalar (single record)
- `SELECT Id, Name FROM Contact WHERE AccountId = :id` → list (multiple records)

The `ROW` keyword is a SOQL extension that:
- Forces `LIMIT 2` at compile time (to detect >1 row error)
- Returns error if more than 1 record is found at runtime
- Returns `null` (not error) if 0 records are found

The first scalar query (`SELECT ROW`) in the array is the implicit default (context record).

#### OVViewComputed is removed

Its functionality is absorbed by `OVViewField` with `expr` set.

### Validation rules (at OV save time)

1. Query name uniqueness — no duplicates
2. Field name uniqueness — no duplicates
3. Expression references valid queries (e.g., `main.Name` → query `main` exists)
4. Field expressions cannot reference list queries (only scalar queries allowed)
5. **DAG validation** — fields form a directed acyclic graph (Kahn's algorithm).
   Cycle → save rejected with error message listing the cycle

### Per-query data endpoint

```
GET /api/v1/view/:ovApiName/query/:queryName?param1=val1&param2=val2
```

- Finds OV by `api_name`, finds query by `name`
- Substitutes URL query params into SOQL `:paramName` placeholders
- Executes via SOQL service (with full security: OLS, FLS, RLS)
- Returns query result with pagination

### Backward compatibility

`OVConfig.UnmarshalJSON` handles the old format:
- `fields: ["name", "email"]` → `fields: [{name: "name"}, {name: "email"}]`
- `computed: [{name, type, expr, when}]` → appended to `fields` as `OVViewField`

### Describe API extension

`FormDescribe` gains a `queries` array (query metadata without SOQL — security).
Type is inferred from SOQL syntax (`SELECT ROW` = scalar, `SELECT` = list):

```json
{
  "queries": [
    {"name": "main", "type": "scalar"},
    {"name": "contacts", "type": "list"}
  ]
}
```

## Consequences

### Positive

- **Explicit data flow.** Every field's data source is declared and traceable.
- **Dashboard pages.** OV can define multiple queries without an implicit record.
- **Unified fields.** Regular and computed fields coexist in one array, simplifying
  sections, highlights, and list columns.
- **Safety.** Cycle detection prevents infinite loops in computed field graphs.
- **Per-query endpoint.** Frontend can independently fetch and paginate each query.

### Negative

- **Breaking change.** `OVViewConfig.Fields` changes from `[]string` to
  `[]OVViewField`. Mitigated by backward-compatible UnmarshalJSON.
- **Migration.** Existing OV configs in the database will be transparently
  converted on read (no schema migration needed — JSONB is flexible).
- **Slightly more verbose.** Simple views need `{name: "X"}` instead of `"X"`.

### Risks

- Expression parsing complexity — mitigated by using simple string prefix matching
  (`query_name.field_name`) rather than full CEL parsing for reference validation.
- Performance of cycle detection — mitigated by limiting field count per OV
  (practical limit ~100 fields, Kahn's algorithm is O(V+E)).
