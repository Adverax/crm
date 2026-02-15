# ADR-0012: Security Caching Strategy

**Status:** Accepted
**Date:** 2026-02-08
**Participants:** @roman_myakotin

## Context

Security enforcement is performed on every data request. Computing permissions
on the fly (recursive CTEs over hierarchies, JOINs across all PermissionSets) creates
unacceptable load. A caching layer with guaranteed consistency is needed.

Key requirements:
- Fast lookup during SOQL/DML (O(1) or a single JOIN)
- Correct invalidation when permissions, roles, or groups change
- Cache sizes must be manageable (not O(users × records))

## Decision

### Closure Tables — Hierarchies

Store all (ancestor, descendant) pairs for fast hierarchical queries.

#### effective_role_hierarchy

```sql
CREATE TABLE security.effective_role_hierarchy (
    ancestor_role_id    UUID NOT NULL REFERENCES iam.user_role(id) ON DELETE CASCADE,
    descendant_role_id  UUID NOT NULL REFERENCES iam.user_role(id) ON DELETE CASCADE,
    computed_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (ancestor_role_id, descendant_role_id)
);
```

Size: O(roles^2) in the worst case, realistically O(roles × depth). Tens to hundreds of rows.

#### effective_territory_hierarchy

```sql
CREATE TABLE security.effective_territory_hierarchy (
    ancestor_territory_id   UUID NOT NULL REFERENCES iam.territory(id) ON DELETE CASCADE,
    descendant_territory_id UUID NOT NULL REFERENCES iam.territory(id) ON DELETE CASCADE,
    computed_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (ancestor_territory_id, descendant_territory_id)
);
```

Size: O(territories × depth). Hundreds to thousands of rows.

#### effective_object_hierarchy

```sql
CREATE TABLE security.effective_object_hierarchy (
    ancestor_object_id   UUID NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    descendant_object_id UUID NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    computed_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (ancestor_object_id, descendant_object_id)
);
```

For `controlled_by_parent` OWD — fast parent-chain lookup.
Size: O(objects × depth). Tens of rows.

### Flattened Group Membership

```sql
CREATE TABLE security.effective_group_members (
    group_id    UUID NOT NULL REFERENCES iam.group(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES iam.user(id) ON DELETE CASCADE,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (group_id, user_id)
);
```

Flattens nested groups into a flat `(group, user)` list.
Size: O(groups × avg_members). Thousands to tens of thousands of rows.

### Effective Visible Owners

```sql
CREATE TABLE security.effective_visible_owner (
    user_id          UUID NOT NULL REFERENCES iam.user(id) ON DELETE CASCADE,
    visible_owner_id UUID NOT NULL REFERENCES iam.user(id) ON DELETE CASCADE,
    permissions      INT NOT NULL DEFAULT 1,  -- role hierarchy = Read only
    computed_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, visible_owner_id)
);

CREATE INDEX ix_evo_user ON security.effective_visible_owner (user_id)
INCLUDE (visible_owner_id, permissions);
CREATE INDEX ix_evo_owner ON security.effective_visible_owner (visible_owner_id);
```

Pre-materialized JOIN: `effective_role_hierarchy × users`.
User A can see records of user B if A's role is an ancestor of B's role.
Permissions = 1 (Read only, ADR-0011).

Size: O(users × avg_subordinates). Tens of thousands of rows.

### Effective User Territories

```sql
CREATE TABLE security.effective_user_territory (
    user_id      UUID NOT NULL REFERENCES iam.user(id) ON DELETE CASCADE,
    territory_id UUID NOT NULL REFERENCES iam.territory(id) ON DELETE CASCADE,
    permissions  INT NOT NULL DEFAULT 0,
    computed_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, territory_id)
);

CREATE INDEX ix_eut_user ON security.effective_user_territory (user_id)
INCLUDE (territory_id, permissions);
CREATE INDEX ix_eut_territory ON security.effective_user_territory (territory_id);
```

All territories for a user (direct + transitive via hierarchy)
with aggregated permissions from `territory_object_default`.

Size: O(users × avg_territories). Thousands to tens of thousands of rows.

### Effective OLS / FLS

```sql
CREATE TABLE security.effective_ols (
    user_id     UUID NOT NULL REFERENCES iam.user(id) ON DELETE CASCADE,
    object_id   UUID NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    permissions INT NOT NULL DEFAULT 0,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, object_id)
);

CREATE TABLE security.effective_fls (
    user_id     UUID NOT NULL REFERENCES iam.user(id) ON DELETE CASCADE,
    field_id    UUID NOT NULL REFERENCES metadata.field_definitions(id) ON DELETE CASCADE,
    permissions INT NOT NULL DEFAULT 0,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, field_id)
);

CREATE TABLE security.effective_field_lists (
    user_id     UUID NOT NULL REFERENCES iam.user(id) ON DELETE CASCADE,
    object_id   UUID NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    mask        INT NOT NULL,             -- 1=readable, 2=writable
    field_names TEXT[] NOT NULL DEFAULT '{}',
    computed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, object_id, mask)
);
```

`effective_ols`: `(OR all grant PS) & ~(OR all deny PS)` (ADR-0010).
Size: O(users × objects). Thousands of rows.

`effective_fls`: `(OR all grant PS) & ~(OR all deny PS)` (ADR-0010).
Size: O(users × fields). Tens to hundreds of thousands of rows.

`effective_field_lists`: pre-computed field lists for the API.
Size: O(users × objects × 2). Thousands of rows.

### Outbox Pattern — Cache Invalidation

```sql
CREATE TABLE security.security_outbox (
    id            BIGSERIAL PRIMARY KEY,
    event_type    VARCHAR(50) NOT NULL,
    entity_type   VARCHAR(50) NOT NULL,
    entity_id     UUID NOT NULL,
    payload       JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at  TIMESTAMPTZ
);

CREATE INDEX ix_outbox_unprocessed
ON security.security_outbox (created_at)
WHERE processed_at IS NULL;
```

Triggers on source tables write events to the outbox. A worker processes them:

```sql
SELECT * FROM security.security_outbox
WHERE processed_at IS NULL
ORDER BY created_at
FOR UPDATE SKIP LOCKED
LIMIT 1;
```

| Event | Invalidates |
|-------|-------------|
| `user_changed` (profile/role) | effective_ols, effective_fls, effective_visible_owner |
| `role_changed` (parent) | effective_role_hierarchy, effective_visible_owner |
| `group_changed` (members) | effective_group_members |
| `permission_set_changed` | effective_ols, effective_fls, effective_field_lists |
| `territory_changed` (parent/model) | effective_territory_hierarchy, effective_user_territory |
| `object_changed` (visibility/parent) | effective_object_hierarchy |

### Cache Size Summary

| Cache | Size | Invalidation |
|-------|------|--------------|
| effective_role_hierarchy | O(roles × depth) | Rare (role structure changes) |
| effective_territory_hierarchy | O(territories × depth) | Rare (territory structure changes) |
| effective_object_hierarchy | O(objects × depth) | Rare (metadata changes) |
| effective_group_members | O(groups × members) | On membership changes |
| effective_visible_owner | O(users × subordinates) | On role/user changes |
| effective_user_territory | O(users × territories) | On territory assignment changes |
| effective_ols | O(users × objects) | On PS/profile changes |
| effective_fls | O(users × fields) | On PS/profile changes |
| effective_field_lists | O(users × objects × 2) | On FLS changes |

No cache has a size of O(users × records) — this is the key difference
from the rejected `effective_rls` approach (ADR-0011).

## Considered Alternatives

### In-memory Cache in Go (rejected for permission caches)

OLS/FLS could have been cached in Go process memory.
However: multi-instance deployment requires distributed invalidation (Redis pub/sub, etc.).
PostgreSQL tables serve as a single source of truth, working for any deployment.

Closure tables and effective_* are stored in PostgreSQL. Hot data (current user)
can additionally be cached in Redis or in-memory with a short TTL.

### Materialized Views (rejected)

PostgreSQL materialized views do not support incremental refresh.
`REFRESH MATERIALIZED VIEW` fully recreates the view. Outbox + tables
allow targeted updates of affected rows.

## Consequences

- All caches are PostgreSQL tables in the `security` schema
- Invalidation via outbox pattern (eventual consistency, typically < 1 second)
- Worker processes events sequentially with `FOR UPDATE SKIP LOCKED`
- On cold start — full recalculation of all caches
- Hot cache in Redis/memory — optional optimization on top of PG tables
- Monitoring: alert if outbox queue > threshold
