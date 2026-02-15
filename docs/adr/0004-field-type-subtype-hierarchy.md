# ADR-0004: Type/Subtype Hierarchy for Field Typing

**Status:** Accepted
**Date:** 2026-02-08
**Participants:** @roman_myakotin

## Context

A metadata-driven CRM supports a variety of field types: text, numbers, dates, references,
picklists, etc. A typing approach is needed that ensures:
- Uniform storage logic in PostgreSQL
- Extensibility without changes to the base infrastructure
- Clear separation of storage logic and semantics/validation
- Simplicity in implementing UI components

## Considered Options

### Option A: Flat type enum

```
field_type: text | textarea | rich_text | email | phone | url |
            number | currency | percent | auto_number |
            date | datetime | time | boolean |
            picklist | multipicklist | association | composition
```

**Pros:**
- Simple — one field, one enum
- Unambiguous type -> behavior mapping

**Cons:**
- Every new type (e.g. `ip_address`, `rating`) expands the enum
- Logic duplication: `text`, `email`, `phone`, `url` are stored identically (VARCHAR),
  but processed as different types — requiring separate branches in every switch/case
- No grouping — the validator, SOQL operator, and UI component cannot handle
  "all string types" in a single block

### Option B: Type/subtype hierarchy (chosen)

```
field_type:    text | number | boolean | datetime | picklist | reference
field_subtype: depends on type (nullable for boolean)
```

`type` defines storage (how to store in PG), `subtype` defines semantics
(how to validate and render).

**Pros:**
- Clear separation: storage concern (type) vs semantic concern (subtype)
- Code is organized by type: one handler for all `text/*`, one for all `number/*`
- Extensibility: a new `text/ip_address` is just a subtype + validator, no storage changes
- UI: base component by `type`, behavior modification by `subtype`

**Cons:**
- Two fields instead of one in metadata
- Requires validation of allowed type+subtype combinations

## Decision

We use a **type/subtype** hierarchy. In the `field_definitions` table:

```sql
field_type    VARCHAR(20) NOT NULL,  -- base type: storage concern
field_subtype VARCHAR(20),           -- semantics: nullable (not required for boolean)
```

### Full type/subtype registry

#### text -> VARCHAR / TEXT

| subtype | PG storage | max_length | Validation | UI |
|---------|-----------|-----------|-----------|-----|
| `plain` | VARCHAR(n) | 1-255 | — | text input |
| `area` | TEXT | — | — | textarea |
| `rich` | TEXT | — | HTML sanitize | rich editor |
| `email` | VARCHAR(255) | 255 | email format | mailto link |
| `phone` | VARCHAR(40) | 40 | phone format | tel link |
| `url` | VARCHAR(2048) | 2048 | URL format | clickable link |

#### number -> NUMERIC

| subtype | PG storage | precision/scale | UI |
|---------|-----------|----------------|-----|
| `integer` | NUMERIC(18,0) | configurable | number input |
| `decimal` | NUMERIC(p,s) | configurable | number input |
| `currency` | NUMERIC(18,2) | fixed | with currency symbol |
| `percent` | NUMERIC(5,2) | fixed | with % symbol |
| `auto_number` | sequence + format | — | display only, read-only |

#### boolean -> BOOLEAN

| subtype | Description |
|---------|----------|
| NULL | subtype is not required |

#### datetime -> DATE / TIMESTAMPTZ / TIME

| subtype | PG storage | UI |
|---------|-----------|-----|
| `date` | DATE | date picker |
| `datetime` | TIMESTAMPTZ | datetime picker |
| `time` | TIME | time picker |

#### picklist -> VARCHAR / VARCHAR[]

| subtype | PG storage | UI |
|---------|-----------|-----|
| `single` | VARCHAR(255) | dropdown / radio |
| `multi` | VARCHAR(255)[] | multi-select / checkboxes |

#### reference -> UUID FK

| subtype | nullable | ON DELETE | record owner | UI |
|---------|----------|-----------|-------------|-----|
| `association` | yes | SET NULL | own | search/select, clearable |
| `composition` | no | CASCADE | inherited from parent | search/select, required |

Terminology is taken from UML/DDD instead of Salesforce-specific `lookup`/`master_detail`:
- **association** — objects are related but independent. Deleting the parent does not destroy the child.
- **composition** — lifecycle dependency. The child does not exist without the parent.

Hierarchical relationship (self-referencing, e.g. User -> User for org chart)
is not a separate subtype — it is an `association` with `referenced_object = self`
and cycle validation in the DML engine.

Details on reference types (cascades, inheritance, restrictions) — ADR-0005.

### Validation of allowed combinations

The metadata engine validates that the (type, subtype) pair is in the registry of allowed
combinations when creating/updating a field. Invalid combinations are rejected.

### `field_definitions` table

```sql
CREATE TABLE field_definitions (
    -- Identification
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    object_id            UUID         NOT NULL REFERENCES object_definitions(id),
    api_name             VARCHAR(100) NOT NULL,
    label                VARCHAR(255) NOT NULL,
    description          TEXT         NOT NULL DEFAULT '',
    help_text            TEXT         NOT NULL DEFAULT '',

    -- Typing
    field_type           VARCHAR(20)  NOT NULL,
    field_subtype        VARCHAR(20),

    -- Reference link (direct column for FK constraint)
    referenced_object_id UUID         REFERENCES object_definitions(id),

    -- Structural constraints
    is_required          BOOLEAN      NOT NULL DEFAULT false,
    is_unique            BOOLEAN      NOT NULL DEFAULT false,

    -- Type-specific parameters (JSONB instead of many nullable columns)
    config               JSONB        NOT NULL DEFAULT '{}',

    -- Classification
    is_system_field      BOOLEAN      NOT NULL DEFAULT false,
    is_custom            BOOLEAN      NOT NULL DEFAULT false,
    is_platform_managed  BOOLEAN      NOT NULL DEFAULT false,
    sort_order           INTEGER      NOT NULL DEFAULT 0,

    -- Timestamps
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),

    UNIQUE (object_id, api_name)
);
```

#### Storing type-specific parameters: separate columns vs JSONB config

**Problem:** Type-specific attributes (`max_length`, `precision`, `scale`, `relationship_name`,
`on_delete`, `is_reparentable`, `auto_number_format`, etc.) are only populated for their own type.
With ~8 nullable columns, a `text/email` field would have 7 out of 9 as NULL. Each new attribute requires a migration.

**Solution:** A single `config JSONB` column. The metadata engine validates the config content
against a JSON schema that depends on `(field_type, field_subtype)` when creating/updating a field.

**Exception:** `referenced_object_id` remains a direct column — the FK constraint on
`object_definitions` ensures integrity at the database level.

#### Config content by type

| type/subtype | config |
|---|---|
| text/plain | `{"max_length": 100, "default_value": ""}` |
| text/email, phone, url | `{"default_value": ""}` |
| text/area, rich | `{"default_value": ""}` |
| number/integer | `{"precision": 18, "scale": 0, "default_value": "0"}` |
| number/decimal | `{"precision": 18, "scale": 2, "default_value": "0.00"}` |
| number/currency | `{"precision": 18, "scale": 2, "default_value": "0.00"}` |
| number/percent | `{"precision": 5, "scale": 2, "default_value": "0.00"}` |
| number/auto_number | `{"format": "INV-{0000}", "start_value": 1}` |
| boolean | `{"default_value": "false"}` |
| datetime/* | `{"default_value": ""}` |
| picklist/single | `{"values": [...], "default_value": ""}` — see Picklist values section |
| picklist/multi | `{"values": [...], "default_value": []}` — see Picklist values section |
| reference/association | `{"relationship_name": "Contacts", "on_delete": "set_null"}` |
| reference/composition | `{"relationship_name": "LineItems", "on_delete": "cascade", "is_reparentable": false}` |
| reference/polymorphic | `{"relationship_name": "Activities"}` |

### Picklist values

Picklist values are stored in `config.values[]` — for both local and global picklists.
A uniform pattern: the metadata engine always reads values from config, not from a separate table.

#### Values format in config

```jsonc
{
  // Reference to global picklist (null for local)
  "picklist_id": "uuid-or-null",
  // Values are always here, regardless of source
  "values": [
    {"id": "uuid1", "value": "new", "label": "New", "sort_order": 1, "is_default": true, "is_active": true},
    {"id": "uuid2", "value": "in_progress", "label": "In Progress", "sort_order": 2, "is_default": false, "is_active": true},
    {"id": "uuid3", "value": "closed", "label": "Closed", "sort_order": 3, "is_default": false, "is_active": true}
  ],
  "default_value": "new"
}
```

Each value has an `id` (UUID) — used as `resource_id` in the `translations` table
(`resource_type = 'PicklistValue'`) for i18n.

`is_active` is necessary for picklist values: a deactivated value is not shown
in the dropdown for new records, but existing records retain it. Deleting a value
would break data.

#### Global picklists (Global Value Sets)

Reusable value sets. Stored in separate tables:

```sql
CREATE TABLE picklist_definitions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_name    VARCHAR(100) NOT NULL UNIQUE,
    label       VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE picklist_values (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    picklist_definition_id UUID         NOT NULL REFERENCES picklist_definitions(id) ON DELETE CASCADE,
    value                  VARCHAR(255) NOT NULL,
    label                  VARCHAR(255) NOT NULL,
    sort_order             INTEGER      NOT NULL DEFAULT 0,
    is_default             BOOLEAN      NOT NULL DEFAULT false,
    is_active              BOOLEAN      NOT NULL DEFAULT true,
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (picklist_definition_id, value)
);
```

#### Synchronizing global picklists into config

When updating a global picklist:
1. Update `picklist_values` in the table
2. Find all `field_definitions` where `config->>'picklist_id' = :id`
3. Overwrite `config.values` with the current data from `picklist_values`

Synchronization occurs during admin operations (infrequently). Runtime reads always come from config.

If `config.picklist_id` is populated — the field is bound to a global picklist and values
are fully synchronized. Local deviations are not allowed. To deviate — unbind from
the global picklist (set `picklist_id` to null), then manage locally.

#### Deferred

- Dependent picklists — not needed for MVP
- Value color/icon — not needed for MVP

## Consequences

- `field_definitions` contains `field_type` + `field_subtype` (nullable) + `config` (JSONB)
- `referenced_object_id` — a direct column with FK for integrity
- Type-specific parameters in `config` — no migrations when adding new attributes
- The metadata engine validates config against a JSON schema for each type+subtype combination
- Structural constraints (`is_required`, `is_unique`) — direct columns (they affect DDL)
- Business validation (CEL expressions) — a separate Validation Rules entity (deferred)
- Validators, SOQL operators, and DML handlers are organized by `field_type`
- Adding a new subtype — registration + validator, no migrations or core changes
- UI components: base by `type`, behavior by `subtype`
- `boolean` has no subtype (`field_subtype = NULL`)
- Reference subtypes: `association`, `composition`, `polymorphic` (details in ADR-0005)
- Hierarchical is not a separate subtype — it is an `association` with self-reference
- i18n for `label`, `description`, `help_text` — via the `translations` table (ADR-0002)
- Picklist values are always in `config.values[]` — uniform reads
- Global picklists: `picklist_definitions` + `picklist_values` tables, synced into config
- Dependent picklists and colors — deferred
