# ADR-0034: Industry Modules (Vertical Extensibility)

**Date:** 2026-02-25

**Status:** Accepted

**Participants:** @roman_myakotin

## Context

### Problem: App Templates are declarative-only

The CRM platform uses App Templates (ADR-0018) to bootstrap domain-specific
object sets (Sales CRM, Recruiting). Templates are Go-embedded structs that
create objects and fields via existing services. This works well for
**declarative content** — objects, fields, OLS/FLS grants.

However, real vertical solutions require more:

| Capability | App Template | Industry Module |
|-----------|-------------|----------------|
| Object + field definitions | Yes | Yes |
| Object Views, Procedures, Automation Rules | No | Yes |
| Custom HTTP endpoints | No | Yes |
| Custom business logic (Go code) | No | Yes |
| DML lifecycle hooks | No | Yes |
| Incremental install (add to existing system) | No (count == 0 guard) | Yes |
| Domain-specific algorithms | No | Yes |

Examples of vertical functionality that cannot be expressed declaratively:

- **Pharmaceutical sales ERP**: regulatory compliance checks, batch/lot
  tracking with expiry validation, controlled substance audit trail,
  custom reporting endpoints.
- **Logistics**: route optimization algorithm, delivery window calculation,
  GPS-based proof-of-delivery endpoint.
- **Real estate**: mortgage calculator, MLS integration endpoint, property
  valuation model.
- **FMCG distribution** (ADR-0033): sync endpoints, offline protocol,
  domain-aware conflict resolution.

### Relationship with existing systems

| System | Role | Unchanged? |
|--------|------|-----------|
| **App Templates** (ADR-0018) | Bootstrap empty system with declarative objects | Yes — modules complement, not replace |
| **Procedure Engine** (ADR-0024) | Declarative automation DSL | Yes — modules can install procedures |
| **Automation Rules** (ADR-0031) | Reactive DML hooks | Yes — modules can install rules |
| **DML Pipeline** (ADR-0020) | 8-stage data processing | Extended — composite post-execute hooks |
| **Modular Monolith** (ADR-0030) | Interface boundaries | Aligned — modules consume MetadataReader |

### Industry context

| Platform | Extension model | Granularity |
|----------|----------------|-------------|
| **Salesforce** | Managed Packages (AppExchange) | Runtime-installed, Apex code |
| **Dynamics 365** | Solutions + Plugins (C#) | Runtime-installed, compiled plugins |
| **Odoo** | Python modules (addons/) | File-system modules, auto-discovered |
| **ERPNext** | Frappe apps | Python apps, pip-installed |
| **Our platform** | **Industry Modules** | Compiled-in Go modules, build-time |

Our approach is closest to Odoo's addons model but with Go's compile-time
safety. Runtime installation (like Salesforce Managed Packages) is deferred —
it requires a sandbox VM, which is premature for the current stage.

## Considered Options

### Option A: Extend App Templates only

Add more capabilities to the existing template structs (procedures, views,
automation rules as data).

**Pros:**
- No new abstraction
- Builds on familiar pattern

**Cons:**
- Cannot express custom Go logic (HTTP endpoints, algorithms)
- Cannot add DML hooks
- Template guard (`count == 0`) prevents incremental installation
- Struct-based templates become unwieldy with 10+ entity types
- No dependency management between templates

### Option B: Package Manager (dynamic loading)

Runtime-installable packages with a plugin interface, similar to Salesforce
Managed Packages.

**Pros:**
- Install/uninstall without recompilation
- Marketplace potential
- Familiar to Salesforce users

**Cons:**
- Go plugin system is fragile (`.so` only, same Go version, same flags)
- Requires sandbox/isolation for untrusted code
- Complex versioning and compatibility matrix
- Massive engineering effort — premature for current stage
- Security implications of running third-party code

### Option C: Industry Modules with Module interface (Chosen)

Compiled-in Go modules that implement a standard interface. Modules combine
declarative content (objects, fields, views, procedures) with custom Go code
(endpoints, hooks, algorithms). Registered at build time, installed at
runtime via admin action.

**Pros:**
- Full Go power — any algorithm, any integration
- Compile-time type safety
- Reuses all platform capabilities (security, SOQL/DML, procedures)
- Clean interface boundary (Module interface)
- Zero duplication — modules call platform services, not raw SQL
- Dual catalog: open-source modules (AGPL) + commercial modules (proprietary)
- Incremental installation — modules add to existing system

**Cons:**
- Requires Go developer to create a module
- API stability commitment — Module and Platform interfaces become public API
- Compiled-in only — no runtime install/uninstall
- Binary size grows with each compiled module

## Decision

### Module interface

Defined in core (`internal/platform/module/`), available to all modules:

```go
package module

// Module is the interface that all industry modules must implement.
type Module interface {
    // Identity
    Name() string           // unique identifier, e.g. "fmcg-distribution"
    Description() string    // human-readable description
    Version() string        // semver, e.g. "1.0.0"
    Dependencies() []string // names of required modules (resolved before install)

    // Declarative content — objects, fields, views, procedures, automation rules.
    // Called once during installation. Platform services handle DDL, constraints,
    // share tables, OLS/FLS grants.
    OnInstall(ctx context.Context, platform Platform) error

    // Custom HTTP endpoints — registered under /api/v1/modules/{name}/.
    // Called on every application start (not just install).
    RegisterRoutes(group *gin.RouterGroup)

    // DML lifecycle hooks — composable chain, not single-slot.
    // Called on every application start. Returned hooks are appended to
    // the composite post-execute hook chain.
    PostExecuteHooks() []dml.PostExecuteHook
}
```

### Platform struct

The Platform struct exposes core services to modules. Modules MUST use these
services — never raw SQL, never direct DB access. This ensures security
enforcement (OLS/FLS/RLS) and metadata consistency.

```go
package module

// Platform provides access to core platform services.
// Modules use this struct to interact with the CRM platform.
type Platform struct {
    Pool              *pgxpool.Pool
    MetadataCache     metadata.MetadataReader
    SOQLService       soql.QueryService
    DMLService        dml.DMLService
    OLSEnforcer       ols.Enforcer
    FLSEnforcer       fls.Enforcer
    ObjectService     metadata.ObjectService
    FieldService      metadata.FieldService
    PermissionService security.PermissionService
    CacheInvalidator  metadata.CacheInvalidator
}
```

Platform struct may grow over time (e.g., ProcedureService, AutomationService).
Additions are backward-compatible (modules that don't use new fields are
unaffected).

### Module Registry and lifecycle

```go
package module

// Registry collects modules and orchestrates their lifecycle.
type Registry struct {
    modules map[string]Module
    order   []string // topological order after dependency resolution
}

func NewRegistry() *Registry { ... }

// Register adds a module to the registry. Called at build time (main.go).
func (r *Registry) Register(m Module) error { ... }

// InstallAll resolves dependencies and installs uninstalled modules.
// Tracks installation status in metadata.module_installations.
func (r *Registry) InstallAll(ctx context.Context, platform Platform) error { ... }

// RegisterAllRoutes registers HTTP routes for all installed modules.
func (r *Registry) RegisterAllRoutes(router *gin.RouterGroup) { ... }

// CollectPostExecuteHooks returns all hooks from all installed modules.
func (r *Registry) CollectPostExecuteHooks() []dml.PostExecuteHook { ... }
```

#### Lifecycle sequence

```
Application Start
│
├── 1. Register: main.go calls registry.Register(module) for each module
│
├── 2. Resolve dependencies: topological sort, detect cycles → fail fast
│
├── 3. Install (once per module):
│      ├── Check metadata.module_installations for each module
│      ├── If not installed → call module.OnInstall(ctx, platform)
│      │   ├── Module creates objects, fields via platform.ObjectService
│      │   ├── Module creates procedures, automation rules via platform services
│      │   └── On success: INSERT into module_installations (name, version, status='installed')
│      └── If already installed → skip
│
├── 4. RegisterRoutes: call module.RegisterRoutes(group) for all installed modules
│      └── Routes mounted under /api/v1/modules/{module.Name()}/
│
└── 5. RegisterHooks: collect PostExecuteHooks from all modules
       └── Append to CompositePostExecuteHook (see below)
```

#### Installation tracking

```sql
CREATE TABLE metadata.module_installations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         TEXT NOT NULL UNIQUE,
    version      TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'installed',  -- installed | failed
    installed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

No `count == 0` guard — modules are designed for incremental installation
into an existing system. A module can be installed alongside App Templates
or other modules.

### CompositePostExecuteHook

The current DML pipeline has a single-slot `PostExecuteHook` (occupied by
AutomationEngine's `DMLPostExecuteHook`). Modules need to add their own
hooks without displacing automation.

```go
package dml

// CompositePostExecuteHook chains multiple PostExecuteHook implementations.
// Hooks execute in registration order. If any hook returns an error,
// execution stops and the error propagates (same TX rollback behavior
// as the existing single hook).
type CompositePostExecuteHook struct {
    hooks []PostExecuteHook
}

func NewCompositePostExecuteHook(hooks ...PostExecuteHook) *CompositePostExecuteHook {
    return &CompositePostExecuteHook{hooks: hooks}
}

func (c *CompositePostExecuteHook) Append(hook PostExecuteHook) {
    c.hooks = append(c.hooks, hook)
}

func (c *CompositePostExecuteHook) AfterDMLExecute(
    ctx context.Context,
    compiled *engine.CompiledDML,
    result *engine.Result,
) error {
    for _, hook := range c.hooks {
        if err := hook.AfterDMLExecute(ctx, compiled, result); err != nil {
            return err
        }
    }
    return nil
}
```

`CompositePostExecuteHook` satisfies the existing `PostExecuteHook` interface.
The DMLService continues to use `SetPostExecuteHook()` — the argument is
now a composite instead of a single hook. **Fully backward-compatible.**

Boot sequence change:

```go
// Before (single hook):
dmlService.SetPostExecuteHook(automationHook)

// After (composite):
composite := dml.NewCompositePostExecuteHook(automationHook)
for _, hook := range moduleRegistry.CollectPostExecuteHooks() {
    composite.Append(hook)
}
dmlService.SetPostExecuteHook(composite)
```

Automation hook always executes first (position 0), module hooks follow
in module registration order.

### File structure

```
internal/platform/module/          ← interface + registry + lifecycle (core, AGPL)
├── module.go                      ← Module interface, Platform struct
├── registry.go                    ← Registry, dependency resolution
└── lifecycle.go                   ← Installation tracking, idempotency

modules/                           ← open-source modules (AGPL v3)
├── fmcg/                          ← example: FMCG Distribution
│   ├── module.go                  ← implements Module interface
│   ├── routes.go                  ← custom endpoints
│   ├── hooks.go                   ← DML hooks
│   ├── install.go                 ← OnInstall: objects, fields, procedures
│   └── migrations/                ← module-specific migrations
│       ├── 000001_sync_schema.up.sql
│       └── 000001_sync_schema.down.sql
└── ...

ee/modules/                        ← commercial modules (Adverax Commercial License)
├── pharma/                        ← example: Pharmaceutical Sales
│   ├── module.go
│   ├── ...
│   └── migrations/
└── ...

cmd/api/
├── modules.go                     ← core module registration
├── modules_enterprise.go          ← ee module registration (//go:build enterprise)
└── modules_stub.go                ← stub (//go:build !enterprise)
```

#### Registration files

```go
// cmd/api/modules.go
package main

func registerCoreModules(registry *module.Registry) {
    // Register open-source modules here
    // registry.Register(fmcg.NewModule())
}
```

```go
// cmd/api/modules_enterprise.go
//go:build enterprise

package main

func registerEnterpriseModules(registry *module.Registry) {
    // Register proprietary modules here
    // registry.Register(pharma.NewModule())
}
```

```go
// cmd/api/modules_stub.go
//go:build !enterprise

package main

func registerEnterpriseModules(registry *module.Registry) {
    // No-op: enterprise modules not available in core build
}
```

### Route namespace

Modules register routes under a dedicated namespace:

```
/api/v1/modules/{module-name}/
```

Examples:
- `/api/v1/modules/fmcg/sync/push`
- `/api/v1/modules/fmcg/sync/pull`
- `/api/v1/modules/pharma/compliance/check`

This prevents route collisions between modules and between modules and
core endpoints. All module routes go through the standard JWT auth
middleware.

### Migration isolation

Each module can have its own `migrations/` directory with a separate
migration tracking table, following the same pattern as `ee/migrations/`:

```bash
# Core migrations
migrate -path migrations/ -database $DB_URL up

# Module migrations (separate tracking table)
migrate -path modules/fmcg/migrations/ \
    -database "$DB_URL&x-migrations-table=module_fmcg_migrations" up
```

This ensures:
- Module migrations don't interfere with core migrations
- Modules can create their own schemas (e.g., `sync` schema for FMCG)
- Module migrations have independent UP/DOWN lifecycle

### Licensing

| Path | License | Header required |
|------|---------|----------------|
| `internal/platform/module/` | AGPL v3 | No (core) |
| `modules/` | AGPL v3 | No (core) |
| `ee/modules/` | Adverax Commercial License | Yes (proprietary header) |

Enterprise modules follow the existing `ee/` licensing pattern (ADR-0014):

```go
// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.
```

### Platform limits

| Parameter | Default | Description |
|-----------|---------|-------------|
| Max modules per installation | 20 | Prevents unbounded binary growth |
| Max DML hooks per module | 5 | Bounds composite hook chain length |
| Module install timeout | 60s | Per-module OnInstall timeout |
| Max route groups per module | 10 | Prevents route namespace pollution |

### Security considerations

| Concern | Mitigation |
|---------|-----------|
| Module bypasses OLS/FLS/RLS | Platform struct exposes only service interfaces — no raw DB access encouraged. Module code review is the ultimate gate (compiled-in). |
| Module route auth | All routes go through JWT middleware (inherited from parent router group). |
| Module migration safety | Separate migration table prevents interference with core schema. |
| Malicious module code | Compiled-in modules require code review before merge. No runtime plugin loading. |
| Dependency cycle | Registry resolves dependencies topologically; cycles are detected and rejected at startup. |

## Consequences

### Positive

- **Full Go power** — modules can implement any algorithm, any integration,
  any custom endpoint while reusing all platform capabilities.
- **Zero duplication** — modules call platform services (SOQL, DML,
  ObjectService) instead of reimplementing business logic.
- **Clean boundaries** — Module interface + Platform struct form a stable
  contract between core and extensions.
- **Dual catalog** — open-source modules in `modules/`, commercial in
  `ee/modules/`, same pattern as existing `ee/` split.
- **Incremental installation** — modules add to an existing system, no
  `count == 0` guard, composable with App Templates.
- **Backward-compatible DML extension** — CompositePostExecuteHook chains
  automation + module hooks without changing the DMLService interface.
- **Compile-time safety** — Go compiler catches interface violations,
  type mismatches, missing dependencies.

### Negative

- **Requires Go developer** — creating a module requires Go programming
  skills; no low-code module builder.
- **API stability commitment** — Module and Platform interfaces become
  public API; breaking changes require migration guides.
- **Compiled-in only** — no runtime install/uninstall; adding a module
  requires rebuild and redeploy.
- **Binary size** — each module increases the binary; bounded by the
  20-module limit.
- **Code review gate** — since modules run in-process with full trust,
  every module must be reviewed before inclusion.

### Future extensions

- **Module marketplace** — curated registry of community modules.
- **Module versioning** — upgrade path when module version changes
  (OnUpgrade method on Module interface).
- **Module configuration** — per-installation settings (e.g., FMCG module
  configured for a specific warehouse layout).
- **Runtime plugins** — if/when Go plugin ecosystem matures or a WASM
  sandbox becomes viable.

### Related ADRs

- ADR-0018: App Templates (declarative-only predecessor, complementary)
- ADR-0020: DML Pipeline Extension (Stage 8 post-execute, composite hook)
- ADR-0024: Procedure Engine (modules can install procedures)
- ADR-0030: Modular Monolith (MetadataReader interface, module boundaries)
- ADR-0031: Automation Rules (first PostExecuteHook consumer, now in composite chain)
- ADR-0033: Offline Sync Protocol (candidate for first FMCG module)
- ADR-0014: Licensing (dual catalog mirrors ee/ pattern)
