# ADR-0009: Security Architecture — Overview

**Status:** Accepted
**Date:** 2026-02-08
**Participants:** @roman_myakotin

## Context

The CRM platform requires an enterprise-grade security system that controls access
at three levels: object, field, and record. The model is inspired by Salesforce architecture,
adapted for our metadata-driven approach (ADR-0003, ADR-0007).

Key requirements:
- Every data operation undergoes a security check
- Administrators configure access without code changes
- The model is extensible (territories, groups, sharing rules) without breaking changes

## Decision

### Three Security Layers

| Layer | Question | Granularity |
|-------|----------|-------------|
| **OLS** (Object-Level Security) | Can the user perform CRUD on this object type? | user x object |
| **FLS** (Field-Level Security) | Can the user read/write this field? | user x field |
| **RLS** (Row-Level Security) | Can the user see this specific record? | user x record |

Check order: OLS → FLS → RLS. If OLS denies — FLS and RLS are not checked.

### Permission Model: Grant + Deny

Two types of PermissionSet:

- **Grant PS** (default) — extends permissions
- **Deny PS** — globally suppresses permissions

Permission sources:

- Profile (base Grant PS) provides the baseline
- Grant PermissionSet **adds** permissions on top of Profile
- Deny PermissionSet **removes** permissions from the result
- Sharing Rules **open** access to records beyond OWD
- Manual Share / Territory **open** access to specific records

OLS/FLS computation:

```
grants  = profile_base_ps | grant_ps1 | grant_ps2 | ...   (bitwise OR)
denies  = deny_ps1 | deny_ps2 | ...                       (bitwise OR)
effective = grants & ~denies
```

Deny always wins over Grant. The order of PS assignment does not matter.
Deny applies only to OLS/FLS. RLS (sharing) remains strictly additive.

### User Model

```
User
  +id           UUID PK
  +profile_id   UUID NOT NULL → Profile
  +role_id      UUID → UserRole (nullable, one per user)
  +...
```

- **One profile** per user (mandatory) — defines the baseline OLS/FLS
- **One role** per user (optional) — defines the position in the hierarchy for RLS
- The role is stored directly in `User.role_id`, without a junction table

### Enforcement Flow

```
HTTP Request
  │ [JWT → UserContext (user_id, profile_id, role_id)]
  ▼
Handler
  │
  ▼
Service → SOQL (read) / DML (write)
  │
  ├─ OLS: profile + permission sets → can CRUD on this object?
  ├─ FLS: profile + permission sets → which fields are accessible?
  ├─ RLS: OWD + hierarchy + sharing → which records are visible?
  │
  ▼
Repository (parameterized SQL)
  │
  ▼
PostgreSQL
```

SOQL/DML is the single enforcement point. Direct database access from handlers/services is prohibited.

### Phasing

| Phase | Components |
|-------|------------|
| Phase 2a | User, Profile, Grant/Deny PermissionSet, OLS, FLS, effective caches (OLS/FLS), outbox worker |
| Phase 2b | UserRole (hierarchy), OWD, share tables, effective_role_hierarchy, effective_visible_owner, sharing rules, manual sharing, SOQL WHERE injection |
| Phase 2c | Groups (personal, role, role_and_subordinates, public), auto-generation, effective_group_members, share → group resolution |
| Phase 3+ | PermissionSetGroup, PS/PSG → Group assignments |
| Phase N | Territory Management, Territory-based Groups, Audit Trail, Auth (JWT) |

Auth (JWT, login, register) is a separate phase. Until Auth is integrated, a
dev middleware with `X-Dev-User-Id` header is used. The enforcement engine works with the
`UserContext` abstraction, the integration point is a single middleware replacement.

Audit Trail is a separate phase. It integrates as a consumer of outbox events.
Existing code is not affected.

## Options Considered

### User.role — Junction Table vs Direct FK

**Option A — junction table `UserRoleAssignment`:**
Allows multiple roles per user. But creates ambiguity:
which role determines record visibility when traversing the hierarchy?

**Option B — direct FK `User.role_id` (chosen):**
One user = one role. Simple, unambiguous model.
Salesforce uses this approach.

### Deny Rules: Muting PS vs Global Deny vs Three-Valued Logic

**Option A — Muting PS (Salesforce):** Deny only within a PermissionSetGroup,
does not affect other sources. Limited scope.

**Option B — Global Deny PS (chosen):** Deny PS globally suppresses permissions
from any source. Formula: `effective = grants & ~denies`. Full control.

**Option C — Three-valued logic (Grant/Deny/Unset):** Each permission is 2 bits.
Maximum flexibility, but doubles the bitmask size and complicates diagnostics.

## Consequences

- Every data request undergoes three checks (OLS → FLS → RLS)
- SOQL/DML engine is the only data access path
- Permissions are computed as `(OR all grants) & ~(OR all denies)`
- One profile and one role per user
- The security engine evolves incrementally without breaking changes
- OLS/FLS details — ADR-0010, RLS — ADR-0011, caching — ADR-0012
