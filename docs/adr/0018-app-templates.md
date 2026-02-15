# ADR-0018: App Templates Instead of Standard Objects

**Status:** Accepted
**Date:** 2026-02-14
**Participants:** @roman_myakotin

## Context

Phase 6 originally assumed the creation of hardcoded "standard objects" (Account, Contact, Opportunity, Task) via seed migrations. However, CRM is a metadata-driven platform (ADR-0003, ADR-0007), and different domains require different sets of objects:

- **Sales CRM**: Account, Contact, Opportunity, Task
- **Recruiting**: Position, Candidate, Application, Interview
- **Real Estate**: Property, Client, Showing, Deal
- **IT Service Desk**: Ticket, Asset, SLA, Knowledge Article

Hardcoding a single set of domain entities contradicts the platform's horizontal architecture. A mechanism is needed that allows the administrator to choose the appropriate set of objects on first launch.

Requirements:
1. Template is applied via Admin UI, not at bootstrap
2. One-time application (if `object_definitions` is not empty — blocked)
3. MVP: 2 templates (Sales CRM, Recruiting)
4. Templates are embedded in the binary (not external files)
5. Object/field creation — via existing ObjectService/FieldService (with DDL, constraints, share tables)

## Options Considered

### Option A — Hardcoded seed migrations

SQL migration that creates standard objects on `migrate up`.

**Pros:**
- Simple to implement
- Objects are available immediately after migration

**Cons:**
- Forces one domain on all users
- Cannot choose the set of objects
- Bypasses ObjectService/FieldService -> no DDL, share tables, OLS/FLS
- Migration is irreversible without manual intervention

### Option B — JSON/YAML template files

Templates stored as JSON/YAML files, read at application time.

**Pros:**
- Easy to add new templates
- Can be edited without recompilation

**Cons:**
- No compile-time validation of structure
- Requires deserialization with error handling
- Files must be shipped alongside the binary
- Harder to test

### Option C — Go code embedded in the binary (chosen)

Templates defined as Go structs, compiled into the binary.

**Pros:**
- Compile-time type safety
- No external dependencies (single binary)
- Easy to test (unit tests on structs)
- IDE auto-completion

**Cons:**
- Requires recompilation to add templates
- Not suitable for user-defined templates (but this is not an MVP requirement)

## Decision

**Option C chosen: Go code embedded in the binary.**

### Architecture

```
internal/platform/templates/
├── types.go         — Template, ObjectTemplate, FieldTemplate
├── registry.go      — Registry (map[string]Template)
├── all.go           — BuildRegistry() → registers all templates
├── applier.go       — Applier: two-pass creation via services
├── sales_crm.go     — SalesCRM() → Template
└── recruiting.go    — Recruiting() → Template
```

### Registry + Applier Pattern

- `Registry` stores all available templates in a `map[string]Template`
- `Applier` applies a template using existing `ObjectService.Create()` and `FieldService.Create()`
- Two-pass creation: first all objects (collecting `map[apiName]UUID`), then all fields (resolving reference -> UUID)
- After creation: OLS (full CRUD) + FLS (full RW) for SystemAdmin PS on all new objects/fields
- Guard: `objectRepo.Count(ctx) > 0` -> `apperror.Conflict`

### API

- `GET /api/v1/admin/templates` — list templates with status (available/applied/blocked)
- `POST /api/v1/admin/templates/:templateId/apply` — apply a template

## Consequences

### Positive
- Phase 6 scope changes from "standard objects" to "App Templates" — the platform remains horizontal
- Administrator chooses the domain on first launch
- Easy to add new templates (HR, Real Estate, IT Service Desk)
- All objects are created via platform services -> automatic DDL, share tables, constraints

### Negative
- One-time application (MVP limitation) — cannot apply a second template
- No UI for customizing the template before application
- No "standard" objects in the classic Salesforce sense (is_platform_managed=true)

### Future Extensions
- User-defined templates (JSON export/import)
- Partial application (choosing objects from a template)
- Template marketplace
