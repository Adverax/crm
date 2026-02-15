# ADR-0007: Table-Per-Object Storage

**Status:** Accepted
**Date:** 2026-02-08
**Participants:** @roman_myakotin

## Context

A metadata-driven CRM allows creating arbitrary objects (custom objects) with arbitrary
fields. It is necessary to determine how record data is stored in PostgreSQL.

Key factors:
- Custom objects are created rarely (admin operations), data queries happen constantly
- SOQL/DML is the unified API for data access, translated into SQL
- Native constraints are important: FK, UNIQUE, NOT NULL, indexes
- PostgreSQL 16 supports transactional DDL (CREATE TABLE, ALTER TABLE are atomic)

## Options Considered

### Option A: Table Per Object (chosen)

Each object gets its own table. Creating an object = `CREATE TABLE`,
adding a field = `ALTER TABLE ADD COLUMN`.

**Pros:**
- Native PostgreSQL performance — indexes, query planner, JOINs
- Strong typing at the DB level (VARCHAR, NUMERIC, UUID, BOOLEAN, TIMESTAMPTZ)
- FK constraints work natively for reference fields
- UNIQUE, NOT NULL constraints at the DB level
- SOQL → SQL translation is trivial (direct mapping object → table, field → column)
- Proven in practice

**Cons:**
- DDL at runtime (CREATE TABLE, ALTER TABLE)
- DDL privileges required for the application
- Number of tables grows with the number of objects

### Option B: EAV (Entity-Attribute-Value)

A single `record_values` table with columns `(record_id, field_id, value_text, value_number, ...)`.

**Pros:** no DDL at runtime, fully dynamic schema.
**Cons:** terrible performance for complex queries (self-JOIN per field),
FK/UNIQUE constraints impossible, SOQL → SQL translation extremely complex.

### Option C: Wide Table (Salesforce approach)

A single `data_rows` table with generic columns `val0..val500`, all values as TEXT.

**Pros:** no DDL, single table.
**Cons:** loss of typing, field count limit, sparse storage, complex casting.

### Option D: JSONB Document

A single `records` table with a `data JSONB` column.

**Pros:** no DDL, JSONB is well-optimized in PG, GIN indexes.
**Cons:** no FK constraints inside JSONB, UNIQUE constraints only via partial index,
aggregations slower than native columns.

## Decision

We adopt **Option A** — table per object.

### Table Placement

The physical location of the table is determined by the `schema_name` and `table_name`
fields in `object_definitions` (ADR-0003). This allows placing data in different PG schemas.

Default convention:

| object_type | schema_name | table_name | Example |
|-------------|-------------|------------|---------|
| standard | `public` | `obj_{api_name}` | `public.obj_account` |
| custom | `public` | `obj_{api_name}` | `public.obj_invoice` |

Administrators can override the schema if needed.

DDL and SOQL/DML engine access the table via `{schema_name}.{table_name}` from metadata,
not by computing the name from `api_name`.

### Object Table Structure

When creating an object, the metadata engine generates DDL:

```sql
CREATE TABLE public.obj_invoice (
    -- System fields (mandatory for every object)
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    UUID        NOT NULL REFERENCES users(id),
    created_by  UUID        NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by  UUID        NOT NULL REFERENCES users(id),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Adding a Field

When creating a field, the metadata engine generates ALTER TABLE:

```sql
-- text/plain
ALTER TABLE obj_invoice ADD COLUMN number VARCHAR(20);

-- number/currency
ALTER TABLE obj_invoice ADD COLUMN amount NUMERIC(18,2);

-- reference/association
ALTER TABLE obj_invoice ADD COLUMN account_id UUID REFERENCES obj_account(id) ON DELETE SET NULL;

-- boolean
ALTER TABLE obj_invoice ADD COLUMN is_paid BOOLEAN NOT NULL DEFAULT false;

-- picklist/single
ALTER TABLE obj_invoice ADD COLUMN status VARCHAR(255);
```

### field_type → DDL Mapping

| field_type | field_subtype | DDL column type |
|-----------|---------------|----------------|
| text | plain | `VARCHAR(n)` — n from config.max_length |
| text | area, rich | `TEXT` |
| text | email | `VARCHAR(255)` |
| text | phone | `VARCHAR(40)` |
| text | url | `VARCHAR(2048)` |
| number | integer | `NUMERIC(p,0)` — p from config.precision |
| number | decimal, currency, percent | `NUMERIC(p,s)` — from config |
| number | auto_number | `INTEGER GENERATED ALWAYS AS IDENTITY` |
| boolean | — | `BOOLEAN` |
| datetime | date | `DATE` |
| datetime | datetime | `TIMESTAMPTZ` |
| datetime | time | `TIME` |
| picklist | single | `VARCHAR(255)` |
| picklist | multi | `TEXT[]` |
| reference | association | `UUID REFERENCES obj_{target}(id) ON DELETE SET NULL` |
| reference | composition | `UUID NOT NULL REFERENCES obj_{target}(id) ON DELETE CASCADE` |
| reference | polymorphic | two columns: `{name}_object_type VARCHAR(100) NOT NULL` + `{name}_record_id UUID NOT NULL` |

### Constraints from Metadata

```sql
-- is_required = true
ALTER TABLE obj_invoice ALTER COLUMN number SET NOT NULL;

-- is_unique = true
ALTER TABLE obj_invoice ADD CONSTRAINT uq_invoice_number UNIQUE (number);
```

### Indexes

The metadata engine automatically creates indexes for:
- FK columns (reference fields)
- Fields with `is_unique = true`
- Composite index for polymorphic reference: `(object_type, record_id)`
- `owner_id` (for RLS queries)

### Deleting a Field

```sql
ALTER TABLE obj_invoice DROP COLUMN amount;
```

### Deleting an Object

```sql
DROP TABLE obj_invoice;
```

Hard delete with confirmation (ADR-0003).

## Consequences

- The metadata engine executes DDL when creating/modifying objects and fields
- The application requires DDL privileges on the data schema
- SOQL → SQL translation: object → `{schema_name}.{table_name}` from metadata, field → column name
- Native PG constraints ensure data integrity
- Schema migration for custom objects is managed by the platform, not by migration files
- Standard objects (Account, Contact, etc.) are created by a seed script during initialization
