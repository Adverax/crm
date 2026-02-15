# ADR-0026: Custom Functions — Named Pure Computations

**Status:** Accepted
**Date:** 2026-02-15
**Participants:** @roman_myakotin

## Context

### Problem: Duplication of CEL Expressions

The platform uses CEL (Common Expression Language) as the unified expression language (ADR-0019) across multiple subsystems:

| Subsystem | Where CEL is Used | Example |
|-----------|-------------------|---------|
| Validation Rules (Phase 7b) | `expression` | `record.amount > 0 && record.amount < 1000000` |
| Default Expressions (Phase 7b) | `default_expr` | `record.tier == "gold" ? record.amount * 0.2 : 0` |
| Object View (ADR-0022) | `visibility_expr` | `record.status == "draft" && record.amount > 10000` |
| Procedure (ADR-0024) | `when`, `input.*` | `$.input.tier == "gold" ? $.input.amount * 0.2 : 0` |
| Scenario (ADR-0025) | `when`, `input.*` | `$.steps.check.tier == "gold"` |
| Automation Rules (ADR-0019) | `condition` | `new.status == "paid" && old.status != "paid"` |
| Dynamic Forms (Phase 9c) | field visibility | `record.type == "enterprise"` |

When the same computational logic is needed in multiple places, the administrator **duplicates the CEL expression**:

```
// Validation rule for Order:
record.tier == "gold" ? record.amount * 0.2 : record.tier == "silver" ? record.amount * 0.1 : 0

// Default expression for discount_amount:
record.tier == "gold" ? record.amount * 0.2 : record.tier == "silver" ? record.amount * 0.1 : 0

// Procedure input:
$.input.tier == "gold" ? $.input.amount * 0.2 : $.input.tier == "silver" ? $.input.amount * 0.1 : 0

// Object View visibility:
(record.tier == "gold" ? record.amount * 0.2 : record.tier == "silver" ? record.amount * 0.1 : 0) > 5000
```

One and the same expression — 4 copies. When a business rule changes (adding a "platinum" tier), all copies must be found and updated. This is:
- **Unreliable** — easy to miss one of the copies
- **Impractical** — complex expressions become unreadable when inlined
- **Violates DRY** — a fundamental development principle

### Procedure Does Not Solve the Problem

Procedure (ADR-0024) is a set of Commands with side effects (CRUD, email, HTTP). For pure computations, Procedure is overkill:

| Aspect | Function (needed) | Procedure (existing) |
|--------|-------------------|----------------------|
| Purpose | Compute a value | Execute actions |
| Side effects | None | Yes (CRUD, email, HTTP) |
| Invocation | Inline from any CEL | `flow.call` from Procedure |
| Return | Value (any) | ProcedureResult |
| Rollback | Not needed | Saga pattern |
| Where available | Everywhere CEL exists | Only as an action |

The function `fn.discount(tier, amount)` is called **inside** a CEL expression. A Procedure is called **instead of** a CEL expression. These are different levels of abstraction.

### Dual-stack: cel-go + cel-js

CEL already works on both sides (ADR-0019):
- **Backend**: cel-go — validation rules, defaults, procedure/scenario engine
- **Frontend**: cel-js — Object View `visibility_expr`, Dynamic Forms field visibility

Custom Functions must be available on **both sides**. Since Functions are pure expressions without side effects, they are portable between cel-go and cel-js without adaptation.

## Considered Options

### Option A — Copy-Paste CEL Expressions (status quo)

The administrator duplicates identical CEL expressions in all usage locations.

**Pros:**
- No new abstractions
- Each expression is self-contained (full logic visible in one place)

**Cons:**
- DRY violation: one change requires updating N locations
- Errors: easy to miss one of the copies
- Unreadability: complex inline expressions are hard to maintain
- Scaling: the more subsystems use CEL, the more duplication

### Option B — Custom Functions as Named CEL Expressions (chosen)

The administrator defines a named function with typed parameters. The function body is a CEL expression. The function is available from any CEL context through the `fn.*` namespace.

**Pros:**
- DRY: one definition — many usages
- Unified mechanism: Functions work everywhere CEL exists (backend + frontend)
- Decomposition: complex logic is broken into understandable named blocks
- Testability: a function can be tested in isolation
- Purity: no side effects — safe, predictable, cacheable
- Minimal implementation: extension of the existing CEL Environment, not a new engine

**Cons:**
- New abstraction: the administrator needs to understand the concept of a "function"
- Debugging: when an error occurs, both the calling code and the function body need to be checked
- Dependencies: deleting/modifying a function can break expressions that use it

### Option C — Macro Expansion (template substitutions)

Instead of runtime functions — templates that are expanded into the CEL expression at compilation time.

**Pros:**
- Transparency: the administrator sees the final "expanded" expression
- No runtime overhead: everything is expanded at compilation time

**Cons:**
- No parameters (or complex substitution): `${discount}` with argument replacement is essentially a homegrown template engine
- Errors in the expanded expression are hard to correlate with the original template
- No type checking at the time of macro definition
- Does not work on frontend (cel-js is unaware of macros)

### Option D — Computed Fields Instead of Functions

Computed fields (Formula Fields, Phase 10) could cover some use cases: `discount_amount = tier == "gold" ? amount * 0.2 : ...`.

**Pros:**
- Already planned (Phase 10)
- Bound to an object — a natural access point

**Cons:**
- Bound to a specific object — cannot be reused across objects
- Cannot pass arbitrary parameters (only fields of the current record)
- Do not work in Procedure/Scenario (different context: `$.input.*`, not `record.*`)
- Duplication: the same formula on different objects

## Decision

**Option B chosen: Custom Functions as named CEL expressions.**

### Function Definition

A Function is stored in `metadata.functions` as JSONB:

```json
{
  "name": "discount",
  "description": "Calculate discount by customer tier",
  "params": [
    { "name": "tier", "type": "string", "description": "Customer tier" },
    { "name": "amount", "type": "number", "description": "Order amount" }
  ],
  "return_type": "number",
  "body": "tier == \"gold\" ? amount * 0.2 : tier == \"silver\" ? amount * 0.1 : 0"
}
```

- **name** — unique identifier (snake_case), invoked as `fn.name()`
- **params** — typed parameters (string, number, boolean, list, map, any)
- **return_type** — return value type (for type checking)
- **body** — CEL expression; parameters are accessible as variables by name

### Invocation from CEL

All Functions are available through the `fn.*` namespace in any CEL context:

```
// Validation rule
fn.discount(record.tier, record.amount) > 5000

// Default expression
fn.discount(record.tier, record.amount)

// Procedure command input
"discount": "fn.discount($.input.tier, $.input.amount)"

// Object View visibility_expr
fn.discount(record.tier, record.amount) > 5000

// Scenario when
fn.discount($.steps.order.tier, $.steps.order.amount) > 10000

// Composition: function calls function
fn.total_with_tax(fn.discount(record.tier, record.amount), record.tax_rate)
```

The `fn.*` namespace separates user-defined functions from built-in ones (`size()`, `has()`, `matches()`), eliminating name conflicts.

### Dual-stack: Loading into cel-go and cel-js

```
metadata.functions (PostgreSQL JSONB)
        |
        +---> Backend startup / cache invalidation
        |    +-- cel-go: env.RegisterFunction("fn.discount", ...)
        |        -> Validation Rules, Defaults, Procedure, Scenario
        |
        +---> GET /api/v1/describe (Describe API)
             +-- response.functions: [{ name, params, body }]
                 +-- cel-js: env.registerFunction("fn.discount", ...)
                     -> visibility_expr, Dynamic Forms
```

**Backend**: Functions are loaded into the cel-go Environment at startup and upon cache invalidation (outbox pattern, ADR-0012).

**Frontend**: The Describe API returns function definitions. cel-js registers them as custom functions. `visibility_expr: "fn.discount(record.tier, record.amount) > 5000"` is evaluated **instantly in the browser** without a round-trip to the server.

### Constructor UI

In the Expression Builder (ADR-0024), Functions appear as a category:

1. **Function picker**: a "Custom Functions" section in the Expression Builder's function catalog
   - Each function with description, parameter types, usage example
   - Auto-completion: selecting `fn.discount` inserts the template `fn.discount(tier, amount)` with placeholders

2. **Function Constructor**: a dedicated admin page for creating/editing functions
   - Name + description
   - Parameters: name + type + description (drag-and-drop for ordering)
   - Body: Expression Builder (the same component) with parameters in the field picker
   - Live preview: test parameter values produce results in real time
   - Validation: type checking of the body upon saving

3. **Dependency view**: where the function is used (list of validation rules, defaults, procedures, Object Views)

### Limits

| Parameter | Limit | Rationale |
|-----------|-------|-----------|
| Body size | 4 KB | A function is a compact expression, not a program |
| Execution time | 100 ms | Called inline, must not block |
| Nesting | 3 levels | `fn.a()` -> `fn.b()` -> `fn.c()` -> stop |
| Parameters | 10 max | More than that is a Procedure |
| Recursion | Forbidden | Call stack tracking, `recursive_function_call` error |
| Number of functions | 200 | Per instance; prevents namespace bloat |

### Safety

| Threat | Protection |
|--------|------------|
| Infinite recursion | Forbidden: call stack tracking, static analysis on save |
| Circular dependencies | Dependency graph checked on save (`fn.a` -> `fn.b` -> `fn.a` = error) |
| Side effects | Impossible: CEL is a pure expression language, no I/O in grammar |
| Deletion with dependencies | Deletion blocked: dependency view shows usages; `DELETE` -> 409 Conflict |
| Rename | Cascading update: find all CEL expressions with `fn.old_name` -> replace |
| Resource exhaustion | Limits: 100ms timeout, 4KB body, 3 nesting levels |

### Storage

Table `metadata.functions`:

| Column | Type | Description |
|--------|------|-------------|
| id | UUID PK | Unique ID |
| name | VARCHAR UNIQUE | Function name (snake_case) |
| description | TEXT | Purpose description |
| params | JSONB | Parameter array `[{name, type, description}]` |
| return_type | VARCHAR | Return value type |
| body | TEXT | CEL expression |
| created_at | TIMESTAMPTZ | Creation time |
| updated_at | TIMESTAMPTZ | Update time |

### API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/admin/functions` | List functions |
| POST | `/api/v1/admin/functions` | Create function |
| GET | `/api/v1/admin/functions/:id` | Get function |
| PUT | `/api/v1/admin/functions/:id` | Update function |
| DELETE | `/api/v1/admin/functions/:id` | Delete (409 if dependencies exist) |
| POST | `/api/v1/admin/functions/:id/test` | Test: input -> result |
| GET | `/api/v1/admin/functions/:id/dependencies` | Where it is used |

### Relationship with Formula Fields (Phase 10)

Formula Fields and Custom Functions are different tools:

| Aspect | Formula Field | Custom Function |
|--------|--------------|-----------------|
| Binding | To a specific object and field | Global, no binding |
| Context | `record.*` (current record fields) | Arbitrary parameters |
| Result | Field value (stored/computed) | Value returned in CEL |
| Reuse | No (per object) | Yes (anywhere in CEL) |
| Where available | SOQL SELECT, record display | Any CEL expression |

A Formula Field **can call** a Custom Function: `fn.discount(tier, amount)` as part of a field formula. But a Formula Field is bound to an object, while a Function is not.

## Consequences

### Positive

- **DRY** — one definition, many usages; change in one place
- **Dual-stack** — the same function works on backend (cel-go) and frontend (cel-js) without adaptation
- **Minimal implementation** — extension of the existing CEL Environment; not a new engine, not a new runtime
- **Decomposition** — complex logic is broken into named, testable blocks
- **Safety** — pure expressions, no side effects, no I/O; protection against recursion and circular deps
- **Instant evaluation on frontend** — `visibility_expr` with `fn.*` is evaluated in the browser without a round-trip
- **Integration with Expression Builder** — functions appear in the catalog; auto-completion of parameters
- **Dependency tracking** — the platform knows where each function is used; protection against deletion

### Negative

- **New abstraction** — the administrator needs to understand the concept of a "function" (Constructor UI lowers the barrier)
- **Two levels of debugging** — the error may be in the calling expression or in the function body
- **Frontend/backend synchronization** — when a function is updated, the cel-js cache needs to be invalidated (Describe API refetch)
- **Cascading errors** — changing a parameter type can break calling expressions (protection: type checking on save)

## Related ADRs

- **ADR-0019** — Declarative business logic: CEL as the unified expression language. Functions extend the CEL Environment with user-defined computations
- **ADR-0022** — Object View: `visibility_expr` can call Functions for complex visibility logic
- **ADR-0024** — Procedure Engine: CEL expressions in `when`, `input.*` can call Functions; Expression Builder shows Functions in the catalog
- **ADR-0025** — Scenario Engine: CEL expressions in `when`, `input.*` can call Functions
