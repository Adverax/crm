# ADR-0022: Object View — Bounded Context Adapter via Role-based UI

**Status:** Accepted
**Date:** 2026-02-15
**Participants:** @roman_myakotin

## Amendment: Read/Write Config Split (2026-02-21)

### Motivation

The original flat OVConfig mixed read and write concerns into a single level. Fields like `queries`, `computed`, `list_fields`, and `highlight_fields` (read-only concerns) lived alongside `validation`, `defaults`, and `mutations` (write concerns). This caused:

1. **Cognitive overload** — administrators editing OV config had to mentally separate "what affects display" from "what affects data entry", with no structural guidance.
2. **Ambiguous API contract** — the frontend received a single config blob and had to decide which parts apply to record viewing vs. record creation/editing.
3. **Unnecessary payload on read-only objects** — objects without create/update operations (e.g. audit logs, reports) still carried empty write-side fields.

### Changes

1. **Split OVConfig into `read` / `write` sub-objects.** All display, query, and presentation concerns move under `read` (fields, actions, queries, computed); all mutation, validation, and default concerns move under `write`.

2. **Rename `virtual_fields` to `computed` in read context (`read.computed`).** The term "virtual fields" was ambiguous — it conflicted with the write-side `computed` (fields evaluated on save). In the read context, these are display-time computed values, now consistently named `OVReadComputed`.

3. **`write` is optional (pointer/nullable).** When an object view describes a read-only context (e.g. a report view, an audit log), `write` is omitted entirely. The API returns `null` / omits the key, and the frontend disables create/edit forms.

4. **`write.fields` is nullable.** When `write.fields` is `null`, the frontend falls back to `read.fields` for form rendering. This avoids duplicating the field list when read and write use the same fields. When `write.fields` is explicitly set, it overrides the read field list for create/edit forms.

5. **Sections removed from OV config.** Per ADR-0027, sections and layout concerns belong to Layout, not Object View. OV defines WHAT data is relevant; Layout defines HOW it is presented (sections, columns, collapsed state). This amendment confirms their removal from the OV config schema.

### New Config JSON Format

```jsonc
{
  // READ side: presentation + display-time data
  "read": {
    "fields": ["client_name", "contact_phone", "deal", "products", "total_amount", "discount"],
    "actions": [
      { "key": "send_proposal", "label": "Send Proposal", "type": "primary", "icon": "mail", "visibility_expr": "record.status == 'draft'" }
    ],
    "queries": [
      { "name": "recent_activities", "soql": "SELECT Id, Subject FROM Activity WHERE WhatId = :recordId LIMIT 5", "when": "record.status != 'cancelled'" }
    ],
    "computed": [
      { "name": "total_with_tax", "type": "float", "expr": "record.amount * (1 + record.tax_rate / 100.0)", "when": "has(record.amount)" }
    ]
  },

  // WRITE side: optional (omitted = read-only object view)
  "write": {
    "fields": null,          // null = fallback to read.fields
    "validation": [
      { "expr": "record.amount > 0", "message": "Amount must be positive", "code": "invalid_amount", "severity": "error" }
    ],
    "defaults": [
      { "field": "status", "expr": "'draft'", "on": "create" }
    ],
    "computed": [
      { "field": "total_with_tax", "expr": "record.amount * (1 + record.tax_rate / 100.0)" }
    ],
    "mutations": [
      { "dml": "INSERT INTO LineItem ...", "when": "record.status == 'confirmed'" }
    ]
  }
}
```

When `write` is omitted, the object view is read-only — no create/edit operations are available in the UI.

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

The config has two sub-objects: **`read`** (presentation + display-time data) and **`write`** (mutation-time data, optional). Sections, highlight fields, related lists, and list configuration are not part of OV config — they belong to Layout (ADR-0027). Actions are part of `read` since they are tied to the read context (record detail page).

#### `read` — display-time data

Per ADR-0019, Object View is a full bounded context adapter. The `read` sub-object defines what data to display for this context.

```jsonc
{
  "read": {
    // Fields visible on the record detail page
    "fields": ["client_name", "contact_phone", "deal", "products", "total_amount", "discount"],

    // Actions available on the record detail page
    "actions": [
      {
        "key": "send_proposal",
        "label": "Send Proposal",
        "type": "primary",                            // primary | secondary | danger
        "icon": "mail",
        "visibility_expr": "record.status == 'draft'" // CEL — conditional visibility
      }
    ],

    // Named SOQL queries scoped to this view
    "queries": [
      {
        "name": "recent_activities",
        "soql": "SELECT Id, Subject, DueDate FROM Activity WHERE WhatId = :recordId ORDER BY DueDate DESC LIMIT 5",
        "when": "record.status != 'cancelled'"     // CEL — conditional execution
      }
    ],

    // Computed fields — display-time expressions (not stored in DB)
    "computed": [
      {
        "name": "total_with_tax",
        "type": "float",                             // string | int | float | bool | timestamp
        "expr": "record.amount * (1 + record.tax_rate / 100.0)",
        "when": "has(record.amount)"
      }
    ]
  }
}
```

#### `write` — mutation-time data (optional)

The `write` sub-object is optional (pointer/nullable). When omitted or `null`, the object view is read-only — no create/edit operations are available in the UI.

`write.fields` is nullable: when `null`, the frontend falls back to `read.fields` for form rendering. When explicitly set, it overrides the read field list for create/edit forms.

```jsonc
{
  "write": {
    // Fields available in create/edit forms (null = fallback to read.fields)
    "fields": null,

    // DML operations scoped to this view
    "mutations": [
      {
        "dml": "INSERT INTO LineItem (order_id, product_id, qty) VALUES (:recordId, :item.product_id, :item.qty)",
        "foreach": "queries.line_items",             // iterate over query results
        "sync": { "key": "product_id", "value": "item.product_id" },  // sync mapping
        "when": "record.status == 'confirmed'"
      }
    ],

    // View-scoped validation rules (additive with metadata-level rules per ADR-0019)
    "validation": [
      {
        "expr": "record.amount > 0",
        "message": "Amount must be positive",
        "code": "invalid_amount",
        "severity": "error",                         // error | warning
        "when": "has(record.amount)"
      }
    ],

    // View-scoped defaults (replace metadata-level defaults per ADR-0019)
    "defaults": [
      {
        "field": "status",
        "expr": "'draft'",
        "on": "create",                              // create | update | create,update
        "when": ""
      }
    ],

    // Computed fields — expressions evaluated on save
    "computed": [
      {
        "field": "total_with_tax",
        "expr": "record.amount * (1 + record.tax_rate / 100.0)"
      }
    ]
  }
}
```

All fields within `read` and `write` are optional (`omitempty` in Go, `?? []` fallback on frontend). Existing Object View records without `read`/`write` sub-objects continue to work unchanged via migration or fallback logic.

### Resolution Rules

When opening a record of object `X` by a user with profile `P`:

```
1. Look for object_views WHERE object_id = X AND profile_id = P
   -> Found? Use it.

2. Look for object_views WHERE object_id = X AND is_default = true
   -> Found? Use it.

3. Fallback: auto-generate from metadata
   -> read.fields: all FLS-accessible fields
   -> read.actions: standard (Save, Delete)
   -> write: non-null with fields=null (fallback to read.fields)
   -> Layout (ADR-0027): highlight_fields, related_lists, list_fields
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

> **Note:** These examples show the complete bounded context including both OV config (`read`/`write`) and Layout properties (`highlight_fields`, `related_lists`, `list_fields`) defined in ADR-0027. In the database, Layout properties are stored in `metadata.layouts`, not in the OV config JSONB.

**Order for Sales Rep (Profile: "Sales"):**
```jsonc
{
  // OV Config (metadata.object_views.config)
  "read": {
    "fields": ["client_name", "contact_phone", "deal", "products", "total_amount", "discount", "shipping_status", "delivery_date"],
    "actions": [
      { "key": "send_proposal", "label": "Send Proposal", "type": "primary" }
    ],
    "queries": [
      { "name": "client_history", "soql": "SELECT Id, Name, Amount FROM Order WHERE ClientId = :record.client_id AND Id != :recordId ORDER BY CreatedDate DESC LIMIT 5" }
    ]
  },
  "write": {
    "fields": null,
    "validation": [
      { "expr": "record.discount <= 20", "message": "Discount cannot exceed 20% for sales reps", "severity": "error" }
    ],
    "defaults": [
      { "field": "status", "expr": "'draft'", "on": "create" }
    ]
  }
  // Layout (ADR-0027, metadata.layouts):
  // highlight_fields: ["order_number", "client_name", "total_amount"]
  // related_lists: [{ "object": "Activity", "label": "Activities" }]
  // list_fields: ["order_number", "client_name", "status", "total_amount"]
}
```

**Order for Warehouse Worker (Profile: "Warehouse"):**
```jsonc
{
  // OV Config (metadata.object_views.config)
  "read": {
    "fields": ["order_number", "client_name", "warehouse", "products", "shipping_status", "tracking", "total_weight", "packages_count"],
    "actions": [
      { "key": "mark_shipped", "label": "Ship", "type": "primary", "visibility_expr": "record.status == 'confirmed'" },
      { "key": "print_label", "label": "Print Label", "type": "secondary" }
    ],
    "computed": [
      { "name": "is_oversized", "type": "bool", "expr": "record.total_weight > 50" }
    ]
  },
  "write": {
    "fields": ["shipping_status", "tracking", "warehouse", "packages_count"],
    "validation": [
      { "expr": "record.tracking != ''", "message": "Tracking number required before shipping", "severity": "error", "when": "record.shipping_status == 'shipping'" }
    ]
  }
  // Layout (ADR-0027, metadata.layouts):
  // highlight_fields: ["order_number", "shipping_status", "warehouse"]
  // related_lists: [{ "object": "InventoryMovement", "label": "Inventory Movements" }]
  // list_fields: ["order_number", "shipping_status", "warehouse", "created_at"]
}
```

**Order for Manager (Profile: "Manager"):**
```jsonc
{
  // OV Config (metadata.object_views.config)
  "read": {
    "fields": ["order_number", "client_name", "status", "total_amount", "cost_price", "margin", "revenue", "discount", "warehouse", "shipping_status", "delivery_date"],
    "actions": [
      { "key": "reassign", "label": "Reassign", "type": "secondary" },
      { "key": "export", "label": "Export", "type": "secondary" }
    ],
    "computed": [
      { "name": "margin_pct", "type": "float", "expr": "record.margin / record.total_amount * 100" }
    ]
  },
  "write": {
    "fields": null,
    "computed": [
      { "field": "revenue", "expr": "record.total_amount - record.cost_price" }
    ]
  }
  // Layout (ADR-0027, metadata.layouts):
  // highlight_fields: ["order_number", "total_amount", "margin"]
  // related_lists: [{ "object": "Activity" }, { "object": "AuditLog" }]
  // list_fields: ["order_number", "client_name", "total_amount", "margin", "status"]
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

Describe API is extended: if an Object View exists for the current profile — the response includes `view` with `read`/`write` sub-objects. The frontend uses `view` for rendering instead of a flat list of fields.

### Storage

Three tables in the `metadata` schema:

- `metadata.object_views` — form/list configuration per (object, profile)
- `metadata.profile_navigation` — sidebar per profile
- `metadata.profile_dashboards` — home page per profile

All configurations are stored in JSONB — flexibility without migrations when extending the schema.

### Implementation Roadmap

```
Phase 9a: Object View Core ✅                 Phase 9b: Navigation + Dashboard
------------------------------------          ----------------------------------
- metadata.object_views table                  - metadata.profile_navigation table
- Admin CRUD API + UI                          - metadata.profile_dashboards table
- Describe API extension                       - Admin UI for navigation/dashboard
- Frontend: render by Object View              - Sidebar per profile
- Fallback logic                               - Home dashboard per profile
- FLS intersection                             - Widget types: list, metric
- Actions with visibility_expr                 - Chart widgets (Phase 15 dependency)
- Config (ADR-0019):
  read (fields, actions, queries, computed),
  write (fields, mutations, validation,
  defaults, computed)
- Admin UI: visual constructor
```

> **Note:** Data contract is stored in the JSONB config and editable through the Admin UI. Runtime execution (query executor, mutation executor) is deferred to a future phase. Phase 9a covers config storage and admin-time editing only.

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
