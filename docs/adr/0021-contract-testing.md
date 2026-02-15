# ADR-0021: Contract Testing — OpenAPI Validation + TS Type Generation

**Status:** Accepted
**Date:** 2026-02-14
**Participants:** @roman_myakotin

## Context

The contract between backend and frontend is defined in `api/openapi.yaml` (OpenAPI 3.0.3, 1700+ lines). Before this decision, there were three points of desynchronization:

1. **OpenAPI spec <-> Go handlers** — the spec was updated manually; nothing verified that the actual JSON response from a handler matched the described schema. Handler tests only checked the status code.
2. **OpenAPI spec <-> TypeScript types** — types in `web/src/types/` were written manually and drifted (we had already caught snake_case/camelCase bugs).
3. **Go generated types <-> OpenAPI spec** — `oapi-codegen` generated Go types, but without `required` arrays in the spec all fields were pointer types, masking real errors.

The cost of desynchronization was growing: 20+ endpoints, 15+ entity schemas, 215 e2e tests. Manual synchronization stopped scaling.

Requirements:
1. Automatic verification that HTTP responses match the OpenAPI schema in Go handler tests
2. A single source of truth for TypeScript types (no manual duplication)
3. A single command to regenerate Go and TS types: `make generate-api`
4. Zero new runtime dependencies (dev/test only)
5. Backward compatibility with existing tests

## Considered Options

### Option A — Manual Synchronization (status quo)

Continue manually maintaining correspondence between spec <-> code <-> types.

**Pros:**
- No additional complexity
- No new tools

**Cons:**
- Drift is inevitable as the API grows (20+ endpoints)
- Errors are discovered only at runtime or in e2e tests
- Double work: describe in spec + write TypeScript interface
- No guarantee that a handler returns what is described in the spec

### Option B — Generate OpenAPI from Code (code-first)

Generate the OpenAPI spec from Go structs or handlers (swaggo, go-swagger).

**Pros:**
- Spec always matches the code
- No need to maintain the spec manually

**Cons:**
- Loss of control over API design (spec as a side effect)
- Spec is tied to implementation, not to the contract
- Complex annotation setup in code
- Contradicts the spec-first approach adopted in the project

### Option C — Spec-first with Contract Validation in Tests (chosen)

The OpenAPI spec remains the source of truth. Go handler tests validate responses against the spec via `kin-openapi`. TypeScript types are generated from the spec via `openapi-typescript`.

**Pros:**
- Spec = single contract, controlled by the developer
- Automatic drift detection in both directions (Go and TS)
- Zero runtime dependencies (`kin-openapi` is already in go.mod as transitive)
- Existing tests gain contract validation without modifications
- TypeScript types are always up to date — manual drift is eliminated

**Cons:**
- Spec needs to be updated with every API change (but that is the goal)
- Adding `required` arrays to the spec changes Go generated types (pointer -> value), requiring one-time handler fixes
- `openapi-typescript` — a new devDependency

### Option D — Third-party Contract Framework (Pact, Dredd)

Use specialized contract testing tools.

**Pros:**
- Rich functionality (consumer-driven contracts, broker)
- Industry standard for microservices

**Cons:**
- Overkill for a monolith with a single frontend
- Additional infrastructure (broker, CI integration)
- Duplicates what `kin-openapi` provides for free
- Learning curve

## Decision

**Option C chosen: Spec-first with contract validation in tests.**

### Architecture

```
api/openapi.yaml                    <- Source of truth (single contract)
    |
    |---> oapi-codegen               -> internal/api/openapi_gen.go (Go types + routes)
    |
    |---> openapi-typescript         -> web/src/types/openapi.d.ts (TS types)
    |    └---> CamelCaseKeys<T>      -> web/src/types/{metadata,auth,...}.ts (derived types)
    |
    └---> kin-openapi (in tests)     -> contractValidationMiddleware (response validation)
```

### Backend: Response validation middleware

File `internal/handler/testutil_contract_test.go` — shared middleware for all handler tests:

- `loadSpec()` loads the OpenAPI spec once via `sync.Once`
- `responseCapture` wraps `gin.ResponseWriter` to intercept the response body
- `contractValidationMiddleware(t)` — Gin middleware that validates every response:
  1. Finds the route in the spec via `gorillamux.Router`
  2. Constructs `RequestValidationInput` + `ResponseValidationInput`
  3. Calls `openapi3filter.ValidateResponse()`
  4. On mismatch — `t.Errorf()` with details

Each `setup*Router()` in tests accepts `t *testing.T` and adds the middleware:
```go
func setupRouter(t *testing.T, h *MetadataHandler) *gin.Engine {
    r := gin.New()
    r.Use(contractValidationMiddleware(t))
    // ...
}
```

### Frontend: Type Generation from OpenAPI

1. `openapi-typescript` generates `web/src/types/openapi.d.ts` from the spec
2. `CamelCaseKeys<T>` (`web/src/types/camelcase.ts`) converts snake_case keys to camelCase (the HTTP client does this at runtime)
3. Each type file (`metadata.ts`, `auth.ts`, `validationRules.ts`, `records.ts`) exports derived types:

```typescript
import type { components } from './openapi'
import type { CamelCaseKeys } from './camelcase'

export type ObjectDefinition = CamelCaseKeys<components['schemas']['ObjectDefinition']>
```

### Makefile: single generation entry point

```makefile
generate-api:
    oapi-codegen -generate gin,types,spec -package api -o internal/api/openapi_gen.go api/openapi.yaml
    cd web && npx openapi-typescript ../api/openapi.yaml -o src/types/openapi.d.ts
```

### Spec strictness: required arrays

Added `required` arrays to response entity schemas (`ObjectDefinition`, `FieldDefinitionSchema`, `ValidationRule`, `TokenPair`, `UserInfo`, `PaginationMeta`, `ObjectNavItem`, `ObjectDescribe`, `FieldDescribe`). This ensures:
- Go: value types instead of pointer types for required fields
- TS: non-optional properties instead of `field?: type`

### Affected Files

| File | Role |
|------|------|
| `internal/handler/testutil_contract_test.go` | New — loadSpec, middleware, responseCapture |
| `internal/handler/*_test.go` (5 files) | setup*Router now accepts `t` + middleware |
| `web/src/types/camelcase.ts` | New — `CamelCaseKeys<T>` utility type |
| `web/src/types/openapi.d.ts` | New — auto-generated types from OpenAPI |
| `web/src/types/{metadata,auth,validationRules,records}.ts` | Derived types instead of manual interfaces |
| `api/openapi.yaml` | Added `required` arrays, `nullable`, `enum` |
| `Makefile` | Updated `generate-api`, added `web-generate-types` |

## Consequences

### Positive
- **Drift is detected instantly**: changing the spec without updating the handler -> test fails; changing the spec without `make generate-api` -> TypeScript compilation error
- **Zero runtime dependencies**: all validation is in tests only
- **28 handler tests** automatically gained contract validation without changes to test logic
- **TypeScript types** are no longer written manually — one `make generate-api` updates everything
- **Contract smoke test**: change a field in the spec -> both Go tests and TS type-check break

### Negative
- When changing the API, the spec must be updated first (spec-first discipline)
- Adding `required` arrays to the existing spec required a one-time refactor of Go handlers (pointer -> value types)
- `openapi-typescript` — another devDependency in the frontend

### Developer Workflow

1. Modify `api/openapi.yaml` (add/change endpoint or schema)
2. `make generate-api` — regenerate Go and TS types
3. Update handler/frontend code to match new types (the compiler will guide you)
4. `go test ./internal/handler/...` — contract validation
5. `cd web && npm run type-check` — TypeScript verification

### Future Extensions
- Request validation middleware (validating incoming requests in tests)
- CI pipeline step: `make generate-api && git diff --exit-code` — verifying that spec and generated code are in sync
- Automatic mock data generation from OpenAPI schemas for e2e tests
