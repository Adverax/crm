# Roadmap: Path to a Salesforce-Grade Platform

**Date:** 2026-02-08
**Stack:** Go 1.25 · PostgreSQL 16 · Vue.js 3 · Redis
**Model:** Open Core (AGPL v3 + Enterprise `ee/`)

---

## Domain Maturity

Current state and target coverage relative to Salesforce Platform.

| Domain | Salesforce | Current State | Target Level |
|--------|-----------|---------------|--------------|
| Metadata Engine | Custom Objects, Fields, Relationships, Record Types, Layouts | Objects, Fields (all types), Relationships (assoc/comp/poly), Table-per-object DDL | 80% SF |
| Security (OLS/FLS) | Profile, Permission Set, Permission Set Group, Muting PS | Profile, Grant/Deny PS, OLS bitmask, FLS bitmask, effective caches | 90% SF |
| Security (RLS) | OWD, Role Hierarchy, Sharing Rules, Manual Sharing, Apex Sharing, Teams, Territory | OWD, Groups (4 types), Share tables, Role hierarchy, Sharing Rules (owner+criteria), Manual Sharing, RLS enforcer, effective caches, Territory Management (ee/) | 85% SF |
| Data Access (SOQL) | SOQL with relationship queries, aggregates, security enforcement | SOQL parser (participle), validator, compiler, executor with OLS+FLS+RLS enforcement, relationship queries, aggregates, date literals, subqueries | 70% SF |
| Data Mutation (DML) | Insert, Update, Upsert, Delete, Undelete, Merge + triggers | INSERT/UPDATE/DELETE/UPSERT, OLS+FLS enforcement, RLS injection for UPDATE/DELETE, batch operations, Custom Functions (fn.* dual-stack), validation rules (CEL), dynamic defaults (CEL) | 65% SF |
| Auth | OAuth 2.0, SAML, MFA, Connected Apps | JWT (access + refresh), login, password reset, rate limiting | JWT + refresh tokens |
| Automation | Flow Builder, Triggers, Workflow Rules, Approval Processes | Not implemented | Triggers + basic Flows |
| UI Framework | Lightning App Builder, LWC, Dynamic Forms | Vue.js admin + metadata-driven CRM UI (AppLayout, dynamic record views, FieldRenderer), Expression Builder (CodeMirror + autocomplete + live preview), Object View (role-based sections, actions, highlights, related lists) | Admin + Record UI + Object Views |
| APIs | REST, SOAP, Bulk, Streaming, Metadata, Tooling, GraphQL | REST admin endpoints (metadata + security + groups + sharing rules) | REST + Streaming |
| Analytics | Reports, Dashboards, Einstein | Not implemented | Basic reports |
| Integration | Platform Events, CDC, External Services | Not implemented | CDC + webhooks |
| Developer Tools | Apex, CLI, Sandboxes, Packaging | — | CLI + migration tools |
| Standard Objects | Account, Contact, Opportunity, Lead, Case, Task, etc. | App Templates (Sales CRM: 4 obj, Recruiting: 4 obj) | 6-8 core objects |

---

## Implementation Phases

### Phase 0: Scaffolding ✅

Project infrastructure.

- [x] Docker + docker-compose for local development
- [x] PostgreSQL 16 + pgTAP
- [x] Makefile, CI (GitHub Actions)
- [x] Project structure (cmd, internal, web, ee, migrations, tests)
- [x] HTTP client, routing (Gin), structured logging (slog)
- [x] Typed errors (apperror), pagination helpers
- [x] Basic Vue.js shell (AdminLayout, UI components)

---

### Phase 1: Metadata Engine ✅

Platform core — dynamic object and field definitions.

- [x] Object Definitions (standard/custom, behavioral flags)
- [x] Field Definitions (type/subtype, config, validation)
- [x] Field types: text, number, boolean, datetime, picklist, reference
- [x] Reference types: association, composition, polymorphic
- [x] Table-per-object: DDL generation (`obj_{api_name}`)
- [x] Constraints: FK, unique, not null, check
- [x] REST API: CRUD objects + fields
- [x] Vue.js admin: objects, fields, detail with tabs
- [x] pgTAP tests for schema

**What differs from Salesforce and will be added later:**

| SF Capability | Our Status | When |
|---------------|-----------|------|
| Record Types | Not implemented | Phase 9c |
| Object Views (role-based layouts) | ✅ Implemented (Phase 9a) | — |
| Compact Layouts (highlight fields) | ✅ Implemented (highlight_fields in OV config) | — |
| Formula Fields | Not implemented | Phase 12 |
| Roll-Up Summary Fields | Not implemented | Phase 12 |
| Validation Rules (formula-based) | Not implemented | Phase 12 |
| Field History Tracking | Not implemented | Phase N (ee/) |
| Custom Metadata Types (`__mdt`) | Not implemented | Phase 14 |
| Big Objects | Not implemented | Far future |
| External Objects | Not implemented | Far future |

---

### Phase 2: Security Engine ✅

Three security layers — the foundation of an enterprise-grade platform.

#### Phase 2a: Identity + Permission Engine ✅

- [x] User Roles (hierarchy via parent_id)
- [x] Permission Sets (grant/deny, bitmask)
- [x] Profiles (auto-created base PS)
- [x] Users (username, email, profile, role, is_active)
- [x] Permission Set Assignments (user ↔ PS)
- [x] Object Permissions (OLS: CRUD bitmask 0-15)
- [x] Field Permissions (FLS: RW bitmask 0-3)
- [x] Effective caches (effective_ols, effective_fls)
- [x] Outbox pattern for cache invalidation
- [x] REST API: full CRUD for all entities
- [x] Vue.js admin: roles, PS, profiles, users, OLS/FLS editor

#### Phase 2b: RLS Core ✅

Row-Level Security — who can see which records.

- [x] Org-Wide Defaults (OWD) per object: private, public_read, public_read_write, controlled_by_parent
- [x] Share tables: `obj_{name}__share` (grantee_id, access_level, share_reason)
- [x] Role Hierarchy: closure table `effective_role_hierarchy`
- [x] Sharing Rules (ownership-based): source group → target group, access level
- [x] Sharing Rules (criteria-based): field conditions → target group, access level
- [x] Manual Sharing: owner/admin shares a record with a specific user/group
- [x] Record ownership model: OwnerId on every record
- [x] Effective visibility cache: `effective_visible_owners`
- [x] REST API: OWD settings, sharing rules CRUD, manual sharing
- [x] Vue.js admin: OWD settings (visibility in object create/edit), sharing rules UI (list/create/detail)
- [x] E2E tests: sharing rules (14 tests), visibility in objects

#### Phase 2c: Groups ✅

Groups — unified grantee for all sharing operations.

- [x] Group types: personal, role, role_and_subordinates, public
- [x] Auto-generation: on user creation → personal group; on role creation → role group + role_and_sub group
- [x] Public group: admin creates, adds members (users, roles, other groups)
- [x] Effective group members cache: `effective_group_members` (closure table)
- [x] Unified grantee (always group_id) for share tables and sharing rules
- [x] REST API: groups CRUD, membership management
- [x] Vue.js admin: group management (list/create/detail + members tab)
- [x] E2E tests: groups (18 tests), sidebar navigation

**What differs from Salesforce and will be added later:**

| SF Capability | Our Status | When |
|---------------|-----------|------|
| Permission Set Groups | Not implemented | Phase 2d |
| Muting Permission Sets | Grant/Deny PS covers this case | — |
| View All / Modify All per object | Not implemented | Phase 2d |
| Implicit Sharing (parent↔child) | Not implemented | Phase 2d |
| Queues (ownership) | Not implemented | Phase 6 |
| Territory Management | ✅ Implemented (ee/) | — |

---

### Phase 3: SOQL — Query Language ✅

Single entry point for all data reads with automatic security enforcement.

- [x] Parser (participle/v2): SELECT, FROM, WHERE, AND, OR, NOT, ORDER BY, LIMIT, OFFSET, GROUP BY, HAVING
- [x] AST: SelectStatement, FieldExpr, WhereClause, OrderByClause, LimitExpr
- [x] Dot-notation for parent fields: `Account.Name` (up to 5 levels)
- [x] Subquery for child relationships: `(SELECT Id FROM Contacts)`
- [x] Literals: string, number, boolean, null, date/datetime
- [x] Date literals: TODAY, YESTERDAY, THIS_WEEK, LAST_N_DAYS:N, etc.
- [x] Aggregate functions: COUNT, COUNT_DISTINCT, SUM, AVG, MIN, MAX
- [x] Built-in functions: UPPER, LOWER, TRIM, CONCAT, LENGTH, ABS, ROUND, COALESCE, NULLIF, etc.
- [x] Operators: =, !=, <>, >, <, >=, <=, IN, NOT IN, LIKE, IS NULL, IS NOT NULL
- [x] FOR UPDATE, WITH SECURITY_ENFORCED, TYPEOF (polymorphic fields)
- [x] Semi-joins: `WHERE Id IN (SELECT ... FROM ...)`
- [x] Field aliases: `SELECT Name AS ContactName`
- [x] Validator: field/object validation via MetadataProvider + AccessController
- [x] Compiler: AST → PostgreSQL SQL with parameterization
- [x] MetadataAdapter: bridge MetadataCache → engine.MetadataProvider
- [x] AccessControllerAdapter: bridge OLS/FLS → engine.AccessController (CanAccessObject, CanAccessField)
- [x] Executor (pgx): SQL execution with RLS WHERE injection
- [x] QueryService: facade parse → validate → compile → execute
- [x] REST API: `GET /api/v1/query?q=...`, `POST /api/v1/query`
- [x] OpenAPI spec: endpoints + schemas

**Salesforce SOQL features for future phases:**

| Capability | Phase |
|------------|-------|
| Cursor-based pagination (queryLocator) | Phase 3d |
| SOSL (full-text search) | Phase 15 |
| `GET /api/v1/soql/describe/{objectName}` | Phase 3d |

---

### Phase 4: DML Engine ✅

Single entry point for all data writes with security enforcement.

- [x] Parser (participle/v2): INSERT INTO, UPDATE SET, DELETE FROM, UPSERT ON
- [x] AST: DMLStatement, InsertStmt, UpdateStmt, DeleteStmt, UpsertStmt
- [x] INSERT: single + multi-row batch (up to 10,000 rows)
- [x] UPDATE: SET clause + WHERE clause, OLS/FLS/RLS enforcement
- [x] DELETE: WHERE clause (required by default), OLS/RLS enforcement
- [x] UPSERT: INSERT ON CONFLICT by external ID field
- [x] WHERE in DML: =, !=, <>, >, <, >=, <=, IN, NOT IN, LIKE, IS NULL, AND, OR, NOT
- [x] Built-in functions in VALUES/SET: UPPER, LOWER, CONCAT, COALESCE, ROUND, etc.
- [x] Validator: field/object validation via MetadataProvider + WriteAccessController
- [x] Compiler: AST → PostgreSQL SQL with RETURNING clause
- [x] MetadataAdapter: bridge MetadataCache → engine.MetadataProvider (with ReadOnly, Required, HasDefault)
- [x] WriteAccessControllerAdapter: bridge OLS/FLS → engine.WriteAccessController (CanCreate/CanUpdate/CanDelete + CheckWritableFields)
- [x] RLS Executor: pgx-based, with RLS WHERE injection for UPDATE/DELETE
- [x] DMLService: facade parse → validate → compile → execute
- [x] REST API: `POST /api/v1/data`
- [x] OpenAPI spec: endpoint + schemas

**DML features for future phases:**

| Capability | Phase |
|------------|-------|
| Automation Rules triggers (before/after insert/update/delete) | Phase 10 |
| Undelete (Recycle Bin) | Phase 4d |
| Merge | Phase 4d |
| Validation Rules (formula-based, pre-DML) | Phase 12 |
| Cascade delete (composition) | Phase 4a |
| Set null on delete (association) | Phase 4a |
| Partial success mode (`allOrNone: false`) | Phase 4d |
| Composite batch API (`/data/composite`) | Phase 4d |

---

### Phase 5: Auth Module ✅

Authentication and session management.

- [x] `POST /auth/login` — sign in with username + password → JWT access + refresh tokens
- [x] `POST /auth/refresh` — access token renewal (with refresh token rotation)
- [x] `POST /auth/logout` — refresh token invalidation
- [x] `GET /auth/me` — current user
- [x] JWT middleware: access token verification (HMAC-SHA256) on every request
- [x] Refresh tokens: SHA-256 hash storage in `iam.refresh_tokens`, rotation on use
- [x] Password hashing: bcrypt (cost=12), `password_hash` in `iam.users`
- [x] Rate limiting: in-memory sliding window per IP (5 attempts / 15 min)
- [x] Password reset flow: `POST /auth/forgot-password` + `POST /auth/reset-password` (token + email)
- [x] Admin password set: `PUT /admin/security/users/:id/password`
- [x] User ↔ security.User integration: JWT claims → UserContext (userId, profileId, roleId)
- [x] Admin-only registration via existing CRUD `POST /admin/security/users`
- [x] Seed admin password: `ADMIN_INITIAL_PASSWORD` env var on first launch
- [x] Vue.js frontend: Login, ForgotPassword, ResetPassword views, auth store (Pinia), router guards, 401 interceptor
- [x] pgTAP tests: password_hash, refresh_tokens, password_reset_tokens
- [x] E2E tests: 15 tests (login, forgot-password, reset-password, guards)

**Auth features for future phases:**

| Capability | Phase |
|------------|-------|
| OAuth 2.0 provider | Phase N |
| SAML 2.0 SSO | Phase N (ee/) |
| MFA (TOTP) | Phase N (ee/) |
| API keys / Connected Apps | Phase N |
| Login IP ranges per profile | Phase N |
| Session management (concurrent limits) | Phase N |

---

### Phase 6: App Templates ✅

Instead of hardcoded standard objects — an application template system (ADR-0018).
Admin selects a template via UI, the platform creates objects and fields through the metadata engine.

#### Implemented

- [x] **App Templates engine**: Registry + Applier pattern, two-pass creation (objects → fields)
- [x] **Sales CRM template**: Account, Contact, Opportunity, Task (4 objects, 36 fields)
- [x] **Recruiting template**: Position, Candidate, Application, Interview (4 objects, 28 fields)
- [x] **REST API**: `GET /api/v1/admin/templates` (list), `POST /api/v1/admin/templates/:id/apply` (apply)
- [x] **Guard**: template can only be applied to an empty database (object_definitions.count == 0)
- [x] **OLS/FLS**: automatic full CRUD + full RW assignment for admin PS
- [x] **Vue.js admin**: templates page with cards, "Apply" buttons
- [x] **E2E tests**: 9 tests (list page + sidebar navigation)
- [x] **Go tests**: 95%+ coverage (applier + registry + template structure validation)
- [x] **OpenAPI spec**: endpoints + schemas

#### Templates — Go Code (Type-Safe)

Templates are embedded in the binary as Go code. Adding a new template = a new file in `internal/platform/templates/`.

**Standard Objects for future phases (additional templates):**

| Template | Objects | Phase |
|----------|---------|-------|
| Customer Support | Case, Knowledge Article, Entitlement | Phase 14 |
| Marketing | Campaign, CampaignMember, Lead | Phase 14 |
| Commerce | Product, PriceBook, Order, OrderItem | Phase 14 |
| Project Management | Project, Milestone, Task, TimeEntry | Phase 14 |

---

### Phase 7: Generic CRUD + Vue.js Frontend ✅

Transition from admin-only to a full CRM interface. Backend: generic CRUD endpoints + DML pipeline extension. Frontend: metadata-driven UI.

#### Phase 7a: Generic CRUD + Metadata-Driven UI ✅

**Backend:**
- [x] Generic CRUD endpoints (single set of handlers for all objects via SOQL/DML)
- [x] Static defaults: inject `FieldConfig.default_value` for missing fields on Create
- [x] System fields injection: `owner_id`, `created_by_id`, `updated_by_id` (RecordService)
- [x] Describe API: `GET /api/v1/describe` (object list), `GET /api/v1/describe/:objectName` (fields + metadata)
- [x] OLS/FLS filtering in Describe API

**Frontend:**
- [x] AppLayout + AppSidebar (CRM zone `/app/*`, navigation from describe API)
- [x] Object list page (dynamic): SOQL-driven table for any object (RecordListView)
- [x] Record detail page (dynamic): fields from metadata + save/delete (RecordDetailView)
- [x] Record create form (dynamic): fields from metadata, pre-fill defaults (RecordCreateView)
- [x] FieldRenderer (type/subtype → input component) + FieldDisplay (read-only formatting)
- [x] Admin ↔ CRM switching (links in sidebars)
- [x] E2E tests: 17 tests (record list, create, detail, sidebar)
- [x] Go unit tests: RecordService 87%+ coverage

*Auth UI (login, forgot-password, auth store, guards) completed in Phase 5.*

#### Phase 7b: CEL Engine + Validation Rules + Dynamic Defaults ✅

**Backend (ADR-0019, ADR-0020):**
- [x] CEL engine integration (`cel-go`) — reusable ProgramCache, StandardEnv/DefaultEnv, EvaluateBool/EvaluateAny
- [x] Validation Rules: `metadata.validation_rules` table, CEL checks in DML (Stage 4b)
- [x] Dynamic defaults: `FieldConfig.default_expr` (CEL), DML Stage 3 dynamic
- [x] DML pipeline extension: typed interfaces (`DefaultResolver`, `RuleValidator`), Option pattern
- [x] Admin REST API: CRUD validation rules (5 endpoints)
- [x] Error mapping: RuleValidationError → 400, DefaultEvalError → 500

**Frontend:**
- [x] Admin UI: validation rules list/create/detail views
- [x] ObjectDetailView: "Validation Rules" tab → link to list
- [x] E2E tests: 14 tests (list, create, detail)
- [ ] Related lists: child objects on detail page (SOQL subqueries)
- [ ] Inline edit: click-to-edit on detail page
- [ ] List views: saved filters (my records, all records, custom)
- [ ] Global search (placeholder → SOSL in Phase 15)
- [ ] Recent items

**UI features for future phases:**

| Capability | Phase |
|------------|-------|
| Kanban view (Opportunity stages) | Phase 11 |
| Calendar view (Events) | Phase 11 |
| Home page with dashboards | Phase 11 |
| Dynamic Forms (visibility rules) | Phase 9c |
| Object Views per profile (role-based UI) | ✅ Phase 9a |
| Navigation + Dashboard per profile | Phase 9b |
| Mobile-responsive layout | Phase 7a (basic) |

---

### Phase 8: Custom Functions (ADR-0026) ✅

Named pure CEL expressions — foundation for reusable computational logic.

#### Phase 8a: Backend + Admin UI ✅

- [x] **metadata.functions table**: id, name, description, params JSONB, return_type, body TEXT
- [x] **CEL integration**: fn.* namespace, cel-go registration, ProgramCache extension
- [x] **Safety**: circular dependency detection (DetectCycles), nesting depth check (max 3 levels)
- [x] **Limits**: 4KB body, 100ms timeout, 3 levels nesting, 10 params, 200 functions max
- [x] **Admin REST API**: CRUD functions (5 endpoints) + deletion protection (409 Conflict, FindUsages)
- [x] **Admin Vue.js UI**: function list/create/detail views
- [x] **Migration + pgTAP tests**: UP/DOWN + schema tests for metadata.functions
- [x] **E2E tests**: admin function CRUD (14 tests)

#### Phase 8b: Expression Builder + cel-js ✅

- [x] **Dual-stack**: cel-js on frontend (`@marcbachmann/cel-js`), FnNamespace pattern for fn.* calls
- [x] **Pinia functions store**: function cache with `ensureLoaded()` / `invalidate()`
- [x] **cel-js wrapper**: `createCelEnvironment()`, `evaluateCel()`, `evaluateCelSafe()` with BigInt→Number conversion
- [x] **useCelEnvironment composable**: reactive, recreates environment on function changes
- [x] **CodeMirror autocomplete**: context-aware (record., old., user., fn., params), `@codemirror/autocomplete`
- [x] **CodeMirrorEditor Compartment**: dynamic extension reconfiguration without recreating editor
- [x] **ExpressionPreview**: live client-side CEL evaluation with 300ms debounce, parameter test inputs
- [x] **FunctionPicker**: catalog of built-in + custom functions (5 groups), insert into editor
- [x] **ExpressionBuilder integration**: Tabs (Fields/Functions), autocomplete, preview toggle
- [x] **Unit tests**: cel-environment (12 tests)
- [x] **E2E tests**: Expression Builder (9 tests — Functions tab, Preview, Autocomplete)

---

### Phase 9: Object View — Role-Based UI (ADR-0022) 🔄

Bounded context adapter: one object — different presentations for different roles.
Users of the same system (Sales, Warehouse, Management) see a role-specific interface without code duplication.

#### Phase 9a: Object View Core ✅

- [x] **object_views table**: `metadata.object_views (object_id, profile_id, config JSONB)`
- [x] **Config schema**: Read (fields, actions, queries, computed) + Write (optional: validation, defaults, computed, mutations)
- [x] **Resolution logic**: profile-specific → default → fallback (auto-generate from FLS)
- [x] **FLS intersection**: Object View fields ∩ FLS-accessible fields
- [x] **Describe API extension**: `GET /api/v1/describe/:objectName` includes resolved `form`
- [x] **Admin REST API**: CRUD for Object Views (5 endpoints)
- [x] **Vue.js Admin UI**: Object View list/create/detail (visual constructor: Read tabs (General, Fields, Actions, Queries, Computed) + Write tabs (Validation, Defaults, Computed, Mutations))
- [x] **Frontend renderer**: RecordDetailView/RecordCreateView render based on Object View config (sections, field order, actions with cel-js visibility)
- [x] **Fallback**: without Object View — current behavior (all FLS-accessible fields, auto-generated form)
- [x] **MetadataCache extension**: object views cached in memory, partial reload
- [x] **Go unit tests**: service + handler with OpenAPI response validation
- [x] **pgTAP tests**: schema tests for metadata.object_views
- [x] **E2E tests**: 24 tests (list, create, detail, sidebar)
- [ ] **Layout (ADR-0027)**: `metadata.layouts` table, Layout per (object_view, form_factor: desktop/tablet/mobile) — deferred to Phase 9d
- [ ] **Layout config**: section_config (columns, collapsed, visibility_expr), field_config (col_span, ui_kind, required_expr, readonly_expr, reference_config), list_columns (width, align, sortable) — deferred to Phase 9d
- [ ] **Form (computed)**: merge OV + Layout → Form in Describe API response. Frontend works only with Form — deferred to Phase 9d
- [ ] **Admin Layout UI**: CRUD layouts, preview per form factor, sync with OV lifecycle — deferred to Phase 9d
- [ ] **ui_kind enum**: 20+ types (auto, text, textarea, badge, lookup, rating, slider, toggle, etc.) — deferred to Phase 9d

#### Phase 9b: Navigation + Dashboard per Profile

- [ ] **profile_navigation table**: sidebar config per profile (groups, items, order)
- [ ] **profile_dashboards table**: home page widgets per profile (list, metric, chart)
- [ ] **Admin UI**: Navigation editor, Dashboard editor per profile
- [ ] **Sidebar per profile**: OLS filtering + profile_navigation grouping
- [ ] **Home dashboard per profile**: widgets with SOQL queries (list, metric)
- [ ] **Fallback**: without config → current behavior (OLS-filtered alphabetical sidebar, default home)

#### Phase 9c: Advanced Metadata

- [ ] **Record Types**: different picklist values and Object View per record type
- [ ] **Dynamic Forms**: field visibility rules (CEL: `record.status == 'closed'`)
- [ ] **Notes & Attachments**: polymorphic note/file objects, linked to any record
- [ ] **Activity History**: unified view of tasks + events for any object with hasActivities
- [ ] **Field History Tracking** (ee/): up to 20 fields per object, changelog table

---

### Phase 10: Procedure Engine + Automation Rules (ADR-0024) ⬜

Declarative automation: from atomic commands to composite procedures.

#### Phase 10a: Procedure Engine Core

- [ ] **Procedure runtime**: JSON DSL parsing, CEL evaluation, command execution
- [ ] **Command types**: `record.*` (CRUD via DML), `notification.*` (email/in-app), `integration.*` (HTTP), `compute.*` (transform/validate/fail), `flow.*` (call/if/match)
- [ ] **Conditional logic**: `when` (per-command), `flow.if` (condition/then/else), `flow.match` (expression/cases)
- [ ] **Rollback (Saga)**: LIFO compensating commands
- [ ] **Security sandbox**: limits (30s timeout, 50 commands, 10 HTTP calls), OLS/FLS/RLS enforcement
- [ ] **Named Credentials (ADR-0028)**: `metadata.credentials` + `credential_tokens` + `credential_usage_log`
- [ ] **Credential encryption**: AES-256-GCM, master key from ENV, unique nonce per record
- [ ] **Credential types**: api_key (header/query), basic (username/password), oauth2_client (auto token refresh)
- [ ] **SSRF protection**: base_url constraint, host match, internal IP blocklist, HTTPS only
- [ ] **Credential Admin API + UI**: CRUD + test connection + usage log + deactivate/activate
- [ ] **Versioning (ADR-0029)**: `metadata.procedure_versions` (draft/published/superseded), auto-increment version counter
- [ ] **Draft/Publish workflow**: save draft → dry-run test → publish; rollback to previous published
- [ ] **Storage**: `metadata.procedures` + `procedure_versions` (definition in versions, not inline)
- [ ] **Admin REST API**: CRUD procedures + versions + test (dry-run on draft) + publish + rollback
- [ ] **Procedure Constructor UI**: visual form-based builder → JSON
- [ ] **pgTAP tests**: schema tests for metadata.procedures + credentials

#### Phase 10b: Automation Rules

- [ ] **Automation Rules**: trigger definitions (before/after insert/update/delete)
- [ ] **Rule conditions**: CEL expression (`new.status != old.status`)
- [ ] **Actions**: invoke Procedure, field update, send notification
- [ ] **Execution order**: sort_order per object per event
- [ ] **Storage**: `metadata.automation_rules` table
- [ ] **Admin REST API + UI**: CRUD automation rules
- [ ] **pgTAP tests + E2E tests**

**Automation features for far future:**

| SF Capability | Equivalent |
|---------------|-----------|
| Apex (custom language) | Go trigger handlers (compiled) |
| Process Builder | Procedure Engine covers this |
| Workflow Rules | Procedure Engine + Automation Rules covers this |
| Flow Builder (visual) | Visual Builder on top of JSON DSL (Phase N) |
| Assignment Rules | Automation Rule + Procedure |
| Escalation Rules | Scenario + timers |

---

### Phase 11: Notifications, Activity & CRM UX ⬜

CRM as a daily working tool. Notifications built on Procedure Engine.

#### Phase 11a: Notifications & Activity

- [ ] **In-app notifications**: bell icon, notification list, read/unread, mark all read
- [ ] **Notification model**: `notification_types` + `notifications` table
- [ ] **Email notifications**: template engine (Go templates), SMTP sender
- [ ] **Trigger integration**: Automation Rules → `notification.email` / `notification.in_app` commands
- [ ] **Activity timeline**: chronological tasks/events on record detail page
- [ ] **Activity model**: `hasActivities` flag on object → polymorphic activity feed

#### Phase 11b: CRM UX Enhancements

- [ ] **Home dashboard**: pipeline chart, tasks due today, recent items
- [ ] **Kanban board**: drag-and-drop for picklist stages (Opportunity, Case)
- [ ] **Calendar view**: events display, day/week/month
- [ ] **Pipeline reports**: grouped by stage, by owner, by period

---

### Phase 12: Formula Engine ⬜

Computed fields and advanced formula-based validation.

- [ ] **Formula parser**: arithmetic, string functions, date math, IF/CASE, cross-object refs
- [ ] **Formula fields**: read-only computed at SOQL level (SQL expression in SELECT)
- [ ] **Roll-Up Summary fields**: COUNT, SUM, MIN, MAX on master-detail parent
- [ ] **Validation Rules (formula-based)**: boolean formula → error message, pre-DML
- [ ] **Default values (formula)**: formula or literal, applied on insert
- [ ] **Auto-number fields**: sequence-based auto-increment with format (INV-{0000})

---

### Phase 13: Scenario Engine + Approval Processes (ADR-0025) ⬜

Long-running process orchestration with durability and approval workflow.

#### Phase 13a: Scenario Engine

- [ ] **Orchestrator**: sequential workflow + goto + loop
- [ ] **Steps**: invoke Procedure, inline Command, wait signal/timer
- [ ] **Signals**: external events (approval, webhook, email confirm)
- [ ] **Timers**: delay, until, timeout, reminder
- [ ] **Rollback (Saga)**: LIFO compensation of completed steps
- [ ] **Durability**: PostgreSQL persistence (`scenario_executions`, `scenario_step_history`)
- [ ] **Recovery**: restart-safe — resume from last checkpoint
- [ ] **Idempotency**: `{executionId}-{stepCode}` key per step
- [ ] **Versioning (ADR-0029)**: `metadata.scenario_versions` (draft/published/superseded), auto-increment version counter
- [ ] **Draft/Publish workflow**: save draft → test → publish; rollback
- [ ] **Run snapshots**: `scenario_run_snapshots` captures procedure_version_id at run start
- [ ] **Signal API**: `POST /executions/{id}/signal`
- [ ] **Admin REST API + Constructor UI**: CRUD scenarios + versions, execution monitoring

#### Phase 13b: Approval Processes

- [ ] Approval definition: entry criteria, steps, approvers
- [ ] Submit for approval → pending → approved/rejected
- [ ] Email notifications per step (via Procedure Engine)
- [ ] Field updates on approve/reject (via Procedure Engine)
- [ ] Approval history on record detail
- [ ] Implemented as built-in Scenario + approval commands

---

### Phase 14: Advanced CRM Objects ⬜

Expanding the set of standard objects for a full-featured CRM.

- [ ] **Product**: name, code, description, is_active, family
- [ ] **PriceBook**: name, is_standard, is_active
- [ ] **PriceBookEntry**: product_id + pricebook_id + unit_price
- [ ] **OpportunityLineItem**: opportunity_id + pricebook_entry_id + quantity + total_price
- [ ] **Order**: account_id, status, order_date, total_amount
- [ ] **OrderItem**: order_id + product_id + quantity + unit_price
- [ ] **Contract**: account_id, status, start_date, end_date, term
- [ ] **Campaign**: name, type, status, start_date, end_date, budget
- [ ] **CampaignMember**: campaign_id + lead_id/contact_id + status
- [ ] **Case**: account_id, contact_id, subject, description, status, priority, origin
- [ ] **Custom Metadata Types** (`__mdt`): deployable config-as-data, queryable via SOQL

---

### Phase 15: Full-Text Search (SOSL) ⬜

Search across all objects simultaneously.

- [ ] PostgreSQL full-text search (tsvector/tsquery) or Elasticsearch/Meilisearch
- [ ] SOSL parser: `FIND {term} IN ALL FIELDS RETURNING Account(Name), Contact(Name, Email)`
- [ ] Indexing: trigger-based search index update on DML
- [ ] REST API: `POST /api/v1/sosl/search`
- [ ] Global search in UI: typeahead with SOSL backend

---

### Phase 16: Streaming & Integration ⬜

Event-driven architecture for integrations.

- [ ] **Change Data Capture (CDC)**: PostgreSQL LISTEN/NOTIFY or WAL-based
- [ ] CDC events: create, update, delete with changed fields
- [ ] **Platform Events**: custom event definitions (metadata), publish/subscribe
- [ ] Event bus: Redis Streams or PostgreSQL pg_notify
- [ ] **Webhooks**: event subscription with HTTP callback
- [ ] **Outbound Messages**: SOAP/REST callout on trigger/flow
- [ ] REST endpoint for publish: `POST /api/v1/events/{eventName}`
- [ ] WebSocket endpoint for subscribe: `WS /api/v1/events/stream`

---

### Phase 17: Analytics — Reports & Dashboards ⬜

Business analytics on top of SOQL.

- [ ] **Report Types**: metadata → which objects and relationships are available
- [ ] **Report Builder** (UI): field, filter, and grouping selection
- [ ] **Formats**: tabular, summary (with groupings), matrix
- [ ] **Aggregate formulas**: SUM, AVG, COUNT, MIN, MAX per group
- [ ] **Cross-filters**: Accounts with/without Opportunities
- [ ] **Dashboard Builder** (UI): components (chart, table, metric), linked to reports
- [ ] **Chart types**: bar, line, pie, donut, funnel, gauge
- [ ] **Scheduled reports**: email delivery on schedule
- [ ] **Dynamic dashboards**: running user = viewing user (RLS-aware)

---

### Phase N: Enterprise Features (ee/) ⬜

Proprietary capabilities requiring a commercial license.

| Capability | Description | SF Equivalent |
|------------|-------------|---------------|
| **Territory Management** ✅ | Territory hierarchy, models (planning/active/archived), user/record assignment, object defaults, assignment rules, territory-based sharing | Territory2 |
| **Audit Trail** | Full journal of all data changes (field-level, 10+ years) | Field Audit Trail (Shield) |
| **SSO (SAML 2.0)** | Single Sign-On via corporate IdP | SAML SSO |
| **Advanced Analytics** | Embedded BI, SAQL-like query language, predictive | CRM Analytics |
| **Encryption at Rest** | Sensitive field encryption at DB level | Platform Encryption (Shield) |
| **Event Monitoring** | Login events, API events, report events for compliance | Event Monitoring (Shield) |
| **Sandbox Management** | Full/partial copy environments for dev/test | Sandboxes |
| **API Governor Limits** | Per-tenant rate limiting, usage metering | API Limits |
| **Multi-org / Multi-tenant** | Single instance for multiple organizations | Multi-tenant kernel |
| **Custom Branding** | White-label UI, custom domain, logo | My Domain, Branding |

---

## Priorities and Dependencies

```
Phase 0 ✅ ──→ Phase 1 ✅ ──→ Phase 2 ✅ ──→ Phase 3 ✅ ──→ Phase 4 ✅ ──→ Phase 5 ✅ ──→ Phase 6 ✅
                                                                                          │
                                                                                          ▼
                                                                                    Phase 7a ✅──→ Phase 7b ✅──→ Phase 8 ✅
                                                                               (generic CRUD)  (CEL+valid.)   (functions)
                                                                                                                   │
                                                                                                             Phase 9a ✅
                                                                                                          (Object View)
                                                                                                                   │
                                                                                                             Phase 10
                                                                                                          (Procedures)
                                                                                                                   │
                                                                                                             Phase 11
                                                                                                       (Notif+CRM UX)
                                                                                                                   │
                                                                                                             Phase 12
                                                                                                            (Formulas)
                                                                                                                   │
                                                                                                             Phase 9c
                                                                                                       (Record Types)
                                                                                                                   │
                                                                                                             Phase 13
                                                                                                           (Scenarios)

                              Phase 15 (SOSL) — independent, after Phase 3
                              Phase 16 (CDC) — independent, after Phase 4
                              Phase 17 (Reports) — after Phase 3 + Phase 12
                              Phase N (ee/) — parallel, after Phase 2
```

### Critical Path (MVP)

Minimum set for a working CRM:

```
Phase 2b/2c ✅ → Phase 3 ✅ → Phase 4 ✅ → Phase 5 ✅ → Phase 6 ✅ → Phase 7 → v0.1.0
```

This covers: security → query → mutation → auth → standard objects → UI.

### Recommended Order After MVP

Principle: **platform before features** — features built on platform layers are cheaper, more flexible, and don't require rewrites.

1. **Phase 8** — Custom Functions (CEL reuse foundation, ADR-0026)
2. **Phase 9a** — Object View core (role-based UI, ADR-0022)
3. **Phase 9b** — Navigation + Dashboard per profile
4. **Phase 10a** — Procedure Engine core (declarative automation, ADR-0024)
5. **Phase 10b** — Automation Rules (trigger → procedure)
6. **Phase 11a** — Notifications & Activity (consumers of Procedure Engine)
7. **Phase 11b** — CRM UX (dashboard, kanban, calendar)
8. **Phase 12** — Formula Engine (computed fields, advanced validation)
9. **Phase 9c** — Record Types + Dynamic Forms
10. **Phase 13a** — Scenario Engine (ADR-0025)
11. **Phase 13b** — Approval Processes
12. **Phase 14** — Advanced CRM Objects
13. **Phase 15** — SOSL (full-text search)
14. **Phase 16** — Streaming & Integration (CDC, webhooks)
15. **Phase 17** — Analytics (reports, dashboards)

---

## What We Deliberately Don't Copy from Salesforce

| SF Feature | Reason for Exclusion | Alternative |
|------------|---------------------|-------------|
| Apex (custom language) | Development and runtime maintenance complexity | Go trigger handlers (compiled, type-safe) |
| Visualforce | Deprecated technology | Vue.js components |
| SOAP API | Legacy, redundant | REST + WebSocket only |
| Multi-tenant kernel | Overengineering for self-hosted | Single-tenant, simple deployment |
| Governor Limits | Not needed for single-tenant | Configurable rate limits |
| Key Prefix (3-char) | UUID v4 covers all cases (ADR-0001) | Polymorphic references via (object_type, record_id) |
| 15/18-char record IDs | UUID v4 | Standard UUID format |
| AppExchange / ISV packaging | Premature | Plugin system in far future |
| Aura Components | Legacy | — |
| Sandboxes (full copy) | Infrastructure complexity | Docker-based dev environments |

---

## Release Versioning

| Version | Phases | What the User Gets |
|---------|--------|-------------------|
| **v0.1.0-alpha** | 0-2 | Metadata engine + full security (OLS/FLS/RLS + Groups + Sharing Rules) + Territory Management (ee/) ✅ |
| **v0.2.0-alpha** | 3-5 | SOQL + DML + Auth — data can be read/written with security enforcement, JWT authentication ✅ |
| **v0.3.0-beta** | 6-7 | App Templates + Record UI — can log in and work with CRM data ✅ |
| **v0.4.0-beta** | 8-9 | Custom Functions + Object Views — CEL reuse, role-based UI |
| **v0.5.0-beta** | 10-11 | Procedure Engine + Notifications — declarative automation, daily CRM tool |
| **v1.0.0** | 12-13 | Formulas + Scenarios + Approvals — production-ready |
| **v1.x** | 14-16 | Advanced objects, search, integration |
| **v2.0** | 17 + N | Analytics, enterprise features |

---

## Platform Maturity Metrics

Criteria for assessing "Salesforce-grade" readiness by domain.

| Domain | Bronze (MVP) | Silver (v1.0) | Gold (v2.0) |
|--------|-------------|---------------|-------------|
| Metadata | Objects + Fields + References | + Record Types + Layouts + Formulas | + Custom Metadata Types + Big Objects |
| Security | OLS + FLS + RLS (OWD + Sharing Rules) | + Groups + Manual Sharing + Implicit | + Territory + Encryption + Audit |
| Data Access | SOQL: basic SELECT/WHERE/JOIN | + Aggregates + Subqueries + Date literals | + SOSL + FOR UPDATE + Polymorphic |
| Data Mutation | Insert + Update + Delete | + Upsert + Triggers + Validation Rules | + Undelete + Merge + Flows |
| UI | Admin + basic Record UI | + Dynamic Forms + List Views + Search | + App Builder + Kanban + Calendar |
| API | REST CRUD | + Composite + Bulk | + Streaming + CDC + GraphQL |
| Automation | — | Trigger handlers + Record-Triggered Flows | + Scheduled Flows + Approvals |
| Analytics | — | Basic reports | + Dashboard Builder + Scheduled reports |

---

### Infrastructure: Modular Monolith Preparation (ADR-0030) ✅

Architectural hygiene for microservices readiness.

- [x] **MetadataReader interface**: all 13 consumers depend on interface, not `*MetadataCache`
- [x] **Identity shared kernel**: `internal/pkg/identity` — `UserContext` without security import
- [x] **CacheBackedMetadataLister**: eliminates cross-schema SQL from security package
- [ ] **P2**: Generalized event bus from outbox pattern
- [ ] **P3**: Per-module narrow interfaces at consumer sites

---

*This document is updated as phases are completed. Last update: 2026-02-21.*
