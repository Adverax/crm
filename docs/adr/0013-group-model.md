# ADR-0013: Group Model

**Status:** Accepted
**Date:** 2026-02-08
**Participants:** @roman_myakotin

## Context

Groups are a key sharing mechanism. Sharing rules and share tables reference access recipients.
The following must be defined:
- Group types and their creation rules
- The grantee model in share tables
- Lifecycle of auto-generated groups

## Considered Alternatives

### Option A — Auto-generated Groups, Unified Grantee (chosen)

All sharing goes through groups. The share table contains only `grantee_id` (always a group).
The system automatically creates groups for roles and users.

Pros: unified model — one grantee type, one cache, one resolution path.
Cons: auto-generation on role/user changes. More rows in group/group_member.

### Option B — Polymorphic Grantee, No Auto-generation

Groups are manual only. Share tables have `grantee_type: 'user' | 'group' | 'role' | 'role_and_subordinates'`.
Each type has its own resolution logic.

Pros: no auto-generation, less data.
Cons: multiple resolution paths in enforcement, more complex WHERE clause.

## Decision

### Group Types

| Type | Creation | `related_role_id` | Membership |
|------|----------|-------------------|------------|
| `personal` | Auto on User creation | NULL | Only this user |
| `role` | Auto on UserRole creation | NOT NULL | Users with this role |
| `role_and_subordinates` | Auto on UserRole creation | NOT NULL | Users with this role + subordinate roles |
| `public` | Manually by admin | NULL | Manual: users + nested groups |

Territory-based types (`territory`, `territory_and_subordinates`) — Phase N.

### Groups Table

```sql
CREATE TABLE iam.groups (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_name          VARCHAR(100) NOT NULL UNIQUE,
    label             VARCHAR(255) NOT NULL,
    group_type        VARCHAR(30) NOT NULL
                      CHECK (group_type IN ('personal', 'role', 'role_and_subordinates', 'public')),
    related_role_id   UUID REFERENCES iam.user_roles(id) ON DELETE CASCADE,
    related_user_id   UUID REFERENCES iam.users(id) ON DELETE CASCADE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

- `related_role_id` — NOT NULL for `role` and `role_and_subordinates`
- `related_user_id` — NOT NULL for `personal`
- Both NULL for `public`

### Group Members Table

```sql
CREATE TABLE iam.group_members (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id        UUID NOT NULL REFERENCES iam.groups(id) ON DELETE CASCADE,
    member_user_id  UUID REFERENCES iam.users(id) ON DELETE CASCADE,
    member_group_id UUID REFERENCES iam.groups(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (
        (member_user_id IS NOT NULL AND member_group_id IS NULL) OR
        (member_user_id IS NULL AND member_group_id IS NOT NULL)
    ),
    UNIQUE (group_id, member_user_id),
    UNIQUE (group_id, member_group_id)
);
```

Supports nested groups: a group can contain users and other groups.
`effective_group_members` (ADR-0012) flattens nested membership into a flat list.

### Auto-generation Lifecycle

| Event | Action |
|-------|--------|
| User created | Auto-create Group type=`personal` with `related_user_id`, add user as member |
| User deleted | Cascade deletion of personal group (ON DELETE CASCADE) |
| UserRole created | Auto-create Group type=`role` + Group type=`role_and_subordinates` with `related_role_id` |
| UserRole deleted | Cascade deletion of related groups |
| User assigned to role | Add to `role` group of this role. Add to all `role_and_subordinates` groups of ancestor roles |
| User removed from role | Remove from corresponding auto-generated groups |
| Role parent_id changed | Recalculate membership of `role_and_subordinates` groups for all affected roles |

All changes → outbox event → recalculation of `effective_group_members` (ADR-0012).

### Unified Grantee in Share Tables

The share table contains only `grantee_id` — always a group_id:

```sql
CREATE TABLE obj_{name}__share (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    record_id       UUID NOT NULL REFERENCES obj_{name}(id) ON DELETE CASCADE,
    grantee_id      UUID NOT NULL REFERENCES iam.groups(id) ON DELETE CASCADE,
    access_level    INT NOT NULL DEFAULT 1,
    reason          VARCHAR(30) NOT NULL
                    CHECK (reason IN ('owner', 'sharing_rule', 'territory', 'manual')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (record_id, grantee_id, reason)
);
```

- Manual share to a specific user → grant to their personal group
- Sharing rule for a role → grant to role/role_and_subordinates group
- Always one resolution path through `effective_group_members`

### Unified WHERE Clause in the SOQL Engine

```sql
-- RLS: share table check
t.id IN (
    SELECT s.record_id
    FROM obj_{name}__share s
    WHERE s.grantee_id IN (
        SELECT group_id
        FROM security.effective_group_members
        WHERE user_id = :user_id
    )
)
```

One path, one cache, one JOIN.

### Sharing Rules — Source/Target via Groups

```sql
-- Sharing rule references groups
source_group_id  UUID NOT NULL REFERENCES iam.groups(id),
target_group_id  UUID NOT NULL REFERENCES iam.groups(id),
```

Instead of `source_type + source_id` / `target_type + target_id` — direct FKs to groups.
"Records of users in the Sales role" → source = Group type=`role` where related_role_id = Sales.
"Accessible to Support role and subordinates" → target = Group type=`role_and_subordinates` where related_role_id = Support.

## Consequences

- Every user and every role automatically receive associated groups
- Share tables have a unified grantee type (group_id), no polymorphic dispatch
- `effective_group_members` is the sole cache for RLS resolution
- Sharing rules reference groups directly (FK), no enum source_type/target_type
- Auto-generation is managed through outbox events
- Public groups are created and populated manually through the admin UI
- Territory-based groups are added in Phase N without changing the model
