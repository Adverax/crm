# Adverax CRM

[![CI](https://github.com/adverax/crm/actions/workflows/ci.yml/badge.svg)](https://github.com/adverax/crm/actions/workflows/ci.yml)

**Open-source CRM platform with Salesforce-grade security and metadata-driven architecture.**

Build custom objects, enforce 3-layer security (OLS/FLS/RLS), and query data through a unified SOQL engine — all self-hosted on PostgreSQL.

---

## Why another CRM?

Salesforce charges **$150+/user/month** and locks you into their cloud. Open-source alternatives (SuiteCRM, EspoCRM) bolt security on as an afterthought and store custom fields in EAV tables that don't scale.

Adverax CRM takes a different approach:

| Problem | How we solve it |
|---------|----------------|
| Custom objects require migrations | **Metadata engine** creates real PostgreSQL tables at runtime — no EAV, no JSON blobs |
| Security is a middleware hack | **3-layer model** (OLS → FLS → RLS) enforced at the query level, not in controllers |
| No standard query language | **SOQL engine** — one language for all reads, with automatic security filtering |
| Vendor lock-in | **Self-hosted**, AGPL v3, your data on your server |
| "Enterprise" means expensive | **Open Core** — full security in the free tier, enterprise add-ons in `ee/` |

---

## Key Features

### Metadata-Driven Objects

Every object — standard or custom — is a real PostgreSQL table (`obj_{api_name}`) with native constraints, indexes, and foreign keys. No EAV overhead. No query complexity. Just SQL under the hood.

```
POST /admin/objects
{
  "api_name": "invoice",
  "label": "Invoice",
  "object_type": "custom"
}
→ Creates table obj_invoice with id, owner_id, created_at, updated_at
→ Registers in metadata cache
→ Ready for CRUD immediately
```

### 3-Layer Security Model

Inspired by Salesforce, implemented from scratch:

```
┌─────────────────────────────────────────┐
│  OLS — Object-Level Security            │
│  Can this profile CRUD this object?     │
├─────────────────────────────────────────┤
│  FLS — Field-Level Security             │
│  Can this profile read/edit this field? │
├─────────────────────────────────────────┤
│  RLS — Row-Level Security               │
│  Can this user see this record?         │
│  (ownership, sharing rules, hierarchy)  │
└─────────────────────────────────────────┘
```

- **Grant + Deny Permission Sets** — `effective = grants & ~denies`
- **Role hierarchy** — managers see subordinates' records
- **Sharing rules** — criteria-based and ownership-based
- **Manual sharing** — share specific records with users/groups
- **4 group types** — personal, role, role & subordinates, public

### SOQL Query Engine

One language for all data reads. Security enforcement is automatic — no way to bypass it.

```sql
SELECT Name, Amount, Account.Name
FROM Deal
WHERE Stage = 'Closed Won' AND Amount > 10000
ORDER BY CloseDate DESC
LIMIT 50
```

- **SOQL Editor** — Rich editor with syntax highlighting, context-aware autocomplete (objects, fields, keywords), server-side validation, and test query execution. Used in Object View queries and reusable across reports.

### App Templates

Start with a pre-built set of objects and fields for your domain — or build your own from scratch. Templates are applied once through the admin UI and create real metadata objects.

- **Sales CRM** — Account, Contact, Opportunity, Task (4 objects, 36 fields)
- **Recruiting** — Position, Candidate, Application, Interview (4 objects, 28 fields)

### Declarative Business Logic

CEL-based validation rules and dynamic defaults — no code required:

- **Validation Rules** — CEL expressions checked on every INSERT/UPDATE. `size(record.Name) > 0` blocks records with empty names.
- **Dynamic Defaults** — CEL expressions for default field values. `user.id` auto-fills owner, `now` sets timestamps.
- **DML Pipeline** — 6-stage pipeline (Parse → Resolve → Defaults → Validate → Compile → Execute) with typed interfaces and Option pattern.

### Generic CRUD + Metadata-Driven UI

One set of REST endpoints and Vue.js views serves **all objects** — no per-object code. The frontend renders forms, tables, and detail pages dynamically from metadata.

- `GET /api/v1/describe` — object list for navigation (OLS-filtered)
- `GET /api/v1/describe/:objectName` — fields and config (FLS-filtered)
- `GET/POST/PUT/DELETE /api/v1/records/:objectName` — generic CRUD
- Two UI zones: `/app/*` (CRM workspace) and `/admin/*` (administration)

### Layout + Form Resolution

Control how records are displayed per Object View, form factor (desktop/tablet/mobile), and mode (edit/view):

- **Layouts** — `metadata.layouts` table, per (object_view_id, form_factor, mode). Configures section grids, field presentation (col_span, ui_kind, reference config), and list columns.
- **Visual Layout Constructor** — tabbed admin editor: Form Layout tab (section canvas with field chips + property panels), List Config tab (DnD column reorder), JSON tab (power-user fallback).
- **Shared Layouts** — `metadata.shared_layouts` reusable configuration snippets (type: field/section/list) referenced via `layout_ref`. Inline overrides win. RESTRICT delete protects referenced shared layouts.
- **Form merge** — Describe API merges OV config + Layout config into a computed Form. Frontend works only with the final Form.
- **Fallback chain** — requested layout -> same form_factor any mode -> desktop same mode -> desktop edit -> auto-generate.
- **Headers** — `X-Form-Factor` and `X-Form-Mode` request headers for layout resolution.

### Table-per-Object Storage

Each object gets a dedicated PostgreSQL table. This means:

- Native `JOIN` performance
- Real foreign keys and constraints
- Standard `EXPLAIN ANALYZE` for optimization
- No N+1 queries from attribute lookups

---

## Architecture

```
┌──────────┐     ┌──────────┐     ┌──────────┐
│ Vue.js 3 │────▶│  Gin API │────▶│  Security│
│ Frontend │     │ Handlers │     │  OLS/FLS │
└──────────┘     └────┬─────┘     └────┬─────┘
                      │                │
                 ┌────▼─────┐     ┌────▼─────┐
                 │  SOQL    │────▶│   RLS    │
                 │  Engine  │     │ Enforcer │
                 └────┬─────┘     └──────────┘
                      │
                 ┌────▼─────┐     ┌──────────┐
                 │ Metadata │     │  Redis   │
                 │  Cache   │     │  Cache   │
                 └────┬─────┘     └──────────┘
                      │
                 ┌────▼──────────────────────┐
                 │       PostgreSQL 16        │
                 │  obj_* tables │ metadata   │
                 │  security    │ migrations  │
                 └───────────────────────────┘
```

**Data flow:** Handler → Service → SOQL/DML → Security → Repository → PostgreSQL

No layer can be bypassed. Every read goes through SOQL. Every write goes through DML. Both enforce all three security layers.

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.25 |
| HTTP | Gin |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| ORM | sqlc (generated type-safe queries) |
| Migrations | golang-migrate |
| DB Tests | pgTAP |
| Frontend | Vue.js 3 + TypeScript + Pinia + Tailwind CSS |
| API | REST + JSON, OpenAPI 3.1 |
| Build | Docker, docker-compose |

---

## Quick Start

### Prerequisites

- Docker and docker-compose
- Go 1.25+ (for local development)
- Node.js 20+ (for frontend)

### Run with Docker

```bash
git clone https://github.com/adverax/crm.git
cd crm

# Start PostgreSQL + Redis + run migrations
make docker-up
make migrate-up

# Start the API server
make run
```

API is available at `http://localhost:8080`.

### Frontend

```bash
cd web
npm install
npm run dev
```

Frontend is available at `http://localhost:5173`.

### Useful Commands

```bash
make build            # Build API binary
make test             # Run Go tests with race detection
make lint             # Run golangci-lint
make test-pgtap       # Run pgTAP schema tests
make sqlc-generate    # Regenerate type-safe queries
make docker-reset     # Reset all data and restart
```

---

## Project Status

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 0 | Done | Scaffolding, Docker, CI, Makefile |
| Phase 1 | Done | Metadata engine — object definitions, field registry, DDL |
| Phase 2a | Done | Identity + Permissions — users, profiles, permission sets, OLS, FLS |
| Phase 2b | Done | RLS core — OWD, share tables, role hierarchy, sharing rules |
| Phase 2c | Done | Groups — personal, role, role & subordinates, public |
| Phase 3 | Done | SOQL parser and executor |
| Phase 4 | Done | DML engine |
| Phase 5 | Done | Auth module — JWT, login, password reset, rate limiting |
| Phase 6 | Done | App Templates — Sales CRM & Recruiting (one-click object/field setup) |
| Phase 7a | Done | Generic CRUD + metadata-driven UI — dynamic record views, describe API |
| Phase 7b | Done | CEL engine, validation rules, dynamic defaults, DML pipeline extension |
| Phase 8 | Done | Custom Functions — fn.* namespace, dual-stack (cel-go + cel-js), Expression Builder |
| Phase 9a | Done | Object View Core — visual constructor, Describe API form resolution, section-based CRM rendering |
| Phase 9b | Done | Navigation per profile + OV Unbinding — grouped sidebar, page nav items |
| Phase 9c | Done | Layout + Form Resolution — layouts + shared_layouts, form merge, fallback chain |
| Phase 10a | Done | Procedure Engine Core — 6 command types, Named Credentials, versioning, Constructor UI |
| Phase 10b | Done | Automation Rules — DML triggers with CEL conditions, procedure_code actions |
| Phase 11a | **Next** | Related Lists + Reference Lookup — child records on detail page, searchable reference fields |
| Phase 11b | Planned | List Search, Sort, Row Actions — search bar, sortable columns, quick actions |
| Phase 11c | Planned | Recycle Bin — soft delete + restore + scheduled purge |
| Phase 11d | Planned | CSV Import — drag-and-drop upload, field mapping, batch DML |
| Phase 12a-d | Planned | Daily Tool — in-app notifications, email, saved list views, activity timeline, bulk actions |
| Phase 13a-d | Planned | Production-Ready — file attachments, export, global search, audit log, formula fields |
| Phase 14+ | Planned | Platform Completeness — scenarios, approvals, record types, advanced objects |

The platform is **fully functional** across 16 completed phases (32 ADRs). It can create objects via metadata engine or App Templates, manage permissions, enforce 3-layer security (OLS/FLS/RLS), query data through SOQL, perform all DML operations with CEL-based validation rules and dynamic defaults, authenticate users via JWT, work with records through a dynamic metadata-driven UI, define reusable Custom Functions with fn.* namespace (dual-stack: cel-go backend + cel-js frontend with Expression Builder), configure **Object Views** as full bounded context adapters per profile, create **Layouts** per Object View + form factor + mode to control page structure and field presentation (with shared layouts for reuse), define **Procedures** with a visual Constructor UI (6 command types, Named Credentials, versioning, Saga rollback), set up **Automation Rules** (DML triggers with CEL conditions), and configure **per-profile Navigation** (grouped sidebar with OLS intersection, page nav items via OV api_name).

Roadmap principle: **user value first** — each next phase delivers visible benefit to end users. Phase 11 makes CRM usable (related lists, search, recycle bin), Phase 12 makes it a daily tool (notifications, list views), Phase 13 makes it production-ready (files, audit, formulas). See [full roadmap](docs/roadmap.md) for details.

---

## Architecture Decisions

Every significant decision is documented as an ADR in [`docs/adr/`](docs/adr/):

| ADR | Decision |
|-----|----------|
| [0001](docs/adr/0001-uuid-as-record-identifier.md) | UUID v4 as primary key |
| [0002](docs/adr/0002-internationalization-strategy.md) | i18n: default value inline + translations table |
| [0003](docs/adr/0003-object-metadata-structure.md) | Metadata structure with behavioral flags |
| [0004](docs/adr/0004-field-type-subtype-hierarchy.md) | Field type/subtype hierarchy |
| [0005](docs/adr/0005-reference-field-types.md) | Reference types: association, composition, polymorphic |
| [0006](docs/adr/0006-relationship-registry-as-cache.md) | Relationship registry as in-memory cache |
| [0007](docs/adr/0007-table-per-object-storage.md) | Table-per-object (not EAV) |
| [0008](docs/adr/0008-admin-panel-placement.md) | Admin panel inside web/ monorepo |
| [0009](docs/adr/0009-security-architecture-overview.md) | 3-layer security: OLS + FLS + RLS |
| [0010](docs/adr/0010-permission-model-ols-fls.md) | Grant + Deny permission sets with bitmasks |
| [0011](docs/adr/0011-row-level-security-model.md) | RLS: OWD + sharing rules + role hierarchy |
| [0012](docs/adr/0012-security-caching-strategy.md) | Closure tables + outbox for cache invalidation |
| [0013](docs/adr/0013-group-model.md) | Groups: 4 types, unified grantee, auto-generation |
| [0014](docs/adr/0014-licensing-and-business-model.md) | Open Core: AGPL v3 + proprietary ee/ |
| [0015](docs/adr/0015-territory-management.md) | Territory Management (ee/) |
| [0016](docs/adr/0016-single-tenant-architecture.md) | Single-tenant architecture |
| [0017](docs/adr/0017-auth-module.md) | JWT auth: access + refresh tokens, bcrypt, rate limiting |
| [0018](docs/adr/0018-app-templates.md) | App Templates: Go-embedded, Registry+Applier pattern |
| [0019](docs/adr/0019-declarative-business-logic.md) | Declarative business logic: 5 subsystems, CEL |
| [0020](docs/adr/0020-dml-pipeline-extension.md) | DML pipeline extension: typed stages with Option pattern |
| [0021](docs/adr/0021-contract-testing.md) | Contract testing: OpenAPI validation + TS type generation |
| [0022](docs/adr/0022-object-view-bounded-context.md) | Object View: role-based UI per profile |
| [0023](docs/adr/0023-action-terminology.md) | Action terminology: Command → Procedure → Scenario |
| [0024](docs/adr/0024-procedure-engine.md) | Procedure Engine: JSON DSL + Constructor UI |
| [0025](docs/adr/0025-scenario-engine.md) | Scenario Engine: durable async workflows |
| [0026](docs/adr/0026-custom-functions.md) | Custom Functions: named pure CEL, fn.* namespace |
| [0027](docs/adr/0027-layout-and-form.md) | Layout + Form: OV (what) + Layout (how) + Form (computed) |
| [0028](docs/adr/0028-named-credentials.md) | Named Credentials: AES-256-GCM encrypted secrets |
| [0029](docs/adr/0029-versioning-strategy.md) | Versioning: Draft/Published for Procedure + Scenario |
| [0030](docs/adr/0030-modular-monolith-strategy.md) | Modular Monolith: MetadataReader interface |
| [0031](docs/adr/0031-automation-rules.md) | Automation Rules: reactive triggers on DML events |
| [0032](docs/adr/0032-profile-navigation-and-dashboard.md) | Profile Navigation + OV Unbinding |

[All 32 ADRs →](docs/adr/)

---

## Comparison

| Feature | Adverax CRM | Salesforce | SuiteCRM | Twenty |
|---------|-------------|------------|----------|--------|
| Custom objects | Real tables (DDL) | Proprietary | EAV | Hardcoded |
| OLS/FLS/RLS | Built-in, 3-layer | Built-in | Partial (roles) | Basic roles |
| Query language | SOQL | SOQL | — | GraphQL |
| Self-hosted | Yes | No | Yes | Yes |
| License | AGPL v3 | Proprietary | AGPL v3 | AGPL v3 |
| Stack | Go + PostgreSQL | Java + Oracle | PHP + MySQL | TypeScript + PostgreSQL |
| Pricing | Free (core) | $25-150/user/mo | Free | Free |

---

## License

- **Core platform** (everything outside `ee/`): [GNU AGPL v3](LICENSE)
- **Enterprise add-ons** (`ee/` directory): [Adverax Commercial License](ee/LICENSE)

Enterprise features: territory management, audit trail, SSO, advanced analytics.

---

## Links

- [Architecture Decision Records](docs/adr/)
- [Roadmap](docs/roadmap.md)
- [API Specification](api/openapi.yaml)
