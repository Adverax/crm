# Roadmap: путь к Salesforce-grade платформе

**Дата:** 2026-02-08
**Стек:** Go 1.25 · PostgreSQL 16 · Vue.js 3 · Redis
**Модель:** Open Core (AGPL v3 + Enterprise `ee/`)

---

## Зрелость по доменам

Карта текущего состояния и целевого покрытия относительно Salesforce Platform.

| Домен | Salesforce | Мы сейчас | Целевой уровень |
|-------|-----------|-----------|-----------------|
| Metadata Engine | Custom Objects, Fields, Relationships, Record Types, Layouts | Objects, Fields (все типы), Relationships (assoc/comp/poly), Table-per-object DDL | 80% SF |
| Security (OLS/FLS) | Profile, Permission Set, Permission Set Group, Muting PS | Profile, Grant/Deny PS, OLS bitmask, FLS bitmask, effective caches | 90% SF |
| Security (RLS) | OWD, Role Hierarchy, Sharing Rules, Manual Sharing, Apex Sharing, Teams, Territory | OWD, Groups (4 типа), Share tables, Role hierarchy, Sharing Rules (owner+criteria), Manual Sharing, RLS enforcer, effective caches, Territory Management (ee/) | 85% SF |
| Data Access (SOQL) | SOQL с relationship queries, aggregates, security enforcement | SOQL parser (participle), validator, compiler, executor с OLS+FLS+RLS enforcement, relationship queries, aggregates, date literals, subqueries | 70% SF |
| Data Mutation (DML) | Insert, Update, Upsert, Delete, Undelete, Merge + triggers | INSERT/UPDATE/DELETE/UPSERT, OLS+FLS enforcement, RLS injection для UPDATE/DELETE, batch operations, functions, validation rules (CEL), dynamic defaults (CEL) | 65% SF |
| Auth | OAuth 2.0, SAML, MFA, Connected Apps | JWT (access + refresh), login, password reset, rate limiting | JWT + refresh tokens |
| Automation | Flow Builder, Triggers, Workflow Rules, Approval Processes | Не реализовано | Triggers + базовые Flows |
| UI Framework | Lightning App Builder, LWC, Dynamic Forms | Vue.js admin + metadata-driven CRM UI (AppLayout, dynamic record views, FieldRenderer) | Admin + Record UI + Object Views |
| APIs | REST, SOAP, Bulk, Streaming, Metadata, Tooling, GraphQL | REST admin endpoints (metadata + security + groups + sharing rules) | REST + Streaming |
| Analytics | Reports, Dashboards, Einstein | Не реализовано | Базовые отчёты |
| Integration | Platform Events, CDC, External Services | Не реализовано | CDC + webhooks |
| Developer Tools | Apex, CLI, Sandboxes, Packaging | — | CLI + migration tools |
| Standard Objects | Account, Contact, Opportunity, Lead, Case, Task и др. | App Templates (Sales CRM: 4 obj, Recruiting: 4 obj) | 6-8 core objects |

---

## Фазы реализации

### Phase 0: Scaffolding ✅

Инфраструктура проекта.

- [x] Docker + docker-compose для локальной разработки
- [x] PostgreSQL 16 + pgTAP
- [x] Makefile, CI (GitHub Actions)
- [x] Структура проекта (cmd, internal, web, ee, migrations, tests)
- [x] HTTP-клиент, роутинг (Gin), structured logging (slog)
- [x] Typed errors (apperror), pagination helpers
- [x] Базовая Vue.js оболочка (AdminLayout, ui-компоненты)

---

### Phase 1: Metadata Engine ✅

Ядро платформы — динамическое определение объектов и полей.

- [x] Object Definitions (standard/custom, поведенческие флаги)
- [x] Field Definitions (type/subtype, config, validation)
- [x] Типы полей: text, number, boolean, datetime, picklist, reference
- [x] Reference types: association, composition, polymorphic
- [x] Table-per-object: DDL генерация (`obj_{api_name}`)
- [x] Constraints: FK, unique, not null, check
- [x] REST API: CRUD objects + fields
- [x] Vue.js admin: objects, fields, detail с табами
- [x] pgTAP тесты на схему

**Что отличает от Salesforce и будет добавлено позже:**

| Возможность SF | Наш статус | Когда |
|----------------|-----------|-------|
| Record Types | Не реализовано | Phase 9c |
| Object Views (role-based layouts) | Не реализовано | Phase 9a (ADR-0022) |
| Compact Layouts (highlight fields) | Не реализовано | Phase 9a (ADR-0022) |
| Formula Fields | Не реализовано | Phase 12 |
| Roll-Up Summary Fields | Не реализовано | Phase 12 |
| Validation Rules (formula-based) | Не реализовано | Phase 12 |
| Field History Tracking | Не реализовано | Phase N (ee/) |
| Custom Metadata Types (`__mdt`) | Не реализовано | Phase 14 |
| Big Objects | Не реализовано | Далёкая перспектива |
| External Objects | Не реализовано | Далёкая перспектива |

---

### Phase 2: Security Engine ✅

Три слоя безопасности — фундамент enterprise-grade платформы.

#### Phase 2a: Identity + Permission Engine ✅

- [x] User Roles (иерархия через parent_id)
- [x] Permission Sets (grant/deny, bitmask)
- [x] Profiles (auto-created base PS)
- [x] Users (username, email, profile, role, is_active)
- [x] Permission Set Assignments (user ↔ PS)
- [x] Object Permissions (OLS: CRUD bitmask 0-15)
- [x] Field Permissions (FLS: RW bitmask 0-3)
- [x] Effective caches (effective_ols, effective_fls)
- [x] Outbox pattern для инвалидации кэшей
- [x] REST API: полный CRUD для всех сущностей
- [x] Vue.js admin: роли, PS, профили, пользователи, OLS/FLS редактор

#### Phase 2b: RLS Core ✅

Row-Level Security — кто видит какие записи.

- [x] Org-Wide Defaults (OWD) per object: private, public_read, public_read_write, controlled_by_parent
- [x] Share tables: `obj_{name}__share` (grantee_id, access_level, share_reason)
- [x] Role Hierarchy: closure table `effective_role_hierarchy`
- [x] Sharing Rules (ownership-based): source group → target group, access level
- [x] Sharing Rules (criteria-based): field conditions → target group, access level
- [x] Manual Sharing: owner/admin расшаривает запись конкретному user/group
- [x] Record ownership model: OwnerId на каждой записи
- [x] Effective visibility cache: `effective_visible_owners`
- [x] REST API: OWD settings, sharing rules CRUD, manual sharing
- [x] Vue.js admin: OWD настройки (visibility в object create/edit), sharing rules UI (list/create/detail)
- [x] E2E тесты: sharing rules (14 тестов), visibility в объектах

#### Phase 2c: Groups ✅

Группы — единый grantee для всех sharing-операций.

- [x] Типы групп: personal, role, role_and_subordinates, public
- [x] Auto-generation: при создании user → personal group; при создании role → role group + role_and_sub group
- [x] Public group: админ создаёт, добавляет участников (users, roles, other groups)
- [x] Effective group members cache: `effective_group_members` (closure table)
- [x] Единый grantee (всегда group_id) для share tables и sharing rules
- [x] REST API: groups CRUD, membership management
- [x] Vue.js admin: управление группами (list/create/detail + members tab)
- [x] E2E тесты: groups (18 тестов), sidebar навигация

**Что отличает от Salesforce и будет добавлено позже:**

| Возможность SF | Наш статус | Когда |
|----------------|-----------|-------|
| Permission Set Groups | Не реализовано | Phase 2d |
| Muting Permission Sets | Grant/Deny PS покрывает этот кейс | — |
| View All / Modify All per object | Не реализовано | Phase 2d |
| Implicit Sharing (parent↔child) | Не реализовано | Phase 2d |
| Queues (ownership) | Не реализовано | Phase 6 |
| Territory Management | ✅ Реализовано (ee/) | — |

---

### Phase 3: SOQL — язык запросов ✅

Единая точка входа для всех чтений данных с автоматическим security enforcement.

- [x] Parser (participle/v2): SELECT, FROM, WHERE, AND, OR, NOT, ORDER BY, LIMIT, OFFSET, GROUP BY, HAVING
- [x] AST: SelectStatement, FieldExpr, WhereClause, OrderByClause, LimitExpr
- [x] Dot-notation для parent fields: `Account.Name` (до 5 уровней)
- [x] Subquery для child relationships: `(SELECT Id FROM Contacts)`
- [x] Литералы: string, number, boolean, null, date/datetime
- [x] Date literals: TODAY, YESTERDAY, THIS_WEEK, LAST_N_DAYS:N и др.
- [x] Aggregate functions: COUNT, COUNT_DISTINCT, SUM, AVG, MIN, MAX
- [x] Built-in functions: UPPER, LOWER, TRIM, CONCAT, LENGTH, ABS, ROUND, COALESCE, NULLIF и др.
- [x] Операторы: =, !=, <>, >, <, >=, <=, IN, NOT IN, LIKE, IS NULL, IS NOT NULL
- [x] FOR UPDATE, WITH SECURITY_ENFORCED, TYPEOF (polymorphic fields)
- [x] Semi-joins: `WHERE Id IN (SELECT ... FROM ...)`
- [x] Field aliases: `SELECT Name AS ContactName`
- [x] Validator: проверка полей/объектов через MetadataProvider + AccessController
- [x] Compiler: AST → PostgreSQL SQL с параметризацией
- [x] MetadataAdapter: мост MetadataCache → engine.MetadataProvider
- [x] AccessControllerAdapter: мост OLS/FLS → engine.AccessController (CanAccessObject, CanAccessField)
- [x] Executor (pgx): выполнение SQL с RLS WHERE injection
- [x] QueryService: фасад parse → validate → compile → execute
- [x] REST API: `GET /api/v1/query?q=...`, `POST /api/v1/query`
- [x] OpenAPI spec: endpoints + schemas

**Salesforce SOQL features для будущих фаз:**

| Возможность | Фаза |
|-------------|------|
| Cursor-based pagination (queryLocator) | Phase 3d |
| SOSL (full-text search) | Phase 15 |
| `GET /api/v1/soql/describe/{objectName}` | Phase 3d |

---

### Phase 4: DML Engine ✅

Единая точка входа для всех записей данных с security enforcement.

- [x] Parser (participle/v2): INSERT INTO, UPDATE SET, DELETE FROM, UPSERT ON
- [x] AST: DMLStatement, InsertStmt, UpdateStmt, DeleteStmt, UpsertStmt
- [x] INSERT: single + multi-row batch (до 10 000 строк)
- [x] UPDATE: SET clause + WHERE clause, OLS/FLS/RLS enforcement
- [x] DELETE: WHERE clause (обязателен по умолчанию), OLS/RLS enforcement
- [x] UPSERT: INSERT ON CONFLICT по external ID field
- [x] WHERE в DML: =, !=, <>, >, <, >=, <=, IN, NOT IN, LIKE, IS NULL, AND, OR, NOT
- [x] Built-in functions в VALUES/SET: UPPER, LOWER, CONCAT, COALESCE, ROUND и др.
- [x] Validator: проверка полей/объектов через MetadataProvider + WriteAccessController
- [x] Compiler: AST → PostgreSQL SQL с RETURNING clause
- [x] MetadataAdapter: мост MetadataCache → engine.MetadataProvider (с ReadOnly, Required, HasDefault)
- [x] WriteAccessControllerAdapter: мост OLS/FLS → engine.WriteAccessController (CanCreate/CanUpdate/CanDelete + CheckWritableFields)
- [x] RLS Executor: pgx-based, с RLS WHERE injection для UPDATE/DELETE
- [x] DMLService: фасад parse → validate → compile → execute
- [x] REST API: `POST /api/v1/data`
- [x] OpenAPI spec: endpoint + schemas

**DML features для будущих фаз:**

| Возможность | Фаза |
|-------------|------|
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

Аутентификация и управление сессиями.

- [x] `POST /auth/login` — вход по username + password → JWT access + refresh tokens
- [x] `POST /auth/refresh` — обновление access token (с ротацией refresh token)
- [x] `POST /auth/logout` — инвалидация refresh token
- [x] `GET /auth/me` — текущий пользователь
- [x] JWT middleware: проверка access token (HMAC-SHA256) на каждом запросе
- [x] Refresh tokens: хранение SHA-256 хэшей в `iam.refresh_tokens`, ротация при использовании
- [x] Password hashing: bcrypt (cost=12), `password_hash` в `iam.users`
- [x] Rate limiting: in-memory sliding window per IP (5 attempts / 15 min)
- [x] Password reset flow: `POST /auth/forgot-password` + `POST /auth/reset-password` (token + email)
- [x] Admin password set: `PUT /admin/security/users/:id/password`
- [x] User ↔ security.User интеграция: JWT claims → UserContext (userId, profileId, roleId)
- [x] Admin-only регистрация через существующий CRUD `POST /admin/security/users`
- [x] Seed admin password: `ADMIN_INITIAL_PASSWORD` env var при первом запуске
- [x] Vue.js frontend: Login, ForgotPassword, ResetPassword views, auth store (Pinia), router guards, 401 interceptor
- [x] pgTAP тесты: password_hash, refresh_tokens, password_reset_tokens
- [x] E2E тесты: 15 тестов (login, forgot-password, reset-password, guards)

**Auth features для будущих фаз:**

| Возможность | Фаза |
|-------------|------|
| OAuth 2.0 provider | Phase N |
| SAML 2.0 SSO | Phase N (ee/) |
| MFA (TOTP) | Phase N (ee/) |
| API keys / Connected Apps | Phase N |
| Login IP ranges per profile | Phase N |
| Session management (concurrent limits) | Phase N |

---

### Phase 6: App Templates ✅

Вместо хардкода стандартных объектов — система шаблонов приложений (ADR-0018).
Админ выбирает шаблон через UI, платформа создаёт объекты и поля через metadata engine.

#### Реализовано

- [x] **App Templates engine**: Registry + Applier pattern, двухпроходное создание (objects → fields)
- [x] **Sales CRM шаблон**: Account, Contact, Opportunity, Task (4 объекта, 36 полей)
- [x] **Recruiting шаблон**: Position, Candidate, Application, Interview (4 объекта, 28 полей)
- [x] **REST API**: `GET /api/v1/admin/templates` (список), `POST /api/v1/admin/templates/:id/apply` (применить)
- [x] **Guard**: шаблон можно применить только на пустую базу (object_definitions.count == 0)
- [x] **OLS/FLS**: автоматическое назначение full CRUD + full RW для admin PS
- [x] **Vue.js admin**: страница шаблонов с карточками, кнопки «Применить»
- [x] **E2E тесты**: 9 тестов (list page + sidebar navigation)
- [x] **Go тесты**: 95%+ покрытие (applier + registry + template structure validation)
- [x] **OpenAPI spec**: endpoints + schemas

#### Шаблоны — Go-код (type-safe)

Шаблоны встроены в бинарник как Go-код. Добавление нового шаблона = новый файл в `internal/platform/templates/`.

**Standard Objects для будущих фаз (дополнительные шаблоны):**

| Шаблон | Объекты | Фаза |
|--------|---------|------|
| Customer Support | Case, Knowledge Article, Entitlement | Phase 14 |
| Marketing | Campaign, CampaignMember, Lead | Phase 14 |
| Commerce | Product, PriceBook, Order, OrderItem | Phase 14 |
| Project Management | Project, Milestone, Task, TimeEntry | Phase 14 |

---

### Phase 7: Generic CRUD + Vue.js Frontend ✅

Переход от admin-only к полноценному CRM-интерфейсу. Backend: generic CRUD endpoints + DML pipeline расширение. Frontend: metadata-driven UI.

#### Phase 7a: Generic CRUD + Metadata-driven UI ✅

**Backend:**
- [x] Generic CRUD endpoints (один набор handlers для всех объектов через SOQL/DML)
- [x] Static defaults: инжект `FieldConfig.default_value` для отсутствующих полей при Create
- [x] System fields injection: `owner_id`, `created_by_id`, `updated_by_id` (RecordService)
- [x] Describe API: `GET /api/v1/describe` (список объектов), `GET /api/v1/describe/:objectName` (поля + metadata)
- [x] OLS/FLS фильтрация в Describe API

**Frontend:**
- [x] AppLayout + AppSidebar (CRM-зона `/app/*`, навигация из describe API)
- [x] Object list page (dynamic): SOQL-driven таблица для любого объекта (RecordListView)
- [x] Record detail page (dynamic): поля из metadata + save/delete (RecordDetailView)
- [x] Record create form (dynamic): поля из metadata, pre-fill defaults (RecordCreateView)
- [x] FieldRenderer (type/subtype → input component) + FieldDisplay (read-only formatting)
- [x] Переключение Admin ↔ CRM (ссылки в sidebar'ах)
- [x] E2E тесты: 17 тестов (record list, create, detail, sidebar)
- [x] Go unit тесты: RecordService 87%+ coverage

*Auth UI (login, forgot-password, auth store, guards) завершено в Phase 5.*

#### Phase 7b: CEL Engine + Validation Rules + Dynamic Defaults ✅

**Backend (ADR-0019, ADR-0020):**
- [x] CEL engine интеграция (`cel-go`) — reusable ProgramCache, StandardEnv/DefaultEnv, EvaluateBool/EvaluateAny
- [x] Validation Rules: таблица `metadata.validation_rules`, CEL-проверки в DML (Stage 4b)
- [x] Dynamic defaults: `FieldConfig.default_expr` (CEL), DML Stage 3 dynamic
- [x] DML pipeline extension: typed interfaces (`DefaultResolver`, `RuleValidator`), Option pattern
- [x] Admin REST API: CRUD validation rules (5 endpoints)
- [x] Error mapping: RuleValidationError → 400, DefaultEvalError → 500

**Frontend:**
- [x] Admin UI: validation rules list/create/detail views
- [x] ObjectDetailView: tab "Правила валидации" → ссылка на list
- [x] E2E тесты: 14 тестов (list, create, detail)
- [ ] Related lists: child objects на detail page (SOQL subqueries)
- [ ] Inline edit: click-to-edit на detail page
- [ ] List views: saved filters (мои записи, все записи, custom)
- [ ] Global search (placeholder → SOSL в Phase 15)
- [ ] Recent items

**UI features для будущих фаз:**

| Возможность | Фаза |
|-------------|------|
| Kanban view (Opportunity stages) | Phase 11 |
| Calendar view (Events) | Phase 11 |
| Home page с dashboards | Phase 11 |
| Dynamic Forms (visibility rules) | Phase 9c |
| Object Views per profile (role-based UI) | Phase 9a |
| Navigation + Dashboard per profile | Phase 9b |
| Mobile-responsive layout | Phase 7a (базовый) |

---

### Phase 8: Custom Functions (ADR-0026) ⬜

Именованные чистые CEL-выражения — фундамент для переиспользования вычислительной логики.

- [ ] **metadata.functions table**: id, name, description, params JSONB, return_type, body TEXT
- [ ] **CEL integration**: fn.* namespace, cel-go registration, ProgramCache extension
- [ ] **Dual-stack**: загрузка в cel-go (backend) + cel-js (frontend) через Describe API
- [ ] **Safety**: circular dependency detection, recursion prevention, call stack tracking
- [ ] **Limits**: 4KB body, 100ms timeout, 3 levels nesting, 10 params, 200 functions max
- [ ] **Admin REST API**: CRUD functions + test endpoint + dependencies view (7 endpoints)
- [ ] **Function Constructor UI**: create/edit + Expression Builder + live preview
- [ ] **Expression Builder**: reusable component — field picker, operator picker, function picker
- [ ] **Dependency tracking**: where-used view, deletion protection (409 Conflict)
- [ ] **Migration + pgTAP tests**: UP/DOWN + schema tests for metadata.functions
- [ ] **E2E tests**: admin function CRUD + Expression Builder

---

### Phase 9: Object View — Role-Based UI (ADR-0022) ⬜

Адаптер bounded context: один объект — разные представления для разных ролей.
Пользователи одной системы (Sales, Warehouse, Management) видят role-specific интерфейс без дублирования кода.

#### Phase 9a: Object View Core

- [ ] **object_views table**: `metadata.object_views (object_id, profile_id, config JSONB)`
- [ ] **Config schema**: sections (field grouping), highlight_fields (compact layout), actions (visibility_expr CEL), related_lists, list_fields, list_default_sort/filter
- [ ] **Resolution logic**: profile-specific → default → fallback (auto-generate from FLS)
- [ ] **FLS intersection**: Object View fields ∩ FLS-доступные поля
- [ ] **Describe API extension**: `GET /api/v1/describe/:objectName` включает resolved Object View
- [ ] **Admin REST API**: CRUD для Object Views (6 endpoints)
- [ ] **Vue.js Admin UI**: Object View list/create/detail + preview
- [ ] **Frontend renderer**: RecordDetailView/RecordCreateView рендерят по Object View config (sections, field order, actions)
- [ ] **Fallback**: без Object View — текущее поведение (все FLS-доступные поля)
- [ ] **Layout (ADR-0027)**: `metadata.layouts` table, Layout per (object_view, form_factor: desktop/tablet/mobile)
- [ ] **Layout config**: section_config (columns, collapsed, visibility_expr), field_config (col_span, ui_kind, required_expr, readonly_expr, reference_config), list_columns (width, align, sortable)
- [ ] **Form (computed)**: merge OV + Layout → Form в Describe API response. Frontend работает только с Form
- [ ] **Admin Layout UI**: CRUD layouts, preview per form factor, sync с OV lifecycle
- [ ] **ui_kind enum**: 20+ типов (auto, text, textarea, badge, lookup, rating, slider, toggle, etc.)

#### Phase 9b: Navigation + Dashboard per Profile

- [ ] **profile_navigation table**: sidebar config per profile (groups, items, order)
- [ ] **profile_dashboards table**: home page widgets per profile (list, metric, chart)
- [ ] **Admin UI**: Navigation editor, Dashboard editor per profile
- [ ] **Sidebar per profile**: OLS-фильтрация + profile_navigation grouping
- [ ] **Home dashboard per profile**: виджеты с SOQL-запросами (list, metric)
- [ ] **Fallback**: без config → текущее поведение (OLS-filtered alphabetical sidebar, default home)

#### Phase 9c: Advanced Metadata

- [ ] **Record Types**: разные picklist values и Object View per record type
- [ ] **Dynamic Forms**: visibility rules на поля (CEL: `record.status == 'closed'`)
- [ ] **Notes & Attachments**: polymorphic note/file объекты, привязка к любой записи
- [ ] **Activity History**: unified view tasks + events для любого объекта с hasActivities
- [ ] **Field History Tracking** (ee/): до 20 полей per object, changelog table

---

### Phase 10: Procedure Engine + Automation Rules (ADR-0024) ⬜

Декларативная автоматизация: от атомарных команд до составных процедур.

#### Phase 10a: Procedure Engine Core

- [ ] **Procedure runtime**: JSON DSL parsing, CEL evaluation, command execution
- [ ] **Command types**: `record.*` (CRUD через DML), `notification.*` (email/in-app), `integration.*` (HTTP), `compute.*` (transform/validate/fail), `flow.*` (call/if/match)
- [ ] **Conditional logic**: `when` (per-command), `flow.if` (condition/then/else), `flow.match` (expression/cases)
- [ ] **Rollback (Saga)**: LIFO compensating commands
- [ ] **Security sandbox**: лимиты (30s timeout, 50 commands, 10 HTTP calls), OLS/FLS/RLS enforcement
- [ ] **Storage**: `metadata.procedures` table (JSONB), snapshot versioning
- [ ] **Admin REST API**: CRUD procedures + test (dry-run)
- [ ] **Procedure Constructor UI**: visual form-based builder → JSON
- [ ] **pgTAP tests**: schema tests for metadata.procedures

#### Phase 10b: Automation Rules

- [ ] **Automation Rules**: trigger definitions (before/after insert/update/delete)
- [ ] **Rule conditions**: CEL expression (`new.status != old.status`)
- [ ] **Actions**: invoke Procedure, field update, send notification
- [ ] **Execution order**: sort_order per object per event
- [ ] **Storage**: `metadata.automation_rules` table
- [ ] **Admin REST API + UI**: CRUD automation rules
- [ ] **pgTAP tests + E2E tests**

**Automation features для далёкой перспективы:**

| Возможность SF | Аналог |
|----------------|--------|
| Apex (custom language) | Go trigger handlers (compiled) |
| Process Builder | Procedure Engine покрывает |
| Workflow Rules | Procedure Engine + Automation Rules покрывает |
| Flow Builder (visual) | Visual Builder поверх JSON DSL (Phase N) |
| Assignment Rules | Automation Rule + Procedure |
| Escalation Rules | Scenario + timers |

---

### Phase 11: Notifications, Activity & CRM UX ⬜

CRM как ежедневный рабочий инструмент. Notifications построены на Procedure Engine.

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

Вычисляемые поля и расширенная валидация на формулах.

- [ ] **Formula parser**: арифметика, строковые функции, date math, IF/CASE, cross-object refs
- [ ] **Formula fields**: read-only computed at SOQL level (SQL expression в SELECT)
- [ ] **Roll-Up Summary fields**: COUNT, SUM, MIN, MAX на master-detail parent
- [ ] **Validation Rules (formula-based)**: boolean formula → error message, pre-DML
- [ ] **Default values (formula)**: formula или literal, applied on insert
- [ ] **Auto-number fields**: sequence-based auto-increment с форматом (INV-{0000})

---

### Phase 13: Scenario Engine + Approval Processes (ADR-0025) ⬜

Оркестрация долгоживущих процессов с durability и approval workflow.

#### Phase 13a: Scenario Engine

- [ ] **Orchestrator**: sequential workflow + goto + loop
- [ ] **Steps**: вызов Procedure, inline Command, wait signal/timer
- [ ] **Signals**: внешние события (approval, webhook, email confirm)
- [ ] **Timers**: delay, until, timeout, reminder
- [ ] **Rollback (Saga)**: LIFO компенсация завершённых steps
- [ ] **Durability**: PostgreSQL persistence (`scenario_executions`, `scenario_step_history`)
- [ ] **Recovery**: restart-safe — возобновление с последнего checkpoint
- [ ] **Idempotency**: `{executionId}-{stepCode}` key per step
- [ ] **Signal API**: `POST /executions/{id}/signal`
- [ ] **Admin REST API + Constructor UI**: CRUD scenarios, execution monitoring

#### Phase 13b: Approval Processes

- [ ] Approval definition: entry criteria, steps, approvers
- [ ] Submit for approval → pending → approved/rejected
- [ ] Email notifications per step (via Procedure Engine)
- [ ] Field updates on approve/reject (via Procedure Engine)
- [ ] Approval history on record detail
- [ ] Реализация как built-in Scenario + approval commands

---

### Phase 14: Advanced CRM Objects ⬜

Расширение набора стандартных объектов для полноценного CRM.

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
- [ ] **Custom Metadata Types** (`__mdt`): deployable config-as-data, queryable через SOQL

---

### Phase 15: Full-Text Search (SOSL) ⬜

Поиск по всем объектам одновременно.

- [ ] PostgreSQL full-text search (tsvector/tsquery) или Elasticsearch/Meilisearch
- [ ] SOSL parser: `FIND {term} IN ALL FIELDS RETURNING Account(Name), Contact(Name, Email)`
- [ ] Индексация: trigger-based обновление search index при DML
- [ ] REST API: `POST /api/v1/sosl/search`
- [ ] Global search в UI: typeahead с SOSL backend

---

### Phase 16: Streaming & Integration ⬜

Event-driven архитектура для интеграций.

- [ ] **Change Data Capture (CDC)**: PostgreSQL LISTEN/NOTIFY или WAL-based
- [ ] CDC events: create, update, delete с changed fields
- [ ] **Platform Events**: custom event definitions (metadata), publish/subscribe
- [ ] Event bus: Redis Streams или PostgreSQL pg_notify
- [ ] **Webhooks**: подписка на события с HTTP callback
- [ ] **Outbound Messages**: SOAP/REST callout при trigger/flow
- [ ] REST endpoint для publish: `POST /api/v1/events/{eventName}`
- [ ] WebSocket endpoint для subscribe: `WS /api/v1/events/stream`

---

### Phase 17: Analytics — Reports & Dashboards ⬜

Бизнес-аналитика поверх SOQL.

- [ ] **Report Types**: metadata → какие объекты и relationships доступны
- [ ] **Report Builder** (UI): выбор полей, фильтров, группировок
- [ ] **Форматы**: tabular, summary (с группировками), matrix
- [ ] **Aggregate формулы**: SUM, AVG, COUNT, MIN, MAX по группам
- [ ] **Cross-filters**: Accounts with/without Opportunities
- [ ] **Dashboard Builder** (UI): компоненты (chart, table, metric), привязка к reports
- [ ] **Chart types**: bar, line, pie, donut, funnel, gauge
- [ ] **Scheduled reports**: email delivery по расписанию
- [ ] **Dynamic dashboards**: running user = viewing user (RLS-aware)

---

### Phase N: Enterprise Features (ee/) ⬜

Проприетарные возможности, требующие коммерческой лицензии.

| Возможность | Описание | Аналог SF |
|-------------|----------|-----------|
| **Territory Management** ✅ | Иерархия территорий, модели (planning/active/archived), user/record assignment, object defaults, assignment rules, territory-based sharing | Territory2 |
| **Audit Trail** | Полный журнал всех изменений данных (field-level, 10+ лет) | Field Audit Trail (Shield) |
| **SSO (SAML 2.0)** | Single Sign-On через corporate IdP | SAML SSO |
| **Advanced Analytics** | Embedded BI, SAQL-like query language, predictive | CRM Analytics |
| **Encryption at Rest** | Шифрование чувствительных полей на уровне БД | Platform Encryption (Shield) |
| **Event Monitoring** | Login events, API events, report events для compliance | Event Monitoring (Shield) |
| **Sandbox Management** | Full/partial copy environments для dev/test | Sandboxes |
| **API Governor Limits** | Per-tenant rate limiting, usage metering | API Limits |
| **Multi-org / Multi-tenant** | Единый instance для нескольких организаций | Multi-tenant kernel |
| **Custom Branding** | White-label UI, custom domain, логотип | My Domain, Branding |

---

## Приоритеты и зависимости

```
Phase 0 ✅ ──→ Phase 1 ✅ ──→ Phase 2 ✅ ──→ Phase 3 ✅ ──→ Phase 4 ✅ ──→ Phase 5 ✅ ──→ Phase 6 ✅
                                                                                          │
                                                                                          ▼
                                                                                    Phase 7a ✅──→ Phase 7b ✅──→ Phase 8
                                                                               (generic CRUD)  (CEL+valid.)   (functions)
                                                                                                                   │
                                                                                                             Phase 9a
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

                              Phase 15 (SOSL) — независимый, после Phase 3
                              Phase 16 (CDC) — независимый, после Phase 4
                              Phase 17 (Reports) — после Phase 3 + Phase 12
                              Phase N (ee/) — параллельно, после Phase 2
```

### Критический путь (MVP)

Минимальный набор для рабочей CRM:

```
Phase 2b/2c ✅ → Phase 3 ✅ → Phase 4 ✅ → Phase 5 ✅ → Phase 6 ✅ → Phase 7 → v0.1.0
```

Это покрывает: security → query → mutation → auth → standard objects → UI.

### Рекомендованный порядок после MVP

Принцип: **платформа перед фичами** — фичи, построенные на платформенных слоях, дешевле, гибче и не требуют переписывания.

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

## Что мы сознательно не копируем у Salesforce

| SF Feature | Причина отказа | Альтернатива |
|------------|---------------|--------------|
| Apex (custom language) | Сложность разработки и поддержки runtime | Go trigger handlers (compiled, type-safe) |
| Visualforce | Устаревшая технология | Vue.js компоненты |
| SOAP API | Legacy, избыточен | Только REST + WebSocket |
| Multi-tenant kernel | Overengineering для self-hosted | Single-tenant, простое развёртывание |
| Governor Limits | Не нужны для single-tenant | Конфигурируемые rate limits |
| Key Prefix (3-char) | UUID v4 покрывает все кейсы (ADR-0001) | Polymorphic ссылки через (object_type, record_id) |
| 15/18-char record IDs | UUID v4 | Standard UUID format |
| AppExchange / ISV packaging | Преждевременно | Plugin system в далёкой перспективе |
| Aura Components | Legacy | — |
| Sandboxes (full copy) | Инфраструктурная сложность | Docker-based dev environments |

---

## Версионирование релизов

| Версия | Фазы | Что пользователь получает |
|--------|-------|--------------------------|
| **v0.1.0-alpha** | 0-2 | Metadata engine + полный security (OLS/FLS/RLS + Groups + Sharing Rules) + Territory Management (ee/) ✅ |
| **v0.2.0-alpha** | 3-5 | SOQL + DML + Auth — данные можно читать/писать с security enforcement, JWT-аутентификация ✅ |
| **v0.3.0-beta** | 6-7 | App Templates + Record UI — можно логиниться и работать с CRM-данными ✅ |
| **v0.4.0-beta** | 8-9 | Custom Functions + Object Views — CEL reuse, role-based UI |
| **v0.5.0-beta** | 10-11 | Procedure Engine + Notifications — declarative automation, daily CRM tool |
| **v1.0.0** | 12-13 | Formulas + Scenarios + Approvals — production-ready |
| **v1.x** | 14-16 | Advanced objects, search, integration |
| **v2.0** | 17 + N | Analytics, enterprise features |

---

## Метрики зрелости платформы

Критерии для оценки «Salesforce-grade» готовности по каждому домену.

| Домен | Bronze (MVP) | Silver (v1.0) | Gold (v2.0) |
|-------|-------------|---------------|-------------|
| Metadata | Objects + Fields + References | + Record Types + Layouts + Formulas | + Custom Metadata Types + Big Objects |
| Security | OLS + FLS + RLS (OWD + Sharing Rules) | + Groups + Manual Sharing + Implicit | + Territory + Encryption + Audit |
| Data Access | SOQL: basic SELECT/WHERE/JOIN | + Aggregates + Subqueries + Date literals | + SOSL + FOR UPDATE + Polymorphic |
| Data Mutation | Insert + Update + Delete | + Upsert + Triggers + Validation Rules | + Undelete + Merge + Flows |
| UI | Admin + basic Record UI | + Dynamic Forms + List Views + Search | + App Builder + Kanban + Calendar |
| API | REST CRUD | + Composite + Bulk | + Streaming + CDC + GraphQL |
| Automation | — | Trigger handlers + Record-Triggered Flows | + Scheduled Flows + Approvals |
| Analytics | — | Basic reports | + Dashboard Builder + Scheduled reports |

---

*Этот документ обновляется по мере завершения фаз. Последнее обновление: 2026-02-15.*
