# ADR-0031: Automation Rules

**Date:** 2026-02-25

**Status:** Accepted

**Participants:** @roman_myakotin

## Context

DML Pipeline (ADR-0020) defines 8 stages. Stages 1-7 are implemented
(parse → resolve → defaults → validate → compute → compile → execute).
Stage 8 (Post-Execute) is a placeholder for reactive business logic —
"when X happens, do Y".

The Procedure Engine (ADR-0024, Phase 10a) is production-ready and provides
the execution backend for complex multi-step actions. Automation Rules
bridge the gap between DML data changes and automated business responses.

### Industry context

Salesforce Workflow Rules, Process Builder, and Record-Triggered Flows
provide similar functionality. Our design is simpler: a flat rule list
with CEL conditions, each rule referencing a published procedure,
covering 90% of use cases without the complexity of a visual flow builder.

### Requirements

1. Rules fire automatically on DML events (insert, update, delete)
2. Rules have a CEL condition that gates execution
3. Rules invoke a published procedure for execution
4. Rules execute within the same transaction for data integrity
5. Side effects (integration, notification) go through async outbox
6. Recursion is bounded to prevent infinite loops

## Considered Options

### Transaction boundary

**Option A: Everything in one TX**
- Pros: Full atomicity
- Cons: Long-running TX for HTTP calls, deadlock risk

**Option B: Each rule in its own TX**
- Pros: Isolation between rules
- Cons: Partial execution on failure, complex rollback

**Option C: Data in TX, side effects async (Chosen)**
- Pros: Data integrity + responsive; integration/notification via outbox
- Cons: Eventual consistency for side effects

### Error handling

**Option A: Best-effort (log and continue)**
- Pros: DML never fails due to automation
- Cons: Silent data corruption

**Option B: Fail-fast (rollback entire DML)**
- Pros: Strong consistency
- Cons: One bad rule blocks all writes

**Option C: Data actions rollback, side effects eventually consistent (Chosen)**
- Pros: Data integrity preserved; side effects retry via outbox
- Cons: Slightly more complex error model

### Execution mode

**Option A: Per-record only**
- Pros: Simple mental model
- Cons: N+1 for batch operations

**Option B: Per-batch only**
- Pros: Efficient
- Cons: Complex CEL expressions for individual record logic

**Option C: Configurable per rule (Chosen)**
- Pros: Flexibility — simple rules per-record, aggregation per-batch
- Cons: Two code paths

## Decision

Implement Automation Rules as DML Pipeline Stage 8 (PostExecutor) with:

### Event types
- `before_insert`, `after_insert`
- `before_update`, `after_update`
- `before_delete`, `after_delete`

### Condition
CEL expression evaluated with variables:
- `new` — new record values (INSERT/UPDATE)
- `old` — old record values (UPDATE/DELETE)
- `user` — current user context
- `now` — current timestamp

NULL condition means "always fire".

### Action: procedure_code

Each automation rule references a published procedure via a plain
`procedure_code TEXT` field. All action logic (field updates,
notifications, integrations) is implemented as procedure commands.

This replaces the earlier multi-action-type design (invoke_procedure,
field_update, send_notification) with a single, simpler pattern:
- Field updates → use `record.update` command in a procedure
- Notifications → use `notification.*` commands in a procedure
- Complex logic → use any combination of procedure commands

### Execution mode
- `per_record` — rule fires once per affected record (default)
- `per_batch` — rule fires once for the entire DML batch

### Transaction boundary (Option C)
- `before_*` + DML execute + `after_*` data actions run in same TX
- Integration and notification commands → async outbox
- If a data action fails → entire DML + automation TX rolls back
- If a side effect fails → logged, retried via outbox

### Recursion
Automation may trigger DML which triggers more automation. Bounded by
a platform-configurable depth limit (default 3).

### Schema

```sql
CREATE TABLE metadata.automation_rules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    object_id       UUID NOT NULL REFERENCES metadata.object_definitions(id),
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    event_type      TEXT NOT NULL,
    condition       TEXT,
    procedure_code  TEXT NOT NULL,
    execution_mode  TEXT NOT NULL DEFAULT 'per_record',
    sort_order      INT NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (object_id, name)
);
```

### Platform limits

| Parameter | Default | Description |
|-----------|---------|-------------|
| Max automation depth | 3 | DML → automation → DML → automation → ... |
| Max rules per object per event | 20 | Active rules evaluated per trigger |
| Automation timeout | 30s | Total time for all rules per DML operation |

### No versioning (ADR-0029)
Automation rules follow the save-as-live pattern (like Functions, VR, OV).
No draft/published workflow.

## Consequences

### Positive
- Reactive business logic without code changes
- Reuses Procedure Engine as the single execution backend
- Data integrity via same-TX execution
- Bounded recursion prevents infinite loops
- Simple mental model: one rule → one procedure

### Negative
- Side effects are eventually consistent (not atomic with DML)
- Per-record mode has O(records × rules) cost
- Recursion depth limit may surprise users with complex workflows

### Related ADRs
- ADR-0019: Declarative business logic (5 subsystems)
- ADR-0020: DML Pipeline Extension (Stage 8 = automation)
- ADR-0024: Procedure Engine (execution backend)
- ADR-0029: Versioning Strategy (save-as-live for AR)
