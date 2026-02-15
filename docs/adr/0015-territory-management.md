# ADR-0015: Territory Management

**Status:** Accepted
**Date:** 2026-02-11
**Participants:** @roman_myakotin

## Context

Territory Management is a mechanism for assigning records to territories (geographic regions,
product lines, verticals) to manage visibility and access. Territories are orthogonal
to role hierarchy: roles define "who you are in the organization", territories define "which
area you are responsible for".

The following must be determined:
- Architecture of territory models (single model vs. multiple models)
- Assignment of users and records to territories
- Mechanism for granting access through territories
- Integration with the existing security model (groups, share tables, effective caches)
- Boundary between core (AGPL) and enterprise (ee/) code

Key constraints:
- Single-tenant architecture (no `tenant_id`, ADR-0007)
- Territories are an enterprise feature (ADR-0014), all code in `ee/`
- Minimal changes in core: only group_type extension + interfaces
- Visibility through territories must work via existing share tables (ADR-0011, ADR-0013)

## Considered Alternatives

### Option A — Single Territory Model (rejected)

One fixed territory hierarchy. No lifecycle, no ability to prepare
a new structure without impacting production.

Pros: simple implementation, no activation complexity.
Cons: no draft/test workflow, seasonal restructuring is impossible, no A/B testing
of territorial divisions.

### Option B — Full Territory Models, ETM2-like (chosen)

Multiple named models with lifecycle (`planning` → `active` → `archived`).
One active model at any time. Models in `planning` can be freely edited.

Pros: draft/test/activate workflow; seasonal restructuring; preparation for M&A;
A/B comparison of geographic vs. industry-based division.
Cons: activation complexity (heavy transaction). Complexity is localized in the activation
service, not spread across the codebase.

### Option C — Territory Groups with `territory_and_subordinates` Type (rejected)

By analogy with `role_and_subordinates` — create two types: `territory` and
`territory_and_subordinates`. Hierarchy propagation through group membership.

Pros: full analogy with the role model.
Cons: impossible to provide per-object access levels — a group grants
the same level of access to all objects. Territory Object Defaults require
different access_levels for different objects within the same territory. Share entries via
ancestor walk solve this with per-object granularity.

## Decision

### Territory Models — Lifecycle

Each model has a status:

| Status | Editing | Affects Access | Transitions |
|--------|---------|----------------|-------------|
| `planning` | Full (CRUD territories, rules, defaults) | No | → `active` |
| `active` | Only assignment rules, record assignments | Yes | → `archived` |
| `archived` | Read only | No | — |

Invariant: at most one active model at any time (enforced by partial unique index).

### SQL Schema

All territory tables are in the `ee` schema (enterprise namespace).
Effective caches are in the `security` schema (common convention, ADR-0012).

#### ee.territory_models

```sql
CREATE TABLE ee.territory_models (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    api_name      VARCHAR(100) NOT NULL UNIQUE,
    label         VARCHAR(255) NOT NULL,
    description   TEXT        NOT NULL DEFAULT '',
    status        VARCHAR(20) NOT NULL DEFAULT 'planning'
                  CHECK (status IN ('planning', 'active', 'archived')),
    activated_at  TIMESTAMPTZ,
    archived_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- At most one active model
CREATE UNIQUE INDEX uq_territory_models_active
ON ee.territory_models (status)
WHERE status = 'active';
```

#### ee.territories

```sql
CREATE TABLE ee.territories (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id    UUID        NOT NULL REFERENCES ee.territory_models(id) ON DELETE CASCADE,
    parent_id   UUID        REFERENCES ee.territories(id) ON DELETE CASCADE,
    api_name    VARCHAR(100) NOT NULL,
    label       VARCHAR(255) NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (model_id, api_name)
);

CREATE INDEX idx_territories_model_id ON ee.territories (model_id);
CREATE INDEX idx_territories_parent_id ON ee.territories (parent_id)
WHERE parent_id IS NOT NULL;
```

`parent_id` must reference a territory in the same model (enforced in service layer).

#### ee.territory_object_defaults

```sql
CREATE TABLE ee.territory_object_defaults (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    territory_id  UUID        NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    object_id     UUID        NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    access_level  VARCHAR(20) NOT NULL CHECK (access_level IN ('read', 'read_write')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (territory_id, object_id)
);

CREATE INDEX idx_territory_object_defaults_territory
ON ee.territory_object_defaults (territory_id);
```

If a territory has no object_default for an object, the territory **does not grant** access
to records of that object (even if records are assigned to the territory).

#### ee.user_territory_assignments

```sql
CREATE TABLE ee.user_territory_assignments (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    territory_id  UUID        NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, territory_id)
);

CREATE INDEX idx_user_territory_assignments_user
ON ee.user_territory_assignments (user_id);
CREATE INDEX idx_user_territory_assignments_territory
ON ee.user_territory_assignments (territory_id);
```

M2M: a user can be assigned to multiple territories simultaneously.

#### ee.record_territory_assignments

```sql
CREATE TABLE ee.record_territory_assignments (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    record_id     UUID        NOT NULL,
    object_id     UUID        NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    territory_id  UUID        NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    reason        VARCHAR(30) NOT NULL DEFAULT 'manual'
                  CHECK (reason IN ('manual', 'assignment_rule')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (record_id, object_id, territory_id)
);

CREATE INDEX idx_record_territory_record
ON ee.record_territory_assignments (record_id, object_id);
CREATE INDEX idx_record_territory_territory
ON ee.record_territory_assignments (territory_id);
```

`record_id` has no FK — records live in different `obj_{name}` tables (ADR-0007).
`object_id` is required to determine which object_default to apply.

#### ee.territory_assignment_rules

```sql
CREATE TABLE ee.territory_assignment_rules (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    territory_id    UUID        NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    object_id       UUID        NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    rule_order      INT         NOT NULL DEFAULT 0,
    criteria_field  VARCHAR(255) NOT NULL,
    criteria_op     VARCHAR(20) NOT NULL
                    CHECK (criteria_op IN ('eq', 'neq', 'in', 'gt', 'lt', 'contains')),
    criteria_value  TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_territory_assignment_rules_territory
ON ee.territory_assignment_rules (territory_id);
CREATE INDEX idx_territory_assignment_rules_object
ON ee.territory_assignment_rules (object_id);
```

Rules are evaluated by `rule_order` on record create/update via the DML engine.
The first matching rule for a territory wins.

### Visibility Mechanism — Share Entries via Ancestor Walk

Territory visibility is implemented through existing share tables (ADR-0011)
and existing `effective_group_members` (ADR-0013). A separate RLS path is not needed.

#### Share Entry Generation Algorithm

When assigning record R (`object_id = O`) to territory T:

1. Build the ancestor chain: `[T, parent(T), grandparent(T), ..., root]`
2. For each territory T' in the chain:
   - Find `territory_object_defaults` for `(T', O)` → `access_level`
   - If object_default **exists**: create a share entry
     `(R, territory_group_T', access_level, reason='territory')`
   - If object_default **does not exist**: skip (no access through this territory)

#### Example

```
EMEA (object_default: Account → read)
└── France (object_default: Account → read_write)
    └── Paris (no object_default for Account)
```

Record Account #42 assigned to Paris:

| Share entry | grantee (territory group) | access_level |
|-------------|--------------------------|--------------|
| 1 | group(France) | read_write |
| 2 | group(EMEA) | read |

- User in Paris: **does not see** Account #42 through Paris (no object_default).
  Sees it only if also assigned to France or EMEA.
- User in France: sees with read_write (share entry 1).
- User in EMEA: sees with read (share entry 2).

#### RLS WHERE Clause

Territory visibility goes through the existing share table path:

```sql
WHERE (
  t.owner_id = :user_id                                      -- 1. Owner
  OR t.owner_id IN (
    SELECT visible_owner_id
    FROM security.effective_visible_owner
    WHERE user_id = :user_id
  )                                                           -- 2. Role hierarchy
  OR t.id IN (
    SELECT s.record_id
    FROM obj_{name}__share s
    WHERE s.grantee_id IN (
      SELECT group_id
      FROM security.effective_group_members
      WHERE user_id = :user_id
    )
  )                                                           -- 3. Sharing (includes territory)
)
```

Share entries with `reason='territory'` participate in the unified resolution through
`effective_group_members`. No separate JOIN for territories is needed.

### Territory Groups

One new group type: `territory`. One group per territory.
Members are users directly assigned to that territory.

The `territory_and_subordinates` type **is not needed** — hierarchy propagation
is provided by share entries (ancestor walk), not group membership.
This gives per-object access granularity that a group cannot provide.

#### Auto-generation

| Event | Action |
|-------|--------|
| Model activation | For each territory: create Group `type='territory'`, `related_territory_id = T.id` |
| User assigned to territory | Add to `group_members` of the territory group |
| User removed from territory | Remove from `group_members` |
| Model archival | Delete all territory groups (CASCADE deletes group_members, share entries) |

All changes → outbox event → recalculation of `effective_group_members` (ADR-0012).

### Effective Caches

#### security.effective_territory_hierarchy

Closure table analogous to `effective_role_hierarchy`:

```sql
CREATE TABLE security.effective_territory_hierarchy (
    ancestor_territory_id   UUID NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    descendant_territory_id UUID NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    depth                   INT  NOT NULL DEFAULT 0,
    computed_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (ancestor_territory_id, descendant_territory_id)
);

CREATE INDEX idx_eth_descendant
ON security.effective_territory_hierarchy (descendant_territory_id);
```

- Self entry: `(T, T, depth=0)` for each territory
- Ancestor entries: `(parent, T, 1)`, `(grandparent, T, 2)`, ...

Size: O(territories * depth). Hundreds to thousands of rows.

#### security.effective_user_territory

```sql
CREATE TABLE security.effective_user_territory (
    user_id        UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    territory_id   UUID NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    computed_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, territory_id)
);

CREATE INDEX idx_eut_user ON security.effective_user_territory (user_id);
CREATE INDEX idx_eut_territory ON security.effective_user_territory (territory_id);
```

Flat list of a user's territories (from `user_territory_assignments`).

Size: O(users * avg_territories). Thousands to tens of thousands of rows.

#### Invalidation (outbox events)

| Event | Invalidates |
|-------|-------------|
| Territory hierarchy change | `effective_territory_hierarchy` |
| User assignment change | `effective_user_territory` + group memberships |
| Record assignment change | Share entries for the record |
| Object defaults change | Share entries for all records in the territory |
| Model activation/archival | Full recalculation of all territory caches |

### Model Activation Algorithm

```
ActivateModel(newModelID):
  1. Verify: newModel.status == 'planning'
  2. Verify: newModel has at least one territory
  3. BEGIN TRANSACTION
     a. Find current active model (oldModel)
     b. If oldModel != nil:
        i.   UPDATE oldModel SET status='archived', archived_at=now()
        ii.  DELETE territory groups (CASCADE → group_members, share entries)
        iii. DELETE effective_territory_hierarchy for old territories
        iv.  DELETE effective_user_territory for old territories
        v.   DELETE share entries with reason='territory' for ALL share tables
     c. UPDATE newModel SET status='active', activated_at=now()
     d. For each territory T in newModel:
        i.   CREATE Group type='territory', related_territory_id=T.id
        ii.  For each user in user_territory_assignments(T):
             INSERT INTO group_members
     e. Rebuild effective_territory_hierarchy for newModel
     f. Rebuild effective_user_territory for all users
     g. Rebuild effective_group_members for new territory groups
     h. For each record_territory_assignment in newModel:
        i.   Ancestor walk → create share entries
  4. COMMIT
  5. Emit outbox events
```

Activation is a heavy operation. For MVP: synchronous in a single transaction.
Optimization: background job with progress tracking.

### Stored Functions — Hybrid Go + PL/pgSQL Approach

#### Problem: Round-trip Overhead

The current pattern for computing effective caches in the project is pure Go:
recursive calls in Go with a separate query for each nesting level,
INSERT in a loop one row at a time. For the security engine (tens of roles, hundreds
of users) this is acceptable.

Territory management scales differently. Activating a model with 50 territories,
200 users, and 10K records generates ~50K round-trips in a single transaction:
creating groups (50), populating group_members (200), closure table (50 × depth),
effective_user_territory (200), share entries (10K × depth ancestor walk ×
checking object_defaults × INSERT). At ~0.1ms/round-trip over loopback this is 5 seconds
**on latency alone**, not counting execution time.

#### Considered Alternatives

**Option A — Pure Go (like the current security engine)**

All code in Go. A separate SQL query for each operation via sqlc.

Pros: single stack, familiar unit tests, debugging in Go.
Cons: ~50K round-trips on activation; O(records × depth) queries for share entry generation; does not scale for production volumes.

**Option B — Pure PL/pgSQL (all logic in stored procedures)**

All business logic, including validation and lifecycle, in stored functions.

Pros: zero round-trips, maximum performance.
Cons: business logic in SQL is hard to test, debug, and code review; duplication of validation between Go (API) and PL/pgSQL; violates project convention (handler → service → repository).

**Option C — Hybrid: PL/pgSQL for Data-intensive Ops, Go for Business Logic (chosen)**

Three stored functions for operations with high round-trip overhead.
All business logic, CRUD, validation, rule evaluation remains in Go.

Pros: ~50K round-trips → ~3 function calls for activation; recursive CTE is a native PostgreSQL strength; business logic stays in Go (testability, debugging); clear separation of responsibilities.
Cons: two languages for the territory engine; stored functions in migrations (versioning).

#### Stored Functions (3 total)

##### 1. `ee.rebuild_territory_hierarchy(p_model_id UUID)`

Recalculates the closure table `security.effective_territory_hierarchy` for a model.
Uses recursive CTE instead of recursive Go calls.

```sql
CREATE FUNCTION ee.rebuild_territory_hierarchy(p_model_id UUID)
RETURNS void AS $$
BEGIN
    DELETE FROM security.effective_territory_hierarchy
    WHERE ancestor_territory_id IN (
        SELECT id FROM ee.territories WHERE model_id = p_model_id
    );

    INSERT INTO security.effective_territory_hierarchy
        (ancestor_territory_id, descendant_territory_id, depth)
    WITH RECURSIVE closure AS (
        -- Self entries
        SELECT id AS ancestor, id AS descendant, 0 AS depth
        FROM ee.territories
        WHERE model_id = p_model_id
        UNION ALL
        -- Walk up: for each territory, add its parent as ancestor
        SELECT t.parent_id AS ancestor, c.descendant, c.depth + 1
        FROM closure c
        JOIN ee.territories t ON t.id = c.ancestor
        WHERE t.parent_id IS NOT NULL
          AND t.model_id = p_model_id
    )
    SELECT ancestor, descendant, depth FROM closure;
END;
$$ LANGUAGE plpgsql;
```

**Instead of:** Go code that loads all territories, builds a parent_map in memory,
walks up for each node, and inserts rows one at a time.
**Benefit:** N × depth INSERT → 1 INSERT...SELECT. Zero round-trips.

##### 2. `ee.generate_record_share_entries(p_record_id UUID, p_object_id UUID, p_territory_id UUID, p_share_table TEXT)`

Generates share entries for a single record when assigned to a territory.
Performs ancestor walk through the closure table, checks object_defaults,
and creates share entries in a single call.

```sql
CREATE FUNCTION ee.generate_record_share_entries(
    p_record_id    UUID,
    p_object_id    UUID,
    p_territory_id UUID,
    p_share_table  TEXT
) RETURNS void AS $$
DECLARE
    rec RECORD;
BEGIN
    FOR rec IN
        SELECT g.id AS group_id, tod.access_level
        FROM security.effective_territory_hierarchy eth
        JOIN ee.territory_object_defaults tod
            ON tod.territory_id = eth.ancestor_territory_id
            AND tod.object_id = p_object_id
        JOIN iam.groups g
            ON g.related_territory_id = eth.ancestor_territory_id
            AND g.group_type = 'territory'
        WHERE eth.descendant_territory_id = p_territory_id
    LOOP
        EXECUTE format(
            'INSERT INTO %I (record_id, group_id, access_level, reason)
             VALUES ($1, $2, $3, $4)
             ON CONFLICT (record_id, group_id, reason) DO UPDATE SET access_level = $3',
            p_share_table
        ) USING p_record_id, rec.group_id, rec.access_level, 'territory';
    END LOOP;
END;
$$ LANGUAGE plpgsql;
```

**Instead of:** Go code with separate queries: 1 SELECT ancestor chain, N SELECT
object_defaults, N INSERT share entries.
**Benefit:** 2 × depth queries per record → 1 function call.
For 10K records: ~60K round-trips → ~10K calls.

##### 3. `ee.activate_territory_model(p_new_model_id UUID)`

Full orchestration of model activation: archiving the old one, creating groups,
populating group_members, recalculating all caches, generating share entries.

```sql
CREATE FUNCTION ee.activate_territory_model(p_new_model_id UUID)
RETURNS void AS $$
DECLARE
    v_old_model_id UUID;
    v_territory RECORD;
    v_assignment RECORD;
    v_group_id UUID;
    v_share_table TEXT;
BEGIN
    -- 1. Archive old active model (if exists)
    SELECT id INTO v_old_model_id
    FROM ee.territory_models WHERE status = 'active';

    IF v_old_model_id IS NOT NULL THEN
        UPDATE ee.territory_models
        SET status = 'archived', archived_at = now(), updated_at = now()
        WHERE id = v_old_model_id;

        -- CASCADE: delete territory groups → group_members → share entries
        DELETE FROM iam.groups
        WHERE related_territory_id IN (
            SELECT id FROM ee.territories WHERE model_id = v_old_model_id
        );

        -- Clean effective caches for old model
        DELETE FROM security.effective_territory_hierarchy
        WHERE ancestor_territory_id IN (
            SELECT id FROM ee.territories WHERE model_id = v_old_model_id
        );

        DELETE FROM security.effective_user_territory
        WHERE territory_id IN (
            SELECT id FROM ee.territories WHERE model_id = v_old_model_id
        );

        -- Clean territory share entries from all share tables
        FOR v_share_table IN
            SELECT table_name || '__share'
            FROM metadata.object_definitions
            WHERE visibility = 'private'
        LOOP
            EXECUTE format(
                'DELETE FROM %I WHERE reason = $1', v_share_table
            ) USING 'territory';
        END LOOP;
    END IF;

    -- 2. Activate new model
    UPDATE ee.territory_models
    SET status = 'active', activated_at = now(), updated_at = now()
    WHERE id = p_new_model_id;

    -- 3. Create territory groups + populate members (batch INSERT...SELECT)
    FOR v_territory IN
        SELECT id, api_name FROM ee.territories WHERE model_id = p_new_model_id
    LOOP
        INSERT INTO iam.groups (api_name, label, group_type, related_territory_id)
        VALUES (
            'territory_' || v_territory.api_name,
            (SELECT label FROM ee.territories WHERE id = v_territory.id),
            'territory',
            v_territory.id
        )
        RETURNING id INTO v_group_id;

        INSERT INTO iam.group_members (group_id, member_user_id)
        SELECT v_group_id, uta.user_id
        FROM ee.user_territory_assignments uta
        WHERE uta.territory_id = v_territory.id;
    END LOOP;

    -- 4. Rebuild effective_territory_hierarchy
    PERFORM ee.rebuild_territory_hierarchy(p_new_model_id);

    -- 5. Rebuild effective_user_territory
    INSERT INTO security.effective_user_territory (user_id, territory_id)
    SELECT uta.user_id, uta.territory_id
    FROM ee.user_territory_assignments uta
    JOIN ee.territories t ON t.id = uta.territory_id
    WHERE t.model_id = p_new_model_id;

    -- 6. Rebuild effective_group_members for territory groups
    INSERT INTO security.effective_group_members (group_id, user_id)
    SELECT gm.group_id, gm.member_user_id
    FROM iam.group_members gm
    JOIN iam.groups g ON g.id = gm.group_id
    WHERE g.group_type = 'territory'
      AND g.related_territory_id IN (
          SELECT id FROM ee.territories WHERE model_id = p_new_model_id
      );

    -- 7. Generate share entries for all record assignments
    FOR v_assignment IN
        SELECT rta.record_id, rta.object_id, rta.territory_id, od.table_name
        FROM ee.record_territory_assignments rta
        JOIN ee.territories t ON t.id = rta.territory_id
        JOIN metadata.object_definitions od ON od.id = rta.object_id
        WHERE t.model_id = p_new_model_id
    LOOP
        PERFORM ee.generate_record_share_entries(
            v_assignment.record_id,
            v_assignment.object_id,
            v_assignment.territory_id,
            v_assignment.table_name || '__share'
        );
    END LOOP;
END;
$$ LANGUAGE plpgsql;
```

**Instead of:** Go activation_service with ~50K round-trips.
**Benefit:** 1 call to `SELECT ee.activate_territory_model(id)` replaces
the entire Go orchestration. All work is server-side, zero network round-trips.

#### What Stays in Go

| Operation | Reason |
|-----------|--------|
| CRUD for all 6 tables | Trivial queries, sqlc handles them |
| Validation (status transitions, parent same model) | Business logic, unit-testable |
| Assignment rule evaluation | Field metadata interpretation, complex criteria matching |
| Outbox event dispatch | Existing pattern, orchestration |
| Service-level coordination | handler → service → repository |
| Calling stored functions | Go service calls `SELECT ee.activate_territory_model($1)` via repository |

#### Testing Stored Functions

- **pgTAP** (`ee/tests/pgtap/functions/`): unit tests for each stored function
  - `rebuild_territory_hierarchy_test.sql`: closure table verification for trees of depth 1, 2, 3
  - `generate_record_share_entries_test.sql`: ancestor walk verification with/without object_defaults
  - `activate_territory_model_test.sql`: full activation flow, archiving old model
- **Go integration tests** (`//go:build integration`): end-to-end via repository → function → assert DB state

### Assignment Rules — DML Integration

```
EvaluateAssignmentRules(objectID, recordID, recordFields):
  1. Get active model → territories
  2. For each territory T:
     a. Get rules for (T, objectID) WHERE is_active ORDER BY rule_order
     b. For each rule:
        i.   Extract field value from recordFields by criteria_field
        ii.  Apply criteria_op to value and criteria_value
        iii. If match:
             - INSERT record_territory_assignment (recordID, objectID, T, 'assignment_rule')
             - Generate share entries (ancestor walk + object_defaults)
             - break (first matching rule per territory)
```

Rules are evaluated synchronously on DML insert/update. Simple criteria for MVP.

### Minimal Changes in Core (AGPL)

1. **Migration**: extend CHECK constraint in `iam.groups` — add `'territory'`
2. **Type constant**: `GroupTypeTerritory = "territory"` in `internal/platform/security/types.go`
3. **Validation**: update `ValidateCreateGroup` to accept `territory`
4. **Interface**: `TerritoryResolver` in rls package (noop implementation `//go:build !enterprise`)
5. **Interface**: `TerritoryAssignmentEvaluator` in dml package (noop implementation)
6. **Outbox worker**: handling of `territory_changed` event type (delegates to enterprise via interface)
7. **Share table DDL**: already supports `reason='territory'` — no changes needed

```go
// internal/platform/security/rls/territory.go (core, AGPL)
type TerritoryResolver interface {
    ResolveTerritoryGroups(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

// internal/platform/security/rls/territory_default.go (core, AGPL)
//go:build !enterprise

type noopTerritoryResolver struct{}

func (r *noopTerritoryResolver) ResolveTerritoryGroups(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
    return nil, nil
}
```

```go
// internal/platform/dml/territory.go (core, AGPL)
type TerritoryAssignmentEvaluator interface {
    EvaluateOnInsert(ctx context.Context, objectID, recordID uuid.UUID, fields map[string]interface{}) error
    EvaluateOnUpdate(ctx context.Context, objectID, recordID uuid.UUID, fields map[string]interface{}) error
}
```

### Extending Groups for Enterprise

```sql
-- EE migration: add FK to territory
ALTER TABLE iam.groups
  ADD COLUMN related_territory_id UUID REFERENCES ee.territories(id) ON DELETE CASCADE;

CREATE INDEX idx_iam_groups_related_territory
ON iam.groups (related_territory_id)
WHERE related_territory_id IS NOT NULL;
```

### ee/ File Structure

```
ee/
├── internal/
│   └── platform/
│       └── territory/
│           ├── types.go                        ← types: TerritoryModel, Territory, etc.
│           ├── inputs.go                       ← input structs
│           ├── validation.go                   ← input data validation
│           ├── repository.go                   ← repository interfaces
│           ├── pg_model_repo.go                ← PG: TerritoryModelRepository
│           ├── pg_territory_repo.go            ← PG: TerritoryRepository
│           ├── pg_object_default_repo.go       ← PG: TerritoryObjectDefaultRepository
│           ├── pg_user_assignment_repo.go      ← PG: UserTerritoryAssignmentRepository
│           ├── pg_record_assignment_repo.go    ← PG: RecordTerritoryAssignmentRepository
│           ├── pg_assignment_rule_repo.go      ← PG: TerritoryAssignmentRuleRepository
│           ├── pg_effective_repo.go            ← PG: TerritoryEffectiveCacheRepository
│           ├── model_service.go                ← CRUD + model lifecycle
│           ├── territory_service.go            ← CRUD territories, hierarchy ops
│           ├── object_default_service.go       ← object defaults management
│           ├── user_assignment_service.go      ← user assignment
│           ├── record_assignment_service.go    ← record assignment
│           ├── assignment_rule_service.go      ← CRUD rules
│           ├── activation_service.go           ← activation logic (stored function call)
│           ├── share_generator.go              ← share entry generation (stored function call)
│           ├── effective_computer.go           ← closure table recalculation (stored function call)
│           ├── resolver.go                     ← TerritoryResolver implementation
│           ├── evaluator.go                    ← TerritoryAssignmentEvaluator implementation
│           ├── model_service_test.go
│           ├── territory_service_test.go
│           ├── activation_service_test.go
│           ├── share_generator_test.go
│           └── effective_computer_test.go
├── internal/
│   └── handler/
│       └── territory_handler.go                ← Gin HTTP handlers
├── migrations/
│   ├── 000001_create_territory_schema.up.sql
│   ├── 000001_create_territory_schema.down.sql
│   ├── 000002_create_effective_territory_caches.up.sql
│   ├── 000002_create_effective_territory_caches.down.sql
│   ├── 000003_alter_groups_add_territory.up.sql
│   ├── 000003_alter_groups_add_territory.down.sql
│   ├── 000004_create_territory_functions.up.sql   ← 3 stored functions
│   └── 000004_create_territory_functions.down.sql
├── sqlc/
│   └── queries/
│       ├── territory_models.sql
│       ├── territories.sql
│       ├── territory_object_defaults.sql
│       ├── user_territory_assignments.sql
│       ├── record_territory_assignments.sql
│       └── territory_assignment_rules.sql
├── tests/
│   └── pgtap/
│       ├── schema/
│       │   ├── territory_models_test.sql
│       │   ├── territories_test.sql
│       │   └── territory_effective_caches_test.sql
│       └── functions/
│           ├── rebuild_territory_hierarchy_test.sql
│           ├── generate_record_share_entries_test.sql
│           └── activate_territory_model_test.sql
└── web/
    └── src/
        ├── views/
        │   ├── TerritoryModelListView.vue
        │   ├── TerritoryModelCreateView.vue
        │   ├── TerritoryModelDetailView.vue
        │   ├── TerritoryTreeView.vue
        │   ├── TerritoryDetailView.vue
        │   └── TerritoryAssignmentRulesView.vue
        ├── components/
        │   ├── TerritoryTree.vue
        │   ├── TerritoryObjectDefaultsEditor.vue
        │   └── TerritoryAssignmentRuleForm.vue
        ├── stores/
        │   └── territory.ts
        └── router/
            └── territory-routes.ts
```

### API Endpoints

```
POST   /api/v1/admin/territory/models              — create model
GET    /api/v1/admin/territory/models              — list models
GET    /api/v1/admin/territory/models/:id          — get model
PUT    /api/v1/admin/territory/models/:id          — update model
DELETE /api/v1/admin/territory/models/:id          — delete model (planning only)
POST   /api/v1/admin/territory/models/:id/activate — activate model

POST   /api/v1/admin/territory/territories              — create territory
GET    /api/v1/admin/territory/territories?model_id=    — list by model
GET    /api/v1/admin/territory/territories/:id          — get territory
PUT    /api/v1/admin/territory/territories/:id          — update territory
DELETE /api/v1/admin/territory/territories/:id          — delete territory

POST   /api/v1/admin/territory/territories/:id/object-defaults        — set object default
GET    /api/v1/admin/territory/territories/:id/object-defaults        — list object defaults
DELETE /api/v1/admin/territory/territories/:id/object-defaults/:objId — delete

POST   /api/v1/admin/territory/territories/:id/users                  — assign user
GET    /api/v1/admin/territory/territories/:id/users                  — list users
DELETE /api/v1/admin/territory/territories/:id/users/:userId          — remove user

POST   /api/v1/admin/territory/territories/:id/records                — assign record
GET    /api/v1/admin/territory/territories/:id/records                — list records
DELETE /api/v1/admin/territory/territories/:id/records/:recordId      — remove record

POST   /api/v1/admin/territory/assignment-rules                       — create rule
GET    /api/v1/admin/territory/assignment-rules?territory_id=         — list rules
PUT    /api/v1/admin/territory/assignment-rules/:id                   — update rule
DELETE /api/v1/admin/territory/assignment-rules/:id                   — delete rule
```

### Implementation Phases

**Phase 1: Core Integration Points (minimal AGPL changes)**
- Core migration: `'territory'` in group_type CHECK
- `GroupTypeTerritory` constant
- Update `ValidateCreateGroup`
- `TerritoryResolver` interface (noop default)
- `TerritoryAssignmentEvaluator` interface (noop default)
- `territory_changed` handler in outbox worker

**Phase 2: Enterprise Schema (ee/ migrations)**
- 6 territory tables
- Effective cache tables
- `related_territory_id` column on `iam.groups`
- 3 stored functions: `rebuild_territory_hierarchy`, `generate_record_share_entries`, `activate_territory_model`

**Phase 3: Enterprise Backend**
- Types, inputs, validation
- Repository interfaces and PG implementations
- Services (Model, Territory, ObjectDefault, UserAssignment, RecordAssignment, AssignmentRule)
- Share generator, effective computer, activation service — stored function calls via repository
- Assignment rule evaluator (Go, synchronous on DML)

**Phase 4: Enterprise Handler + Routes**

**Phase 5: Enterprise Frontend** (Pinia store, model views, tree view, detail views)

**Phase 6: Tests** (pgTAP, Go unit, E2E)

## Consequences

- All territory tables are in the `ee` schema (enterprise namespace)
- Effective caches are in the `security` schema (common convention, ADR-0012)
- The only change in core: `'territory'` in group_type CHECK + interfaces with noop defaults
- Share entries with `reason='territory'` use unified enforcement through `effective_group_members` (ADR-0013)
- Hybrid Go + PL/pgSQL approach: 3 stored functions for data-intensive operations (closure table, share generation, activation), all business logic in Go
- Model activation is performed by a single call to `SELECT ee.activate_territory_model(id)` — zero network round-trips
- Stored functions are tested via pgTAP (`ee/tests/pgtap/functions/`), Go services via unit tests
- Assignment rules are evaluated synchronously on DML insert/update in Go; simple criteria for MVP
- Territory groups are auto-generated on model activation, deleted on archival
- Share entry recalculation is needed when: assigning a record, changing object_defaults, activating a model
- Frontend: enterprise views via dynamic imports + `VITE_ENTERPRISE` flag
- Build tag `//go:build enterprise` on all Go files in `ee/`
- One group type `territory` (no `territory_and_subordinates`) — hierarchy propagation via share entries
