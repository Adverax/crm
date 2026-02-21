# ADR-0030: Modular Monolith — Microservices Readiness

**Date:** 2026-02-21

**Status:** Accepted

## Context

The CRM platform is a well-structured monolith with clean layering
(handler → service → SOQL/DML → repository). However, three coupling
hotspots prevent future extraction into independent services:

1. **`*metadata.MetadataCache` is a concrete type** — 13 consumers in 6
   packages depend on the struct directly, not an interface. This is the
   #1 coupling point in the system.
2. **`security.UserContext`** — defined in the `security` package, forcing
   every consumer (SOQL, DML, record service, middleware) to import
   `security` for a 3-field struct.
3. **`PgMetadataFieldLister`** — the `security` package reads
   `metadata.*` tables directly via raw SQL (5 queries), violating schema
   ownership boundaries.

### Drivers

- **Team scaling:** independent ownership zones for Identity, Metadata,
  Security, and Data Engine.
- **Technical scaling:** Data Engine (SOQL/DML) may need to scale
  independently of Admin/Metadata operations.
- **Architectural hygiene:** testability (mock interfaces instead of
  concrete structs), clear module boundaries, schema ownership.

### Five Bounded Contexts Identified

| Bounded Context   | Owns                        | Extraction Difficulty |
|-------------------|-----------------------------|-----------------------|
| Identity & Auth   | `iam.*` tables, JWT, login  | MEDIUM                |
| Metadata Admin    | `metadata.*` tables, cache  | HARD                  |
| Security Engine   | `security.*`, OLS/FLS/RLS   | VERY HARD             |
| Data Engine       | `public.obj_*`, SOQL/DML    | VERY HARD             |
| Automation (future)| procedures, scenarios      | EASY                  |

## Decision

**Modular Monolith** — extract interfaces and shared kernel NOW, keep
single binary. This gives ~90% of architectural benefits at ~10% of
operational cost compared to microservices.

### What we implement now (P0–P1)

1. **P0: `MetadataReader` interface** in `internal/platform/metadata/`.
   All consumers depend on this interface instead of `*MetadataCache`.
   The concrete cache satisfies it automatically.

2. **P0: `internal/pkg/identity` shared kernel.** Move `UserContext`,
   `ContextWithUser`, `UserFromContext` to a lightweight package with
   zero dependencies. The `security` package re-exports via type/var
   aliases for backward compatibility.

3. **P1: Replace `PgMetadataFieldLister`** (raw SQL against
   `metadata.*`) with `CacheBackedMetadataLister` that delegates to
   `MetadataReader`. Eliminates cross-schema SQL from the security
   package.

### What we defer (P2–P3)

4. **P2: Generalized event bus** from the existing outbox pattern.
5. **P3: Per-module narrow interfaces** at consumer sites (e.g.,
   `soql.MetadataProvider` wrapping only the methods SOQL needs).

### Split Criteria — When to actually extract a service

- Independent scaling requirement (Data Engine vs Admin).
- Independent deployment cadence (separate teams).
- Organizational boundary (separate ownership).
- Technology divergence (e.g., workflow runtime for Automation).

**Important:** Start extraction with Automation (EASY), NOT
Metadata/Security (VERY HARD).

## Consequences

### Positive

- All 13 consumers of `*MetadataCache` depend on an interface → trivial
  to mock in tests, swap implementations, or route through a cache proxy.
- `security.UserContext` lives in a shared kernel → no forced imports of
  the security package for a 3-field struct.
- Security package no longer runs raw SQL against `metadata.*` tables →
  clean schema ownership.
- Single binary, single deploy, shared PostgreSQL — zero operational
  overhead increase.
- Future extraction path is clear: interface boundaries are already in
  place.

### Negative

- Interface adds one level of indirection (negligible performance cost).
- Shared kernel (`identity`) is a new package to maintain.
- Type aliases in `security` for backward compatibility add slight
  complexity.

### Anti-patterns to avoid

- **Distributed monolith:** sync calls between "services" sharing a DB.
- **Premature gRPC/protobuf** between in-process modules.
- **Extracting the wrong service first:** start with Automation, NOT
  Metadata/Security.
- **Breaking the SOQL/DML security invariant:** security enforcement
  must remain in the data path regardless of module boundaries.
