# ADR-0022: Object View — Bounded Context Adapter via Role-based UI

**Status:** Accepted
**Date:** 2026-02-15
**Participants:** @roman_myakotin

## Context

### Problem: same data — different contexts

The platform is metadata-driven: a single object (e.g. `Order`) serves different business roles. Each role operates within its own **bounded context** (DDD term) — with its own mental model, set of fields, actions, and related objects:

| Aspect | Sales Rep | Warehouse Worker | Manager |
|--------|-----------|------------------|---------|
| Focus | Client, deal, sale | Shipping, warehouse, movement | Revenue, margin, conversion |
| Fields | client_name, products, discount, total_amount | warehouse, shipping_status, tracking, packages | margin, cost_price, revenue, conversion |
| Actions | send_proposal, create_task | mark_shipped, print_label | export_report, reassign |
| Related | Activities, Files | InventoryMovements | Reports, Subordinate deals |
| Sidebar | Accounts, Contacts, Deals, Tasks | Orders, Warehouses, Shipments | Dashboard, Reports, Users |

Today OLS/FLS/RLS (ADR-0009..0012) control **data access**, but not **presentation**:

| Layer | What it decides | What is missing |
|-------|----------------|-----------------|
| OLS | Which objects a profile can see | Which sections/actions to show |
| FLS | Which fields are accessible | In what order, how to group them |
| RLS | Which records are visible | Which related lists to show |

Result: all users see the same form with all available (per FLS) fields in the `sort_order` from metadata. This creates:

1. **Cognitive overload** — 30+ fields on a form when a role needs 8-10
2. **No role-specific workflow** — a warehouse worker cannot have a "Ship" button without custom code
3. **Monolithic UI** — one layout for everyone, instead of a bounded context per role
4. **Barrier for CRM+ERP scenarios** — impossible to give a warehouse role an "ERP-like" interface on the same objects

### Relationship with ADR-0019

ADR-0019 defined Object View as the 4th subsystem of declarative business logic with a three-level cascade `Metadata -> Object View -> Layout`. However, ADR-0019 focused on cascade semantics (additive validation, replace defaults) and did not detail:

- Binding Object View to profile/role (bounded context adapter)
- Structure of sections, actions, related lists
- Sidebar per profile
- Dashboard per role
- Fallback mechanism when no view exists

This ADR details Object View as a **full-fledged bounded context adapter**, transforming unified data into role-specific UI without code duplication.

### Industry Context

| Platform | Mechanism | Binding |
|----------|-----------|---------|
| **Salesforce** | Page Layouts + Record Types + App | Profile + Record Type |
| **Dynamics 365** | Forms + Views + App Modules | Security Role + App |
| **ServiceNow** | UI Policies + UI Actions + Views | Role |
| **HubSpot** | Record Customization + Views | Team + Pipeline |

All enterprise platforms solve the bounded context challenge by binding presentation to a role/profile. This is not an optional feature — it is the foundation of scalable UI.

## Considered Options

### Option A — Hardcoded views per role in frontend

Separate Vue components for each role: `OrderSalesView.vue`, `OrderWarehouseView.vue`.

**Pros:**
- Quick implementation for 2-3 roles
- Full freedom in layout

**Cons:**
- Does not scale: N objects x M roles = NxM components
- Contradicts the metadata-driven architecture
- Custom objects do not get role-based views
- Code is duplicated across views

### Option B — Object View as metadata-driven configuration (chosen)

Object View — a JSON configuration in metadata, bound to `(object, profile)`. The frontend renders UI based on the configuration. Fallback: no Object View -> all FLS-accessible fields in `sort_order`.

**Pros:**
- Scales to any number of objects and roles
- Administrator configures through UI without code
- Custom objects get role-based views automatically
- Single renderer — one component handles any Object View
- Fits into the three-level cascade of ADR-0019
- Inherits security: Object View cannot show a field forbidden by FLS

**Cons:**
- Requires new metadata storage + Admin UI
- Declarative configuration is limited — complex customizations are impossible without code
- Additional metadata request when loading a form

### Option C — Layout Builder (drag-and-drop)

Visual form builder with drag-and-drop (like Salesforce Lightning App Builder).

**Pros:**
- Maximum flexibility for the administrator
- Visual configuration without knowledge of JSON/configs

**Cons:**
- Enormous implementation complexity (6+ months)
- Not needed at the current stage (80/20: JSON config covers 90% of cases)
- Can be added on top of Option B later (Builder = visual editor for the same JSON)

### Option D — Profile-specific CSS/visibility rules

Hide/show elements through CSS classes or visibility rules on the frontend.

**Pros:**
- Minimal backend changes
- Quick implementation

**Cons:**
- Security through obscurity — data still arrives in the payload
- No grouping of fields into sections
- No role-specific actions and related lists
- Does not scale

## Decision

**Option B chosen: Object View as metadata-driven configuration bound to profile.**

### Conceptual Model

```
+-------------------------------------------------------------+
|                     User logs in                              |
|                           |                                  |
|                     JWT -> Profile + Role                      |
|                           |                                  |
|              +------------+----------------+                 |
|              v            v                v                  |
|         +---------+  +---------+   +--------------+         |
|         |   OLS   |  | Object  |   |   Dashboard  |         |
|         | filter  |  |  View   |   |  per profile |         |
|         | sidebar |  | resolve |   |              |         |
|         +----+----+  +----+----+   +------+-------+         |
|              |            |               |                  |
|              v            v               v                  |
|         Sidebar      Record Form      Home Page              |
|        (only         (sections,       (role                  |
|         accessible   order,           widgets)               |
|         objects)     actions)                                 |
+-------------------------------------------------------------+
```

### Object View Structure

Object View — a record in metadata describing how to display an object for a specific profile:

```
metadata.object_views
+-- id               UUID PK
+-- object_id        FK -> object_definitions.id
+-- profile_id       FK -> iam.profiles.id (nullable — default view)
+-- api_name         VARCHAR UNIQUE (e.g. "order_sales", "order_warehouse")
+-- label            VARCHAR
+-- description      TEXT
+-- is_default       BOOLEAN (fallback view for a profile without a specific view)
+-- config           JSONB (see below)
+-- created_at       TIMESTAMPTZ
+-- updated_at       TIMESTAMPTZ

UNIQUE(object_id, profile_id)  — one view per (object, profile) pair
```

### Config JSON Schema

```jsonc
{
  // Form sections — field grouping
  "sections": [
    {
      "key": "client_info",
      "label": "Client Information",
      "columns": 2,                    // 1 or 2 columns
      "collapsed": false,              // collapsed by default
      "fields": [
        "client_name",                 // field api_name
        "contact_phone",
        "deal"                         // reference field
      ]
    },
    {
      "key": "products",
      "label": "Products",
      "columns": 1,
      "fields": ["products", "total_amount", "discount"]
    }
  ],

  // Highlight panel — key fields at the top of the record card (Compact Layout)
  "highlight_fields": ["order_number", "status", "total_amount"],

  // Actions (buttons) on the record card
  "actions": [
    {
      "key": "send_proposal",
      "label": "Send Proposal",
      "type": "primary",                // primary | secondary | danger
      "icon": "mail",
      "visibility_expr": "record.status == 'draft'"  // CEL — when to show
    },
    {
      "key": "mark_shipped",
      "label": "Ship",
      "type": "primary",
      "icon": "truck",
      "visibility_expr": "record.status == 'confirmed'"
    }
  ],

  // Related Lists — child objects at the bottom of the record card
  "related_lists": [
    {
      "object": "Activity",
      "label": "Activities",
      "fields": ["subject", "type", "due_date", "status"],
      "filter": "WhatId = :recordId",
      "sort": "due_date DESC",
      "limit": 10
    }
  ],

  // List View — which columns to show in the list table
  "list_fields": ["order_number", "client_name", "status", "total_amount", "created_at"],

  // Default sort in the list
  "list_default_sort": "created_at DESC",

  // Default filters in the list
  "list_default_filter": "owner_id = :currentUserId"
}
```

### Resolution Rules

When opening a record of object `X` by a user with profile `P`:

```
1. Look for object_views WHERE object_id = X AND profile_id = P
   -> Found? Use it.

2. Look for object_views WHERE object_id = X AND is_default = true
   -> Found? Use it.

3. Fallback: auto-generate from metadata
   -> sections: one section "Details" with all FLS-accessible fields
   -> highlight_fields: first 3 fields
   -> actions: standard (Save, Delete)
   -> related_lists: all child objects (composition/association)
   -> list_fields: first 5 fields
```

The fallback guarantees that **the system works without a single Object View** — current behavior is preserved. Object View is an optional enhancement.

### Interaction with Security

Object View **does not expand** access — it only **narrows the presentation**:

```
Visible fields = Object View fields ∩ FLS-accessible fields
```

If Object View includes a field forbidden by FLS — the field is not displayed (FLS wins).
If FLS allows a field but Object View does not include it — the field is not displayed (View narrows).

```
+------------------------------------------+
|             FLS-accessible fields         |
|  +------------------------------------+  |
|  |     Object View fields            |  |
|  |  +--------------------------+     |  |
|  |  |  Displayed fields       |     |  |
|  |  |  (intersection)         |     |  |
|  |  +--------------------------+     |  |
|  +------------------------------------+  |
+------------------------------------------+
```

Actions undergo an analogous check:
- `send_proposal` requires OLS Update on Order -> if absent — button is hidden
- `delete` requires OLS Delete -> if absent — button is hidden

### Integration with ADR-0019 Cascade

Object View occupies the second level of the cascade:

```
Metadata (base)
   | additive validation, inherit defaults
Object View (bounded context)        <- THIS ADR
   | additive validation, replace defaults, override visibility
Layout (presentation, future)
```

| Aspect | Metadata -> Object View | Mechanism |
|--------|------------------------|----------|
| **Validation Rules** | Additive (AND) | OV adds rules, does not remove metadata-level ones |
| **Default Expressions** | Replace | OV can override the default for a field |
| **Field visibility** | Restrict | OV shows a subset of fields from metadata |
| **Actions** | Define | OV defines available actions |
| **Related Lists** | Define | OV defines child objects for display |

### Sidebar per Profile

OLS already filters objects by profile. Object View supplements:

```
metadata.profile_navigation
+-- id               UUID PK
+-- profile_id       FK -> iam.profiles.id
+-- config           JSONB
+-- created_at       TIMESTAMPTZ
+-- updated_at       TIMESTAMPTZ

config = {
  "groups": [
    {
      "label": "Sales",
      "items": ["Account", "Contact", "Opportunity"]  // object api_names
    },
    {
      "label": "Documents",
      "items": ["Order", "Contract", "Quote"]
    }
  ]
}
```

Fallback: no record in `profile_navigation` -> sidebar from OLS-accessible objects in alphabetical order (current behavior).

### Dashboard per Profile

Home page adapts to the profile:

```
metadata.profile_dashboards
+-- id               UUID PK
+-- profile_id       FK -> iam.profiles.id
+-- config           JSONB
+-- created_at       TIMESTAMPTZ
+-- updated_at       TIMESTAMPTZ

config = {
  "widgets": [
    {
      "type": "list",                      // list | chart | metric | calendar
      "label": "My Open Tasks",
      "query": "SELECT Id, Subject, DueDate FROM Task WHERE OwnerId = :currentUserId AND Status != 'Completed' ORDER BY DueDate LIMIT 10",
      "size": "half"                        // full | half | third
    },
    {
      "type": "metric",
      "label": "Deals This Month",
      "query": "SELECT COUNT(Id) FROM Opportunity WHERE CreatedDate = THIS_MONTH",
      "size": "third"
    }
  ]
}
```

Fallback: no dashboard config -> standard dashboard with recent items and tasks.

### Example: one object — three bounded contexts

**Order for Sales Rep (Profile: "Sales"):**
```jsonc
{
  "sections": [
    { "key": "client", "label": "Client", "fields": ["client_name", "contact_phone", "deal"] },
    { "key": "products", "label": "Products", "fields": ["products", "total_amount", "discount"] },
    { "key": "delivery", "label": "Delivery", "fields": ["shipping_status", "delivery_date"] }
  ],
  "highlight_fields": ["order_number", "client_name", "total_amount"],
  "actions": [
    { "key": "send_proposal", "label": "Send Proposal", "type": "primary" }
  ],
  "related_lists": [
    { "object": "Activity", "label": "Activities" }
  ],
  "list_fields": ["order_number", "client_name", "status", "total_amount"]
}
```

**Order for Warehouse Worker (Profile: "Warehouse"):**
```jsonc
{
  "sections": [
    { "key": "order", "label": "Order", "fields": ["order_number", "client_name"] },
    { "key": "shipping", "label": "Shipping", "fields": ["warehouse", "products", "shipping_status", "tracking"] },
    { "key": "dimensions", "label": "Weight and Dimensions", "fields": ["total_weight", "packages_count"] }
  ],
  "highlight_fields": ["order_number", "shipping_status", "warehouse"],
  "actions": [
    { "key": "mark_shipped", "label": "Ship", "type": "primary", "visibility_expr": "record.status == 'confirmed'" },
    { "key": "print_label", "label": "Print Label", "type": "secondary" }
  ],
  "related_lists": [
    { "object": "InventoryMovement", "label": "Inventory Movements" }
  ],
  "list_fields": ["order_number", "shipping_status", "warehouse", "created_at"]
}
```

**Order for Manager (Profile: "Manager"):**
```jsonc
{
  "sections": [
    { "key": "overview", "label": "Overview", "fields": ["order_number", "client_name", "status", "total_amount"] },
    { "key": "financials", "label": "Financials", "fields": ["cost_price", "margin", "revenue", "discount"] },
    { "key": "execution", "label": "Execution", "fields": ["warehouse", "shipping_status", "delivery_date"] }
  ],
  "highlight_fields": ["order_number", "total_amount", "margin"],
  "actions": [
    { "key": "reassign", "label": "Reassign", "type": "secondary" },
    { "key": "export", "label": "Export", "type": "secondary" }
  ],
  "related_lists": [
    { "object": "Activity", "label": "Activities" },
    { "object": "AuditLog", "label": "Change History" }
  ],
  "list_fields": ["order_number", "client_name", "total_amount", "margin", "status"]
}
```

Three profiles, one URL `/app/Order/123` — three different interfaces. Without a single line of hardcoded logic.

### API

```
GET  /api/v1/describe/:objectName          — includes resolved Object View for the current profile
GET  /api/v1/admin/object-views            — list all Object Views (admin)
POST /api/v1/admin/object-views            — create Object View
GET  /api/v1/admin/object-views/:id        — get Object View
PUT  /api/v1/admin/object-views/:id        — update Object View
DELETE /api/v1/admin/object-views/:id      — delete Object View
GET  /api/v1/admin/profile-navigation/:id  — profile navigation
PUT  /api/v1/admin/profile-navigation/:id  — update navigation
GET  /api/v1/admin/profile-dashboards/:id  — profile dashboard
PUT  /api/v1/admin/profile-dashboards/:id  — update dashboard
```

Describe API is extended: if an Object View exists for the current profile — the response includes `view` with sections, actions, related lists. The frontend uses `view` for rendering instead of a flat list of fields.

### Storage

Three tables in the `metadata` schema:

- `metadata.object_views` — form/list configuration per (object, profile)
- `metadata.profile_navigation` — sidebar per profile
- `metadata.profile_dashboards` — home page per profile

All configurations are stored in JSONB — flexibility without migrations when extending the schema.

### Implementation Roadmap

```
Phase 9a: Object View Core                    Phase 9b: Navigation + Dashboard
------------------------------------          ----------------------------------
- metadata.object_views table                  - metadata.profile_navigation table
- Admin CRUD API + UI                          - metadata.profile_dashboards table
- Describe API extension                       - Admin UI for navigation/dashboard
- Frontend: render by Object View              - Sidebar per profile
- Fallback logic                               - Home dashboard per profile
- FLS intersection                             - Widget types: list, metric
- Actions with visibility_expr                 - Chart widgets (Phase 15 dependency)
```

## Consequences

### Positive

- **Bounded context without duplication** — one object, N presentations, zero code per view
- **Graceful degradation** — the system works without Object Views (fallback = current behavior)
- **Security-first** — Object View narrows but does not expand access (FLS intersection)
- **Administrator configures, not developer** — Admin CRUD UI for Object Views
- **CRM+ERP without ERP** — a warehouse role gets an "ERP-like" interface through Object View
- **Fits into ADR-0019 cascade** — validation additive, defaults replace, visibility restrict
- **Extensibility** — Layout Builder (drag-and-drop) can be added on top as a visual editor
- **App Templates** can include Object Views per profile — out-of-the-box role-specific UI

### Negative

- Additional metadata request when loading a record (cached on the frontend)
- Complexity: Object View configuration can be non-trivial for an inexperienced admin
- Actions are currently declarative only — actual logic requires Automation Rules (Phase 13)
- Dashboard widgets with SOQL — potential performance concern with complex queries (addressed through SOQL query limits)

### Related ADRs

- **ADR-0009..0012** — Security layers (OLS/FLS/RLS): Object View is built on top, does not bypass
- **ADR-0019** — Declarative business logic: Object View = second level of the cascade
- **ADR-0020** — DML Pipeline: Object View can add validation rules (additive) and override defaults (replace)
- **ADR-0010** — Permission model: Profile = key binding for Object View
- **ADR-0018** — App Templates: can include Object View definitions
- **ADR-0027** — Layout + Form: Layout defines presentation (HOW) on top of Object View (WHAT). Form = computed merge of OV + Layout for the frontend
