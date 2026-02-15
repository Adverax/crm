# ADR-0014: Licensing and Business Model

**Status:** Accepted
**Date:** 2026-02-08
**Participants:** @roman_myakotin

## Context

The project is approaching the public release stage. The following must be determined:
- Distribution model (SaaS, open source, open core, source available)
- Code license
- Boundary between free and paid components
- Technical organization of differently-licensed code in a single repository

Key constraints:
- Minimal budget — no resources for separate SaaS infrastructure at launch
- The security engine (OLS/FLS/RLS) is deeply integrated into the core (SOQL, DML) — it cannot be extracted without significantly complicating the architecture
- Target audience — B2B companies where compliance and legal risks are critical

## Considered Alternatives

### Option A — Pure SaaS (closed source)

A fully closed product available only as a cloud service.

Pros: full control, straightforward monetization, IP protection.
Cons: infrastructure budget required from day one, no community effect, high competition with Salesforce/Bitrix24/amoCRM without differentiation.

### Option B — Pure Open Source (AGPL)

All code under AGPL v3. Monetization through support and consulting.

Pros: maximum trust, community contributions, rapid adoption.
Cons: hard to monetize — the support model scales linearly with people. A competitor can take the code and sell a hosted version (AGPL requires opening code but does not prohibit commercial use).

### Option C — Open Core: AGPL + Proprietary `ee/` in a Single Repository (chosen)

The core (including the full security engine) is AGPL v3. Enterprise add-ons are under a proprietary license in the `ee/` directory. A single public repository.

Pros: a fully functional self-hosted CRM attracts users; enterprise add-ons are monetized through licenses; AGPL protects against hosted competitors; proven model (GitLab, Mattermost, Grafana); single repo — simple development and CI.
Cons: proprietary code is technically accessible (but legally protected); the license boundary must be carefully marked.

### Option D — Source Available (BSL / ELv2)

All code under Business Source License or Elastic License 2.0. Prohibition on competing managed services.

Pros: simple protection against competitors, all code is visible.
Cons: not OSI-approved — community perceives it negatively, fewer contributions, less trust.

## Decision

### Distribution Model: Open Core

A single public repository with two licenses:

| Scope | License | Directory |
|-------|---------|-----------|
| Core platform | AGPL v3 | Everything outside `ee/` |
| Enterprise add-ons | Adverax Commercial License | `ee/` |

### Boundary Between Free and Paid Components

**AGPL v3 (free, self-hosted):**

Platform:
- Metadata engine (custom objects <= 20, custom fields per object <= 50)
- SOQL parser and executor
- DML engine
- Standard objects (contacts, accounts, deals, tasks)
- REST API (<= 1000 req/min)
- Vue.js frontend
- Self-hosted deployment (Docker)
- Webhooks (outbound, all events, retry 3x)
- Data export (CSV)

Security:
- OLS + FLS fully included
- RLS fully included (OWD, share tables, role hierarchy, sharing rules, manual sharing)
- Groups (all 4 types: personal, role, role_and_subordinates, public)
- Security caching (closure tables, effective caches)

Auth:
- JWT (access + refresh tokens)
- Login, register, password reset
- MFA (TOTP, WebAuthn)
- Basic login history (login log, date, IP, user-agent)

**Adverax Commercial License (paid):**

Security & Access Control:
- Territory management (territorial hierarchy, territory-based groups)
- PermissionSetGroups (permission set grouping)
- Delegated administration (delegating admin rights by department)
- IP whitelist / login restrictions
- Advanced session management (force logout, session policies)

Auth & Identity:
- SSO / SAML 2.0
- LDAP / Active Directory sync
- OAuth2 provider (CRM as IdP)

Compliance & Audit:
- Audit Trail (full log of all record changes)
- Field History Tracking (history of individual field changes)
- Data retention policies (auto-cleanup, GDPR compliance)
- Security analytics (login geo-analytics, anomaly detection)

Automation:
- Workflow rules (field update, email alert, record creation)
- Approval processes (approval chains)
- Scheduled jobs / batch processing (background processing)

Analytics & Reporting:
- Custom reports builder (visual report designer)
- Dashboards (configurable dashboards, drag-and-drop)
- Scheduled report delivery (email report delivery)

Platform:
- Multi-org / multi-tenant mode
- Sandbox environments (dev/staging organization copy)
- Increased limits: custom objects > 20, custom fields > 50, API > 1000 req/min

Services:
- Managed cloud hosting (SaaS)
- Priority support + SLA
- Professional services / onboarding

**Boundary rationale:**

- The security engine (OLS/FLS/RLS) is deeply integrated into SOQL/DML — separation would require a complex plugin architecture. Full security stays in core.
- MFA, webhooks, CSV export, basic login history are in core for trust and user attraction. Security by default, no vendor lock-in.
- Enterprise features are what large companies need: compliance (audit), advanced auth (SSO/LDAP), automation, analytics, territories.
- Limits in the free tier (20 objects, 50 fields, 1000 req/min) are sufficient for small/medium businesses. Enterprise removes the restrictions.

### Repository Structure

The `ee/` directory mirrors the main project structure and contains all layers of enterprise code: Go packages, Vue components, SQL migrations, tests.

```
crm/
├── LICENSE                            ← AGPL v3 (default for everything)
├── internal/                          ← AGPL v3: core platform
│   ├── platform/
│   │   ├── security/                  ← full RLS/OLS/FLS
│   │   ├── metadata/
│   │   ├── soql/
│   │   └── dml/
│   ├── modules/
│   ├── handler/
│   └── service/
├── migrations/                        ← AGPL v3: core schema
├── web/                               ← AGPL v3: core frontend
│   └── src/
├── ee/                                ← Adverax Commercial License
│   ├── LICENSE                        ← proprietary license
│   ├── internal/
│   │   ├── platform/
│   │   │   ├── territory/             ← Go: territory hierarchy, territory-based groups
│   │   │   └── audit/                 ← Go: audit trail engine
│   │   ├── modules/
│   │   │   └── sso/                   ← Go: SSO / SAML / LDAP
│   │   ├── handler/                   ← Go: enterprise API endpoints
│   │   └── service/                   ← Go: enterprise business logic
│   ├── migrations/                    ← SQL: enterprise-only tables
│   ├── sqlc/
│   │   └── queries/                   ← SQL: enterprise queries
│   ├── web/
│   │   └── src/
│   │       ├── views/                 ← Vue: enterprise pages
│   │       ├── components/            ← Vue: enterprise components
│   │       └── stores/                ← Vue: enterprise Pinia stores
│   └── tests/
│       └── pgtap/                     ← pgTAP: enterprise schema tests
└── ...
```

### Integration Principle: Interfaces in Core, Implementations in `ee/`

The core defines interfaces (extension points). Community Edition uses default implementations (no-op / stubs). Enterprise Edition substitutes full implementations via build tags.

```go
// internal/platform/security/rls/territory.go (core, AGPL)
// Interface for territory-based access resolution.
type TerritoryResolver interface {
    ResolveTerritoryGroups(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

// internal/platform/security/rls/territory_default.go (core, AGPL)
//go:build !enterprise

// Default: territory is not used, returns nil.
type noopTerritoryResolver struct{}

func (r *noopTerritoryResolver) ResolveTerritoryGroups(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
    return nil, nil
}
```

```go
// ee/internal/platform/territory/resolver.go (enterprise, proprietary)
//go:build enterprise

// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License. See ee/LICENSE for details.

type territoryResolver struct { ... }

func (r *territoryResolver) ResolveTerritoryGroups(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
    // Full implementation: territory hierarchy, territory-based groups,
    // effective_user_territory cache lookup.
}
```

Vue components are connected similarly — via dynamic imports with feature flag checking:

```typescript
// web/src/router/index.ts (core, AGPL)
const routes = [
  ...coreRoutes,
  // Enterprise routes are loaded dynamically if available
  ...(import.meta.env.VITE_ENTERPRISE === 'true' ? enterpriseRoutes : []),
]
```

```typescript
// ee/web/src/router/enterprise-routes.ts (enterprise, proprietary)
export const enterpriseRoutes = [
  { path: '/admin/territories', component: () => import('../views/TerritoryManager.vue') },
  { path: '/admin/audit-log', component: () => import('../views/AuditLog.vue') },
  { path: '/admin/sso', component: () => import('../views/SSOConfig.vue') },
]
```

Enterprise migrations are run via a separate migration path:

```makefile
# Community Edition migrations
migrate-up:
	migrate -path migrations/ -database $(DB_URL) up

# Enterprise Edition migrations (core + enterprise)
migrate-up-ee:
	migrate -path migrations/ -database $(DB_URL) up
	migrate -path ee/migrations/ -database $(DB_URL) up
```

### License Marking in Code

Files in `ee/` contain the header:

```go
// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.
```

Files outside `ee/` optionally contain:

```go
// Copyright 2026 Adverax.
// Licensed under AGPL v3. See LICENSE for details.
```

### Build

Enterprise features are connected via Go build tags:

```go
// ee/territory/manager.go
//go:build enterprise

package territory
```

Two build variants:

```makefile
# Community Edition (default)
build:
	go build -o crm ./cmd/api

# Enterprise Edition
build-ee:
	go build -tags enterprise -o crm-ee ./cmd/api
```

### Legal Protection

- **AGPL on core** — a competitor hosting a modified version must open all of their code
- **Proprietary license on `ee/`** — use without payment = copyright violation
- **B2B context** — target clients (companies) comply with licenses due to legal risks
- **Precedents** — GitLab, Mattermost, Sourcegraph have successfully used this model for years

### Industry Precedents

| Project | Model | Core License | Enterprise |
|---------|-------|--------------|------------|
| GitLab | Open Core, single repo | MIT | `ee/` — proprietary |
| Mattermost | Open Core, single repo | MIT + Apache 2.0 | `enterprise/` — proprietary |
| Grafana | Open Core, single repo | AGPL v3 | Enterprise plugins — proprietary |
| Sourcegraph | Open Core, single repo | Apache 2.0 | `enterprise/` — proprietary |

## Consequences

- The entire security engine (OLS, FLS, RLS, groups, caching) is implemented in the core under AGPL — no splitting
- The `ee/` directory mirrors the main structure: `ee/internal/`, `ee/migrations/`, `ee/sqlc/`, `ee/web/`, `ee/tests/`
- The core defines interfaces (extension points), Community Edition uses no-op stubs (`//go:build !enterprise`), Enterprise substitutes full implementations (`//go:build enterprise`)
- Vue enterprise components are connected via dynamic imports and feature flag `VITE_ENTERPRISE`
- Enterprise migrations use a separate migration path (`ee/migrations/`), run after core migrations
- `LICENSE` file (AGPL v3) is in the repository root
- `ee/LICENSE` file (Adverax Commercial License) is in the `ee/` directory
- Build tag `enterprise` is used for conditional compilation of Go enterprise code
- Makefile gets targets: `build-ee`, `migrate-up-ee`, `test-pgtap-ee`
- Current development (Phase 2: Security engine) is not affected — all security goes into core
- Public release is planned after Phase 5-6 (auth + standard objects)
- First enterprise feature (territory management) is implemented in Phase N
