# ADR-0025: Scenario Engine — Orchestration of Long-lived Business Processes

**Status:** Accepted
**Date:** 2026-02-15
**Participants:** @roman_myakotin

## Context

### Problem: coordination of multi-step business processes

The platform supports synchronous business logic through DML Pipeline (ADR-0020) and Procedure Engine (ADR-0024): validation, defaults, computed fields, Command chains. However, real business processes often extend beyond a single HTTP request:

| Process | Duration | Characteristics |
|---------|----------|-----------------|
| Client onboarding | Hours to days | Waiting for email confirmation, setting up integrations |
| Discount approval | Days | Human-in-the-loop, escalation on timeout |
| Order processing | Minutes to hours | Saga: reservation, payment, shipping, rollback on error |
| Contract signing | Days to weeks | External signals (DocuSign), reminders |

The current architecture does not solve:

1. **Durability** -- process state is stored in memory; on server restart, context is lost and the process "hangs"
2. **Consistency** -- when an error occurs at step 4 of 7, there is no mechanism for automatic rollback of completed steps 1-3
3. **Waiting for external events** -- a process cannot "sleep" and wake up upon receiving a webhook or user action
4. **Observability** -- no unified execution history; logs are scattered, impossible to answer "at which step is the request stuck?"
5. **Idempotency** -- on retry after a failure, duplications are possible (double charges, repeated emails)

### Operational and Financial Risks Without Orchestration

| Risk | Consequence |
|------|------------|
| "Stuck" operations | Client paid, but order was not created -- no automatic rollback |
| Context loss on failure | After restart, it is unclear which operations completed |
| Double charges | Re-execution without idempotency check |
| Missed deals | Process "hangs", manager forgets, client goes to competitor |

### Relationship with ADR-0023 Terminology

ADR-0023 defined the hierarchy of executable logic: Action (umbrella) -> Command (atomic operation) -> Procedure (synchronous chain) -> **Scenario** (asynchronous long-lived process). Scenario is the top level of the hierarchy, providing durability and coordination. Each Scenario Step invokes a Procedure, inline Command, or built-in operation (`wait signal`, `wait timer`).

### Relationship with Object View (ADR-0022) and Automation Rules (ADR-0019)

- **Object View**: action type `scenario` launches a Scenario from a UI button on the record card
- **Automation Rules**: a "record after update" trigger can launch a Scenario as a reaction (post-execute stage, ADR-0020)

## Considered Options

### Option A -- Direct Service Calls (status quo)

Each process is implemented as a chain of calls in a Go service: `createAccount() -> sendEmail() -> setupIntegration()`. State is in the variables of the current request.

**Pros:**
- No new abstractions
- Simple implementation for 2-3 processes
- No overhead for state serialization

**Cons:**
- No durability: restart = context loss
- No rollback: error at step N leaves the system in an inconsistent state
- Cannot wait for external events (approval, webhook)
- Boilerplate: retry, timeout, state management -- repeated in every service
- Does not scale: 10+ processes = technical debt
- "Only Vasya knows how this process works"

### Option B -- Event-driven Choreography (pub/sub)

Services communicate through events: `AccountCreated -> EmailService.SendWelcome -> IntegrationService.Setup`. Each service reacts to events from others.

**Pros:**
- Loose coupling between services
- Natural scalability
- Each service evolves independently

**Cons:**
- Control flow is distributed across services -- no single place to understand the process
- Debugging through event tracing -- harder than a single observation point
- Failure handling in each service separately -- no centralized rollback
- Circular dependencies between events -- hard-to-discover bugs
- Difficult to add conditional logic ("if amount > 100k, different approver")
- Requires separate infrastructure (message broker)

### Option C -- JSON DSL + Constructor UI (chosen)

A central coordinator executes scenarios described declaratively (JSON, JSONB in PostgreSQL). The administrator assembles a scenario through Constructor UI (analogous to the Procedure Constructor from ADR-0024). Saga pattern for rollback.

**Pros:**
- Constructor-first: administrator assembles a scenario through forms, does not write JSON
- JSON is native to the stack: `encoding/json` (Go), JSONB (PostgreSQL), TypeScript
- Explicit control flow: the entire process is visible in one place
- Centralized error handling and rollback
- Durability out of the box (PostgreSQL)
- Built-in signals, timers, retry policies
- Easy debugging: single observation point, complete execution history
- CEL as cross-cutting expression language (ADR-0019, Phase 7b)

**Cons:**
- Overhead for simple operations (single-step -- Scenario is not needed)
- PostgreSQL as the only backend for durability
- Constructor UI -- additional frontend investment
- Declarativeness is limiting: for complex logic, a Procedure is needed (ADR-0024)

### Option D -- Temporal/Cadence (external workflow engine)

Using a production-grade platform (Temporal.io) for orchestration.

**Pros:**
- Production-ready, proven at scale (Uber, Netflix, Stripe)
- Deterministic replay
- Fork/join, child workflows, versioning
- Active community and documentation

**Cons:**
- External dependency: separate service, cluster, monitoring
- Contradicts the platform's self-hosted focus (ADR-0016) -- the user must deploy Temporal
- Learning curve is significantly higher (Temporal SDK, worker concept, activity vs workflow)
- Workflow code is in Go, not declarative -- administrator cannot configure
- Overkill for 80% of CRM scenarios (linear approval flows)
- Vendor lock-in

## Decision

**Option C chosen: JSON DSL + Constructor UI, persistence in PostgreSQL (JSONB), and Saga pattern for rollback.**

The administrator assembles a scenario through Constructor UI (Scenario Constructor). JSON is the internal representation (IR). Power users can edit JSON directly. As needs grow (fork/join, deterministic replay), migration to Temporal is possible -- the declarative DSL can be compiled into a Temporal workflow.

### Architectural Principle: Orchestration over Choreography

For critical business processes, orchestration is chosen:
- A central coordinator (Orchestrator) knows all the steps
- Explicit control flow is visible in one place
- Centralized error handling and rollback
- Single observation point for monitoring and debugging

### Execution Model: Workflow

**Sequential Workflow** -- the primary model. Steps execute sequentially with conditional skipping (`when`).

Workflow extensions:
- **`goto`** -- jump to an arbitrary step (creating loops, returns)
- **`loop`** -- repeat a group of steps while a condition holds (`while`)

**State Machine is deferred.** Workflow + `goto` + `wait signal` covers 95% of business scenarios. State Machine is planned for future versions (document workflows with statuses, subscriptions with lifecycle).

```
Workflow with extensions:

  +---------+
  | Step 1  |
  +----+----+
       |
       v
  +---------+     +---------+
  | Step 2  |---->| Step 4  |  (goto)
  +----+----+     +---------+
       | when
       v
  +---------+
  | Step 3  |<---+
  +----+----+    | (loop)
       +---------+
```

### Scenario Structure

```json
{
  "code": "order_fulfillment",
  "name": "Order Fulfillment",
  "version": 1,
  "description": "Full cycle: reservation -> payment -> shipping",
  "input": [
    { "name": "orderId", "type": "uuid", "required": true },
    { "name": "amount", "type": "number", "required": true }
  ],
  "steps": [],
  "procedures": {},
  "onError": "compensate",
  "settings": {
    "timeout": "30d",
    "retryPolicy": { "maxAttempts": 3, "delay": "5s", "backoff": 2 }
  },
  "meta": {}
}
```

### Step Structure

Each Step is a unit of work inside a scenario. In accordance with ADR-0023, the `procedure` field (formerly `handler`) points to a Procedure or inline Command.

```json
{
  "code": "charge_payment",
  "name": "Charge Payment",
  "procedure": "process_payment",
  "input": {
    "orderId": "$.input.orderId",
    "amount": "$.input.amount"
  },
  "rollback": {
    "procedure": "refund_payment",
    "input": { "paymentId": "$.steps.charge_payment.paymentId" }
  },
  "retry": { "maxAttempts": 3, "delay": "5s", "backoff": 2 },
  "timeout": "30s",
  "when": "$.steps.reserve.success",
  "goto": null,
  "meta": {}
}
```

Formats for specifying `procedure` in a step:

| Format | Interpretation | Example |
|--------|---------------|---------|
| String identifier | Reference to a Procedure | `"procedure": "create_customer"` |
| String with namespace | External Procedure | `"procedure": "integrations.sync_1c"` |
| Inline Command (JSON) | Single command | `"procedure": { "type": "notification.email", ... }` |
| Array of commands (JSON) | Multiple commands | `"procedure": [{ "type": "record.create", ... }]` |

Name resolution order: first local `procedures` of the scenario, then the global Procedure registry (ADR-0024).

### Execution Lifecycle

```
  pending ---------> running ----------> completed
     |                  |
     |                  +---> waiting ---> running
     |                  |
     |                  +---> compensating ---> failed
     |
     +---> cancelled
```

| Status | Description |
|--------|-------------|
| `pending` | Created, awaiting launch |
| `running` | Executing (active step) |
| `waiting` | Waiting for signal or timer |
| `compensating` | Executing rollback (Saga) |
| `completed` | Successfully completed |
| `failed` | Error (after rollback or fail_fast) |
| `cancelled` | Cancelled manually or via API |

### Signals

**Signal** -- an external event affecting execution. Execution is paused (`status=waiting`) until a signal is received.

```json
{
  "code": "wait_approval",
  "procedure": { "type": "wait.signal", "signalType": "approval_decision", "timeout": "24h" },
  "input": {}
}
```

Signal API:

```
POST /api/v1/executions/{executionId}/signal
{
  "type": "approval_decision",
  "payload": { "approved": true, "comment": "OK" }
}
```

Typical signals: `approval_decision`, `email_confirmed`, `payment_completed`, `document_signed`.

### Timers

Four timer types:

| Type | Purpose | Example |
|------|---------|---------|
| `delay` | Pause for N time | Wait 1 hour before follow-up |
| `until` | Wait until timestamp | Activate on start date |
| `timeout` | Limit signal wait time | Cancel if no response in 7 days |
| `reminder` | Periodic reminder | Remind every 24 hours |

### Rollback (Saga Pattern)

On error at step N, all completed steps are automatically rolled back in reverse order (LIFO):

```
Forward:  Step1 --> Step2 --> Step3 --> Step4 (X Error!)
Rollback:          Comp3 <-- Comp2 <-- Comp1
```

Rollback is executed on a best-effort basis: if a compensating action fails, the error is logged and rollback continues with the next step. Not all steps are required to have a rollback (an email sent cannot be unsent).

Error handling strategies:

| Strategy | Behavior |
|----------|----------|
| `fail_fast` | On first error -- immediately `failed`, without rollback |
| `retry` | Retry per policy, then fail_fast |
| `compensate` | Retry, then rollback all completed steps |

### Durability and Recovery

Execution state is persisted in PostgreSQL after each step. On application restart:

1. Find executions with status `running`, `waiting`, `compensating`
2. For `running` -- determine the last completed step, retry the current one with the same idempotency key
3. For `waiting` -- check for received signals/timers, resume if any
4. For `compensating` -- continue rollback from the current step

A step is recorded as completed only after successful execution. On recovery -- retry the current step, do not repeat completed ones.

### Idempotency

The platform automatically generates an idempotency key for each step:

```
idempotencyKey = {executionId}-{stepCode}
```

This key is passed into the Procedure and further into external calls. A repeated call with the same key must produce the same result. All built-in platform Procedures (`record.*`, `notification.*`) are idempotent by design. For HTTP integrations, the key is passed in the request header.

### Context (Context Model)

Context accumulates as execution progresses and is accessible via CEL expressions:

| Path | Description | Mutability |
|------|-------------|------------|
| `$.input` | Scenario input parameters | Immutable |
| `$.steps.<code>` | Step result (null if skipped by `when`) | Append-only |
| `$.steps.<code>.meta` | Step metadata from the definition | Immutable |
| `$.signals` | Received signals | Append-only |
| `$.meta` | Scenario metadata | Immutable |
| `$.execution` | System data (startedAt, attempt) | Read-only |
| `$.user` | Current user | Immutable |
| `$.now` | Current time | Computed |

Skipped steps (condition `when` = false): `$.steps.<code>` = `null`. Safe access: `$.steps.x != null && $.steps.x.field == "value"`.

### Architectural Components

| Component | Responsibility |
|-----------|---------------|
| **Scenario Registry** | Stores scenario definitions (JSON/JSONB in DB, Go-embedded for built-in) |
| **Orchestrator** | Launches and coordinates executions, manages lifecycle |
| **Step Executor** | Executes a specific step: resolves Procedure, passes input, saves output |
| **Compensator** | Executes compensating actions in reverse order (LIFO) |
| **Signal Handler** | Receives external signals via API, wakes up waiting executions |
| **Timer Scheduler** | Schedules and triggers deferred events, checks timeouts |
| **Execution Repository** | Persists execution state and step history in PostgreSQL |

### Storage

Two tables in PostgreSQL:

**`scenario_executions`** -- execution state:

| Column | Type | Description |
|--------|------|-------------|
| id | UUID PK | Unique execution ID |
| scenario_code | VARCHAR | Scenario code |
| scenario_version | INT | Version at launch time |
| status | VARCHAR | pending/running/waiting/compensating/completed/failed/cancelled |
| input | JSONB | Input parameters |
| context | JSONB | Accumulated context (step results, signals) |
| current_step | VARCHAR | Current step code |
| error | JSONB | Error information (if any) |
| started_at | TIMESTAMPTZ | Start time |
| completed_at | TIMESTAMPTZ | Completion time |
| created_at | TIMESTAMPTZ | Creation time |
| updated_at | TIMESTAMPTZ | Update time |

**`scenario_step_history`** -- step execution history:

| Column | Type | Description |
|--------|------|-------------|
| id | UUID PK | Unique record ID |
| execution_id | UUID FK | Reference to execution |
| step_code | VARCHAR | Step code |
| status | VARCHAR | completed/failed/skipped/compensated |
| input | JSONB | Step input data |
| output | JSONB | Execution result |
| error | JSONB | Error (if any) |
| attempt | INT | Attempt number |
| started_at | TIMESTAMPTZ | Start time |
| completed_at | TIMESTAMPTZ | Completion time |
| created_at | TIMESTAMPTZ | Creation time |

### Limits

| Parameter | Limit | Rationale |
|-----------|-------|-----------|
| Maximum steps in a scenario | 50 | Prevent excessively complex processes |
| Maximum goto depth | 100 iterations | Protection against infinite loops |
| Default execution timeout | 30 days | Protection against "forgotten" executions |
| Default retry maxAttempts | 3 | Balance of reliability and resources |
| Context size (JSONB) | 1 MB | Prevent state bloat |

### State Machine (not in MVP)

Planned for future versions. Workflow + goto + wait signal covers 95% of CRM business scenarios. State Machine will be needed for:
- Document workflows with statuses (draft -> review -> approved -> active)
- Subscriptions with lifecycle (trial -> active -> paused -> cancelled)
- Processes where state matters more than step sequence

Planned syntax: `mode: state_machine`, `states` block instead of `steps`, transitions by events (`on`). Alternative in MVP: Workflow + goto + wait signal for event-driven logic.

## Consequences

### Positive

- **Durability** -- execution state survives restarts; recovery from the last checkpoint; no "stuck" operations
- **Saga guarantees** -- automatic rollback of all completed steps on error; system returns to a consistent state
- **Observability** -- complete execution history of every step with input/output/error; single point for diagnostics
- **Self-service for administrators** -- new processes are assembled through Constructor UI without development; deployment time: days instead of weeks
- **Built-in primitives** -- signals, timers, retry policies, idempotency -- out of the box, without boilerplate in every service
- **Fits into ADR-0023 hierarchy** -- Scenario = top level (Action -> Command -> Procedure -> Scenario); Step invokes Procedure
- **CEL as cross-cutting expression language** -- unified language from Object View to Scenario (`when`, `input.*`)
- **Incremental implementation** -- Phase 13b; does not block Phase 9a (Object View) or Phase 13a (Procedure Engine)
- **Migration path to Temporal** -- as needs grow, the declarative DSL can be compiled into a Temporal workflow

### Negative

- **Overhead for simple operations** -- single-step processes without waiting should not be Scenarios (use Procedure or direct service call)
- **PostgreSQL dependency** -- durability is tied to PostgreSQL; under high execution loads, a separate DB may be needed
- **Learning curve** -- lifecycle, signals, retry policies, Saga pattern -- new concepts for the team (Constructor UI lowers the entry barrier)
- **Declarative limitations** -- for complex computational logic within a step, a Procedure is needed (ADR-0024), not an inline expression

## Related ADRs

- **ADR-0019** -- Declarative business logic: Automation Rules (post-execute trigger) launch Scenarios; CEL as expression language
- **ADR-0020** -- DML Pipeline: post-execute stage can launch a Scenario through Automation Rules
- **ADR-0022** -- Object View: action type `scenario` launches a Scenario from a UI button on the record card
- **ADR-0023** -- Action terminology: Scenario in the hierarchy Action -> Command -> Procedure -> Scenario; Step invokes Procedure
- **ADR-0024** -- Procedure Engine: Steps execute Procedures; Procedure = synchronous Command chain
- **ADR-0029** -- Versioning: Scenario definition is stored in `scenario_versions`. Draft/Published lifecycle. Scenario run captures procedure versions at start through `scenario_run_snapshots`
