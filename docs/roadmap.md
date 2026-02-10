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
| Security (RLS) | OWD, Role Hierarchy, Sharing Rules, Manual Sharing, Apex Sharing, Teams, Territory | ADR приняты (0011-0013), не реализовано | 80% SF |
| Data Access (SOQL) | SOQL с relationship queries, aggregates, security enforcement | Не реализовано | 70% SF |
| Data Mutation (DML) | Insert, Update, Upsert, Delete, Undelete, Merge + triggers | Не реализовано | 60% SF |
| Auth | OAuth 2.0, SAML, MFA, Connected Apps | Не реализовано (отложено) | JWT + refresh tokens |
| Automation | Flow Builder, Triggers, Workflow Rules, Approval Processes | Не реализовано | Triggers + базовые Flows |
| UI Framework | Lightning App Builder, LWC, Dynamic Forms | Vue.js admin для metadata + security | Admin + Record UI |
| APIs | REST, SOAP, Bulk, Streaming, Metadata, Tooling, GraphQL | REST admin endpoints (metadata + security) | REST + Streaming |
| Analytics | Reports, Dashboards, Einstein | Не реализовано | Базовые отчёты |
| Integration | Platform Events, CDC, External Services | Не реализовано | CDC + webhooks |
| Developer Tools | Apex, CLI, Sandboxes, Packaging | — | CLI + migration tools |
| Standard Objects | Account, Contact, Opportunity, Lead, Case, Task и др. | Не реализовано | 6-8 core objects |

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
| Record Types | Не реализовано | Phase 9 |
| Page Layouts | Не реализовано | Phase 9 |
| Compact Layouts | Не реализовано | Phase 9 |
| Formula Fields | Не реализовано | Phase 10 |
| Roll-Up Summary Fields | Не реализовано | Phase 10 |
| Validation Rules (formula-based) | Не реализовано | Phase 10 |
| Field History Tracking | Не реализовано | Phase N (ee/) |
| Custom Metadata Types (`__mdt`) | Не реализовано | Phase 11 |
| Big Objects | Не реализовано | Далёкая перспектива |
| External Objects | Не реализовано | Далёкая перспектива |

---

### Phase 2: Security Engine 🔧 (в процессе)

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

#### Phase 2b: RLS Core ⬜

Row-Level Security — кто видит какие записи.

- [ ] Org-Wide Defaults (OWD) per object: private, public_read, public_read_write, controlled_by_parent
- [ ] Share tables: `obj_{name}__share` (grantee_id, access_level, share_reason)
- [ ] Role Hierarchy: closure table `effective_role_hierarchy`
- [ ] Sharing Rules (ownership-based): source group → target group, access level
- [ ] Sharing Rules (criteria-based): field conditions → target group, access level
- [ ] Manual Sharing: owner/admin расшаривает запись конкретному user/group
- [ ] Record ownership model: OwnerId на каждой записи
- [ ] Effective visibility cache: `effective_visible_owners`
- [ ] REST API: OWD settings, sharing rules CRUD, manual sharing
- [ ] Vue.js admin: OWD настройки, sharing rules UI

#### Phase 2c: Groups ⬜

Группы — единый grantee для всех sharing-операций.

- [ ] Типы групп: personal, role, role_and_subordinates, public
- [ ] Auto-generation: при создании user → personal group; при создании role → role group + role_and_sub group
- [ ] Public group: админ создаёт, добавляет участников (users, roles, other groups)
- [ ] Effective group members cache: `effective_group_members` (closure table)
- [ ] Единый grantee (всегда group_id) для share tables и sharing rules
- [ ] REST API: groups CRUD, membership management
- [ ] Vue.js admin: управление группами

**Что отличает от Salesforce и будет добавлено позже:**

| Возможность SF | Наш статус | Когда |
|----------------|-----------|-------|
| Permission Set Groups | Не реализовано | Phase 2d |
| Muting Permission Sets | Grant/Deny PS покрывает этот кейс | — |
| View All / Modify All per object | Не реализовано | Phase 2d |
| Implicit Sharing (parent↔child) | Не реализовано | Phase 2d |
| Queues (ownership) | Не реализовано | Phase 6 |
| Territory Management | Не реализовано | Phase N (ee/) |

---

### Phase 3: SOQL — язык запросов ⬜

Единая точка входа для всех чтений данных с автоматическим security enforcement.

#### Phase 3a: Lexer + Parser

- [ ] Tokenizer: SELECT, FROM, WHERE, AND, OR, NOT, ORDER BY, LIMIT, OFFSET, GROUP BY, HAVING
- [ ] AST: SelectStatement, FieldExpr, WhereClause, OrderByClause, LimitExpr
- [ ] Поддержка dot-notation для parent fields: `Account.Name`
- [ ] Subquery для child relationships: `(SELECT Id FROM Contacts)`
- [ ] Bind-переменные: `:paramName`
- [ ] Литералы: string, number, boolean, null, date literals (TODAY, LAST_N_DAYS:n)

#### Phase 3b: Executor

- [ ] AST → SQL translation с использованием metadata registry
- [ ] Автоматическое разрешение `obj_{api_name}` → реальная таблица
- [ ] Автоматическое разрешение field api_name → column name
- [ ] JOIN для relationship queries (parent-to-child, child-to-parent)
- [ ] FLS enforcement: исключение полей без доступа из SELECT
- [ ] RLS enforcement: добавление WHERE-условий на основе visibility
- [ ] Aggregate functions: COUNT, SUM, AVG, MIN, MAX
- [ ] Pagination: LIMIT + OFFSET (позже — cursor-based)
- [ ] Parameterized queries (защита от SQL injection)

#### Phase 3c: REST API

- [ ] `POST /api/v1/soql/query` — выполнение SOQL-запроса
- [ ] `GET /api/v1/soql/describe/{objectName}` — описание объекта (доступные поля с учётом FLS)
- [ ] Query result format: `{ totalSize, done, records: [...] }`
- [ ] Пагинация через `nextRecordsUrl` (queryLocator)

**Salesforce SOQL features для будущих фаз:**

| Возможность | Фаза |
|-------------|------|
| FOR UPDATE (row locking) | Phase 3d |
| TYPEOF (polymorphic resolution) | Phase 3d |
| WITH SECURITY_ENFORCED / USER_MODE | Встроено с Phase 3b |
| Semi-joins / Anti-joins | Phase 3d |
| Relationship queries (5 levels deep) | Phase 3b: 2 уровня, 3d: 5 уровней |
| Date literals (THIS_QUARTER, LAST_N_DAYS) | Phase 3a |
| SOSL (full-text search) | Phase 12 |

---

### Phase 4: DML Engine ⬜

Единая точка входа для всех записей данных с security + trigger pipeline.

#### Phase 4a: Core Operations

- [ ] `INSERT`: создание записей с OLS/FLS enforcement
- [ ] `UPDATE`: обновление с OLS/FLS + ownership check
- [ ] `DELETE`: удаление с OLS check + cascade (composition)
- [ ] `UPSERT`: insert-or-update по external ID или record ID
- [ ] Bulk operations: операции над массивом записей (до 200 за вызов)
- [ ] Partial success mode: `allOrNone: false` — возврат ошибок per record

#### Phase 4b: Trigger Pipeline

- [ ] Before triggers: before insert, before update, before delete
- [ ] After triggers: after insert, after update, after delete
- [ ] Trigger context: `new`, `old`, `newMap`, `oldMap`, `isInsert`, `isUpdate`, `isDelete`
- [ ] Trigger handler registry: один handler per object, регистрация в metadata
- [ ] Order of execution: validation → before triggers → DML → after triggers → sharing recalc

#### Phase 4c: REST API

- [ ] `POST /api/v1/data/{objectName}` — insert
- [ ] `PATCH /api/v1/data/{objectName}/{id}` — update
- [ ] `DELETE /api/v1/data/{objectName}/{id}` — delete
- [ ] `PUT /api/v1/data/{objectName}/{externalIdField}/{externalIdValue}` — upsert
- [ ] `POST /api/v1/data/composite` — batch operations

**DML features для будущих фаз:**

| Возможность | Фаза |
|-------------|------|
| Undelete (Recycle Bin) | Phase 4d |
| Merge | Phase 4d |
| Validation Rules (formula-based, pre-DML) | Phase 10 |
| Auto-number fields (sequence on insert) | Phase 4a |
| Default values | Phase 4a |
| Cascade delete (composition) | Phase 4a |
| Set null on delete (association) | Phase 4a |

---

### Phase 5: Auth Module ⬜

Аутентификация и управление сессиями.

- [ ] `POST /auth/login` — вход по username + password → JWT access + refresh tokens
- [ ] `POST /auth/register` — регистрация (admin-only или self-service)
- [ ] `POST /auth/refresh` — обновление access token
- [ ] `POST /auth/logout` — инвалидация refresh token
- [ ] JWT middleware: проверка access token на каждом запросе
- [ ] Refresh tokens: хранение хэшей в БД, ротация при использовании
- [ ] Password hashing: bcrypt/argon2
- [ ] Rate limiting: login attempts per IP/username
- [ ] Password reset flow (email + token)
- [ ] User ↔ security.User интеграция: auth middleware → context с userId, profileId, roleId

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

### Phase 6: Standard Objects ⬜

Предустановленные объекты для CRM-сценариев. Создаются через metadata engine (не хардкод).

#### Объекты и поля

| Объект | Ключевые поля | Relationships |
|--------|--------------|---------------|
| **Account** | name, industry, type, phone, website, billing_address, shipping_address | parent_account_id (self-ref) |
| **Contact** | first_name, last_name, email, phone, title, department, mailing_address | account_id (association) |
| **Opportunity** | name, stage, amount, close_date, probability, type | account_id (association) |
| **Lead** | first_name, last_name, company, email, phone, status, source, rating | — |
| **Task** | subject, description, status, priority, due_date | who_id (polymorphic: Contact/Lead), what_id (polymorphic: Account/Opportunity) |
| **Event** | subject, location, start_datetime, end_datetime, is_all_day | who_id, what_id (polymorphic) |

#### Дополнительно

- [ ] Seed-миграция: создание standard objects через metadata API (не raw SQL)
- [ ] Picklist values: стандартные значения для stage, status, industry, type
- [ ] Флаг `is_platform_managed: true` — запрет удаления standard objects
- [ ] System fields: created_by, updated_by, owner_id на всех объектах

**Standard Objects для будущих фаз:**

| Объект | Фаза |
|--------|------|
| Campaign | Phase 8 |
| Case | Phase 8 |
| Product / PriceBook / PriceBookEntry | Phase 11 |
| Order / OrderItem | Phase 11 |
| Contract | Phase 11 |
| Note / Attachment | Phase 9 |
| ActivityHistory (unified) | Phase 9 |

---

### Phase 7: Vue.js Frontend — Record UI ⬜

Переход от admin-only к полноценному CRM-интерфейсу.

#### Phase 7a: Shell + Auth UI

- [ ] Login page, register page
- [ ] Auth store (Pinia): JWT management, auto-refresh
- [ ] Protected routes (navigation guard)
- [ ] App shell: top nav, user menu, global search placeholder

#### Phase 7b: Dynamic Record UI

- [ ] Object list page (dynamic): SOQL-driven таблица для любого объекта
- [ ] Record detail page (dynamic): поля из metadata + FLS
- [ ] Record create/edit form (dynamic): поля из metadata, validation из field config
- [ ] Related lists: child objects на detail page (SOQL subqueries)
- [ ] Inline edit: click-to-edit на detail page
- [ ] Record owner display + manual sharing UI

#### Phase 7c: Navigation & Search

- [ ] App navigation: tabs для каждого объекта (из metadata, ordered)
- [ ] List views: saved filters (мои записи, все записи, custom)
- [ ] Global search (placeholder → SOSL в Phase 12)
- [ ] Recent items

**UI features для будущих фаз:**

| Возможность | Фаза |
|-------------|------|
| Kanban view (Opportunity stages) | Phase 8 |
| Calendar view (Events) | Phase 8 |
| Home page с dashboards | Phase 8 |
| Dynamic Forms (visibility rules) | Phase 9 |
| Page Layouts per profile/record type | Phase 9 |
| Mobile-responsive layout | Phase 7b (базовый) |

---

### Phase 8: Notifications, Dashboard, Activity ⬜

CRM становится рабочим инструментом.

- [ ] In-app notifications: bell icon, notification list, read/unread
- [ ] Email notifications: template engine, trigger-based sending
- [ ] Home dashboard: pipeline chart, tasks due today, recent items
- [ ] Activity timeline: хронология tasks/events на record detail
- [ ] Kanban board для Opportunity stages
- [ ] Calendar view для Events
- [ ] Pipeline reports: grouped by stage, by owner, by close_date month

---

### Phase 9: Advanced Metadata ⬜

Расширение metadata engine до Salesforce-level гибкости.

- [ ] **Record Types**: разные picklist values и page layouts для одного объекта
- [ ] **Page Layouts**: JSON-описание расположения полей, секций, related lists
- [ ] **Compact Layouts**: какие поля показывать в highlight panel
- [ ] **Dynamic Forms**: visibility rules на поля (IF field=value THEN show)
- [ ] **Notes & Attachments**: polymorphic note/file объекты, привязка к любой записи
- [ ] **Activity History**: unified view tasks + events для любого объекта с hasActivities
- [ ] **Field History Tracking** (ee/): до 20 полей per object, changelog table

---

### Phase 10: Formula Engine + Validation Rules ⬜

Вычисляемые поля и декларативная валидация.

- [ ] **Formula parser**: арифметика, строковые функции, date math, IF/CASE, cross-object refs
- [ ] **Formula fields**: read-only computed на уровне SOQL (SQL expression в SELECT)
- [ ] **Roll-Up Summary fields**: COUNT, SUM, MIN, MAX на master-detail parent
- [ ] **Validation Rules**: boolean formula → error message, checked before DML save
- [ ] **Default values**: formula или literal, applied on insert
- [ ] **Auto-number fields**: sequence-based auto-increment с форматом (INV-{0000})

---

### Phase 11: Advanced CRM Objects ⬜

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

### Phase 12: Full-Text Search (SOSL) ⬜

Поиск по всем объектам одновременно.

- [ ] PostgreSQL full-text search (tsvector/tsquery) или Elasticsearch/Meilisearch
- [ ] SOSL parser: `FIND {term} IN ALL FIELDS RETURNING Account(Name), Contact(Name, Email)`
- [ ] Индексация: trigger-based обновление search index при DML
- [ ] REST API: `POST /api/v1/sosl/search`
- [ ] Global search в UI: typeahead с SOSL backend

---

### Phase 13: Automation Engine ⬜

Декларативная и программная автоматизация.

#### Phase 13a: Trigger Framework (Go-based)

- [ ] Trigger registry: metadata-driven регистрация handlers
- [ ] Trigger interface: `BeforeInsert(ctx, records)`, `AfterUpdate(ctx, old, new)`, etc.
- [ ] Bulkification: handler получает slice записей, не одну
- [ ] Order of execution: documented, deterministic
- [ ] Recursion prevention: max depth, `TriggerContext.isExecuting`

#### Phase 13b: Flow Engine (декларативный)

- [ ] Flow definition (JSON/YAML): nodes, edges, conditions, actions
- [ ] Record-Triggered Flows: before save, after save
- [ ] Scheduled Flows: cron-based выполнение с фильтром записей
- [ ] Flow actions: create record, update record, send email, invoke REST
- [ ] Flow Builder UI (Vue.js): visual drag-and-drop editor

#### Phase 13c: Approval Processes

- [ ] Approval definition: entry criteria, steps, approvers
- [ ] Submit for approval → pending → approved/rejected
- [ ] Email notifications на каждом шаге
- [ ] Field updates on approve/reject
- [ ] Approval history на record detail

**Automation features для далёкой перспективы:**

| Возможность SF | Аналог |
|----------------|--------|
| Apex (custom language) | Go trigger handlers (compiled) |
| Process Builder | Flow Engine покрывает |
| Workflow Rules | Flow Engine покрывает |
| Assignment Rules | Record-Triggered Flow + Queue ownership |
| Escalation Rules | Scheduled Flow + criteria |

---

### Phase 14: Streaming & Integration ⬜

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

### Phase 15: Analytics — Reports & Dashboards ⬜

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
| **Territory Management** | Иерархия территорий, привязка accounts, territory-based sharing | Territory2 |
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
Phase 0 ✅ ──→ Phase 1 ✅ ──→ Phase 2 🔧 ──→ Phase 3 ──→ Phase 4 ──→ Phase 5
                                  │                │          │          │
                                  │                ▼          ▼          ▼
                                  │           Phase 10    Phase 13   Phase 7a
                                  │           (formulas)  (automation)(auth UI)
                                  │
                                  ▼
                              Phase 6 ──→ Phase 7b ──→ Phase 8
                           (std objects)  (record UI)  (dashboards)
                                                          │
                                                          ▼
                                            Phase 9 ──→ Phase 11
                                          (adv meta)   (adv objects)

                              Phase 12 (SOSL) — независимый, после Phase 3
                              Phase 14 (CDC) — независимый, после Phase 4
                              Phase 15 (Reports) — после Phase 3 + Phase 7
                              Phase N (ee/) — параллельно, после Phase 2
```

### Критический путь (MVP)

Минимальный набор для рабочей CRM:

```
Phase 2b/2c → Phase 3 → Phase 4 → Phase 5 → Phase 6 → Phase 7 → v0.1.0
```

Это покрывает: security → query → mutation → auth → standard objects → UI.

### Рекомендованный порядок после MVP

1. **Phase 8** — notifications + dashboard (CRM становится ежедневным инструментом)
2. **Phase 10** — formulas + validation (data quality)
3. **Phase 13a** — trigger framework (extensibility)
4. **Phase 14** — CDC + webhooks (integrations)
5. **Phase 12** — SOSL (search)
6. **Phase 9** — record types + layouts (multi-scenario)
7. **Phase 15** — reports (analytics)
8. **Phase 11** — advanced objects (full CRM suite)
9. **Phase 13b/c** — flows + approvals (no-code automation)

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
| **v0.1.0-alpha** | 0-2 | Metadata engine + security admin (текущее состояние) |
| **v0.2.0-alpha** | 3-4 | SOQL + DML — данные можно читать и писать через платформу |
| **v0.3.0-beta** | 5-6 | Auth + standard objects — можно логиниться и работать с CRM-данными |
| **v0.4.0-beta** | 7 | Полноценный UI — CRM можно использовать через браузер |
| **v0.5.0-beta** | 8 | Notifications + dashboards — CRM как рабочий инструмент |
| **v1.0.0** | 9-10 | Record types, formulas, validation — production-ready |
| **v1.x** | 11-15 | Advanced objects, search, automation, reports, integration |
| **v2.0** | N | Enterprise features, multi-tenant, advanced analytics |

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

*Этот документ обновляется по мере завершения фаз. Последнее обновление: 2026-02-08.*
