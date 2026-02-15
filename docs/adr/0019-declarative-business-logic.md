# ADR-0019: Declarative Business Logic for Objects

**Status:** Accepted
**Date:** 2026-02-14
**Participants:** @roman_myakotin

## Context

The platform is metadata-driven (ADR-0003, ADR-0007): objects are defined through metadata, stored in real PG tables `obj_{api_name}`, SOQL/DML provides read/write with security (OLS/FLS/RLS — ADR-0009..0012).

The current metadata model covers **storage structure**: field types, required fields, uniqueness, references (ADR-0004, ADR-0005). However, behavioral logic — business rules, computed fields, dynamic defaults — is missing:

| What exists (metadata) | What is missing |
|---|---|
| Field type (text, number, boolean) | Cross-field validation (`close_date > created_at`) |
| `is_required` flag | Conditional required (`feedback required when status=completed`) |
| `default_value` as a static string | Dynamic defaults (`owner_id = current_user.id`) |
| — | Computed display fields (`full_name = first + last`) |
| — | Unified rules for frontend and backend |

### Current DML Engine Limitations

1. **`DefaultValue`** — `*string` in `FieldConfig` JSONB. At the DDL level it is only used for booleans. DML Engine does not inject values — it only skips the required-check when `HasDefault=true`
2. **Pipeline** (parse -> validate -> compile -> execute) has no hooks: no pre-insert, post-update, no trigger system
3. **Validation** is limited: required-fields + type-compatibility. Cross-field validation is impossible

### Platform Needs

For a metadata-driven CRM it is not enough to describe only the storage schema. The platform needs a **declarative behavioral logic layer** that:

1. Defines business rules (validation) declaratively, without per-object code
2. Computes derived values (formula fields) based on expressions
3. Sets dynamic default values
4. Works identically on backend and frontend (unified expression language)
5. Allows different contexts (forms, API endpoints) to have different rule sets
6. Guarantees a minimum level of data integrity regardless of the calling context

### Industry Context

All major CRM platforms split behavioral logic into independent subsystems:

| Platform | Validation | Computed Fields | Automation | Queries |
|-----------|-----------|-----------------|------------|---------|
| **Salesforce** | Validation Rules | Formula Fields | Flow / Apex Triggers | SOQL |
| **Dynamics 365** | Business Rules | Calculated Fields | Power Automate | FetchXML |
| **HubSpot** | Property Rules | Calculated Properties | Workflows | — |
| **Zoho CRM** | Validation Rules | Formula Fields | Workflows | — |

None of them combine everything into a monolithic object. Each subsystem has its own lifecycle, storage, and extension model.

## Options Considered

### Option A — Monolithic Declarative Object

A single YAML/JSON document per object describing everything: queries, computed fields, validation, defaults, mutations, automation.

**Pros:**
- Single abstraction — all object logic in one place
- Maximum declarativeness

**Cons:**
- God object (7+ responsibilities): data loading, transformation, validation, persistence, automation
- Mixes per-object logic (validation — must always work) with per-view logic (queries — depend on UI context)
- Blocks Phase 7a: requires CEL engine, executor, storage, YAML parser, dependency resolver — months of work
- Does not match the industry pattern

### Option B — Decomposition into Subsystems with a Three-Level Cascade (chosen)

Split behavioral logic into independent subsystems (Validation Rules, Default Expressions, Formula Fields, Object View, Automation Rules). Connect levels (Metadata -> Object View -> Layout) with a cascading inheritance model.

**Pros:**
- Each subsystem is independently useful and testable
- Incremental implementation — does not block the current phase
- Reusability (validation rules work with any write method: API, import, integration)
- Cascade with inheritance — DRY + flexibility
- Matches the industry pattern (Salesforce, Dynamics)

**Cons:**
- No single document for all object logic
- More separate ADRs (each subsystem = a separate decision)
- Composition layer (Object View) is deferred

### Option C — Defer Entirely

Build Phase 7a without abstractions, hardcode logic per-object.

**Pros:**
- Fast delivery of Phase 7a

**Cons:**
- Technical debt as the number of objects grows
- Validation duplication on frontend/backend
- Refactoring later will be more expensive
- Contradicts the platform's metadata-driven architecture

### Option D — Minimal Metadata Extensions

Extend `FieldConfig` for static defaults and simple validation without an expression engine.

**Pros:**
- Fast delivery, minimal changes

**Cons:**
- Static defaults are insufficient for dynamic values (`owner_id = current_user.id`)
- Cross-field validation is impossible without an expression engine
- Logic duplication on the frontend

## Decision

**Option B chosen: Decomposition into independent subsystems with a three-level cascade.**

### Three-Level Cascade

Rules and settings are defined at three levels. Each subsequent level **inherits** rules from the previous one:

```
Metadata (base)
   ↓ inherits
Object View (business context)
   ↓ inherits
Layout (presentation)
```

#### Cascade Semantics by Type

| Aspect | Metadata -> Object View | Object View -> Layout | Mechanism |
|--------|------------------------|----------------------|----------|
| **Validation** | Additive (AND) | Additive (AND) | New rules can only be added |
| **Defaults** | Replace | Replace | Last level wins |
| **Formula Fields** | Inherit (read-only) | Inherit (read-only) | Cannot be overridden |
| **Field visibility** | N/A | Override | Layout hides/shows |

#### Validation: Additive Model (tightening only)

Validation rules at each cascade level can only be **added**, never removed or replaced. The effective set is a conjunction (AND) of all rules:

```
effective_validation = metadata_rules AND object_view_rules AND layout_rules
```

This mathematically guarantees tightening: adding any new condition via AND narrows the set of valid values.

**Why not allow loosening?** Programmatic verification that one CEL expression is "stricter" than another reduces to a theorem proving problem: for all input: A(input)=true -> B(input)=true. For arbitrary expressions this is undecidable. Even for a restricted subset an SMT solver would be required — overkill for a CRM. The additive model eliminates the problem: no verification is needed, AND guarantees tightening automatically.

**Cascade example:**

```yaml
# Metadata (universal invariant):
- expr: 'discount <= 50'          # data integrity

# Object View "partner_portal" (adds business rule):
- expr: 'discount <= 20'          # business context

# Layout "mobile_form" (adds UI rule):
- expr: 'has(discount)'           # field required on this form

# Effective: discount <= 50 AND discount <= 20 AND has(discount)
# = discount is required AND no more than 20%
```

**Consequence for rule design:** if validation can differ across contexts, it belongs in **Object View**, not metadata. In metadata — only universal invariants whose violation = data corruption.

| Where to define | Criterion | Examples |
|---|---|---|
| **Metadata** | Universal invariant, violation = data corruption | `amount >= 0`, FK integrity, type safety |
| **Object View** | Business context, may differ between views | Phone format, conditional required, `discount <= N` |
| **Layout** | UI-specific tightening | Field required on a specific form |

#### Defaults: Replace (last level wins)

A default is "what value to substitute if not provided". Replacement is safe: the final validation will still check the value's correctness.

```yaml
# Metadata:       status default = "new"
# Object View:    status default = "draft"      ← replaces
# Layout:         (does not override)
# Effective:      "draft"
```

### Subsystems

Behavioral logic is split into **6 independent subsystems**:

```
┌─────────────────────────────────────────────────────────────────┐
│  GLOBAL (cross-cutting)                                          │
│  Available at all cascade levels, on backend and frontend        │
│                                                                  │
│  ┌───────────────────────────────────────────────────────────┐   │
│  │  Custom Functions (ADR-0026)                               │   │
│  │  Named pure CEL expressions: fn.discount(tier, amt)       │   │
│  │  Dual-stack: cel-go + cel-js. No side effects.            │   │
│  │  Callable from any CEL context below.                     │   │
│  └───────────────────────────────────────────────────────────┘   │
└──────────────────────────────┬──────────────────────────────────┘
                               │ available as fn.*
                               ▼
┌─────────────────────────────────────────────────────────┐
│              PER-OBJECT (metadata level)                 │
│   Base rules, inherited by all Object Views/Layouts     │
│                                                         │
│  ┌───────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  Validation   │  │   Default    │  │   Formula    │  │
│  │  Rules        │  │   Expressions│  │   Fields     │  │
│  │               │  │              │  │              │  │
│  │  CEL expr     │  │  CEL expr    │  │  CEL expr    │  │
│  │  per-object   │  │  per-field   │  │  per-field   │  │
│  │  in metadata  │  │  in metadata │  │  in metadata │  │
│  └───────────────┘  └──────────────┘  └──────────────┘  │
└──────────────────────────┬──────────────────────────────┘
                           │ inherits (additive validation,
                           │            replace defaults)
                           ▼
┌─────────────────────────────────────────────────────────┐
│           PER-VIEW (Object View level)                   │
│   Business context: specific UI screen or API endpoint  │
│                                                         │
│  ┌───────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  + Validation │  │  + Default   │  │  Queries     │  │
│  │  (additive)   │  │  (replace)   │  │  (SOQL)      │  │
│  └───────────────┘  └──────────────┘  └──────────────┘  │
│  ┌───────────────┐  ┌──────────────┐                    │
│  │  Virtual      │  │  Mutations   │                    │
│  │  Fields (CEL) │  │  (DML)       │                    │
│  └───────────────┘  └──────────────┘                    │
└──────────────────────────┬──────────────────────────────┘
                           │ inherits (additive validation,
                           │            replace defaults)
                           ▼
┌─────────────────────────────────────────────────────────┐
│           PER-LAYOUT (Layout level)                      │
│   Presentation: visual arrangement and UI rules         │
│                                                         │
│  ┌───────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  + Validation │  │  + Default   │  │  Field       │  │
│  │  (additive)   │  │  (replace)   │  │  visibility  │  │
│  └───────────────┘  └──────────────┘  │  & ordering  │  │
│                                       └──────────────┘  │
└─────────────────────────────────────────────────────────┘
```

#### 1. Validation Rules (per-object, metadata level)

CEL expressions that validate data before writes. Stored in the metadata schema, applied by the DML Engine on every operation.

```
metadata.validation_rules (object_id, expr, message, code, severity, when_expr, sort_order)
```

- Integration: DML Engine (validate step)
- Evaluation: CEL runtime (cel-go backend, cel-js frontend)
- Variables: `record` (current values), `old` (previous values on update), `user`, `now`

#### 2. Default Expressions (per-field, metadata level)

CEL expressions for computing default values. Extension of the existing `FieldConfig.DefaultValue` to dynamic expressions.

```
field_definitions.config.default_expr — CEL expression (nullable)
field_definitions.config.default_on   — "create" | "update" | "create,update"
```

- Integration: DML Engine (pre-validate step, missing field injection)
- Variables: `record`, `user`, `now`
- Static defaults (`"draft"`, `true`) remain as `default_value`

#### 3. Formula Fields (per-field, metadata level, read-only)

CEL expressions for computed fields that are not stored in the DB. Computed at read time. Analogous to Salesforce Formula Fields.

```
field_definitions.config.formula_expr — CEL expression (nullable)
field_definitions.field_type = "formula"
field_definitions.field_subtype = "string" | "number" | "boolean" | "datetime"
```

- Integration: SOQL Executor (post-fetch computation)
- Variables: `record` (all fields of the current record)
- Frontend: computed locally using the same CEL

#### 4. Object View (per-view, future)

Composition layer for a specific UI screen or API endpoint. Inherits validation and defaults from metadata (additive / replace). Contains its own components:

- Queries — named SOQL queries with cross-references
- Virtual Fields — view-specific computed fields (CEL)
- Mutations — DML operation orchestration (foreach, sync)
- Validation overrides — additional rules (additive only)
- Default overrides — alternative defaults (replace)

**Architectural role: bounded context adapter (DDD).** The same object (e.g. `Order`) serves different business roles: sales manager, warehouse worker, manager. Each role operates in its own bounded context — with its own set of fields, actions, related lists, and sidebar. Object View, bound to a profile (`profile_id`), adapts unified data to the context of a specific role without code duplication. OLS/FLS/RLS controls *data access*, Object View controls *data presentation*. At the same time, Object View only narrows visibility (FLS intersection), but does not expand access.

Details: [ADR-0022](0022-object-view-bounded-context.md) — config structure, resolution logic, sidebar/dashboard per profile, role-based UI examples.

#### 5. Automation Rules (per-object, future)

Reactive logic: "when X happens, do Y". Analogous to Salesforce Flow / Process Builder.

- Trigger condition — CEL expression (`new.Stage == "Closed Won" && old.Stage != "Closed Won"`)
- Action — reference to a Procedure (synchronous, ADR-0024) or Scenario (asynchronous, ADR-0025)
- Terminology: ADR-0023

Separate ADR if needed.

#### 6. Custom Functions (global, cross-cutting, ADR-0026)

Named pure CEL expressions with typed parameters. Eliminate CEL logic duplication across subsystems.

```
metadata.functions (name, params JSONB, return_type, body TEXT)
```

- **Pure**: no side effects — only computations, no CRUD/IO
- **Global**: not bound to an object, callable from any CEL context via `fn.*` namespace
- **Dual-stack**: loaded into cel-go (backend) and cel-js (frontend) — identical behavior
- **Reusable**: one definition -> use in validation rules, defaults, formulas, visibility, procedure/scenario input
- **Composable**: `fn.total(fn.discount(tier, amount), tax_rate)` — functions can call each other (max 3 levels)

Difference from Formula Fields: a Formula Field is bound to an object and a field; a Function is global and accepts arbitrary parameters.

Details: [ADR-0026](0026-custom-functions.md).

### Execution Context: DML Engine

When performing an operation, the DML Engine receives an **effective ruleset** — the result of cascading rule merging depending on the calling context:

```
Call without Object View (raw DML, import, integration):
  effective_validation = metadata_rules
  effective_defaults   = metadata_defaults

Call through Object View (UI form, specific API endpoint):
  effective_validation = metadata_rules AND object_view_rules
  effective_defaults   = merge(metadata_defaults, object_view_defaults)

Call through Layout:
  effective_validation = metadata_rules AND object_view_rules AND layout_rules
  effective_defaults   = merge(metadata_defaults, object_view_defaults, layout_defaults)
```

"Bare" DML (without Object View) always applies metadata-level rules — the minimum guaranteed level of data integrity protection.

### CEL as the Expression Language

CEL — Common Expression Language (Google) — is used for all expressions (validation, defaults, formulas, virtual fields).

| Criterion | CEL | Alternatives |
|----------|-----|-------------|
| Go runtime | `cel-go` (official) | Expr, Govaluate |
| JS runtime | `cel-js` | Only CEL has both |
| Security | Sandboxed, no side effects | Expr — similar, Govaluate — no |
| Typing | Static type checking | Expr — runtime only |
| Standard | Google, K8s, Firebase | None |
| Syntax | C-like, intuitive | Expr — similar |

CEL integration is introduced with the implementation of Validation Rules (Phase 7b).

### Roadmap

```
Phase 7a                  Phase 7b                Phase 10                Phase 9a/9b
──────────────────    ──────────────────    ──────────────────────    ──────────────────
Generic CRUD          CEL engine (cel-go)   Custom Functions          Object Views
+ Metadata-driven UI  + Validation Rules    + fn.* namespace          + Query composition
+ Static defaults     + Dynamic defaults    + Function Constructor    + Actions
+ System fields       + DML pipeline ext.   + Expression Builder      + Automation Rules
                      + Frontend CEL eval   + Formula Fields          + Layout cascade
                                            + SOQL integration
```

**Phase 7a.** Generic metadata-driven REST endpoints: a single set of handlers serves all objects via SOQL (reads) and DML (writes). Frontend renders forms from metadata. Validation: required + type constraints (already in DML Engine). Static defaults: injection of `FieldConfig.default_value` for missing fields. System fields (`owner_id`, `created_by_id`, `created_at`, `updated_at`). No CEL.

**Phase 7b — CEL + Validation Rules + Dynamic Defaults.** Integration of `cel-go`. Table `metadata.validation_rules`. Extension of `FieldConfig` for `default_expr`. Integration into the DML pipeline (Stage 3 dynamic + Stage 4b). Frontend library for CEL evaluation (`cel-js`).

**Phase 10 — Custom Functions + Formula Fields.** Custom Functions (ADR-0026): global named CEL expressions with `fn.*` namespace, dual-stack (cel-go + cel-js), Function Constructor + Expression Builder integration. Formula Fields: `field_type = "formula"`, CEL expression in config, SOQL executor computes after fetch, frontend computes locally. Formula Fields can call Custom Functions.

**Phase 9a/9b — Object Views + Automation + Layout Cascade.** Full composition with the three-level cascade. Cascading merge (metadata + Object View + Layout). ADR-0022 (Object View), ADR-0023 (Action terminology), ADR-0024 (Procedure Engine), ADR-0025 (Scenario Engine).

## Consequences

### Positive

- Phase 7a is not blocked — generic CRUD endpoints are built on existing infrastructure (SOQL + DML + MetadataCache)
- Each subsystem is independently useful and testable
- Three-level cascade (Metadata -> Object View -> Layout) provides DRY + flexibility
- Additive validation model guarantees tightening without programmatic expression verification
- Validation Rules work with any write method (API, import, integration)
- "Bare" DML without Object View is still protected by metadata-level rules
- Incremental value delivery
- Matches the industry pattern (Salesforce, Dynamics, Zoho)

### Negative

- No single document for all object logic (deliberate trade-off in favor of SoC)
- More ADRs (each subsystem = a separate architectural decision)
- Composition layer (Object View + Layout cascade) is deferred to Phase N+2
- Additive validation model does not allow loosening a rule — if a rule can differ across contexts, it should be placed in Object View from the start, not in metadata

### Related ADRs

- ADR-0003 — Object metadata structure (extended with validation rules and default expressions)
- ADR-0004 — Field type/subtype hierarchy (extended with formula type)
- ADR-0007 — Table-per-object storage (generic CRUD works with `obj_{api_name}` tables)
- ADR-0009..0012 — Security layers (validation rules supplement but do not replace OLS/FLS/RLS)
- ADR-0018 — App Templates (create schema; subsystems from this ADR define behavior)
- ADR-0020 — DML Pipeline Extension (typed stages — subsystem integration points in DML Engine)
- ADR-0022 — Object View as bounded context adapter (details for subsystem 4: role-based UI, config schema, resolution logic)
- ADR-0023 — Action terminology: unified hierarchy Action -> Command -> Procedure -> Scenario + Function (orthogonal)
- ADR-0024 — Procedure Engine: JSON DSL + Constructor UI for synchronous business logic (Mutations -> Action type: procedure)
- ADR-0025 — Scenario Engine: JSON DSL + Constructor UI for asynchronous long-running processes
- ADR-0026 — Custom Functions (details for subsystem 6: fn.* namespace, dual-stack, constraints, Constructor UI)
- ADR-0027 — Layout + Form (third cascade level: Layout defines presentation on top of Object View, Form = computed merge for frontend)
