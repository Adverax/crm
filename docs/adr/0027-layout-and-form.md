# ADR-0027: Layout + Form — Presentation Layer for Portal

**Status:** Accepted (revised 2026-02-26)
**Date:** 2026-02-15
**Revised:** 2026-02-26
**Participants:** @roman_myakotin

## Context

### Problem: Portal Defines WHAT, but Not HOW

Portal (ADR-0022) solves the bounded context problem: one object — different representations for different profiles. Portal defines **which** sections, fields, actions, and related lists each role sees.

But Portal does not answer presentation questions:

| Question | Portal answers? |
|----------|-------------|
| Which fields does Sales see? | Yes |
| How is the page structured — sidebar, full width? | No |
| How many columns should a section render in? | No |
| What col_span does a specific field have? | No |
| How does the status field look — text or badge? | No |
| Is the discount field required when amount > 10000? | No |
| How does the list view look — column widths, sorting? | No |
| How does the form adapt to mobile devices? | No |
| Is this field visible in view mode but hidden in edit mode? | No |

### Problem: God Object When Extending Portal

If all layout attributes are added to the Portal config, Portal becomes a God Object with two incompatible responsibilities:

- **Bounded context** (WHAT) — business decision about which fields/actions are available
- **Presentation** (HOW) — visual decision about how to render on different devices

These responsibilities change independently. Different people may be responsible for each.

### Problem: One Portal — Multiple Devices and Modes

Desktop, tablet, and mobile require **structurally different** representations. Edit and view modes also differ: view mode uses display components (Badge, Avatar, RelativeTime), edit mode uses input components (Select, DatePicker).

A single Portal cannot contain multiple grid configurations and mode-specific settings without turning into a God Object. Hence the need for a **separate Layout entity** bound to form factor and mode.

### Problem: Page Structure Is Hardcoded

The current record detail page has a hardcoded structure: highlights at top, then sections, then related lists. But different profiles and use cases need different page structures: sidebar with related lists, two-column layout, query widgets alongside field sections.

The page structure must be configurable through the same Layout mechanism.

### Problem: Frontend Needs a Unified Contract

The frontend should not know about Portal and Layout as separate concepts. It needs **one object** (Form) containing everything for rendering. The Describe API should return a ready-made Form.

### Relationship with ADR-0019 and ADR-0022

ADR-0019 defined a three-level cascade:

```
Metadata (base)
   | additive validation, inherit defaults
Portal (bounded context)        <- ADR-0022
   | additive validation, replace defaults, override visibility
Layout (presentation)                <- THIS ADR
   | merge
Form (frontend contract)             <- THIS ADR
```

## Decision

**Layout per Portal + form factor + mode, Form as the computed frontend contract.**

### Three Entities, Three Responsibilities

```
+--------------------------+
|      Portal         |  Stored in metadata.portals
|   (bounded context)      |  Per profile (global api_name)
|                          |
|  WHAT: sections, fields, |
|  actions, related lists, |
|  queries, validation     |
+------------+-------------+
             | 1:N
             v
+--------------------------+
|        Layout            |  Stored in metadata.layouts
|    (presentation)        |  Per (portal, form_factor, mode)
|                          |
|  HOW: page structure,    |
|  grid, col_span, ui_kind,|
|  conditions, list config |
+------------+-------------+
             | resolve + merge
             v
+--------------------------+
|         Form             |  Computed (not stored)
|  (frontend contract)     |  Describe API response
|                          |
|  EVERYTHING: structure + |
|  presentation +          |
|  conditional exprs       |
+--------------------------+

+--------------------------+
|    Shared Layouts        |  Stored in metadata.shared_layouts
|  (reusable fragments)    |  Global, referenced via layout_ref
|                          |
|  DRY: common field,      |
|  section, list configs   |
+--------------------------+
```

### Layout Config Structure

Layout config has two levels:

1. **`root`** — component tree defining page structure (WHAT goes WHERE)
2. **`*_config`** maps — detailed configuration per component (HOW it renders)

```jsonc
{
  // Component tree — page structure
  "root": {
    "type": "grid",
    "columns": 12,
    "children": [
      {
        "type": "group",
        "col_span": 8,
        "children": [
          { "type": "highlights" },
          { "type": "field_section", "key": "basic_info" },
          { "type": "field_section", "key": "address" }
        ]
      },
      {
        "type": "group",
        "col_span": 4,
        "children": [
          { "type": "related_list", "key": "tasks", "limit": 5 },
          { "type": "activity_feed" }
        ]
      },
      {
        "type": "related_list",
        "col_span": 12,
        "key": "contacts"
      }
    ]
  },

  // Section presentation
  "section_config": {
    "basic_info": {
      "columns": 2,
      "collapsed": false,
      "collapsible": true
    },
    "address": {
      "columns": 2,
      "collapsed": true,
      "visibility_expr": "record.status != 'cancelled'"
    }
  },

  // Field presentation
  "field_config": {
    "client_name": {
      "col_span": 2,
      "ui_kind": "lookup",
      "reference": {
        "display_fields": ["name", "email"],
        "search_fields": ["name", "email", "phone"],
        "target": "popup"
      }
    },
    "status": {
      "col_span": 1,
      "ui_kind": { "type": "badge", "color_map": {"active": "green", "closed": "gray"} }
    },
    "discount": {
      "col_span": 1,
      "required_expr": "record.amount > 10000",
      "readonly_expr": "record.status == 'closed'"
    },
    "email": {
      "col_span": 2,
      "ui_kind": "email"
    }
  },

  // List column presentation
  "list_config": {
    "columns": [
      {"field": "order_number", "width": "15%", "sortable": true},
      {"field": "client_name", "width": "30%", "sortable": true},
      {"field": "status", "width": "100px", "align": "center", "ui_kind": "badge"},
      {"field": "total_amount", "width": "15%", "align": "right", "sortable": true},
      {"field": "created_at", "width": "15%", "sortable": true, "sort_dir": "desc"}
    ],
    "search": {
      "fields": ["order_number", "client_name"],
      "placeholder": "Search orders..."
    }
  }
}
```

### Page Structure: GridLayout + GroupLayout

Instead of a separate PageLayout entity, the page structure is defined by the `root` component tree using two container types:

- **GridLayout** (`type: "grid"`) — CSS Grid container with configurable columns. Can nest other grids.
- **GroupLayout** (`type: "group"`) — vertical stack of components within a grid cell.

The app shell (navigation sidebar, breadcrumbs) is fixed. Layout's `root` controls only the **main content area**.

```
App Shell (fixed):
┌────────┬──────────────────────────────────┐
│        │ Breadcrumbs                      │
│  Nav   ├──────────────────────────────────┤
│  bar   │                                  │
│        │   root (GridLayout + children)   │
│        │   ← Layout controls this zone    │
│        │                                  │
└────────┴──────────────────────────────────┘
```

Common page structures are expressed naturally:

```jsonc
// Full width (single column)
{ "type": "grid", "columns": 1, "children": [...] }

// Main + sidebar (70/30)
{ "type": "grid", "columns": 12, "children": [
    { "type": "group", "col_span": 8, "children": [...] },
    { "type": "group", "col_span": 4, "children": [...] }
]}

// Two equal columns
{ "type": "grid", "columns": 2, "children": [
    { "type": "group", "col_span": 1, "children": [...] },
    { "type": "group", "col_span": 1, "children": [...] }
]}
```

**Presets**: common page structures are offered as starting points in the admin UI (not a first-class entity — just pre-filled `root` configs).

**Fallback**: if `root` is not defined, auto-generate a full-width layout: highlights at top, all Portal sections in order, related lists at bottom.

### Component Types

Components are placed inside GridLayout/GroupLayout children:

| type | Description | Data source |
|------|-------------|-------------|
| `highlights` | Key field values at top of page | Portal.read.highlight_fields |
| `field_section` | Group of fields with label | Portal.read.sections[key] |
| `related_list` | Table of child records | Portal.read.related_lists[key] |
| `query_widget` | SOQL query result (table/metric) | Portal.read.queries[name] |
| `actions_bar` | Action buttons | Portal.read.actions |
| `activity_feed` | Activity timeline | Built-in |
| `tabs` | Tab container grouping child components | Container |

Components reference Portal data by key/name. A component that references a non-existent Portal key is silently skipped ( Portal is the source of truth).

The set is extensible — new component types can be added in future phases.

### Form Factor + Mode

Layout is scoped to `(portal_id, form_factor, mode)`:

| Dimension | Values | Purpose |
|-----------|--------|---------|
| form_factor | `desktop`, `tablet`, `mobile` | Device adaptation |
| mode | `edit`, `view` | Interaction context |

**Mode distinction:**

| Aspect | edit | view |
|--------|------|------|
| Fields | Input components (TextField, Select, DatePicker) | Display components (Text, Badge, Link) |
| Actions | Save, Cancel, validation | Edit, Delete, custom |
| Page structure | May differ (e.g., simpler layout for data entry) | May include more related lists, activity |

**Fallback chain:**

```
Requested (form_factor, mode)
  → same form_factor, any mode
  → desktop, same mode
  → desktop, edit (ultimate fallback)
  → auto-generate
```

### Shared Layouts and layout_ref

Reusable layout fragments are stored in `metadata.shared_layouts` and referenced via `layout_ref`:

```jsonc
// Shared layout (stored in metadata.shared_layouts)
{
  "api_name": "customer_lookup",
  "type": "field",
  "config": {
    "ui_kind": "lookup",
    "reference": {
      "display_fields": ["name", "email", "phone"],
      "search_fields": ["name", "email", "phone"],
      "target": "popup"
    }
  }
}

// Usage in Layout field_config — reference + local overrides
{
  "field_config": {
    "customer_id": { "layout_ref": "customer_lookup", "col_span": 2 }
  }
}

// Resolved (what frontend receives in Form):
{
  "field": "customer_id",
  "col_span": 2,                              // from inline override
  "ui_kind": "lookup",                         // from shared layout
  "reference": { "display_fields": ["name", "email", "phone"], ... }  // from shared layout
}
```

**Merge strategy**: shallow merge, inline fields override shared layout fields.

**Shared layout types:**

| type | Reuses | Example |
|------|--------|---------|
| `field` | Field presentation config | Lookup with display_fields, search_fields |
| `section` | Section with fields and grid | Address section (country, city, street, zip) |
| `list` | List column configuration | Standard contact list columns |

**Deletion**: RESTRICT — cannot delete a shared layout while it is referenced.

### UIKind

UIKind specifies the UI component for a field. Supports short form (string) and long form (object with options):

```jsonc
// Short form
{ "ui_kind": "email" }

// Long form with options
{ "ui_kind": { "type": "badge", "color_map": {"active": "green", "closed": "gray"} } }

// Long form with currency
{ "ui_kind": { "type": "currency", "currency": "USD", "decimals": 2 } }
```

**Input components** (edit mode):

| ui_kind | Description |
|---------|-------------|
| `text` | Single-line text input |
| `textarea` | Multi-line text input |
| `number` | Numeric input |
| `select` | Dropdown |
| `multi_select` | Multiple selection |
| `checkbox` | Checkbox |
| `toggle` | Toggle switch |
| `radio` | Radio buttons |
| `date` | Date picker |
| `datetime` | DateTime picker |
| `time` | Time picker |
| `lookup` | Reference field (search + select) |
| `file_upload` | File upload |

**Display components** (view mode, lists):

| ui_kind | Description |
|---------|-------------|
| `badge` | Colored badge (with color_map) |
| `avatar` | Image/initials |
| `link` | Clickable hyperlink |
| `progress` | Progress bar |
| `rating` | Star rating |
| `relative_time` | "2 hours ago" |
| `currency` | Formatted currency |
| `percent` | Formatted percentage |
| `color` | Color indicator |
| `template` | String template with field substitution |

When `ui_kind` is not specified, auto-determined from field type/subtype in metadata.

### Reference Config

Reference fields (lookup) support detailed configuration:

```jsonc
{
  "ui_kind": "lookup",
  "reference": {
    "display_fields": ["name", "email", "phone"],
    "search_fields": ["name", "email"],
    "target": "popup",
    "hint": "Search customer...",
    "filter": {
      "items": [
        { "field": "is_active", "operator": "==", "value": true },
        { "field": "country_id", "operator": "==", "value": "{record.country_id}" }
      ]
    }
  }
}
```

| Property | Description |
|----------|-------------|
| `display_fields` | Fields shown in lookup popup/inline |
| `search_fields` | Fields searched when typing |
| `target` | Display mode: `popup` \| `inline` \| `link` \| `drawer` |
| `hint` | Placeholder text |
| `filter` | Pre-filter with static values or `{record.field}` placeholders |

### Conditional Logic

Fields, sections, and components support CEL-based conditional behavior:

```jsonc
// Field-level conditions
{
  "discount": {
    "required_expr": "record.amount > 10000",
    "readonly_expr": "record.status == 'closed'",
    "visibility_expr": "record.type == 'premium'"
  }
}

// Section-level conditions
{
  "section_config": {
    "financial": {
      "visibility_expr": "user.roles.exists(r, r == 'accountant' || r == 'admin')"
    }
  }
}
```

| Property | Type | Applies to |
|----------|------|-----------|
| `visibility_expr` | CEL → bool | Fields, sections, actions, components |
| `required_expr` | CEL → bool | Fields |
| `readonly_expr` | CEL → bool | Fields |

**CEL context variables**: `record` (current data), `user` (current user), `original` (before edit, null for new).

**Custom Functions** (ADR-0026): `fn.*` are available in all expressions (e.g., `fn.is_premium(record.tier)`).

**Evaluation**: backend evaluates during Form resolution (for access control); frontend re-evaluates via cel-js for instant reactivity.

### Actions

Actions support confirm dialogs and commands:

```jsonc
// Portal defines actions (WHAT)
{
  "actions": [
    {
      "key": "approve",
      "label": "Approve",
      "type": "success",
      "icon": "check",
      "visibility_expr": "record.status == 'pending'",
      "confirm": {
        "title": "Confirm approval",
        "description": "This action cannot be undone.",
        "confirm_text": "Approve",
        "cancel_text": "Cancel"
      },
      "command": {
        "type": "procedure",
        "procedure_code": "approve_order"
      }
    }
  ]
}
```

**Action types**: `primary` | `secondary` | `danger` | `success` | `link` | `icon` | `menu`

**Action targets**: `self` | `modal` | `drawer` | `new_tab` | `redirect` | `download`

### List Config

List view configuration with columns, sorting, filtering, search, and view modes:

```jsonc
{
  "list_config": {
    "view": "table",
    "columns": [
      {
        "field": "name",
        "label": "Name",
        "width": "30%",
        "sortable": true,
        "sort_dir": "asc",
        "filterable": true
      },
      {
        "field": "amount",
        "label": "Amount",
        "width": "15%",
        "align": "right",
        "sortable": true,
        "ui_kind": { "type": "currency", "currency": "USD" },
        "filter": { "type": "number", "operators": ["==", ">", "<", "between"] }
      },
      {
        "field": "status",
        "label": "Status",
        "width": "100px",
        "align": "center",
        "ui_kind": "badge",
        "filter": {
          "type": "select",
          "options": [
            { "label": "Active", "value": "active" },
            { "label": "Closed", "value": "closed" }
          ]
        }
      }
    ],
    "sort_by": [
      { "field": "created_at", "direction": "desc" }
    ],
    "search": {
      "fields": ["name", "email"],
      "placeholder": "Search..."
    },
    "row_actions": [
      { "id": "view", "type": "link", "target": "self" },
      { "id": "delete", "type": "icon", "icon": "trash", "confirm": { "title": "Delete?" } }
    ]
  }
}
```

**View modes**: `table` | `card` | `kanban` | `timeline`

**Column filter types**: `text` | `number` | `select` | `multi_select` | `date` | `date_range` | `boolean`. Auto-determined from field type if not specified.

### Layout for Mobile

Same Portal, different Layout with mode=edit:

```jsonc
// Layout (form_factor=mobile, mode=edit)
{
  "root": {
    "type": "grid",
    "columns": 1,
    "children": [
      { "type": "highlights" },
      { "type": "field_section", "key": "basic_info" },
      { "type": "field_section", "key": "address" }
    ]
  },

  "section_config": {
    "basic_info": { "columns": 1, "collapsed": false },
    "address": { "columns": 1, "collapsed": true }
  },

  "field_config": {
    "client_name": {
      "col_span": 1,
      "ui_kind": "lookup",
      "reference": {
        "display_fields": ["name"],
        "target": "link"
      }
    },
    "email": { "visibility_expr": "false" }
  }
}
```

### Form — Computed Frontend Contract

Form is built by the server during a Describe API request:

```
GET /api/v1/describe/Order
Authorization: Bearer <jwt>        -> determines profile -> Portal
X-Form-Factor: desktop             -> determines form_factor
X-Form-Mode: edit                  -> determines mode

Response: {
  "object": { ... },
  "fields": [ ... ],
  "form": {
    "root": {
      "type": "grid",
      "columns": 12,
      "children": [
        {
          "type": "group",
          "col_span": 8,
          "children": [
            { "type": "highlights" },
            { "type": "field_section", "key": "client_info" }
          ]
        },
        {
          "type": "group",
          "col_span": 4,
          "children": [
            { "type": "related_list", "key": "activities" }
          ]
        }
      ]
    },
    "sections": [
      {
        "key": "client_info",
        "label": "Client Information",
        "columns": 2,
        "collapsed": false,
        "fields": [
          {"field": "client_name", "col_span": 2, "ui_kind": "lookup", "reference": {...}},
          {"field": "contact_phone", "col_span": 1},
          {"field": "email", "col_span": 2, "ui_kind": "email"}
        ]
      }
    ],
    "highlight_fields": ["order_number", "status", "total_amount"],
    "actions": [...],
    "related_lists": [...],
    "list_config": { "columns": [...], "search": {...} }
  }
}
```

**The frontend receives Form and works only with it.** All `layout_ref` references are resolved, all shared layouts are merged, all FLS intersections are applied.

### Storage

```sql
-- Layouts — presentation per (portal, form_factor, mode)
CREATE TABLE metadata.layouts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    portal_id  UUID NOT NULL REFERENCES metadata.portals(id) ON DELETE CASCADE,
    form_factor     VARCHAR(20) NOT NULL DEFAULT 'desktop',
    mode            VARCHAR(20) NOT NULL DEFAULT 'edit',
    config          JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT layouts_form_factor_check
        CHECK (form_factor IN ('desktop', 'tablet', 'mobile')),
    CONSTRAINT layouts_mode_check
        CHECK (mode IN ('edit', 'view')),
    CONSTRAINT layouts_ov_ff_mode_unique
        UNIQUE (portal_id, form_factor, mode)
);

CREATE INDEX idx_layouts_portal_id ON metadata.layouts(portal_id);

-- Shared layouts — reusable fragments referenced via layout_ref
CREATE TABLE metadata.shared_layouts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_name   VARCHAR(63) NOT NULL,
    type       VARCHAR(20) NOT NULL,
    config     JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT shared_layouts_api_name_unique UNIQUE (api_name),
    CONSTRAINT shared_layouts_type_check CHECK (type IN ('field', 'section', 'list'))
);
```

### Resolution Rules

```
1. Resolve Portal:
   a. Find Portal by api_name (from navigation config or route)
   b. Fallback: auto-generate Portal (all FLS-accessible fields, one section)

2. Resolve Layout:
   a. layouts WHERE portal_id = Portal.id
      AND form_factor = requested AND mode = requested -> found? use
   b. Same form_factor, any mode -> fallback
   c. desktop, same mode -> fallback
   d. desktop, edit -> ultimate fallback
   e. No Layout -> auto-generate (full-width grid, col_span=1, ui_kind=auto)

3. Resolve layout_ref:
   a. For each field_config/section_config entry with layout_ref:
      shared_layouts WHERE api_name = layout_ref -> merge (inline overrides shared)
   b. Shared layout not found -> skip reference, use inline only

4. Merge -> Form:
   a. root from Layout (or auto-generate if absent)
   b. Portal sections + Layout section_config -> merged sections
   c. Portal fields + Layout field_config (with resolved layout_ref) -> merged fields
   d. Portal actions (as-is, already have visibility_expr)
   e. Portal list_fields + Layout list_config -> merged list config
   f. FLS intersection: remove fields denied by FLS
```

### Interaction with Security

Layout **does not extend** access:

```
Visible fields = Portal fields ∩ FLS-accessible fields ∩ Layout visibility
```

- If FLS denies a field -> the field is not in Form (security wins)
- If Portal does not include a field -> the field is not in Form (bounded context wins)
- If Layout hides a field (`visibility_expr: "false"`) -> the field is not rendered (presentation)

Layout can only **narrow** visibility, not extend it.

### Lifecycle: Portal -> Layout Synchronization

```
Portal field added (discount added to section "products")
  -> Layout does not change
  -> Form merge: discount appears with default presentation (col_span=1, ui_kind=auto)
  -> Admin can enrich the Layout for the new field

Portal field removed (discount removed from section)
  -> Layout field_config.discount remains (orphan — no effect)
  -> Form merge: discount not in Portal -> does not appear in Form
  -> Orphan cleanup: periodic or on Layout save

Portal section added
  -> Layout section_config does not contain new section -> defaults
  -> root may not include the new section -> add to end of main region
```

**Principle: Portal is the source of truth for structure. Layout supplements, but cannot show what is not in Portal.**

### API

```
-- Layout CRUD (Admin)
GET    /api/v1/admin/layouts?portal_id=:portalId     — list layouts for Portal
POST   /api/v1/admin/layouts                           — create layout
GET    /api/v1/admin/layouts/:id                       — get layout
PUT    /api/v1/admin/layouts/:id                       — update layout
DELETE /api/v1/admin/layouts/:id                       — delete layout

-- Shared Layouts CRUD (Admin)
GET    /api/v1/admin/shared-layouts                    — list shared layouts
POST   /api/v1/admin/shared-layouts                    — create shared layout
GET    /api/v1/admin/shared-layouts/:id                — get shared layout
PUT    /api/v1/admin/shared-layouts/:id                — update shared layout
DELETE /api/v1/admin/shared-layouts/:id                — delete (RESTRICT if referenced)

-- Form (User-facing — through Describe API)
GET    /api/v1/describe/:objectName                    — includes resolved Form
       Header: X-Form-Factor: desktop|tablet|mobile
       Header: X-Form-Mode: edit|view
```

### Constructor UI

**Layout Constructor** — dedicated admin screen, accessible from Portal detail:

1. **Page tab**: root component tree editor (drag components into grid, resize col_span)
2. **Sections tab**: per-section config (columns slider, collapsed toggle, visibility_expr)
3. **Fields tab**: per-field config (col_span, ui_kind picker, required_expr, readonly_expr, reference)
4. **List tab**: column config (width, align, sortable, filterable, ui_kind)
5. **Preview**: live preview with test data (desktop/tablet/mobile + edit/view switching)
6. **Presets**: start from common page structures (full-width, sidebar-right, two-columns)

Navigation: Portal detail -> button "Layout (desktop, edit)" -> Layout Constructor.

### Limits

| Parameter | Limit | Rationale |
|-----------|-------|-----------|
| Layouts per Portal | 6 max (3 form_factors × 2 modes) | Fixed combinations |
| field_config entries | Unlimited | Depends on Portal fields |
| root nesting depth | 3 levels | Prevent over-complex grids |
| grid columns | 1-12 | CSS grid column span |
| visibility_expr size | 1 KB | CEL expression, not a program |
| shared_layouts | Unlimited | Global library |

## Consequences

### Positive

- **Clean separation**: Portal = WHAT (bounded context), Layout = HOW (presentation), Form = unified contract
- **Composable pages**: GridLayout + GroupLayout express any page structure without a separate PageLayout entity
- **Multi-platform**: one Portal — different Layouts for desktop/tablet/mobile
- **Edit/view modes**: different presentations for editing vs viewing
- **Per-profile conditional behavior**: Layout per Portal -> each profile can have its own conditions
- **Shared layouts**: reusable field/section/list configs via layout_ref (DRY, consistency)
- **Frontend simplicity**: receives Form, does not know about Portal, Layout, or shared layouts
- **Graceful degradation**: no Layout -> default presentation; no root -> auto-generate full-width
- **Dual-stack CEL**: expressions evaluated via cel-js on frontend for instant reactivity
- **Custom Functions (ADR-0026)**: `fn.*` available in all Layout expressions
- **Extensible components**: new component types can be added without schema changes

### Negative

- **Additional tables + API**: metadata.layouts + metadata.shared_layouts with CRUD endpoints
- **Admin workflow**: two screens ( Portal editor + Layout editor) instead of one
- **Orphan config**: field removed from Portal -> Layout config has orphan entries (cleanup needed)
- **Merge complexity**: Form resolution requires merging Portal + Layout + shared layouts + FLS (cacheable)
- **layout_ref resolution**: additional DB lookups during Form resolution (cacheable in MetadataCache)

## Related ADRs

- **ADR-0019** — Declarative business logic: Layout = third level of cascade (Metadata -> Portal -> Layout -> Form)
- **ADR-0022** — Portal: Layout builds on top of Portal. Form = merge of Portal + Layout
- **ADR-0026** — Custom Functions: fn.* available in visibility_expr, required_expr, readonly_expr
- **ADR-0009..0012** — Security: Layout does not extend access. Form = Portal ∩ FLS ∩ Layout visibility
