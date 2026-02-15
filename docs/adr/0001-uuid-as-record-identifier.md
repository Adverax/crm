# ADR-0001: UUID as Record Identifier

**Status:** Accepted
**Date:** 2026-02-08
**Participants:** @roman_myakotin

## Context

A CRM platform with metadata-driven architecture requires a uniform identifier format for records.
The chosen format must provide:
- System-wide uniqueness
- Unpredictability (protection against IDOR attacks)
- Performance for indexing and join operations
- Support for polymorphic references (a record may reference objects of different types)

## Considered Options

### Option A: Composite ID with key_prefix (Salesforce-style)

Format: `{3-character prefix}{16 random characters}`, e.g. `001a1B2c3D4e5F6g7`.

The prefix encodes the object type — from the ID you can determine that `001...` is an Account.

**Pros:**
- Polymorphic reference with a single field — the object type is encoded in the ID itself
- Debugging convenience — the type is immediately visible
- URL routing without specifying the type: `/record/001xxxx`

**Cons:**
- String PK (19 characters) — index and join degradation compared to native UUID (16 bytes)
- FK constraints are impossible for polymorphic references — validation only in code
- Requires a prefix registry, unique ID generator, collision handling
- A historical Salesforce pattern (1999), not justified by modern requirements

### Option B: UUID v4 (chosen)

Native PostgreSQL `uuid` type. For polymorphic references, an explicit pair of fields is used.

**Pros:**
- Native PostgreSQL support: `uuid` type = 16 bytes, optimized B-tree
- Built-in unpredictability — protection against enumeration
- FK constraints work for direct (non-polymorphic) references
- No additional infrastructure — standard `gen_random_uuid()`
- Ecosystem compatibility (ORMs, clients, tools)

**Cons:**
- Polymorphic reference requires two fields instead of one

## Decision

We use **UUID v4** as the primary key for all records.

For polymorphic references (Activities, Notes, Feed) — an explicit pair of fields:

```sql
related_object_type  VARCHAR(100) NOT NULL  -- Object API name: 'Account', 'Contact'
related_record_id    UUID         NOT NULL  -- Record UUID
```

A composite index `(related_object_type, related_record_id)` ensures fast lookups.

## Consequences

- All tables use `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- System fields `owner_id`, `created_by`, `updated_by` — type `UUID` with FK to `users.id`
- Polymorphic tables (activities, notes, attachments, feed) contain the pair `(object_type, record_id)`
- The SOQL engine performs a lookup by `object_type` when resolving polymorphic fields
