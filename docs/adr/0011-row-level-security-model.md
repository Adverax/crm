# ADR-0011: Row-Level Security — OWD and Sharing

**Status:** Accepted
**Date:** 2026-02-08
**Participants:** @roman_myakotin

## Context

RLS determines which **specific records** a user can see. This is the most complex
security layer. The following must be defined:
- Organization-Wide Defaults (OWD) — the baseline record visibility level for an object
- Access extension mechanisms (sharing rules, manual sharing)
- The role of role hierarchy in record visibility
- Physical storage of row-level grants

## Decision

### OWD — Organization-Wide Defaults

Each object has a `visibility` — the baseline level of access to records
for users who have OLS access (Read) to the object.

```sql
ALTER TABLE metadata.object_definitions
ADD COLUMN visibility VARCHAR(30) NOT NULL DEFAULT 'private'
CHECK (visibility IN ('private', 'public_read', 'public_read_write', 'controlled_by_parent'));
```

| Value | Read | Edit | Semantics |
|-------|------|------|-----------|
| `private` | owner + hierarchy + sharing | owner + sharing | Maximum restriction |
| `public_read` | all with OLS Read | owner + sharing | Everyone can see, edit is restricted |
| `public_read_write` | all with OLS Read | all with OLS Update | Fully open |
| `controlled_by_parent` | inherited from parent | inherited from parent | For composition (ADR-0005) |

`controlled_by_parent` applies to child objects in a composition relationship.
Access to a child record is determined by access to the parent record.

### Role Hierarchy — Read Only

A manager (parent role) can see records owned by users in subordinate roles.
Access is granted **for reading only** (permissions = 1).

Applies for OWD `private` and `public_read`:
- `private`: manager **sees** subordinates' records (but cannot edit)
- `public_read`: everyone can already see, hierarchy adds nothing new
- `public_read_write`: everyone can already see and edit

For edit access via hierarchy — use sharing rules or manual sharing.

### Share Tables — Per-Object Grant Storage

Instead of full materialization of `(user × record)` — per-object tables
with compact grants. Created by the DDL engine when creating an object
with OWD != `public_read_write`.

```sql
CREATE TABLE obj_{name}__share (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    record_id       UUID NOT NULL REFERENCES obj_{name}(id) ON DELETE CASCADE,
    grantee_id      UUID NOT NULL REFERENCES iam.groups(id) ON DELETE CASCADE,
    access_level    INT NOT NULL DEFAULT 1,   -- bitmask: 1=R, 5=R+U, etc.
    reason          VARCHAR(30) NOT NULL
                    CHECK (reason IN ('owner', 'sharing_rule', 'territory', 'manual')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (record_id, grantee_id, reason)
);

CREATE INDEX ix_{name}__share_grantee
ON obj_{name}__share (grantee_id);
```

`grantee_id` — always a group_id. No polymorphic `grantee_type` (ADR-0013).
- Manual share to a specific user → grant to their personal group
- Sharing rule for a role → grant to role/role_and_subordinates group
- Unified resolution through `effective_group_members`

`reason` enables targeted revoke: deleting a sharing rule removes only
entries with `reason = 'sharing_rule'`, without affecting manual shares.

### Sharing Rules

Stored in a common table with a rule type:

```sql
CREATE TABLE security.sharing_rules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    object_id       UUID NOT NULL REFERENCES metadata.object_definitions(id),
    rule_type       VARCHAR(20) NOT NULL
                    CHECK (rule_type IN ('owner_based', 'criteria_based')),

    -- Source (whose records are being shared) — always a group (ADR-0013)
    source_group_id UUID NOT NULL REFERENCES iam.groups(id),

    -- Target (who gets access) — always a group (ADR-0013)
    target_group_id UUID NOT NULL REFERENCES iam.groups(id),

    -- Access level
    access_level    INT NOT NULL DEFAULT 1,  -- 1=R, 5=R+U

    -- Criteria (only for criteria_based, NULL for owner_based)
    criteria_field_id   UUID REFERENCES metadata.field_definitions(id),
    criteria_operator   VARCHAR(10)
                        CHECK (criteria_operator IN ('eq', 'neq', 'in', 'gt', 'lt')),
    criteria_value      TEXT,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

Source and target are direct FKs to groups. No polymorphic `source_type`/`target_type`.

**Owner-based:** "Account records owned by group(type=role, Sales) are accessible to group(type=role_and_subordinates, Support) with Read permission."

**Criteria-based:** "Account records where `status = 'Active'` are accessible to group(type=public, Partners) with Read permission."

When a sharing rule is created or modified, grants are asynchronously generated in share tables
(via outbox pattern, ADR-0012).

### Manual Record Sharing

Direct grants in the share table with `reason = 'manual'`. Created via UI or API.
Not automatically removed when sharing rules change.

### Query-time RLS

The SOQL engine builds a WHERE clause depending on the object's OWD:

**OWD = `public_read_write`:** No WHERE clause needed.

**OWD = `public_read`:** WHERE clause for write operations.

**OWD = `private`:**
```sql
WHERE (
  -- 1. Owner
  t.owner_id = :user_id
  -- 2. Role hierarchy (read-down)
  OR t.owner_id IN (
    SELECT visible_owner_id
    FROM security.effective_visible_owner
    WHERE user_id = :user_id
  )
  -- 3. Share table grants (unified path through groups, ADR-0013)
  OR t.id IN (
    SELECT s.record_id
    FROM obj_{name}__share s
    WHERE s.grantee_id IN (
      SELECT group_id
      FROM security.effective_group_members
      WHERE user_id = :user_id
    )
  )
  -- 4. Territory access (Phase N)
  -- OR t.id IN (
  --   SELECT rta.record_id
  --   FROM iam.record_territory_assignment rta
  --   JOIN security.effective_user_territory eut
  --     ON eut.territory_id = rta.territory_id
  --   WHERE eut.user_id = :user_id AND rta.object_id = :object_id
  -- )
)
```

**OWD = `controlled_by_parent`:** Access to the parent record is checked recursively.

## Considered Alternatives

### Full effective_rls Materialization (rejected)

A table `(user_id, object_id, record_id) → permissions`. Size = O(users × records).
With 1,000 users and 1M records = 1B rows. Does not scale.
Recalculation when a sharing rule changes affects millions of rows.

### Query-time via Closure Tables (considered)

No pre-materialization, only closure tables + runtime JOIN.
Works, but `effective_visible_owner` provides a query-time advantage
by using a single lookup instead of two JOINs (ADR-0012).

### Share Tables + Effective Caches (chosen)

Compact per-object share tables (grants, not a full matrix) +
pre-materialized helper caches. Best balance between size and speed.

## Consequences

- Every object with OWD != `public_read_write` gets a share table via the DDL engine
- Sharing rules are stored in a common table, grants are generated asynchronously
- Role hierarchy provides Read only — for edit an explicit grant is needed
- The SOQL engine builds the WHERE clause dynamically based on OWD and caches
- Territory management is Phase N, but the model already supports it
- Criteria-based sharing rules use nullable fields in the common table (no separate criteria table)
