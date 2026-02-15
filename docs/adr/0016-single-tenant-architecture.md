# ADR-0016: Single-Tenant Architecture

**Status:** Accepted
**Date:** 2026-02-11
**Participants:** @roman_myakotin

## Context

The platform is designed as a self-hosted CRM for B2B companies (ADR-0014: Open Core, AGPL).
We need to establish the tenancy model, as it fundamentally affects:
- database schema (shared schema vs isolated DB)
- security engine (RLS policies, data isolation)
- metadata engine (table-per-object, DDL at admin time — ADR-0007)
- infrastructure and deployment model

Key factors:

1. **Self-hosted focus.** The product is positioned as self-hosted: each customer deploys their own instance. Multi-tenant architecture is excessive for this model.

2. **Simplicity for MVP.** Multi-tenancy adds cross-cutting complexity: tenant_id in every table, tenant-aware caches, tenant isolation in SOQL/DML, tenant-specific migrations. This would significantly slow down delivery of Phase 3–8.

3. **Security isolation.** B2B customers require full data isolation (compliance, GDPR, regulatory requirements). Physical isolation (separate DB per instance) is simpler for auditing and certification than logical isolation via row-level tenant filtering.

4. **Architectural compatibility.** Table-per-object (ADR-0007) generates DDL when creating objects. In a multi-tenant shared schema, one tenant's DDL affects everyone. This creates locks and complicates migrations.

## Options Considered

### Option A — Multi-tenant: shared database, shared schema

All tenants in one DB, isolation via a `tenant_id` column in every table.

**Pros:**
- Resource savings with a large number of small customers
- Single deployment, one DB to maintain

**Cons:**
- Cross-cutting complexity: `tenant_id` in every query, every index, every cache
- Risk of data leakage from filtering mistakes
- DDL from table-per-object (ADR-0007) blocks all tenants
- Metadata engine must be tenant-aware (separate object_definitions per tenant)
- Complicates SOQL/DML: every query must filter by tenant_id
- Cannot give the customer superuser access to the DB
- Auditing and compliance significantly harder

### Option B — Multi-tenant: shared database, separate schemas

Each tenant gets a separate PostgreSQL schema (`tenant_123.contacts`).

**Pros:**
- Better isolation than shared schema
- Native PostgreSQL `search_path` support

**Cons:**
- DDL from table-per-object is still risky (catalog-level locks)
- Migrations must run across all schemas — O(tenants) time
- Caches (metadata, security) must be per-schema
- PostgreSQL limitations: thousands of schemas with thousands of tables slow down `pg_catalog`
- Does not simplify deployment — still one DB

### Option C — Single-tenant: one instance per customer (chosen)

Each customer gets a fully isolated application instance + DB.

**Pros:**
- Full data isolation — trivial for compliance and auditing
- No `tenant_id` anywhere — simpler code, fewer bugs, higher performance
- Table-per-object DDL is safe — affects only one customer
- Metadata engine, SOQL/DML, security caches — all work without tenant-awareness
- Customer can get superuser access to their own DB
- Independent migrations, backups, scaling
- Natural fit for the self-hosted model (ADR-0014)
- Simpler for MVP — can focus on business logic

**Cons:**
- More infrastructure in a SaaS model (separate DB per customer)
- No resource sharing between customers
- Managed SaaS would require an orchestrator (Kubernetes, Terraform)

## Decision

**Option C chosen — single-tenant architecture.**

One application instance serves one organization. Each deployment includes:
- Its own PostgreSQL instance (or a separate DB)
- Its own Redis
- Its own API server

### Consequences for Code

- **No `tenant_id`** — not in tables, not in queries, not in caches
- **Metadata engine** (ADR-0003, ADR-0007) works without changes
- **Security engine** (ADR-0009–0013) — RLS/OLS/FLS without tenant filtering
- **SOQL/DML** — queries contain no tenant predicates
- **Migrations** — standard `golang-migrate`, no tenant loop
- **Configuration** — via environment variables (`.env`), unique per instance

### Path to SaaS (if needed)

Single-tenant does not close the path to a managed cloud offering:
- **Kubernetes + Helm chart** — each customer = namespace with a separate deployment
- **Database-per-tenant** pattern (as opposed to schema-per-tenant) is cheap to orchestrate
- **Terraform/Pulumi** — provisioning automation
- Many enterprise SaaS products (Atlassian Data Center, GitLab Dedicated) use single-tenant

## Related Decisions

- [ADR-0007: Table-per-object](0007-table-per-object-storage.md) — DDL at admin time, safe in single-tenant
- [ADR-0009: Security architecture](0009-security-architecture-overview.md) — 3-layer security without tenant-awareness
- [ADR-0014: Open Core](0014-licensing-and-business-model.md) — self-hosted focus confirms single-tenant
