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
| Data Mutation (DML) | Insert, Update, Upsert, Delete, Undelete, Merge + triggers | INSERT/UPDATE/DELETE/UPSERT, OLS+FLS enforcement, RLS injection for UPDATE/DELETE, batch operations, Custom Functions (fn.* dual-stack), validation rules (CEL), dynamic defaults (CEL), Automation Rules (triggers) | 70% SF |
| Auth | OAuth 2.0, SAML, MFA, Connected Apps | JWT (access + refresh), login, password reset, rate limiting | JWT + refresh tokens |
| Automation | Flow Builder, Triggers, Workflow Rules, Approval Processes | Automation Rules (before/after triggers, CEL conditions, procedure_code), Procedure Engine (6 command types, Named Credentials) | Triggers + basic Flows |
| UI Framework | Lightning App Builder, LWC, Dynamic Forms | Vue.js admin + metadata-driven CRM UI (AppLayout, dynamic record views, FieldRenderer), Expression Builder (CodeMirror + autocomplete + live preview), Object View (role-based sections, actions, highlights, related lists), Profile Navigation (grouped sidebar, page views via OV + Navigation), Layout + Form Resolution (per form_factor + mode, shared layouts, fallback chain) | Admin + Record UI + Object Views + Navigation + Layouts |
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
| Record Types | Not implemented | Phase 14b |
| Object Views (role-based layouts) | ✅ Implemented (Phase 9a) | — |
| Compact Layouts (highlight fields) | ✅ Implemented (highlight_fields in OV config) | — |
| Formula Fields | Not implemented | Phase 13d |
| Roll-Up Summary Fields | Not implemented | Phase 13d |
| Validation Rules (formula-based) | ✅ CEL-based (Phase 7b) | — |
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
| Automation Rules triggers (before/after insert/update/delete) | ✅ Phase 10b |
| Undelete (Recycle Bin) | Phase 11c |
| Merge | Phase 14+ |
| Validation Rules (formula-based, pre-DML) | ✅ Phase 7b (CEL-based) |
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
| Customer Support | Case, Knowledge Article, Entitlement | Phase 14c |
| Marketing | Campaign, CampaignMember, Lead | Phase 14c |
| Commerce | Product, PriceBook, Order, OrderItem | Phase 14c |
| Project Management | Project, Milestone, Task, TimeEntry | Phase 14c |

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
- [ ] Related lists: child objects on detail page (Phase 11a)
- [ ] Reference lookup: searchable dropdown (Phase 11a)
- [ ] List search, sort, row actions (Phase 11b)
- [ ] Saved list views / filters (Phase 12c)
- [ ] Global search (Phase 13b)

**UI features for future phases:**

| Capability | Phase |
|------------|-------|
| Related lists + Reference lookup | Phase 11a |
| List search, sort, row actions | Phase 11b |
| Recycle Bin (soft delete + restore) | Phase 11c |
| CSV Import | Phase 11d |
| Saved List Views (filters) | Phase 12c |
| Activity timeline + Bulk actions | Phase 12d |
| Kanban view (Opportunity stages) | Phase 14+ |
| Calendar view (Events) | Phase 14+ |
| Dynamic Forms (visibility rules) | Phase 14b |
| Object Views per profile (role-based UI) | ✅ Phase 9a |
| Navigation per profile + OV Unbinding | ✅ Phase 9b |
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

### Phase 9: Object View — Role-Based UI (ADR-0022) ✅ (9a-9c done, 9d → Phase 14b)

Bounded context adapter: one object — different presentations for different roles.
Users of the same system (Sales, Warehouse, Management) see a role-specific interface without code duplication.

#### Phase 9a: Object View Core ✅

- [x] **object_views table**: `metadata.object_views (api_name UNIQUE, profile_id, config JSONB)`
- [x] **Config schema**: Read (fields, actions, queries, computed) + Write (optional: validation, defaults, computed, mutations)
- [x] **View endpoint**: `GET /api/v1/view/:ovApiName` — returns OV config by api_name
- [x] **FLS intersection**: Object View fields ∩ FLS-accessible fields
- [x] **Describe API**: always returns fallback form (all FLS-accessible fields)
- [x] **Admin REST API**: CRUD for Object Views (5 endpoints)
- [x] **Vue.js Admin UI**: Object View list/create/detail (visual constructor: Read tabs (General, Fields, Actions, Queries, Computed) + Write tabs (Validation, Defaults, Computed, Mutations))
- [x] **Frontend renderer**: RecordDetailView/RecordCreateView render based on Object View config (sections, field order, actions with cel-js visibility)
- [x] **Fallback**: without Object View — current behavior (all FLS-accessible fields, auto-generated form)
- [x] **MetadataCache extension**: object views cached in memory, partial reload
- [x] **Go unit tests**: service + handler with OpenAPI response validation
- [x] **pgTAP tests**: schema tests for metadata.object_views
- [x] **E2E tests**: 24 tests (list, create, detail, sidebar)
- [x] **Layout + Form Resolution (ADR-0027 revised)**: moved to Phase 9c (see below)

#### Phase 9b: Navigation per Profile + OV Unbinding ✅

- [x] **OV Unbinding**: `object_id` removed from object_views — OV is no longer bound to a specific object
- [x] **profile_navigation table**: `metadata.profile_navigation` (migration 000031), UNIQUE(profile_id), JSONB config
- [x] **Navigation config extension**: `ov_api_name` field on nav items, `page` item type (renders OV as a standalone page)
- [x] **Navigation service**: CRUD + `ResolveForProfile`, validation (max 20 groups, 50 items/group, URL safety)
- [x] **Admin REST API**: 5 navigation endpoints on `/admin/profile-navigation`
- [x] **Resolution endpoints**: `GET /navigation` (OLS intersection, fallback to flat list), `GET /view/:ovApiName` (OV config by api_name)
- [x] **Sidebar per profile**: AppSidebar enhanced — grouped navigation from config, collapsible groups, fallback to OLS-filtered flat list
- [x] **Welcome page**: `/app` home page shows a welcome page instead of dashboard
- [x] **Dashboard removed**: separate dashboard entity eliminated — page-like dashboards achievable via OV + `page` nav items
- [x] **Admin UI**: Navigation list/create/detail (3 views)
- [x] **Go unit tests**: navigation service, table-driven
- [x] **pgTAP tests**: schema tests for profile_navigation
- [x] **E2E tests**: admin-navigation + app-sidebar-navigation
- [x] **OpenAPI spec**: navigation + view endpoints + schemas

#### Phase 9c: Layout + Form Resolution (ADR-0027 revised) ✅

- [x] **metadata.layouts table**: Layout per (object_view_id, form_factor, mode), UNIQUE constraint, ON DELETE CASCADE
- [x] **metadata.shared_layouts table**: reusable configuration snippets (type: field/section/list), api_name UNIQUE, RESTRICT delete protects referenced layouts
- [x] **Layout config**: section_config (columns, collapsed, visibility_expr), field_config (col_span, ui_kind, required_expr, readonly_expr, reference_config), list_columns (width, align, sortable)
- [x] **Form merge**: OV config + Layout config → computed Form in Describe API response. Frontend works only with Form
- [x] **Fallback chain**: requested layout → same form_factor any mode → desktop same mode → desktop edit → auto-generate
- [x] **X-Form-Factor / X-Form-Mode headers**: request headers for layout resolution in Describe API
- [x] **Shared layouts with layout_ref**: field/section/list references, inline overrides win, RESTRICT delete
- [x] **Admin REST API**: CRUD layouts (5 endpoints) + CRUD shared layouts (5 endpoints)
- [x] **Admin Vue.js UI**: Layout list/create/detail, Shared Layout list/create/detail
- [x] **Visual Layout Constructor**: tabbed editor (Form Layout / List Config / JSON) — section canvas with field chips, section/field property panels, list column DnD reorder (vue-draggable-plus), shared layout ref dropdown, JSON fallback editor
- [x] **CRM rendering**: RecordDetailView col_span/collapsible/grid sections, RecordListView layout-driven columns
- [x] **pgTAP tests**: schema tests for metadata.layouts + metadata.shared_layouts
- [x] **E2E tests**: 41 tests (layouts + shared layouts + CRM rendering)

**Technical Debt (Layout Constructor):**
- Root component tree editor: Layout stores `root` (page structure) but frontend doesn't render it yet. Visual tree editor deferred.
- Root component tree rendering: RecordDetailView should render `root` component tree instead of flat sections. Deferred.
- ExpressionBuilder integration: visibility/required/readonly expressions currently use plain textarea. Should integrate ExpressionBuilder.
- Live preview: show production-like preview of how the form will look to end user.
- Field order override: Layout can't reorder fields within sections (OV controls order). Consider optional field order in Layout.

#### Phase 9d: Advanced Metadata (deferred → Phase 14b)

Moved to Phase 14b (Record Types + Dynamic Forms) as part of the roadmap revision.
Notes & Attachments → Phase 13a (File Attachments). Activity History → Phase 12d.

---

### Phase 10: Procedure Engine + Automation Rules (ADR-0024) ✅

Declarative automation: from atomic commands to composite procedures.

#### Phase 10a: Procedure Engine Core ✅

- [x] **Procedure runtime**: JSON DSL parsing, CEL evaluation, command execution
- [x] **Command types**: `record.*` (CRUD via DML), `notification.*` (stub), `integration.*` (HTTP), `compute.*` (transform/validate/fail), `flow.*` (call/if/match), `wait.*` (stub)
- [x] **Conditional logic**: `when` (per-command), `flow.if` (condition/then/else), `flow.match` (expression/cases)
- [x] **Rollback (Saga)**: LIFO compensating commands
- [x] **Security sandbox**: limits (30s timeout, 50 commands, 10 HTTP calls), OLS/FLS/RLS enforcement
- [x] **Named Credentials (ADR-0028)**: `metadata.credentials` + `credential_tokens` + `credential_usage_log`
- [x] **Credential encryption**: AES-256-GCM, master key from ENV, unique nonce per record
- [x] **Credential types**: api_key (header/value), basic (username/password), oauth2_client (auto token refresh)
- [x] **SSRF protection**: base_url constraint, host match, internal IP blocklist, HTTPS only
- [x] **Credential Admin API + UI**: CRUD + test connection + usage log + deactivate/activate
- [x] **Versioning (ADR-0029)**: `metadata.procedure_versions` (draft/published/superseded), auto-increment version counter
- [x] **Draft/Publish workflow**: save draft → dry-run test → publish; rollback to previous published
- [x] **Storage**: `metadata.procedures` + `procedure_versions` (definition in versions, not inline)
- [x] **Admin REST API**: CRUD procedures + versions + test (dry-run on draft) + publish + rollback
- [x] **Procedure Constructor UI**: visual form-based builder (CommandEditor, CommandPicker, DryRunPanel)
- [x] **pgTAP tests**: schema tests for metadata.procedures + credentials
- [x] **E2E tests**: 24 procedure tests + 18 credential tests (42 total)

#### Phase 10b: Automation Rules ✅

- [x] **ADR-0031**: Automation Rules architecture (event types, conditions, actions, TX boundary, recursion limits)
- [x] **Automation Rules**: trigger definitions (before/after insert/update/delete)
- [x] **Rule conditions**: CEL expression (`new.status != old.status`)
- [x] **Actions**: procedure_code referencing a published procedure (simplified from 3 action types)
- [x] **Execution modes**: per_record and per_batch (configurable per rule)
- [x] **Execution order**: sort_order per object per event
- [x] **Storage**: `metadata.automation_rules` table (migration 000030)
- [x] **Automation Engine**: rule evaluation + dispatch, DML Pipeline Stage 8 (PostExecuteHook)
- [x] **Recursion depth limit**: configurable (default 3)
- [x] **Admin REST API**: 5 endpoints (list/create/get/update/delete)
- [x] **Vue.js Admin UI**: list (with object selector), create, detail views
- [x] **pgTAP tests**: 19 assertions for schema
- [x] **E2E tests**: 20 tests (list, create, detail, sidebar)

**Automation features for far future:**

| SF Capability | Equivalent |
|---------------|-----------|
| Apex (custom language) | Go trigger handlers (compiled) |
| Process Builder | Procedure Engine covers this |
| Workflow Rules | Procedure Engine + Automation Rules covers this |
| Flow Builder (visual) | Visual Builder on top of JSON DSL (Phase N) |
| Assignment Rules | Automation Rule + Procedure |
| Escalation Rules | Scenario + timers |

#### SOQL Editor (cross-cutting enhancement) ✅

- [x] **Backend**: `POST /admin/soql/validate` endpoint — server-side query validation with error position (line/column)
- [x] **CodeMirror refactoring**: CodeMirrorEditor language-agnostic (language prop), reused by CEL and SOQL
- [x] **SOQL syntax highlighting**: StreamLanguage parser — keywords, functions, date literals, parameters, strings, comments
- [x] **Context-aware autocomplete**: clause detection (SELECT→fields, FROM→objects, WHERE→fields+dates, ORDER BY→fields+ASC/DESC)
- [x] **SoqlEditor component**: toolbar (validate, test query, mode toggle, object/field picker), CodeMirror + autocomplete
- [x] **Integration**: OVQueriesTab uses SoqlEditor instead of plain textarea
- [x] **OpenAPI spec**: endpoint + schemas (SoqlValidateRequest/Response/Error)
- [x] **E2E tests**: 4 tests (editor visibility, validate POST, error display, test query results)

---

### Phase 11: Usable CRM Core — "Can show to a client" ⬜

Making CRM usable as a real working tool. Each sub-phase delivers visible user value.

#### Phase 11a: Related Lists + Reference Lookup

**Backend:**
- [ ] Describe handler: populate `related_lists` via `GetReverseRelationships()` (currently returns `[]`)
- [ ] RefConfig already exists in LayoutFieldConfig (DisplayFields, SearchFields, Target, Filter) — pass through to form

**Frontend:**
- [ ] `RelatedListPanel` component: mini-table of child records, SOQL query `SELECT ... FROM {child} WHERE {fk} = '{recordId}'`
- [ ] Integration into `RecordDetailView` below sections
- [ ] `ReferencePicker` component: searchable dropdown, SOQL search on target object
- [ ] Replace `<Input type="text">` with `ReferencePicker` in `FieldRenderer` for reference fields
- [ ] `FieldDisplay` for reference: display Name instead of UUID

**Dependencies:** none (SOQL subqueries and metadata already work)
**Complexity:** M (2-3 weeks)
**Value:** Detail page shows related records. Reference fields have search.

#### Phase 11b: List Search, Sort, Row Actions

**Backend:**
- [ ] Add query params `search`, `sort_by`, `sort_dir` to `GET /records/:objectName`
- [ ] `RecordService.List`: inject WHERE LIKE and ORDER BY into SOQL
- [ ] Per-page size selector (10/20/50/100)

**Frontend:**
- [ ] Search bar in `RecordListView` (reads `ListSearchConfig` from form)
- [ ] Clickable column headers with sort icons (reads `Sortable` from `ListColumnConfig`)
- [ ] Row actions (edit, delete, custom) from `ListConfig.RowActions`

**Dependencies:** none
**Complexity:** S (1-2 weeks)
**Value:** Users can search and sort records. Quick actions on rows.

#### Phase 11c: Recycle Bin (Soft Delete + Restore)

**Backend:**
- [ ] Migration: `is_deleted BOOLEAN DEFAULT FALSE`, `deleted_at TIMESTAMPTZ` on all `obj_*` tables
- [ ] DDL generator: new tables automatically get these columns
- [ ] DML engine: DELETE → `UPDATE SET is_deleted=true, deleted_at=NOW()`
- [ ] SOQL engine: default filter `is_deleted = false` (analogous to RLS injection)
- [ ] Endpoints: `GET /records/:obj/deleted`, `POST /records/:obj/:id/undelete`, `DELETE /records/:obj/:id/purge`
- [ ] Scheduled purge: delete records older than 15 days

**Frontend:**
- [ ] "Recycle Bin" page in sidebar
- [ ] Restore button, permanent delete with confirmation

**Dependencies:** none
**Complexity:** M (2-3 weeks)
**Value:** Safe deletion. Users don't fear losing data.

#### Phase 11d: CSV Import + Field Mapping

**Backend:**
- [ ] `POST /records/:objectName/import` (multipart/form-data CSV)
- [ ] CSV parser (encoding detection UTF-8/Windows-1252)
- [ ] Batch DML execution (batches of 200)
- [ ] Import result: total/success/error + downloadable error log

**Frontend:**
- [ ] `ImportWizard`: drag-and-drop upload → field mapping → preview 5 rows → confirm
- [ ] Progress bar + result summary

**Dependencies:** Phase 11c (recycle bin for rolling back failed imports)
**Complexity:** M (2-3 weeks)
**Value:** Onboarding — user uploads data from Excel/CSV.

---

### Phase 12: Daily Tool — "Use it every day" ⬜

Features that make CRM a daily working tool. Notifications, saved views, activity timeline.

#### Phase 12a: In-App Notifications

- [ ] Tables: `iam.notifications` (user_id, type, title, body, is_read, record_object, record_id)
- [ ] `notification.in_app` command in Procedure Engine (replace stub)
- [ ] API: list, mark read, mark all read, unread count
- [ ] Frontend: bell icon + badge in header, notification panel, click → navigate to record
- [ ] Polling or SSE for real-time

**Complexity:** M | **Value:** Automation Rules become visible to the user.

#### Phase 12b: Email Notifications (SMTP + Templates)

- [ ] Email sender: SMTP (gomail), ENV config
- [ ] `metadata.email_templates` (subject_template, body_template, object_api_name)
- [ ] `notification.email` command in Procedure Engine (replace stub)
- [ ] Admin UI: template CRUD + preview
- [ ] Dev mode: console sender (stdout)

**Complexity:** M | **Value:** Automated emails from Automation Rules.

#### Phase 12c: Saved List Views (Filters)

- [ ] `metadata.list_views` (object_api_name, name, owner_id, filter JSONB, columns, sort, visibility)
- [ ] Built-in: "All Records", "My Records" (`OwnerId = :currentUserId`)
- [ ] Filter builder: field + operator + value
- [ ] Frontend: dropdown switcher above the table

**Dependencies:** Phase 11b
**Complexity:** M | **Value:** Saved filters — "My Opportunities", "Open Tasks".

#### Phase 12d: Activity Timeline + Bulk Actions

- [ ] Activity timeline: `ActivityFeed` on detail page for objects with `hasActivities=true`
- [ ] Quick-add task/event inline
- [ ] Bulk actions: checkbox column, selection bar (delete, change owner, update field)
- [ ] Backend: `POST /records/:objectName/bulk` (action + ids + fields)

**Dependencies:** Phase 11a, 11c
**Complexity:** M | **Value:** Daily workflow: timeline + bulk operations.

---

### Phase 13: Production-Ready — "Trust it in production" ⬜

Features required for production deployment: files, export, audit, formulas.

#### Phase 13a: File Attachments

- [ ] Local FS (MVP) with S3-compatible interface
- [ ] `metadata.attachments` (record_object, record_id, filename, content_type, size_bytes, storage_path)
- [ ] Upload/download/delete API, 25MB limit
- [ ] Frontend: file zone on detail page, attachment list
- [ ] Security: OLS + RLS on parent record

**Complexity:** M | **Value:** Documents, contracts, images on records.

#### Phase 13b: CSV/JSON Export + Global Search

- [ ] Export: `GET /records/:objectName/export?format=csv|json` (stream SOQL → CSV/JSON)
- [ ] Respects current list view filters
- [ ] Global search: PostgreSQL `tsvector/tsquery` on Name fields of searchable objects
- [ ] DDL: `search_vector tsvector` + trigger on `obj_*` tables
- [ ] API: `GET /search?q=term` → results grouped by object (top 5 each)
- [ ] Frontend: search bar in header, dropdown with grouping

**Complexity:** L | **Value:** Export for reports. Quick search across all CRM data.

#### Phase 13c: Basic Audit Log (Core)

- [ ] `iam.audit_log` (user_id, action, object_api_name, record_id, old_values JSONB, new_values JSONB, ip_address)
- [ ] DML post-execute hook: capture old/new values
- [ ] API: record history + admin audit viewer (filters: user, object, action, date)
- [ ] Frontend: "History" tab on detail page, admin audit page
- [ ] Retention: 90 days (configurable)
- [ ] Note: this is core (AGPL). ee/ will add field-level audit trail with long retention.

**Complexity:** M | **Value:** "Who changed what" — compliance and troubleshooting.

#### Phase 13d: Formula Fields

- [ ] New subtype `formula` for each FieldType
- [ ] `FieldConfig.formula_expr` (CEL → SQL in SOQL SELECT)
- [ ] Roll-Up Summary: COUNT/SUM/MIN/MAX on parent from child records
- [ ] Read-only in DML, computed at query time
- [ ] Admin UI: formula editor with Expression Builder

**Dependencies:** Phase 8 (CEL), Phase 11a (related lists)
**Complexity:** L | **Value:** Computed fields, aggregates on parent records.

---

### Phase 14: Platform Completeness ⬜

Completing the platform with advanced features.

#### Phase 14a: Scenario Engine + Approval Processes (ADR-0025)

Long-running process orchestration with durability and approval workflow.

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
- [ ] Approval definition: entry criteria, steps, approvers
- [ ] Submit for approval → pending → approved/rejected
- [ ] Email notifications per step (via Procedure Engine)
- [ ] Approval history on record detail

#### Phase 14b: Record Types + Dynamic Forms

- [ ] **Record Types**: different picklist values and Object View per record type
- [ ] **Dynamic Forms**: field visibility rules (CEL: `record.status == 'closed'`)
- [ ] **Field History Tracking** (ee/): up to 20 fields per object, changelog table

#### Phase 14c: Advanced CRM Objects

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

### Dependency Graph (Phase 11+)

```
COMPLETED (0–10b, 9a–9c)
    │
    ├─→ 11a [Related Lists + Ref Lookup]
    │     ├─→ 11b [Search, Sort, Row Actions]
    │     │     ├─→ 11c [Recycle Bin]
    │     │     │     └─→ 11d [CSV Import]
    │     │     └─→ 12c [Saved List Views]
    │     │           └─→ 13b [Export + Global Search]
    │     └─→ 12d [Activity + Bulk Actions]
    │
    ├─→ 12a [In-App Notifications] ─→ 12b [Email]
    │
    ├─→ 13a [File Attachments]       (independent)
    ├─→ 13c [Audit Log]              (independent)
    └─→ 13d [Formula Fields]         (after 11a)
```

**Parallel tracks:**
- **Track A (Core UX):** 11a → 11b → 11c → 11d
- **Track B (Notifications):** 12a → 12b
- **Track C (Independent):** 13a, 13c — can start any time

### Full Dependency Graph (all phases)

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
                                                                                                             Phase 9b ✅
                                                                                                        (Nav+OV Unbinding)
                                                                                                                   │
                                                                                                             Phase 9c ✅
                                                                                                       (Layout+Form)
                                                                                                                   │
                                                                                                             Phase 10a ✅
                                                                                                          (Procedures)
                                                                                                                   │
                                                                                                             Phase 10b ✅
                                                                                                       (Automation Rules)
                                                                                                                   │
                                                                                     ┌──────────────────┬──────────┼───────────┐
                                                                                     ▼                  ▼          ▼           ▼
                                                                               Phase 11a          Phase 12a   Phase 13a   Phase 13c
                                                                          (Related+Ref)    (In-App Notif)  (Files)     (Audit)
                                                                                     │                  │
                                                                                     ▼                  ▼
                                                                               Phase 11b          Phase 12b
                                                                          (Search+Sort)     (Email Notif)
                                                                                     │
                                                                           ┌─────────┤
                                                                           ▼         ▼
                                                                     Phase 11c   Phase 12c ──→ Phase 13b (Export+Search)
                                                                    (Recycle Bin) (List Views)
                                                                           │
                                                                           ▼
                                                                     Phase 11d
                                                                    (CSV Import)

                              Phase 12d (Activity+Bulk) — after 11a + 11c
                              Phase 13d (Formulas) — after 11a + 8
                              Phase 14a (Scenarios) — after 12b
                              Phase 14b (Record Types) — after 13d
                              Phase 15 (SOSL) — independent, after Phase 3
                              Phase 16 (CDC) — independent, after Phase 4
                              Phase 17 (Reports) — after Phase 3 + Phase 13d
                              Phase N (ee/) — parallel, after Phase 2
```

### Critical Path (MVP — already done)

Minimum set for a working CRM (completed):

```
Phase 2b/2c ✅ → Phase 3 ✅ → Phase 4 ✅ → Phase 5 ✅ → Phase 6 ✅ → Phase 7 ✅ → v0.1.0
```

This covers: security → query → mutation → auth → standard objects → UI.

### Recommended Order After Phase 10b

Principle: **user value first** — each phase should deliver visible benefit to end users.

**Completed (platform foundation):**
1. ~~Phase 8~~ ✅ — Custom Functions
2. ~~Phase 9a~~ ✅ — Object View core
3. ~~Phase 9b~~ ✅ — Navigation per profile + OV Unbinding
4. ~~Phase 9c~~ ✅ — Layout + Form Resolution
5. ~~Phase 10a~~ ✅ — Procedure Engine core
6. ~~Phase 10b~~ ✅ — Automation Rules

**Next (usable CRM):**
7. **Phase 11a** — Related Lists + Reference Lookup (detail page becomes useful)
8. **Phase 11b** — List Search, Sort, Row Actions (list page becomes useful)
9. **Phase 11c** — Recycle Bin (safe deletion)
10. **Phase 11d** — CSV Import (onboarding)
11. **Phase 12a** — In-App Notifications (automation becomes visible)
12. **Phase 12b** — Email Notifications (external communication)
13. **Phase 12c** — Saved List Views (daily productivity)
14. **Phase 12d** — Activity Timeline + Bulk Actions (daily workflow)
15. **Phase 13a** — File Attachments (documents on records)
16. **Phase 13b** — CSV/JSON Export + Global Search (data access)
17. **Phase 13c** — Audit Log (compliance)
18. **Phase 13d** — Formula Fields (computed data)
19. **Phase 14a** — Scenario Engine + Approvals
20. **Phase 14b** — Record Types + Dynamic Forms
21. **Phase 14c** — Advanced CRM Objects
22. **Phase 15** — SOSL (full-text search)
23. **Phase 16** — Streaming & Integration (CDC, webhooks)
24. **Phase 17** — Analytics (reports, dashboards)

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
| **v0.4.0-beta** | 8-10b, 9a-9c | Functions + Views + Procedures + Automation (current) ✅ |
| **v0.5.0-beta** | 11a-11d | **Usable CRM**: related lists, reference lookup, search/sort, recycle bin, CSV import |
| **v0.6.0-beta** | 12a-12d | **Daily tool**: notifications, email, saved list views, activity timeline, bulk actions |
| **v0.7.0-rc** | 13a-13d | **Production-ready**: files, export, global search, audit log, formula fields |
| **v1.0.0** | 14a-14b | Scenarios + Approvals + Record Types |
| **v1.1** | 14c | Advanced CRM Objects |
| **v1.x** | 15-16 | Full-text search, streaming, integration |
| **v2.0** | 17 + N | Analytics, enterprise features |

---

## Platform Maturity Metrics

Criteria for assessing "Salesforce-grade" readiness by domain.

| Domain | Bronze (v0.4) ✅ | Silver (v0.7) | Gold (v2.0) |
|--------|-------------|---------------|-------------|
| Metadata | Objects + Fields + References + Layouts + OV | + Record Types + Formulas | + Custom Metadata Types + Big Objects |
| Security | OLS + FLS + RLS + Groups + Sharing Rules | + Audit Log + Recycle Bin | + Territory + Encryption + Field Audit Trail |
| Data Access | SOQL: SELECT/WHERE/JOIN/Aggregates/Subqueries | + Global Search (tsvector) | + SOSL + FOR UPDATE + Polymorphic |
| Data Mutation | Insert + Update + Delete + Upsert + Triggers + Validation Rules | + Soft Delete + Undelete + CSV Import | + Merge + Flows |
| UI | Admin + Record UI + Object Views + Navigation + Layouts | + Related Lists + Search/Sort + List Views + Activity | + App Builder + Kanban + Calendar |
| API | REST CRUD + SOQL + DML | + Bulk + Export + Import | + Streaming + CDC + GraphQL |
| Automation | Procedure Engine + Automation Rules | + Notifications + Email | + Scenarios + Approvals |
| Analytics | — | — | + Dashboard Builder + Scheduled reports |

---

### Infrastructure: Modular Monolith Preparation (ADR-0030) ✅

Architectural hygiene for microservices readiness.

- [x] **MetadataReader interface**: all 13 consumers depend on interface, not `*MetadataCache`
- [x] **Identity shared kernel**: `internal/pkg/identity` — `UserContext` without security import
- [x] **CacheBackedMetadataLister**: eliminates cross-schema SQL from security package
- [ ] **P2**: Generalized event bus from outbox pattern
- [ ] **P3**: Per-module narrow interfaces at consumer sites

---

*This document is updated as phases are completed. Last update: 2026-02-26 (roadmap revision: user-value-first approach).*
