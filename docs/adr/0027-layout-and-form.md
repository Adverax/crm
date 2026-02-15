# ADR-0027: Layout + Form — Presentation Layer for Object View

**Status:** Accepted
**Date:** 2026-02-15
**Participants:** @roman_myakotin

## Context

### Problem: Object View Defines WHAT, but Not HOW

Object View (ADR-0022) solves the bounded context problem: one object — different representations for different profiles. OV defines **which** sections, fields, actions, and related lists each role sees.

But OV does not answer presentation questions:

| Question | OV answers? |
|----------|-------------|
| Which fields does Sales see? | Yes |
| How many columns should a section render in? | No (only `columns: 2` — minimum) |
| What col_span does a specific field have? | No |
| How does the status field look — text or badge? | No |
| Is the discount field required when amount > 10000? | No |
| How does the list view look — column widths, sorting? | No |
| How does the form adapt to mobile devices? | No |

### Problem: God Object When Extending OV

If all layout attributes are added to the OV config, Object View becomes a God Object with two incompatible responsibilities:

- **Bounded context** (WHAT) — business decision about which fields/actions are available to a role
- **Presentation** (HOW) — visual decision about how to render on different devices

These responsibilities change independently: the bounded context changes when the business process changes, presentation — when UI/UX requirements change. Different people may be responsible for each.

### Problem: One OV — Multiple Devices

Desktop, tablet, and mobile require **structurally different** representations:

```
Desktop (2 columns, 3 sections):
+------------------+------------------+
| Client Name      | Phone            |
| col_span: 1      | col_span: 1      |
+------------------+------------------+
| Email (col_span: 2)                 |
+-------------------------------------+

Mobile (1 column, compact):
+-------------------------------------+
| Client Name                         |
+-------------------------------------+
| Phone                               |
+-------------------------------------+
| Email                               |
+-------------------------------------+
```

The grid is physically different. This is not a CSS media query — these are different col_span values, different collapsed sections, potentially a different set of visible fields (mobile may hide secondary ones).

A single OV cannot contain multiple grid configurations without turning into a God Object. Hence the need for a **separate Layout entity** bound to a form factor.

### Problem: Frontend Needs a Unified Contract

The frontend should not know about OV and Layout as separate concepts. It needs **one object** (Form) containing everything for rendering: structure, presentation, conditional behavior. The Describe API should return a ready-made Form, not two fragments to be assembled on the client.

### Relationship with ADR-0019 and ADR-0022

ADR-0019 defined a three-level cascade:

```
Metadata (base)
   | additive validation, inherit defaults
Object View (bounded context)        <- ADR-0022
   | additive validation, replace defaults, override visibility
Layout (presentation)                <- THIS ADR
   | merge
Form (frontend contract)             <- THIS ADR
```

Layout is the third level of the cascade. Form is the result of cascading resolution.

## Considered Options

### Option A — Extend Object View Config (God Object)

All layout attributes (grid, col_span, ui_kind, conditional behavior, list columns) are added to the OV config JSONB.

**Pros:**
- One entity, one table, one API
- No synchronization problem

**Cons:**
- God Object: OV is responsible for both bounded context and presentation
- Impossibility of different representations for desktop/mobile (one config)
- SRP violation: business decisions and visual decisions in one place
- Cognitive complexity: the administrator configures everything in one screen

### Option B — Layout Per Object (shared across profiles)

Layout is bound to the object, not the profile. One Layout for Order, all profiles use it.

**Pros:**
- DRY: one visual configuration per object
- Simplicity: one Layout per object

**Cons:**
- Conditional behavior (`required_expr`, `readonly_expr`) cannot be per-profile
- Example: Sales — discount required when amount > 10000; Manager — discount always optional. One Layout does not cover both cases
- Section-level config (columns, collapsed) is also per-profile: Sales — 2 columns, Warehouse — 1 column

### Option C — Layout Per Object View + Form Factor (chosen)

Layout is bound to a specific Object View and form factor. One OV can have multiple Layouts (desktop, tablet, mobile). Form is a computed merge of OV + Layout.

**Pros:**
- Clean separation: OV = WHAT, Layout = HOW
- Per-profile conditional behavior: each OV has its own Layouts with its own conditions
- Multi-platform: different Layouts for different devices
- Unified frontend contract: Form contains everything for rendering
- OV works without Layout (fallback = default presentation)
- Layout Builder (drag-and-drop) in the future — a clean editing point

**Cons:**
- Additional table and API
- Sync on OV change: adding a field to OV must be reflected in Layout
- Administrator works with two screens (OV + Layout)

### Option D — CSS-only Responsive

Responsiveness through CSS media queries. Minimal configuration.

**Pros:**
- Simplest implementation

**Cons:**
- Administrator cannot configure grid per section
- No conditional field behavior
- No ui_kind overrides
- No list column configuration

## Decision

**Option C chosen: Layout per Object View + form factor, Form as the computed frontend contract.**

### Three Entities, Three Responsibilities

```
+--------------------------+
|      Object View         |  Stored in metadata.object_views
|   (bounded context)      |  Per (object, profile)
|                          |
|  WHAT: sections, fields, |
|  actions, related lists  |
+------------+-------------+
             | 1:N
             v
+--------------------------+
|        Layout            |  Stored in metadata.layouts
|    (presentation)        |  Per (object_view, form_factor)
|                          |
|  HOW: grid, col_span,   |
|  ui_kind, conditions,    |
|  list columns            |
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
```

### Object View Config (WHAT — unchanged from ADR-0022)

```jsonc
{
  // Sections — field grouping
  "sections": [
    {
      "key": "client_info",
      "label": "Client Information",
      "fields": ["client_name", "contact_phone", "email"]
    },
    {
      "key": "products",
      "label": "Products",
      "fields": ["products", "total_amount", "discount"]
    }
  ],

  // Key fields at the top of the card
  "highlight_fields": ["order_number", "status", "total_amount"],

  // Actions
  "actions": [
    {
      "key": "send_proposal",
      "label": "Send Proposal",
      "type": "primary",
      "icon": "mail",
      "visibility_expr": "record.status == 'draft'"
    }
  ],

  // Child objects
  "related_lists": [
    {
      "object": "Activity",
      "label": "Activities",
      "fields": ["subject", "type", "due_date", "status"],
      "sort": "due_date DESC",
      "limit": 10
    }
  ],

  // List columns (which ones — without visual details)
  "list_fields": ["order_number", "client_name", "status", "total_amount", "created_at"],
  "list_default_sort": "created_at DESC"
}
```

### Layout Config (HOW)

```jsonc
{
  // Section presentation
  "section_config": {
    "client_info": {
      "columns": 2,
      "collapsed": false
    },
    "products": {
      "columns": 2,
      "collapsed": false,
      "visibility_expr": "record.status != 'cancelled'"
    }
  },

  // Field presentation
  "field_config": {
    "client_name": {
      "col_span": 2,
      "ui_kind": "lookup",
      "reference_config": {
        "display_fields": ["name", "email"],
        "search_fields": ["name", "email", "phone"],
        "target": "popup"
      }
    },
    "status": {
      "col_span": 1,
      "ui_kind": "badge"
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
  "list_columns": {
    "order_number": {"width": "15%", "sortable": true},
    "client_name": {"width": "30%", "sortable": true},
    "status": {"width": "100px", "align": "center", "ui_kind": "badge"},
    "total_amount": {"width": "15%", "align": "right", "sortable": true},
    "created_at": {"width": "15%", "sortable": true, "sort_dir": "desc"}
  }
}
```

### Layout for Mobile Platform

Same Object View, different Layout:

```jsonc
// Layout (Order Sales, mobile)
{
  "section_config": {
    "client_info": {
      "columns": 1,            // 1 column instead of 2
      "collapsed": false
    },
    "products": {
      "columns": 1,
      "collapsed": true,       // collapsed on mobile
      "visibility_expr": "record.status != 'cancelled'"
    }
  },

  "field_config": {
    "client_name": {
      "col_span": 1,           // full width (1 column)
      "ui_kind": "lookup",
      "reference_config": {
        "display_fields": ["name"],   // fewer fields for mobile
        "target": "link"              // link instead of popup
      }
    },
    "email": {
      "col_span": 1,
      "visibility_expr": "false"      // hidden on mobile
    }
  },

  "list_columns": {
    "order_number": {"width": "30%"},
    "status": {"width": "30%", "ui_kind": "badge"},
    "total_amount": {"width": "40%", "align": "right"}
  }
}
```

### Example: Business Process Phases (doctor visit)

One OV + one Layout — different behavior by phase through `visibility_expr`:

```jsonc
// OV (Visit, Doctor) — all fields for the Doctor profile
{
  "sections": [
    {"key": "patient", "label": "Patient", "fields": ["patient_name", "age", "diagnosis"]},
    {"key": "recommendations", "label": "Recommendations", "fields": ["notes", "medications"]},
    {"key": "review", "label": "Visit Outcome", "fields": ["outcome", "next_visit", "rating"]}
  ],
  "actions": [
    {"key": "start_visit", "label": "Start Appointment", "visibility_expr": "record.status == 'scheduled'"},
    {"key": "complete_visit", "label": "Complete Appointment", "visibility_expr": "record.status == 'in_progress'"}
  ]
}

// Layout (Visit, Doctor, desktop)
{
  "section_config": {
    "patient": {"columns": 2, "collapsed": false},
    "recommendations": {
      "columns": 1,
      "visibility_expr": "record.status == 'scheduled' || record.status == 'in_progress'"
    },
    "review": {
      "columns": 2,
      "visibility_expr": "record.status == 'completed'"
    }
  },
  "field_config": {
    "outcome": {"required_expr": "record.status == 'completed'"},
    "next_visit": {"ui_kind": "date"},
    "rating": {"ui_kind": "rating", "readonly_expr": "record.status != 'completed'"}
  }
}
```

Before the visit: the "Recommendations" section is visible, "Visit Outcome" is hidden.
After the visit: "Visit Outcome" is visible with a required outcome, "Recommendations" is hidden.

One OV, one Layout. The frontend evaluates `visibility_expr` through cel-js and adapts the form instantly.

### Form — Computed Frontend Contract

Form is built by the server during a Describe API request:

```
GET /api/v1/describe/Order
Authorization: Bearer <jwt>        -> determines profile
X-Form-Factor: desktop             -> determines platform

Response: {
  "object": { ... },
  "fields": [ ... ],
  "form": {                         <- merged OV + Layout
    "sections": [
      {
        "key": "client_info",
        "label": "Client Information",
        "columns": 2,
        "collapsed": false,
        "fields": [
          {"field": "client_name", "col_span": 2, "ui_kind": "lookup", "reference_config": {...}},
          {"field": "contact_phone", "col_span": 1},
          {"field": "email", "col_span": 2, "ui_kind": "email"}
        ]
      }
    ],
    "highlight_fields": ["order_number", "status", "total_amount"],
    "actions": [
      {"key": "send_proposal", "label": "Send", "type": "primary", "visibility_expr": "..."}
    ],
    "related_lists": [...],
    "list_columns": [
      {"field": "order_number", "width": "15%", "sortable": true},
      {"field": "total_amount", "width": "15%", "align": "right"}
    ]
  }
}
```

**The frontend receives Form and works only with it.** It does not know about OV and Layout as separate concepts.

### Storage

```sql
-- Layouts — presentation per (object_view, form_factor)
CREATE TABLE metadata.layouts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    object_view_id  UUID NOT NULL REFERENCES metadata.object_views(id) ON DELETE CASCADE,
    form_factor     VARCHAR(20) NOT NULL DEFAULT 'desktop',  -- desktop | tablet | mobile
    config          JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT layouts_form_factor_check
        CHECK (form_factor IN ('desktop', 'tablet', 'mobile')),
    CONSTRAINT layouts_object_view_form_factor_unique
        UNIQUE (object_view_id, form_factor)
);

CREATE INDEX idx_layouts_object_view_id ON metadata.layouts(object_view_id);
```

### Resolution Rules

```
1. Resolve Object View:
   a. object_views WHERE object_id = X AND profile_id = P -> found? use
   b. object_views WHERE object_id = X AND is_default = true -> found? use
   c. Fallback: auto-generate OV (all FLS-accessible fields, one section)

2. Resolve Layout:
   a. layouts WHERE object_view_id = OV.id AND form_factor = requested -> found? use
   b. layouts WHERE object_view_id = OV.id AND form_factor = 'desktop' -> fallback to desktop
   c. No Layout -> auto-generate (col_span=1, ui_kind=auto, all fields visible)

3. Merge -> Form:
   a. OV sections + Layout section_config -> merged sections
   b. OV fields + Layout field_config -> merged fields with presentation
   c. OV actions (as-is, already have visibility_expr)
   d. OV list_fields + Layout list_columns -> merged list columns
   e. FLS intersection: remove fields denied by FLS
```

### Interaction with Security

Layout **does not extend** access:

```
Visible fields = OV fields ∩ FLS-accessible fields ∩ Layout visibility
```

- If FLS denies a field -> the field is not in Form (security wins)
- If OV does not include a field -> the field is not in Form (bounded context wins)
- If Layout hides a field (`visibility_expr: "false"`) -> the field is not rendered (presentation)

Layout can only **narrow** visibility, not extend it.

### Lifecycle: OV -> Layout Synchronization

When Object View changes (field addition/removal):

```
OV field added (discount added to section "products")
  -> Layout does not change
  -> Form merge: discount field appears with default presentation (col_span=1, ui_kind=auto)
  -> Admin can enrich the Layout for the new field

OV field removed (discount removed from section)
  -> Layout field_config.discount remains (orphan — has no effect)
  -> Form merge: discount field is not in OV -> does not appear in Form
  -> Orphan cleanup: periodic or on Layout save

OV section added
  -> Layout section_config does not contain the new section -> default (columns=1, collapsed=false)
  -> Admin can enrich the Layout for the new section
```

**Principle: OV is the source of truth for structure. Layout supplements, but cannot show what is not in OV.**

### API

```
-- Layout CRUD (Admin)
GET    /api/v1/admin/layouts?object_view_id=:ovId    — list layouts for OV
POST   /api/v1/admin/layouts                          — create layout
GET    /api/v1/admin/layouts/:id                      — get layout
PUT    /api/v1/admin/layouts/:id                      — update layout
DELETE /api/v1/admin/layouts/:id                      — delete layout

-- Form (User-facing — through Describe API)
GET    /api/v1/describe/:objectName                   — includes resolved Form
       Header: X-Form-Factor: desktop|tablet|mobile
```

### Constructor UI

**Layout Constructor** — a dedicated admin screen, accessible from Object View detail:

1. **Section tab**: per-section config (columns slider, collapsed toggle, visibility_expr)
2. **Fields tab**: per-field config (col_span slider, ui_kind picker, required_expr, readonly_expr, reference_config)
3. **List tab**: column config (width, align, sortable, filterable)
4. **Preview**: live preview with test data (desktop/tablet/mobile switching)

Navigation: Object View detail -> button "Layout (desktop)" / "Layout (mobile)" -> Layout Constructor.

### Limits

| Parameter | Limit | Rationale |
|-----------|-------|-----------|
| Layouts per OV | 3 (desktop, tablet, mobile) | Fixed form factors |
| field_config entries | Unlimited | Depends on the number of fields in OV |
| visibility_expr size | 1 KB | CEL expression, not a program |
| Nesting (col_span) | 1-12 | CSS grid column span |

### ui_kind Types

| ui_kind | Description | Applicable to |
|---------|-------------|---------------|
| `auto` | Auto-determined from field type/subtype | Default |
| `text` | Text input | string |
| `textarea` | Multiline text input | text/long_text |
| `number` | Numeric input | number |
| `currency` | Numeric with currency symbol | number/currency |
| `percent` | Numeric with % | number/percent |
| `email` | Email with icon and clickable link | string/email |
| `phone` | Phone with clickable link | string/phone |
| `url` | URL with clickable link | string/url |
| `date` | Date picker | datetime/date |
| `datetime` | DateTime picker | datetime |
| `checkbox` | Checkbox | boolean |
| `toggle` | Toggle switch | boolean |
| `select` | Dropdown | picklist |
| `radio` | Radio buttons | picklist (<=5 options) |
| `badge` | Colored badge | picklist/status |
| `lookup` | Search and select field | reference |
| `rating` | Stars/scale | number (1-5) |
| `slider` | Slider | number (range) |
| `color` | Color picker | string/color |
| `rich_text` | Rich text editor | text/rich |

When `ui_kind: "auto"`, the component type is determined from the field type/subtype (metadata). Override through Layout allows the administrator to choose an alternative component.

## Consequences

### Positive

- **Clean separation**: OV = WHAT (bounded context), Layout = HOW (presentation), Form = unified contract
- **Multi-platform**: one OV — different Layouts for desktop/tablet/mobile
- **Per-profile conditional behavior**: Layout per OV -> each profile can have its own required_expr, readonly_expr
- **Business process phases**: visibility_expr on sections/fields — one form adapts to the record state
- **Frontend simplicity**: receives Form, does not know about OV and Layout
- **Graceful degradation**: no Layout -> default presentation from OV; no OV -> auto-generate from metadata + FLS
- **Layout Builder**: in the future — visual drag-and-drop editor for Layout (a clean editing point)
- **Dual-stack CEL**: visibility_expr, required_expr, readonly_expr are evaluated through cel-js on the frontend instantly
- **Custom Functions (ADR-0026)**: `fn.*` are available in Layout expressions (`fn.is_premium(record.tier)`)

### Negative

- **Additional table + API**: metadata.layouts with CRUD endpoints
- **Admin workflow**: two screens (OV editor + Layout editor) instead of one
- **Orphan config**: when a field is removed from OV, Layout field_config contains orphan entries (cleanup needed)
- **Merge complexity**: Form resolution requires merging OV + Layout + FLS intersection (cacheable)
- **Three concepts**: the administrator must understand OV, Layout, Form (Constructor UI lowers the barrier)

## Related ADRs

- **ADR-0019** — Declarative business logic: Layout = the third level of the cascade (Metadata -> OV -> Layout -> Form)
- **ADR-0022** — Object View: Layout is built on top of OV, does not replace it. Form = merge of OV + Layout
- **ADR-0026** — Custom Functions: fn.* are available in visibility_expr, required_expr, readonly_expr
- **ADR-0009..0012** — Security: Layout does not extend access. Form = OV ∩ FLS ∩ Layout visibility
