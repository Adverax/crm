# ADR-0006: Relationship Registry as Cache, Not a Separate Table

**Status:** Accepted
**Date:** 2026-02-08
**Participants:** @roman_myakotin

## Context

Reference fields in `field_definitions` (ADR-0004, ADR-0005) implicitly define relationships
between objects. For example, a `Contact.account_id` field of type `reference/association`
creates an Account → Contacts relationship.

The platform requires **relationship navigation** for:
- SOQL: `SELECT (SELECT Name FROM Contacts) FROM Account` — reverse relationship
- Admin UI: "Show all relationships of the Account object"
- Cascade analysis: "What will be deleted when Account is deleted?"
- Junction detection: an object with 2+ composition = junction

The question: is a separate `relationship_definitions` table needed, or can relationships
be derived from existing metadata?

## Options Considered

### Option A: Separate relationship_definitions Table

```sql
CREATE TABLE relationship_definitions (
    id                UUID PRIMARY KEY,
    parent_object_id  UUID REFERENCES object_definitions(id),
    child_object_id   UUID REFERENCES object_definitions(id),
    child_field_id    UUID REFERENCES field_definitions(id),
    relationship_name VARCHAR(100),
    relationship_type VARCHAR(20),
    created_at        TIMESTAMPTZ
);
```

**Pros:**
- Fast direct queries by parent/child
- Convenient graph navigation

**Cons:**
- Denormalization: all information already exists in `field_definitions` + `polymorphic_targets`
- Two sources of truth → synchronization required on every field change
- Desynchronization = bugs in SOQL resolution and cascades

### Option B: No Registry — Everything from field_definitions

**Pros:**
- Single source of truth, no synchronization

**Cons:**
- Reverse queries ("all children of Account") require scanning `field_definitions` across all objects
- SOQL resolution of reverse relationships is slower

### Option C: In-Memory Cache Built from field_definitions (chosen)

The relationship graph is **computed** from `field_definitions` + `polymorphic_targets`
and stored in an in-memory cache. Invalidated on metadata changes.

**Pros:**
- Single source of truth (field_definitions) — no desynchronization
- Fast access: O(1) lookup by (object_id, relationship_name)
- Metadata changes rarely (admin operations), reads on every SOQL/DML — cache is ideal
- Metadata caching is inevitable for performance — the relationship graph becomes
  part of the general metadata cache, not a separate subsystem

**Cons:**
- Cache invalidation mechanism needed on metadata changes
- Cold start requires building the graph from DB (one-time operation)

## Decision

We adopt **Option C**. No separate table. The relationship graph is part of
the in-memory metadata cache.

### Cache Structure

```
MetadataCache
├── objects: map[api_name] → ObjectDefinition
├── fields:  map[object_id] → []FieldDefinition
└── relationships:
    ├── forward:  map[object_id][field_api_name] → RelationshipInfo
    └── reverse:  map[object_id][relationship_name] → RelationshipInfo
```

`RelationshipInfo` contains:
- `parent_object_id` — parent object
- `child_object_id` — child object
- `child_field_id` — reference field
- `relationship_name` — relationship name (for SOQL)
- `relationship_type` — association / composition / polymorphic
- `on_delete` — set_null / cascade / restrict

### Cache Construction

On application startup (or invalidation):

1. Load all `field_definitions` with `field_type = 'reference'`
2. Load all `polymorphic_targets`
3. Build forward map: `(child_object, field) → parent_object`
4. Build reverse map: `(parent_object, relationship_name) → child_object + field`
5. For polymorphic: one entry in the reverse map for each target object

### Invalidation

- Any change in `field_definitions` (CREATE/UPDATE/DELETE reference field) → rebuild
- Any change in `polymorphic_targets` → rebuild
- Change in `object_definitions` (DELETE object) → rebuild
- Metadata changes rarely, full rebuild is acceptable (tens/hundreds of objects)

### Usage

```
// SOQL: SELECT (SELECT Name FROM Contacts) FROM Account
// Resolving "Contacts":
rel := cache.relationships.reverse["Account"]["Contacts"]
// → {child_object: "Contact", child_field: "account_id", type: "association"}

// Cascade analysis when deleting Account:
children := cache.relationships.reverse["Account"]
// → filter by on_delete == "cascade"

// All relationships of an object (for Admin UI):
forward := cache.relationships.forward["Contact"]   // what it references
reverse := cache.relationships.reverse["Contact"]   // what references it
```

## Consequences

- No `relationship_definitions` table — single source of truth
- Metadata cache is a mandatory platform component (not an optional optimization)
- SOQL engine, DML engine, Admin UI work with the cache, not with direct queries to field_definitions
- Cache invalidation on any metadata change
- In a clustered configuration — invalidation via pub/sub (Redis, PG NOTIFY) — to be designed later
