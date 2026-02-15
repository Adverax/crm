# ADR-0010: Permission Model — OLS/FLS

**Status:** Accepted
**Date:** 2026-02-08
**Participants:** @roman_myakotin

## Context

It is necessary to determine how Object-Level Security (OLS) and
Field-Level Security (FLS) are stored and computed. The key question: Profile and PermissionSet
have the same permissions structure — should they be stored separately or unified?

## Options Considered

### Option A — Profile = Special PermissionSet (chosen)

Profile contains `base_permission_set_id` — a reference to a PermissionSet.
OLS/FLS are stored **only** in PermissionSet tables.
Single source of truth, single enforcement path.

Pros: no logic duplication, unified computation mechanism.
Cons: slightly less obvious mental model for the admin.

### Option B — Separate Storage

Profile has its own `ProfileObjectPermissions` / `ProfileFieldPermissions`.
PermissionSet has its own `ObjectPermissions` / `FieldPermissions`.

Pros: explicit separation of baseline and additive.
Cons: two tables with identical structure, two code paths, double maintenance.

## Decision

### Profile as a Special PermissionSet

```
Profile
  +id                      UUID PK
  +api_name                VARCHAR(100) UNIQUE
  +label                   VARCHAR(255)
  +description             TEXT
  +base_permission_set_id  UUID NOT NULL → PermissionSet
  +created_at              TIMESTAMPTZ
  +updated_at              TIMESTAMPTZ

PermissionSet
  +id           UUID PK
  +api_name     VARCHAR(100) UNIQUE
  +label        VARCHAR(255)
  +description  TEXT
  +ps_type      VARCHAR(10) NOT NULL DEFAULT 'grant'
                CHECK (ps_type IN ('grant', 'deny'))
  +created_at   TIMESTAMPTZ
  +updated_at   TIMESTAMPTZ
```

- `ps_type = 'grant'` — extends permissions (default)
- `ps_type = 'deny'` — globally suppresses permissions

Profile is a separate entity (assigned to a user, mandatory),
its permissions reside in the linked PermissionSet (always `ps_type = 'grant'`).

### OLS — Object Permissions

```
ObjectPermissions
  +id                  UUID PK
  +permission_set_id   UUID NOT NULL → PermissionSet
  +object_id           UUID NOT NULL → object_definitions
  +permissions         INT NOT NULL DEFAULT 0
  +UNIQUE (permission_set_id, object_id)
```

Bitmask `permissions`:

| Bit | Value | Operation |
|-----|-------|-----------|
| 1   | 0x01  | Read      |
| 2   | 0x02  | Create    |
| 4   | 0x04  | Update    |
| 8   | 0x08  | Delete    |

Examples: `1` = Read only, `3` = Read + Create, `15` = Full CRUD.

### FLS — Field Permissions

```
FieldPermissions
  +id                  UUID PK
  +permission_set_id   UUID NOT NULL → PermissionSet
  +field_id            UUID NOT NULL → field_definitions
  +permissions         INT NOT NULL DEFAULT 0
  +UNIQUE (permission_set_id, field_id)
```

Bitmask `permissions`:

| Bit | Value | Operation |
|-----|-------|-----------|
| 1   | 0x01  | Read      |
| 2   | 0x02  | Write     |

Examples: `0` = Hidden, `1` = Read only, `3` = Read + Write.

### Assigning PermissionSets to Users

```
PermissionSetToUser
  +id                  UUID PK
  +permission_set_id   UUID NOT NULL → PermissionSet
  +user_id             UUID NOT NULL → User
  +UNIQUE (permission_set_id, user_id)
```

### Computing Effective Permissions

**Effective OLS** for a user on an object:

```
-- Step 1: collect all grant PS (profile base + assigned grant PS)
grants = profile.base_ps.permissions[object]
       | grant_ps1.permissions[object]
       | grant_ps2.permissions[object]
       | ...

-- Step 2: collect all deny PS
denies = deny_ps1.permissions[object]
       | deny_ps2.permissions[object]
       | ...

-- Step 3: deny wins over grant
effective_ols(user, object) = grants & ~denies
```

**Effective FLS** — analogously:

```
grants = OR(all grant PS field_permissions[field])
denies = OR(all deny PS field_permissions[field])
effective_fls(user, field) = grants & ~denies
```

If a field is not mentioned in any grant PS — access is `0` (Hidden).
A Deny PS on a field that is not in any grant — has no effect.

**Example:**

```
Profile base PS:       Account = 15 (CRUD)
Grant PS "Sales":      Account = 15 (CRUD)
Deny PS "No Delete":   Account = 8  (Delete)

grants  = 15 | 15 = 15  (0b1111)
denies  = 8             (0b1000)
effective = 15 & ~8 = 7 (0b0111) → Read + Create + Update, NO Delete
```

### Caching

Effective permissions are cached in `effective_ols`, `effective_fls`,
`effective_field_lists` tables (see ADR-0012). Invalidation via the outbox pattern.

## Consequences

- Profile and PermissionSet use shared tables `ObjectPermissions` / `FieldPermissions`
- Single enforcement path: collect grant PS → OR, collect deny PS → OR, result = `grants & ~denies`
- Bitmask encoding enables efficient computation and storage
- Deny PS globally suppresses permissions from any source (ADR-0009)
- Deny applies only to OLS/FLS; RLS (sharing) remains strictly additive
- PermissionSetGroup (PS container) — deferred to Phase 2b
- Assigning PS to groups (`PermissionSetToGroup`) — deferred to Phase 2b
