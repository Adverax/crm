# ADR-0023: Executable Logic Terminology — Action, Command, Procedure, Scenario, Function

**Status:** Accepted
**Date:** 2026-02-15
**Participants:** @roman_myakotin

## Context

### Problem: terminological conflict

The platform is developing several subsystems that describe reactions to user actions:

1. **Object View** (ADR-0022) — buttons on the record card ("Ship", "Send Proposal")
2. **Procedure Engine** (ADR-0024) — declarative YAML DSL for business logic with CEL expressions
3. **Scenario Engine** (ADR-0025) — orchestration of long-lived business processes with durability
4. **DML Pipeline** (ADR-0020) — data processing stages during writes
5. **Automation Rules** (ADR-0019) — reactive logic "when X -> do Y"

Terminological conflicts arose between them:

| Term | Where used | Meaning 1 | Meaning 2 |
|------|-----------|-----------|-----------|
| **Action** | handler.md | Unit of work (`record.create`, `email send`) | — |
| **Action** | ADR-0022 | UI button on the record card | — |
| **Handler** | handler.md | Named set of actions (YAML DSL) | — |
| **Handler** | Go code | HTTP handler (Gin) | — |
| **Mutation** | ADR-0019 | DML orchestration in Object View | — |
| **Step** | scenario.md | Atomic step in a scenario | — |

Three key conflicts:

- **Action** is overloaded: UI button vs atomic operation
- **Handler** is overloaded: declarative YAML block vs HTTP handler
- **Mutation** is isolated in ADR-0019, not connected to the overall model

Without unified terminology:
- Developers get confused during discussions ("which action do you mean?")
- Documentation contradicts itself
- Architectural decisions do not align across ADRs

### Requirements for Terminology

1. **Consistency** — one term = one meaning across all documents
2. **Hierarchy** — terms form a clear hierarchy from abstract to concrete
3. **Intuitiveness** — a new developer understands the term without consulting a glossary
4. **Compatibility** — does not conflict with established terms in Go (handler), SQL (procedure), HTTP (request)
5. **Extensibility** — new types of executable logic integrate without breaking terminology

### Industry Context

| Platform | UI button | Atomic operation | Set of operations | Long-running process |
|----------|-----------|-----------------|-------------------|----------------------|
| **Salesforce** | Quick Action | — | Apex Trigger / Flow Element | Flow / Process |
| **Dynamics 365** | Command | Action Step | Action | Business Process Flow |
| **ServiceNow** | UI Action | Activity | Workflow Activity | Workflow |
| **Temporal** | — | Activity | — | Workflow |
| **n8n** | — | Node | — | Workflow |

The industry has no single standard, but a pattern is evident: **hierarchy from simple to complex** with different levels of durability.

## Considered Options

### Option A — Minimal Refactoring (rename only conflicts)

Rename only the conflicting terms, leaving the rest as-is.

- handler.md "Action" -> "Operation"
- handler.md "Handler" -> "Logic Block"

**Pros:**
- Minimal changes to documentation
- Quick

**Cons:**
- Does not create a unified hierarchy
- "Operation" and "Logic Block" — generic, carry no semantic weight
- Mutation from ADR-0019 remains isolated

### Option B — Unified Hierarchy with Action as Umbrella Term (chosen)

Action = umbrella term. All types of executable logic are subtypes of Action. Conflicting terms get new names.

**Pros:**
- Unified hierarchy: Action -> Command -> Procedure -> Scenario
- Each term is unambiguous
- Mutation is absorbed (action type: procedure)
- Extensible: new action types are added without breaking anything

**Cons:**
- Requires refactoring docs/private/handler.md and docs/private/scenario.md
- "Procedure" may be associated with SQL stored procedure (but context differentiates)
- ADR-0022 config will need updating upon implementation

### Option C — Salesforce Terminology

Adopt Salesforce terminology: Quick Action, Flow, Apex Trigger.

**Pros:**
- Familiar to Salesforce users
- Industry-proven

**Cons:**
- Ties to another brand
- "Apex Trigger" is inapplicable (we have no Apex)
- "Flow" conflicts with Go control flow
- Not all SF concepts map 1:1 to our architecture

### Option D — Workflow-centric Terminology

Everything is built around "Workflow": Workflow Action, Workflow Step, Workflow.

**Pros:**
- Single root
- Clear

**Cons:**
- "Workflow" is overloaded in the industry (GitHub Actions, n8n, Temporal)
- Does not distinguish synchronous (procedure) from asynchronous (scenario)
- Excessively long compound terms

## Decision

**Option B chosen: Unified hierarchy with Action as umbrella term.**

### Term Hierarchy

The platform separates **executable logic** (actions with side effects) and **computational logic** (pure computations):

```
Executable logic (Action hierarchy)          Computational logic (CEL ecosystem)
-------------------------------------        -------------------------------------
Action (umbrella term)                       CEL Expression (inline, one-off)
|                                            |
+-- type: navigate    -> URL transition      Function (named, reusable)
+-- type: field_update -> atomic DML           fn.discount(tier, amount)
+-- type: procedure   -> Procedure (sync)      fn.is_high_value(amount)
+-- type: scenario    -> Scenario (async)      Callable from any CEL context
```

**Function** is orthogonal to the Action hierarchy. Functions do not perform actions — they compute values. Functions are **called within** CEL expressions, which are used at all levels of both hierarchies.

### Glossary

| Term | Definition | Analog (Salesforce) | Durability level |
|------|-----------|--------------------|--------------------|
| **Action** | System reaction to a trigger. Umbrella term, defined by type (`navigate`, `field_update`, `procedure`, `scenario`). Can be invoked from Object View (button), Automation Rule (trigger), API (endpoint), or another Action | Quick Action / Button | — |
| **Command** | Atomic operation inside a Procedure: `record.create`, `notification.email`, `POST url`, `transform`, `validate`. Executed synchronously. Has no own state | Flow Element / Action Step | None (in-memory) |
| **Procedure** | Named set of Commands, described declaratively (JSON + CEL). Assembled through Constructor UI or edited as JSON. Stored as JSONB. Executed synchronously within a single request. Supports conditional logic (`when`, `if/else`, `match`), rollback (Saga pattern), calling other Procedures (`call`). Analog of a stored procedure, but safe (sandbox, limits) | Invocable Action / Autolaunched Flow | None (transaction) |
| **Scenario** | Long-lived business process coordinating a sequence of Steps with durability guarantees (state survives restarts), consistency (rollback on errors), and observability (complete history). Executed asynchronously. Supports Signals (waiting for external events), Timers (deferred actions), Checkpoints | Screen Flow / Record-Triggered Flow | Yes (PostgreSQL) |
| **Step** | Unit of work inside a Scenario. Invokes a Procedure, inline Command, or built-in operation (`wait signal`, `wait timer`). Has input/output mapping, retry policy, rollback | Flow Step | Yes (persisted) |
| **Function** | Named pure CEL expression with typed parameters. Called via the `fn.*` namespace from any CEL context (validation rules, defaults, visibility, procedure input, scenario when). No side effects. Dual-stack: cel-go (backend) + cel-js (frontend). Not an Action — it is a computational unit, orthogonal to the Action hierarchy | Custom Formula Function | — |

### Mapping of Old Terms

| Old term | Document | New term | Rationale |
|----------|----------|----------|-----------|
| Action (unit of work) | handler.md | **Command** | Imperative name for an atomic operation |
| Action (UI button) | ADR-0022 | **Action** | Stays — this is the umbrella term |
| Handler (YAML DSL block) | handler.md | **Procedure** | Named set of commands, analog of stored procedure |
| Handler (HTTP) | Go code | **Handler** | Stays — this is a Go/HTTP term, does not conflict in context |
| Mutation (DML orchestration) | ADR-0019 | **Action type: procedure** | Absorbed — mutation = procedure with DML commands |
| Action Type (record, notification) | handler.md | **Command Type** | Category of command: `record.*`, `notification.*`, `integration.*` |

### Relationship Between Levels

```
+-------------------------------------------------------------+
|                        Object View                               |
|   actions: [                                                     |
|     { key: "ship", type: "field_update", ... }                  |
|     { key: "send", type: "procedure", procedure: "send_prop" }  |
|     { key: "fulfill", type: "scenario", scenario: "order_ful" } |
|   ]                                                              |
+----------+------------------+------------------+----------------+
           |                  |                  |
           v                  v                  v
     +-----------+    +--------------+    +--------------+
     |   DML     |    |  Procedure   |    |  Scenario    |
     |  Engine   |    |  Engine      |    |  Engine      |
     |           |    |              |    |              |
     |  UPDATE   |    |  commands:   |    |  steps:      |
     |  SET ...  |    |   - record.* |    |   - proc     |
     |           |    |   - email    |    |   - wait     |
     |           |    |   - POST     |    |   - signal   |
     +-----------+    +--------------+    +--------------+
     Synchronous      Synchronous         Asynchronous
     Transaction      Transaction         Durable
```

### Action Types: Details

#### navigate

Client-side navigation. Does not call the backend. Executed by the frontend router.

```jsonc
{
  "key": "create_task",
  "label": "Create Task",
  "type": "navigate",
  "navigate_to": "/app/Task/new?related_to=:recordId"
}
```

Available: Phase 9a (Object View core).

#### field_update

Atomic field update via DML. One operation, one transaction. Does not require the Procedure Engine.

```jsonc
{
  "key": "mark_shipped",
  "label": "Ship",
  "type": "field_update",
  "updates": {
    "status": "shipped",
    "shipped_at": "now()"
  },
  "visibility_expr": "record.status == 'confirmed'"
}
```

Execution: `DML UPDATE obj_order SET status='shipped', shipped_at=NOW() WHERE id=:recordId` with OLS/FLS/RLS enforcement.

Available: Phase 9a (Object View core).

#### procedure

Invocation of a named Procedure (formerly Handler). Synchronous execution of a Command chain.

```jsonc
{
  "key": "send_proposal",
  "label": "Send Proposal",
  "type": "procedure",
  "procedure": "send_proposal",
  "visibility_expr": "record.status == 'draft'"
}
```

Procedure `send_proposal` (JSON, ADR-0024):
```json
{
  "name": "send_proposal",
  "commands": [
    {
      "type": "record.update",
      "object": "Order",
      "id": "$.input.recordId",
      "data": { "status": "\"proposal_sent\"", "proposal_sent_at": "$.now" }
    },
    {
      "type": "notification.email",
      "to": "$.input.record.client_email",
      "template": "proposal",
      "data": {
        "order_number": "$.input.record.order_number",
        "total_amount": "$.input.record.total_amount"
      }
    }
  ],
  "result": { "status": "\"proposal_sent\"" }
}
```

Available: Phase 13a (Procedure Engine).

#### scenario

Launching a long-lived Scenario. Asynchronous execution (fire-and-forget). Returns `execution_id`.

```jsonc
{
  "key": "start_fulfillment",
  "label": "Start Fulfillment",
  "type": "scenario",
  "scenario": "order_fulfillment",
  "visibility_expr": "record.status == 'paid'"
}
```

Available: Phase 13b (Scenario Engine).

### Relationship with Automation Rules (ADR-0019)

Automation Rules use the same Action hierarchy:

```json
{
  "object": "Order",
  "trigger": "record.after_update",
  "condition": "new.status == 'paid' && old.status != 'paid'",
  "action": {
    "type": "scenario",
    "scenario": "order_fulfillment",
    "input": { "orderId": "record.id" }
  }
}
```

An Automation Rule is not a separate concept — it is a **trigger that invokes an Action**. The trigger defines *when*, the Action defines *what*.

### Command Types (formerly Action Types)

Categories of atomic operations inside a Procedure:

| Command Type | Prefix | Examples |
|-------------|--------|---------|
| **record** | `record.*` | `record.create`, `record.update`, `record.delete`, `record.get`, `record.query` |
| **notification** | `notification.*` | `notification.email`, `notification.sms`, `notification.push` |
| **integration** | `integration.*` | `POST url`, `GET url`, `webhook` |
| **compute** | `compute.*` | `transform`, `validate`, `aggregate`, `fail` |
| **flow** | `flow.*` | `call` (procedure), `start` (scenario) |
| **wait** | `wait.*` | `wait signal`, `wait timer`, `wait until` |

### Incremental Implementation

| Phase | What is available | Action types / CEL |
|-------|-------------------|-------------------|
| **Phase 9a** | Object View core | `navigate`, `field_update` |
| **Phase 10** | Custom Functions (ADR-0026) | `fn.*` in any CEL context |
| **Phase 13a** | Procedure Engine | + `procedure` |
| **Phase 13b** | Scenario Engine | + `scenario` |
| **Phase 13c** | Approval Processes | Scenario + built-in approval commands |

Phase 9a starts with `navigate` and `field_update` — they do not require the Procedure/Scenario Engine. Custom Functions (Phase 10) eliminate duplication in CEL expressions. When the Engine arrives, Object View gains new action types **without architectural changes**.

### CEL as Cross-cutting Expression Language

All levels use CEL (ADR-0019, Phase 7b). Custom Functions (ADR-0026) eliminate duplication of CEL expressions:

| Level | Where CEL is used | Example with Function |
|-------|-------------------|-----------------------|
| **Object View** | `visibility_expr` — when to show a button | `fn.is_high_value(record.amount)` |
| **Validation Rule** | `expression` — validation on save | `fn.discount(record.tier, record.amount) < 10000` |
| **Default Expression** | `default_expr` — default value | `fn.discount(record.tier, record.amount)` |
| **Procedure** | `when`, `input.*` — conditions and mapping | `fn.discount($.input.tier, $.input.amount)` |
| **Scenario** | `when`, `input.*` — conditions and mapping | `fn.is_vip($.steps.order.tier)` |
| **Automation Rule** | `condition` — trigger condition | `fn.needs_approval(new.amount, new.tier)` |

A unified expression language from UI to backend — cel-go (backend) + cel-js (frontend). Functions are available on both sides (dual-stack, ADR-0026).

### Terminology in Code

| Term | Go package | DB table | API endpoint |
|------|-----------|----------|-------------|
| Action (definition) | `internal/platform/action` | `metadata.action_definitions` | `/api/v1/admin/actions` |
| Procedure | `internal/platform/procedure` | `metadata.procedures` | `/api/v1/admin/procedures` |
| Command | `internal/platform/procedure/command` | (inline in procedure JSON) | — |
| Scenario | `internal/platform/scenario` | `metadata.scenarios` + `scenario_executions` | `/api/v1/admin/scenarios` |
| Step | `internal/platform/scenario/step` | `scenario_step_history` | — |
| Function | `internal/platform/function` | `metadata.functions` | `/api/v1/admin/functions` |

## Consequences

### Positive

- **Unified vocabulary** — one term = one meaning across all ADRs, documents, code, and discussions
- **Hierarchy is intuitively clear**: Action (what?) -> Command (atomic) -> Procedure (chain) -> Scenario (long-running process)
- **Handler no longer conflicts** — in Go it remains an HTTP handler, declarative block = Procedure
- **Mutation is absorbed** — no separate term needed, it is action type: procedure
- **Extensibility** — new action types (e.g. `approval`, `batch`) are added to the hierarchy without breaking it
- **Incrementality** — Phase 9a works with navigate + field_update, Procedure/Scenario Engine are added later

### Negative

- Transition period: in existing code (if any) old terms may appear — needs a "deprecated terminology" annotation
- "Procedure" may be associated with SQL stored procedure — differentiated by context (metadata vs database)
- Transition period: old terms in documents are not yet updated — needs a "deprecated terminology" annotation

### Related ADRs

- **ADR-0019** — Declarative business logic: Automation Rules use Action types; Object View -> Action binding; the term "Mutation" is replaced by "Action type: procedure"
- **ADR-0020** — DML Pipeline: field_update action type is executed through DML Engine
- **ADR-0022** — Object View: actions config uses the typing from this ADR (navigate, field_update, procedure, scenario)

- **ADR-0024** — Procedure Engine: JSON DSL + Constructor UI. Terminology mapping: "Handler" -> **Procedure**, "Action" -> **Command**, "Action Type" -> **Command Type**
- **ADR-0025** — Scenario Engine: JSON DSL + Constructor UI. Terminology mapping: "Scenario" -> **Scenario** (no change), "Step" -> **Step** (no change), "Handler" (in step context) -> **Procedure**
- **ADR-0026** — Custom Functions: named pure CEL expressions. **Function** is orthogonal to the Action hierarchy (computations, not actions). `fn.*` namespace, dual-stack (cel-go + cel-js)
