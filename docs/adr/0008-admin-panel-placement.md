# ADR-0008: Admin Panel Placement in the Monorepo

**Status:** Accepted
**Date:** 2026-02-08
**Participants:** @roman_myakotin

## Context

The CRM platform requires an administrative interface (analogous to Salesforce Setup) for:
- Managing metadata (objects, fields, layouts)
- Managing security (profiles, roles, OLS/FLS/RLS)
- Managing users and the organization
- Platform monitoring and configuration

The admin panel and the user-facing CRM interface are different functional areas with different
target audiences (administrators vs operators), but they work with a single backend and a single
authorization model (JWT + profiles).

It is necessary to determine where in the monorepo to place the admin panel code.

## Options Considered

### Option A: Routes Inside `web/` (chosen)

The admin panel is placed as a set of `/admin/*` routes inside the existing Vue application `web/`.
Separate layout, lazy-loaded views, guard based on user profile.

```
web/src/
├── layouts/
│   ├── DefaultLayout.vue      ← CRM layout
│   └── AdminLayout.vue        ← Admin layout
├── views/
│   ├── admin/                 ← admin panel
│   │   ├── metadata/
│   │   ├── security/
│   │   └── users/
│   ├── contacts/              ← CRM
│   ├── deals/
│   └── ...
```

**Pros:**
- One build, one deploy, one domain — minimal infrastructure overhead
- Shared components, stores, API client, types — no duplication
- Shared auth flow (JWT, single profile-based guard)
- Lazy-loading of routes solves the bundle size problem
- Easy migration: if the admin panel grows, `views/admin/` can be easily extracted into a separate application

**Cons:**
- Cannot deploy the admin panel separately from the CRM
- Cannot restrict access at the infrastructure level (separate domain, VPN)
- Over time `web/` may become large

### Option B: Separate Application `admin/`

A completely separate Vue application at the monorepo root.

```
crm/
├── web/        ← CRM frontend
├── admin/      ← Admin frontend (separate package.json)
```

**Pros:** full isolation, independent deployment, access can be restricted at the infra level.
**Cons:** duplication of dependencies, types, API client; two CI pipelines; two builds.

### Option C: npm Workspaces (frontend monorepo)

```
web/
├── apps/crm/       ← CRM
├── apps/admin/     ← Admin
├── packages/shared/ ← shared types, UI, API client
```

**Pros:** isolation + reuse via shared packages.
**Cons:** significant infrastructure overhead (workspace config, shared package builds),
justified only for large teams.

## Decision

We adopt **Option A** — routes inside `web/`.

### Rationale

1. **MVP pragmatism** — at the current stage (one developer, Phase 1), the overhead of a separate
   application is not justified
2. **Single backend** — one API server (`cmd/api`), one JWT — a single frontend is logical
3. **Access control via profiles** — access to `/admin/*` is controlled by a Vue Router guard
   based on the user profile, which aligns with the OLS security model (Phase 2)
4. **Lazy-loading** — admin modules are loaded only when navigating to `/admin/*`,
   they do not affect the initial bundle size for regular users

### Conventions

- All admin routes: `/admin/*`
- Separate layout: `AdminLayout.vue`
- Views: `web/src/views/admin/{module}/`
- Admin-specific components: `web/src/components/admin/`
- Router guard: profile/role check before rendering admin routes

### Migration Criteria to Option C

Transitioning to workspaces is justified when:
- The admin panel exceeds ~30 views
- A dedicated frontend development team for the admin panel appears
- Independent deployment is required (separate domain, VPN)

## Consequences

- The admin panel resides in `web/src/views/admin/` as part of the main Vue application
- Access is controlled at the route level (guard), not at the infrastructure level
- Shared components and stores are reused without duplication
- As the project grows, migration to a workspace structure is possible without losing code
