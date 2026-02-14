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
| Phase 8 | Next | Notifications, dashboard, analytics |

The platform is **fully functional** across 9 completed phases (20 ADRs). It can create objects via metadata engine or App Templates, manage permissions, enforce 3-layer security (OLS/FLS/RLS), query data through SOQL, perform all DML operations with CEL-based validation rules and dynamic defaults, authenticate users via JWT, and work with records through a dynamic metadata-driven UI.

---

## Architecture Decisions

Every significant decision is documented as an ADR in [`docs/adr/`](docs/adr/):

| ADR | Decision |
|-----|----------|
| [0001](docs/adr/0001-uuid-as-record-identifier.md) | UUID v4 as primary key |
| [0003](docs/adr/0003-object-metadata-structure.md) | Metadata structure with behavioral flags |
| [0007](docs/adr/0007-table-per-object-storage.md) | Table-per-object (not EAV) |
| [0009](docs/adr/0009-security-architecture-overview.md) | 3-layer security: OLS + FLS + RLS |
| [0010](docs/adr/0010-permission-model-ols-fls.md) | Grant + Deny permission sets with bitmasks |
| [0011](docs/adr/0011-row-level-security-model.md) | RLS: OWD + sharing rules + role hierarchy |
| [0012](docs/adr/0012-security-caching-strategy.md) | Closure tables + outbox for cache invalidation |
| [0014](docs/adr/0014-licensing-and-business-model.md) | Open Core: AGPL v3 + proprietary ee/ |
| [0017](docs/adr/0017-auth-module.md) | JWT auth: access + refresh tokens, bcrypt, rate limiting |
| [0018](docs/adr/0018-app-templates.md) | App Templates: Go-embedded templates instead of hardcoded standard objects |
| [0019](docs/adr/0019-declarative-business-logic.md) | Declarative business logic: validation rules, defaults, formulas (CEL) |
| [0020](docs/adr/0020-dml-pipeline-extension.md) | DML pipeline extension: typed stages with Option pattern |

[All 20 ADRs →](docs/adr/)

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
- [API Specification](api/openapi.yaml)
