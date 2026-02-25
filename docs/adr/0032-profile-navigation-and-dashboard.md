# ADR-0032: Profile Navigation and Dashboard

**Date:** 2026-02-25

**Status:** Accepted

**Participants:** @roman_myakotin

## Context

### Problem: one-size-fits-all navigation and home page

After Phase 9a (Object View, ADR-0022), record forms adapt to the user's
profile — a Sales Rep sees different fields and actions than a Warehouse Worker.
However, the **sidebar** and **home page** are still identical for everyone:

| Component | Current behavior | Desired behavior |
|-----------|-----------------|-----------------|
| Sidebar | Flat alphabetical list, OLS-filtered | Grouped, ordered, profile-specific |
| Home page | Empty (`<RouterView />` placeholder) | Dashboard with profile-specific widgets |

This creates friction:
1. A Sales Rep sees "Application", "Candidate", "Interview" in sidebar —
   objects irrelevant to their role (but accessible via OLS).
2. No home dashboard — users land on an empty page and must click an object.
3. No way for an admin to curate navigation for each department.

### Relationship with ADR-0022

ADR-0022 defined `profile_navigation` and `profile_dashboards` as Phase 9b
deliverables with high-level schema sketches. This ADR details the
implementation decisions: API design, widget execution model, caching,
security, and fallback behavior.

### Industry context

| Platform | Sidebar mechanism | Dashboard mechanism |
|----------|------------------|-------------------|
| **Salesforce** | Lightning App → Tab Set per app per profile | Home Page Layout per profile, components |
| **Dynamics 365** | App Module → Site Map (groups + areas) | Dashboard per security role |
| **HubSpot** | Navigation preset per team | Dashboard builder (reports) |

All enterprise CRMs provide profile/role-scoped navigation and dashboards.

## Considered Options

### Sidebar configuration granularity

**Option A: Per-profile JSON config in a single table (Chosen)**
- One row per profile with JSONB config containing groups and items.
- Pros: Simple, consistent with Object View pattern, one query per login.
- Cons: No sharing of navigation configs between profiles.

**Option B: Normalized tables (nav_groups, nav_items)**
- Separate tables for groups and items with FK to profile.
- Pros: Granular queries, reusable groups across profiles.
- Cons: Over-engineered for the current scale; more queries, more joins.

**Option C: App-based navigation (Salesforce model)**
- Users belong to an "App" (Sales Cloud, Service Cloud), each app defines tabs.
- Pros: Familiar to Salesforce users; apps can span multiple profiles.
- Cons: Introduces a new "App" concept, premature abstraction for MVP.

### Dashboard widget execution

**Option A: SOQL-driven widgets (Chosen)**
- Each widget has a SOQL query; backend executes it with full RLS/FLS.
- Pros: Leverages existing SOQL infrastructure; security enforced.
- Cons: Multiple SOQL queries per dashboard load.

**Option B: Pre-computed materialized views**
- Dashboard data computed on schedule and cached.
- Pros: Fast reads, no runtime SOQL.
- Cons: Stale data, complex invalidation, overkill for MVP.

**Option C: Frontend-only widgets**
- Frontend fetches data via existing record list API.
- Pros: No new backend endpoints.
- Cons: Cannot aggregate across objects; limited to existing list API.

### Dashboard storage

**Option A: Single JSONB config per profile (Chosen)**
- Mirrors profile_navigation pattern. One row, one config blob.
- Pros: Simple, consistent, flexible schema evolution.
- Cons: No sharing of dashboard configs.

**Option B: Widget-per-row normalized model**
- Each widget is a separate row with FK to dashboard.
- Pros: Granular CRUD, reusable widgets.
- Cons: More tables, more complexity, premature for MVP.

## Decision

### Schema

Two tables in the `metadata` schema, following the Object View JSONB pattern:

```sql
-- Profile Navigation: sidebar configuration per profile
CREATE TABLE metadata.profile_navigation (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id      UUID NOT NULL REFERENCES iam.profiles(id) ON DELETE CASCADE,
    config          JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (profile_id)
);

-- Profile Dashboard: home page widgets per profile
CREATE TABLE metadata.profile_dashboards (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id      UUID NOT NULL REFERENCES iam.profiles(id) ON DELETE CASCADE,
    config          JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (profile_id)
);
```

Key: `UNIQUE(profile_id)` — one navigation config and one dashboard per profile.
`ON DELETE CASCADE` — deleting a profile removes its navigation and dashboard.

### Navigation config schema

```jsonc
{
  "groups": [
    {
      "key": "sales",                    // unique within config
      "label": "Sales",
      "icon": "briefcase",              // lucide icon name (optional)
      "items": [
        {
          "type": "object",             // object | link | divider
          "object_api_name": "Account"  // for type=object
        },
        {
          "type": "object",
          "object_api_name": "Contact"
        },
        {
          "type": "link",               // custom external/internal link
          "label": "Reports",
          "url": "/admin/reports",
          "icon": "bar-chart-2"
        },
        {
          "type": "divider"             // visual separator
        }
      ]
    }
  ]
}
```

#### Navigation item types

| Type | Description | Required fields |
|------|-------------|----------------|
| `object` | Link to an object list page `/app/{api_name}` | `object_api_name` |
| `link` | Custom URL (internal or external) | `label`, `url` |
| `divider` | Visual separator line | none |

#### OLS intersection

Navigation config references objects by `api_name`. At render time,
each `object` item is intersected with OLS:

```
Visible items = config.items.filter(item =>
    item.type != "object" || OLS.canRead(item.object_api_name)
)
```

If a group becomes empty after OLS filtering — the group is hidden.
This ensures navigation cannot expose objects forbidden by security.

### Dashboard config schema

```jsonc
{
  "widgets": [
    {
      "key": "open_tasks",              // unique within config
      "type": "list",                   // list | metric | link_list
      "label": "My Open Tasks",
      "size": "half",                   // full | half | third
      "query": "SELECT Id, subject, due_date FROM Task WHERE owner_id = :currentUserId AND status != 'Completed' ORDER BY due_date LIMIT 10",
      "columns": ["subject", "due_date"],
      "object_api_name": "Task"         // for row click navigation
    },
    {
      "type": "metric",
      "key": "deals_count",
      "label": "Deals This Month",
      "size": "third",
      "query": "SELECT COUNT(Id) FROM Opportunity WHERE created_at >= THIS_MONTH",
      "format": "number"                // number | currency | percent
    },
    {
      "type": "link_list",
      "key": "quick_actions",
      "label": "Quick Actions",
      "size": "third",
      "links": [
        { "label": "New Account", "url": "/app/Account/new", "icon": "building" },
        { "label": "New Contact", "url": "/app/Contact/new", "icon": "user-plus" }
      ]
    }
  ]
}
```

#### Widget types (Phase 9b scope)

| Type | Data source | Rendering |
|------|------------|-----------|
| `list` | SOQL query | Table with columns, row click → record detail |
| `metric` | SOQL aggregate query (single value) | Large number with label |
| `link_list` | Static links in config | List of clickable links with icons |

Chart widgets (`bar`, `pie`, `line`) are deferred to Phase 17 (Analytics),
as they require a charting library and more complex data transformation.

#### SOQL variables in dashboard queries

| Variable | Resolves to | Example |
|----------|------------|---------|
| `:currentUserId` | Authenticated user's UUID | `WHERE owner_id = :currentUserId` |

Variables are substituted server-side before SOQL execution.
Full RLS/FLS enforcement applies — a widget cannot show data
the user cannot see via normal record access.

### API design

#### Navigation

```
GET    /api/v1/admin/profile-navigation              — list all (admin)
POST   /api/v1/admin/profile-navigation              — create for a profile
GET    /api/v1/admin/profile-navigation/:id           — get by ID
PUT    /api/v1/admin/profile-navigation/:id           — update config
DELETE /api/v1/admin/profile-navigation/:id           — delete

GET    /api/v1/navigation                             — resolved for current user
```

The `/api/v1/navigation` endpoint:
1. Get profile_id from JWT (UserContext).
2. Look up `profile_navigation` for this profile_id.
3. If found: apply OLS intersection, return filtered config.
4. If not found: return fallback (OLS-filtered objects, alphabetical, no groups).

#### Dashboard

```
GET    /api/v1/admin/profile-dashboards               — list all (admin)
POST   /api/v1/admin/profile-dashboards               — create for a profile
GET    /api/v1/admin/profile-dashboards/:id            — get by ID
PUT    /api/v1/admin/profile-dashboards/:id            — update config
DELETE /api/v1/admin/profile-dashboards/:id            — delete

GET    /api/v1/dashboard                               — resolved + executed for current user
```

The `/api/v1/dashboard` endpoint:
1. Get profile_id from JWT.
2. Look up `profile_dashboards` for this profile_id.
3. If found: execute each widget's SOQL query (with RLS/FLS), return data.
4. If not found: return fallback (empty dashboard or default widgets).

Dashboard execution response:

```jsonc
{
  "widgets": [
    {
      "key": "open_tasks",
      "type": "list",
      "label": "My Open Tasks",
      "size": "half",
      "object_api_name": "Task",
      "columns": ["subject", "due_date"],
      "data": {
        "records": [
          { "id": "...", "subject": "Call client", "due_date": "2026-02-26" }
        ],
        "total_count": 5
      }
    },
    {
      "key": "deals_count",
      "type": "metric",
      "label": "Deals This Month",
      "size": "third",
      "format": "number",
      "data": {
        "value": 12
      }
    }
  ]
}
```

### Caching strategy

| Data | Cache layer | Invalidation |
|------|------------|--------------|
| Navigation config | MetadataCache (in-memory) | On admin save (same as Object View) |
| Dashboard config | MetadataCache (in-memory) | On admin save |
| Dashboard widget data | No cache (real-time SOQL) | N/A — always fresh |

Navigation config is loaded once per login and cached on the frontend.
Dashboard data is fetched on each home page visit (real-time).

### Fallback behavior

| Scenario | Navigation fallback | Dashboard fallback |
|----------|--------------------|--------------------|
| No `profile_navigation` row | OLS-filtered alphabetical list (current behavior) | Empty state with "Welcome" message |
| Profile deleted | CASCADE removes config | CASCADE removes config |
| Object in nav deleted | Item filtered out at render time | Widget with deleted object returns empty data |
| SOQL query error in widget | N/A | Widget shows error state, other widgets unaffected |

Fallback guarantees **zero-config operation**: the system works identically
to today if no navigation or dashboard configs are created.

### Validation rules

#### Navigation config validation (on save)

1. `groups` must be an array (can be empty).
2. Each group must have a unique `key`.
3. Each `object` item must reference an existing object `api_name`.
4. Each `link` item must have non-empty `label` and `url`.
5. `url` for links: must start with `/` (internal) or `https://` (external). No `javascript:`.
6. Max 20 groups, max 50 items per group.

#### Dashboard config validation (on save)

1. `widgets` must be an array (can be empty).
2. Each widget must have a unique `key`.
3. `type` must be one of: `list`, `metric`, `link_list`.
4. `list` and `metric` widgets must have a non-empty `query`.
5. `list` widgets must have non-empty `columns` and `object_api_name`.
6. `metric` widgets: `format` must be one of: `number`, `currency`, `percent`.
7. `link_list` widgets must have non-empty `links` array.
8. SOQL query syntax is validated (parsed, not executed) on save.
9. Max 12 widgets per dashboard.

### Security

- Admin endpoints require OLS admin access (same as Object Views).
- Navigation resolution endpoint (`/api/v1/navigation`): any authenticated user.
- Dashboard resolution endpoint (`/api/v1/dashboard`): any authenticated user.
- SOQL queries in widgets execute under the requesting user's security context
  (full RLS/FLS enforcement via existing QueryService).
- Navigation items undergo OLS intersection — no object exposure beyond OLS grants.

### Platform limits

| Parameter | Default | Description |
|-----------|---------|-------------|
| Max groups per navigation | 20 | Navigation groups per profile |
| Max items per group | 50 | Items within a single group |
| Max widgets per dashboard | 12 | Widgets per profile dashboard |
| Dashboard SOQL timeout | 5s | Per-widget query timeout |
| Dashboard total timeout | 15s | Total time for all widget queries |

## Consequences

### Positive

- **Role-specific workspace** — each profile gets curated navigation and dashboard
- **Zero-config fallback** — works identically to today without any configuration
- **Security preserved** — OLS intersection for nav, RLS/FLS for dashboard queries
- **Consistent pattern** — same JSONB-in-metadata approach as Object Views
- **Admin-configurable** — no code changes needed to customize per profile
- **Real-time dashboard data** — SOQL queries execute with current security context

### Negative

- **Multiple SOQL queries per dashboard load** — bounded by 12 widgets and 5s timeout
- **No shared configs** — each profile needs its own navigation and dashboard
- **No chart widgets in Phase 9b** — deferred to Phase 17 (Analytics)
- **SOQL in config** — admin must know SOQL syntax for list/metric widgets

### Related ADRs

- ADR-0022: Object View (Phase 9a, bounded context adapter — this ADR implements Phase 9b)
- ADR-0009..0012: Security layers (OLS/FLS/RLS enforcement)
- ADR-0019: Declarative business logic (three-level cascade)
- ADR-0027: Layout + Form (presentation layer, future)
