# ADR-0005: Reference Field Types

**Status:** Accepted
**Date:** 2026-02-08
**Participants:** @roman_myakotin

## Context

A metadata-driven CRM requires describing relationships between objects. Relationships are defined
through reference fields in metadata. The following must be designed:
- Relationship kinds (subtypes within the `reference` type, see ADR-0004)
- Behavior when a parent record is deleted
- Restrictions (depth, cycles, self-reference)
- Polymorphic references (a field pointing to records of different objects)

## Decision

### Three subtypes for the `reference` type

#### association — soft relationship

Objects are related but independent.

| Aspect | Value |
|--------|----------|
| PG storage | `UUID` (nullable) |
| FK constraint | yes |
| on_delete | `set_null` or `restrict` |
| record owner | own |
| sharing/security | own |
| reparenting | always |
| self-reference | allowed (e.g. Account.parent_account_id) |
| max per object | no limit |

#### composition — hard relationship (lifecycle dependency)

The child does not exist without the parent. Part of a whole.

| Aspect | Value |
|--------|----------|
| PG storage | `UUID NOT NULL` |
| FK constraint | yes |
| on_delete | `cascade` or `restrict` |
| record owner | inherited from parent |
| sharing/security | inherited from parent |
| reparenting | controlled by `is_reparentable` flag (default: false) |
| self-reference | **prohibited** (recursive cascade) |
| max per object | unlimited, but chain depth <= 2 |

#### polymorphic — reference to different object types

A field can point to records of different objects. The list of allowed
target objects is stored explicitly.

| Aspect | Value |
|--------|----------|
| PG storage | two columns: `VARCHAR(100)` + `UUID` |
| FK constraint | **no** (validation in DML engine) |
| on_delete | depends on context, validated in code |
| record owner | own |
| self-reference | allowed |
| max per object | no limit |

### Reference field metadata

Additional attributes in `field_definitions` (beyond the common ones):

```sql
-- For association and composition:
referenced_object_id UUID     REFERENCES object_definitions(id),
relationship_name    VARCHAR(100),  -- reverse relationship name for SOQL
on_delete            VARCHAR(20) NOT NULL DEFAULT 'set_null'
                     CHECK (on_delete IN ('set_null', 'cascade', 'restrict')),
is_reparentable      BOOLEAN NOT NULL DEFAULT true,

-- For polymorphic:
-- referenced_object_id = NULL (multiple targets)
-- relationship_name = reverse relationship name
-- on_delete = behavior defined per target or by default
```

### `polymorphic_targets` table

```sql
CREATE TABLE polymorphic_targets (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    field_id   UUID NOT NULL REFERENCES field_definitions(id) ON DELETE CASCADE,
    object_id  UUID NOT NULL REFERENCES object_definitions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (field_id, object_id)
);
```

An explicit list of allowed target objects for each polymorphic field.
No special flags (`is_target_any`, `is_platform_managed`) —
they can be added later if needed.

### Delete behavior (on_delete)

| on_delete | Description | Available for |
|-----------|----------|-------------|
| `set_null` | Set the reference to null on child records | association |
| `cascade` | Delete child records | composition |
| `restrict` | Prevent parent deletion if children exist | association, composition |

### Restrictions

| Restriction | Value | Rationale |
|-------------|----------|-------------|
| Self-reference composition | **prohibited** | Recursive cascade on deletion |
| Composition chain depth | **<= 2** | A->B->C is allowed, A->B->C->D is not. Limits cascading complexity |
| Cycles in composition | **prohibited** | Metadata engine validates on field creation |
| Max compositions per object | no limit | Chain depth is already restricted |

### Data storage in object tables

```sql
-- association (e.g. Contact.account_id):
account_id UUID REFERENCES obj_account(id) ON DELETE SET NULL

-- composition (e.g. DealLineItem.deal_id):
deal_id UUID NOT NULL REFERENCES obj_deal(id) ON DELETE CASCADE

-- polymorphic (e.g. Task.what):
what_object_type VARCHAR(100) NOT NULL
what_record_id   UUID         NOT NULL
-- + composite index (what_object_type, what_record_id)
-- + DML validation: object_type IN polymorphic_targets
```

## Consequences

- The reference type has three subtypes: `association`, `composition`, `polymorphic`
- The metadata engine validates restrictions when creating reference fields
- The DML engine checks referential integrity for polymorphic references (no FK)
- The schema generator creates different DDL depending on the subtype
- Polymorphic targets are stored in a separate normalized table
- Security inheritance (owner/sharing) for composition — Phase 2
