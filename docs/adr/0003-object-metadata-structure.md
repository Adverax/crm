# ADR-0003: Object Metadata Structure

**Status:** Accepted
**Date:** 2026-02-08
**Participants:** @roman_myakotin

## Context

A metadata-driven CRM requires a formal description of every object (Account, Contact,
custom objects). The description defines identification, classification, allowed
operations, and connected subsystems.

Key decisions:
- Classification: minimal enum `standard | custom` + behavioral flags (instead of
  a rich enum with types like `system`, `junction`, etc.)
- Security (`default_sharing_model`) deferred to Phase 2 (Security engine)
- Object soft delete (`is_active`) deferred — not needed for MVP
- Record soft delete (`is_deleted`, `deleted_at`, `deleted_by`) deferred — cascading
  delete/restore creates ambiguities, requires a separate subsystem
- i18n: `label`, `plural_label`, `description` are stored as default values,
  translations via the `translations` table (ADR-0002)

## Considered Options

### Classification: rich enum vs minimal enum + flags

**Option A — rich enum:** `standard | custom | system | junction`

Pros: explicit filtering by type.
Cons: rigid, adding a new type requires changing the enum. Junction is derived
from relationships, system is a behavior rather than a type.

**Option B — minimal enum + flags (chosen):**
`object_type: standard | custom`, and `system` behavior is determined through
`is_platform_managed` and other flags.

Pros: flexible, extensible, does not require anticipating all future types.

### Object soft delete (is_active)

Deferred. Cost: cascading behavior (what happens to fields, relationships, records on deactivation),
`WHERE is_active = true` filtering in every metadata query.
For MVP: standard objects are always active, custom objects are removed via hard delete.
Adding `is_active` later requires a single migration.

### Record soft delete (is_deleted / deleted_at / deleted_by)

Deferred. The main issue is the ambiguity of cascading delete and restore:

**Scenario 1:** A detail record is manually deleted, then the master is cascade-deleted.
When restoring the master, the detail that was previously deleted manually should not be restored,
but the system cannot distinguish it from cascade-deleted details without additional metadata
(`delete_reason`, `delete_operation_id`).

**Scenario 2:** The master is deleted, all details are cascade-deleted.
Restoring a single detail without the master results in a broken FK. Restoring the master requires
deciding which details to restore.

A correct implementation requires:
- `delete_reason`: `user_action` | `cascade`
- `delete_operation_id`: UUID for grouping cascade deletions
- Restore logic that accounts for the dependency tree
- UI: deletion tree display, conflict resolver, bulk restore with preview

This is a separate subsystem (Recycle Bin / Archive) that will be designed in a dedicated ADR.
For MVP: hard delete with UI confirmation.

## Decision

### `object_definitions` table

```sql
CREATE TABLE object_definitions (
    -- Identification
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_name                 VARCHAR(100)  NOT NULL UNIQUE,
    label                    VARCHAR(255)  NOT NULL,
    plural_label             VARCHAR(255)  NOT NULL,
    description              TEXT          NOT NULL DEFAULT '',

    -- Physical storage (ADR-0007)
    schema_name              VARCHAR(63)   NOT NULL DEFAULT 'public',
    table_name               VARCHAR(63)   NOT NULL,

    -- Classification
    object_type              VARCHAR(20)   NOT NULL CHECK (object_type IN ('standard', 'custom')),

    -- Behavioral flags (schema level — what can be done with the object itself)
    is_platform_managed      BOOLEAN       NOT NULL DEFAULT false,
    is_visible_in_setup      BOOLEAN       NOT NULL DEFAULT true,
    is_custom_fields_allowed BOOLEAN       NOT NULL DEFAULT true,
    is_deleteable_object     BOOLEAN       NOT NULL DEFAULT true,

    -- Record capabilities (what can be done with records of this object)
    is_createable            BOOLEAN       NOT NULL DEFAULT true,
    is_updateable            BOOLEAN       NOT NULL DEFAULT true,
    is_deleteable            BOOLEAN       NOT NULL DEFAULT true,
    is_queryable             BOOLEAN       NOT NULL DEFAULT true,
    is_searchable            BOOLEAN       NOT NULL DEFAULT true,

    -- Features (connected subsystems)
    has_activities            BOOLEAN       NOT NULL DEFAULT false,
    has_notes                 BOOLEAN       NOT NULL DEFAULT false,
    has_history_tracking      BOOLEAN       NOT NULL DEFAULT false,
    has_sharing_rules         BOOLEAN       NOT NULL DEFAULT false,

    -- System timestamps
    created_at               TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ   NOT NULL DEFAULT now()
);
```

### System fields for records (in each object's data table, not in metadata)

Every object data table automatically contains:

```sql
id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
owner_id    UUID        NOT NULL REFERENCES users(id),
created_by  UUID        NOT NULL REFERENCES users(id),
created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
updated_by  UUID        NOT NULL REFERENCES users(id),
updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
```

Record soft delete (`is_deleted`, `deleted_at`, `deleted_by`) is deferred.
For MVP — hard delete with confirmation. Recycle Bin / Archive is a separate feature
with its own ADR, including design of cascading delete and restore.

### Configuration examples

| Object | object_type | is_platform_managed | is_deleteable_object | is_custom_fields_allowed |
|--------|-------------|---------------------|----------------------|--------------------------|
| Account | standard | false | false | true |
| Contact | standard | false | false | true |
| User | standard | true | false | true |
| Profile | standard | true | false | false |
| Invoice__c | custom | false | true | true |

## Consequences

- The `object_definitions` table is the central registry of all system objects
- The metadata engine checks these flags during SOQL/DML operations for validation
- Security (`default_sharing_model`, OLS, FLS, RLS) is added in Phase 2 via separate tables
- `is_active` / object soft delete is added later if needed
- Record soft delete is deferred — Recycle Bin / Archive will be a separate ADR
- For MVP: hard delete of records with UI confirmation
- i18n for `label`, `plural_label`, `description` — via the `translations` table (ADR-0002)
