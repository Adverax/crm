# ADR-0020: DML Pipeline Extension (Typed Stages)

**Status:** Accepted
**Date:** 2026-02-14
**Participants:** @roman_myakotin

## Context

The DML Engine (Phase 4) implements a 4-stage pipeline:

```
parse → validate → compile → execute
```

Where `validate` performs: field existence check, required-fields check, type-compatibility, FLS (writable fields). This is sufficient for structural validation but insufficient for the behavioral logic defined in ADR-0019:

| Subsystem (ADR-0019) | What is required from DML | Current status |
|---|---|---|
| Default Expressions | Inject missing fields before validation | Missing. `DefaultValue` is only used for boolean DDL; DML does not inject |
| Validation Rules | CEL checks after defaults | Missing. Only required + type |
| Computed Fields (stored) | Recompute derived values before compile | Missing |
| Automation Rules | Reactive logic after execute | Missing |

The pipeline needs to be extended to integrate these subsystems.

### Requirements for Extension

1. **Clear execution order** — each stage has a defined place in the pipeline, no ambiguity
2. **Typed interfaces** — each stage = interface with a specific signature, not an arbitrary callback
3. **Incrementality** — stages are added according to the ADR-0019 roadmap, not all at once
4. **Effective ruleset** — pipeline accepts the calling context (metadata / Object View / Layout) for cascading rule merging (ADR-0019)
5. **Testability** — each stage is tested in isolation via mock

## Options Considered

### Option A — Generic hooks (middleware pattern)

Arbitrary callbacks registered on events (before insert, after update):

```go
engine.Before("insert", func(ctx context.Context, record Record) error { ... })
engine.After("update", func(ctx context.Context, old, new Record) error { ... })
```

**Pros:**
- Maximum flexibility
- Familiar pattern (Express middleware, Gin handlers)

**Cons:**
- No execution order guarantees between hooks
- Arbitrary code — hard to test, debug
- "Which hook broke the record?" — a classic Salesforce problem (35+ steps in the order of execution)
- Violates the platform's declarative principle
- Hooks can conflict with each other

### Option B — Typed pipeline stages (chosen)

A fixed set of stages with typed interfaces. Each stage is responsible for one task and has a defined place in the pipeline.

**Pros:**
- Clear, predictable execution order
- Each stage = typed interface -> compile-time safety, easy to test
- Declarative subsystems (CEL) plug in through these interfaces
- No "hook A conflicts with hook B" problem
- Simple debugging: each stage can be logged separately

**Cons:**
- Less flexible than arbitrary hooks
- For non-standard scenarios, a custom handler at the Automation Rules level is needed, not inside the pipeline

### Option C — Salesforce-style order of execution

A fixed order of 10+ steps with a clear description of each:

**Pros:**
- Proven in production (Salesforce)

**Cons:**
- Salesforce order of execution — 35+ steps, a notoriously known problem
- Triggers can cause recursion (trigger -> DML -> trigger)
- Excessive complexity for our platform

## Decision

**Option B chosen: Typed pipeline stages.**

### Extended Pipeline

```
┌──────────────────────────────────────────────────────────┐
│                     DML Pipeline                          │
│                                                          │
│  1. PARSE            ← AST from DML expression           │
│     Existing parser (Participle)                         │
│                                                          │
│  2. RESOLVE          ← load metadata + context           │
│     ├─ ObjectMeta + FieldMeta (MetadataProvider)         │
│     └─ Effective ruleset (cascade per ADR-0019)          │
│                                                          │
│  3. DEFAULTS         ← inject missing fields             │
│     ├─ static: FieldConfig.default_value                 │
│     └─ dynamic: FieldConfig.default_expr (CEL)           │
│     Only for INSERT. Only for fields absent              │
│     from the statement. Cascade: metadata → OV → Layout. │
│                                                          │
│  4. VALIDATE                                             │
│     a) Metadata constraints: required, type, unique      │
│     b) Validation Rules: CEL expressions (AND)           │
│     c) FLS: writable fields check                        │
│     Cascade: metadata_rules AND ov_rules AND layout_rules │
│                                                          │
│  5. COMPUTE          ← recompute stored computed fields  │
│     CEL expressions from FieldConfig.formula_expr        │
│     Only for fields with field_type="formula" + stored   │
│     Adds computed values to the statement                │
│                                                          │
│  6. COMPILE          ← SQL generation                    │
│     Existing compiler (parameterized SQL)                │
│                                                          │
│  7. EXECUTE          ← pgx                               │
│     Existing executor + RLS injection                    │
│                                                          │
│  8. POST-EXECUTE     ← reactive logic (future)           │
│     Automation Rules: trigger conditions (CEL)           │
│     → Handler (sync) / Scenario (async)                  │
│     Runs after successful execute, before commit         │
│     or after commit (depending on the action type)       │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

### Stages and Operations

Not all stages apply to all DML operations:

| Stage | INSERT | UPDATE | DELETE | UPSERT |
|---|---|---|---|---|
| 1. Parse | Yes | Yes | Yes | Yes |
| 2. Resolve | Yes | Yes | Yes | Yes |
| 3. Defaults | Yes | Conditional (default_on=update) | No | Yes (insert part) |
| 4a. Metadata validate | Yes | Yes | No | Yes |
| 4b. Validation Rules | Yes | Yes | Conditional | Yes |
| 4c. FLS | Yes | Yes | Yes | Yes |
| 5. Compute | Yes | Yes | No | Yes |
| 6. Compile | Yes | Yes | Yes | Yes |
| 7. Execute | Yes | Yes | Yes | Yes |
| 8. Post-execute | Yes | Yes | Yes | Yes |

### Typed Interfaces

Each new stage is an interface, plugged in via the Option pattern (like the existing `WithMetadata`, `WithExecutor`):

```go
// Stage 3: Default injection
type DefaultResolver interface {
    ResolveDefaults(ctx context.Context, object string, operation Operation, fields map[string]Value) (map[string]Value, error)
}

// Stage 4b: Validation Rules
type RuleValidator interface {
    ValidateRules(ctx context.Context, object string, operation Operation, record, old map[string]Value) []ValidationError
}

// Stage 5: Computed fields
type ComputeEngine interface {
    ComputeFields(ctx context.Context, object string, record map[string]Value) (map[string]Value, error)
}

// Stage 8: Post-execute reactions
type PostExecutor interface {
    AfterExecute(ctx context.Context, object string, operation Operation, result *Result) error
}
```

Plugged in via Options:

```go
engine := dml.NewEngine(
    dml.WithMetadata(metadataAdapter),
    dml.WithWriteAccessController(flsEnforcer),
    dml.WithExecutor(rlsExecutor),
    // New stages:
    dml.WithDefaultResolver(celDefaultResolver),
    dml.WithRuleValidator(celRuleValidator),
    dml.WithComputeEngine(celComputeEngine),
    dml.WithPostExecutor(automationDispatcher),
)
```

Each stage is optional. If the interface is not provided, the stage is skipped. This enables incremental addition per the ADR-0019 roadmap.

### Execution Context

The DML Engine accepts a calling context that determines the effective ruleset (cascade from ADR-0019):

```go
type ExecutionContext struct {
    ObjectViewID *uuid.UUID  // nil = raw DML (metadata rules only)
    LayoutID     *uuid.UUID  // nil = no layout-level rules
}
```

The Resolve stage uses the context for cascading merge:

```
// Validation: additive (AND)
effective_rules = metadata_rules
if ctx.ObjectViewID != nil {
    effective_rules = append(effective_rules, ov_rules...)
}
if ctx.LayoutID != nil {
    effective_rules = append(effective_rules, layout_rules...)
}

// Defaults: replace (last wins)
effective_defaults = metadata_defaults
if ctx.ObjectViewID != nil {
    effective_defaults = merge(effective_defaults, ov_defaults)
}
if ctx.LayoutID != nil {
    effective_defaults = merge(effective_defaults, layout_defaults)
}
```

When called without context (raw DML, import, integration) only metadata-level rules apply — the minimum guaranteed level of protection.

### Validation Rules: Variables in the CEL Environment

| Variable | Type | Availability | Description |
|---|---|---|---|
| `record` | map | INSERT, UPDATE, UPSERT | Current field values (after defaults) |
| `old` | map | UPDATE | Previous values (before change) |
| `user` | map | Always | Current user (`id`, `profile_id`, `role_id`) |
| `now` | timestamp | Always | Current UTC time |

For INSERT: `old` = nil. Validation rules with `old` in the expression are automatically skipped on INSERT.

### Default Expressions: Application Order

1. Determine fields absent from the DML statement
2. For each absent field, check for a default:
   - First `default_value` (static) — lower priority
   - Then `default_expr` (CEL) — overrides static
   - Cascade: Layout > Object View > Metadata
3. If `default_on` does not match the current operation — skip
4. CEL expression is evaluated with variables `record`, `user`, `now`
5. Result is added to `record` before the validate step

### Errors

Each stage returns typed errors:

| Stage | Error code | HTTP | Description |
|---|---|---|---|
| Defaults | `default_eval_error` | 500 | Error evaluating a default CEL expression |
| Validation (metadata) | `missing_required_field` | 400 | Required field is missing |
| Validation (metadata) | `type_mismatch` | 400 | Incompatible value type |
| Validation (rules) | `validation_rule_failed` | 400 | CEL validation failed (code from rule) |
| Validation (rules) | `rule_eval_error` | 500 | Error evaluating a rule CEL expression |
| Compute | `compute_eval_error` | 500 | Error computing a computed field |
| Post-execute | `automation_error` | 500 | Error in an automation rule |

Validation rules with `severity: warning` do NOT block execution — they are collected in `Result.Warnings`.

### Non-Recursion

Automation Rules (post-execute) can perform DML operations on other objects. To prevent recursion:

- Automation Rules CANNOT modify the object that triggered them
- Maximum DML call nesting depth from automation: 2 (analogous to Salesforce trigger depth limit)
- DML calls from automation execute with `ExecutionContext = nil` (metadata rules only)

## Roadmap

Stages are added incrementally, in accordance with ADR-0019:

| Phase | Stages added | Dependencies |
|---|---|---|
| **7a** | 3. Defaults (static `default_value` only) | — |
| **7b** | 3. Defaults (dynamic `default_expr`) + 4b. Validation Rules | CEL engine (cel-go) |
| **N+1** | 5. Compute | CEL engine (already available after 7b) |
| **N+2** | 2. Resolve (cascade) + 8. Post-execute | Object View storage, Automation Rules |

**Phase 7a** — static defaults: injection of `FieldConfig.default_value` for missing fields + system fields (`owner_id`, `created_by_id`, `created_at`, `updated_at`). No CEL. Pipeline is extended with Stage 3 in a minimal variant.

**Phase 7b** — CEL engine: `cel-go` integration, `default_expr`, table `metadata.validation_rules`, Stage 4b. Validation Rules and dynamic defaults use a shared CEL runtime.

## Consequences

### Positive

- Predictable execution order — no "which hook broke the record?"
- Each stage is tested in isolation via interface mock
- Stage optionality via the Option pattern — incremental addition
- Non-recursion of automation is guaranteed (depth limit)
- Declarative approach preserved (CEL, not arbitrary Go code in the pipeline)
- Errors are typed and tied to a specific stage

### Negative

- Less flexible than generic hooks — edge cases are solved via Automation Rules (Go handler), not via pipeline injection
- Fixed stage order — cannot insert a custom stage "between validate and compute"
- CEL dependency for stages 3, 4b, 5 (but this is a decision from ADR-0019)

### Related ADRs

- ADR-0019 — Declarative business logic (defines subsystems and cascade; this ADR defines integration points in DML)
- ADR-0004 — Field type/subtype (extended with formula type for stage 5)
- ADR-0009 — Security architecture (FLS remains in stage 4c, not mixed with validation rules)
