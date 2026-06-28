# Adverax CRM — Полная техническая спецификация (ТЗ)

**Статус:** реверс-спецификация по существующей кодовой базе
**Назначение:** документ может трактоваться как техническое задание для воспроизведения платформы с нуля.
**Версия платформы на момент составления:** v0.4.x-beta (Phase 0–10b, 9a–9c завершены; 37 ADR).
**Стек:** Go 1.25 · PostgreSQL 16 · Redis 7 · Vue 3 / TypeScript · Docker.
**Лицензия:** Open Core — ядро AGPL v3, каталог `ee/` под коммерческой лицензией.

---

## 1. Назначение и видение

Adverax CRM — самостоятельно размещаемая (self-hosted) CRM-платформа уровня Salesforce: метаданные управляют структурой данных, трёхслойная модель безопасности (OLS/FLS/RLS) встроена на уровне запросов, единый язык запросов SOQL и единый движок изменений DML. Цель — дать предприятию возможность создавать кастомные объекты, поля, связи, права доступа, бизнес-логику и UI **без программирования**, при этом без EAV: каждый объект — настоящая таблица PostgreSQL.

### 1.1 Ключевые принципы (обязательные к воспроизведению)

1. **Metadata-driven**: каждый объект (стандартный или кастомный) — реальная таблица `obj_{api_name}` с нативными ограничениями, FK, индексами. Никакого EAV, никаких JSON-блобов под данные.
2. **Security-by-construction**: OLS/FLS/RLS нельзя обойти. Любое чтение проходит через SOQL, любая запись — через DML; оба движка инжектируют ограничения безопасности на уровне SQL.
3. **Единый язык выражений**: CEL (Common Expression Language) используется везде — валидации, значения по умолчанию, видимость, условия автоматизаций, процедуры. Двойной стек: `cel-go` (бэкенд) + `cel-js` (фронтенд) с идентичной семантикой.
4. **Декларативность**: бизнес-логика (валидации, дефолты, процедуры, автоматизации, layout-ы) хранится как данные (строки/JSONB), а не как код.
5. **Модульный монолит**: единый бинарь и единая БД, но чёткие границы модулей с обменом через интерфейсы (`MetadataReader`, `UserContext` в shared kernel) — путь к будущему выделению сервисов.
6. **Open Core**: полная безопасность — в бесплатном ядре; платные расширения (территории, аудит, SSO) — в `ee/`, подключаются через build-теги.

### 1.2 Out of scope (сознательно не копируем у Salesforce)

Apex/собственный язык (заменён компилируемыми Go-хуками), Visualforce/Aura, SOAP API, мультитенантность (платформа однотенантна, ADR-0016), governor limits, 15/18-символьные ID (используем UUID v4, ADR-0001).

---

## 2. Общая архитектура

```
Vue 3 SPA  ──HTTP/JSON──►  Gin API  ──►  Middleware (RequestID, Logger, Recovery, JWTAuth)
                                              │
                              ┌───────────────┼─────────────────┐
                              ▼               ▼                 ▼
                        Handlers          Services         Modules (auth, …)
                              │               │
                              ▼               ▼
                   ┌──────────────────────────────────────────────┐
                   │  Platform engines                            │
                   │  SOQL │ DML │ CEL │ Procedure │ Automation    │
                   │  Metadata cache │ Security (OLS/FLS/RLS)      │
                   │  Credential │ Templates                       │
                   └──────────────────────────────────────────────┘
                              │
                   sqlc-репозитории (pgx/v5)
                              │
                        PostgreSQL 16  (schemas: metadata, iam, security, public.obj_*, ee)
                              │
                        Outbox worker (LISTEN/NOTIFY) → пересчёт effective-кэшей
```

**Поток данных:** `Handler → Service → SOQL/DML → Security → Repository → PostgreSQL`.
Ни один слой нельзя обойти: чтение всегда через SOQL, запись всегда через DML, оба применяют все три слоя безопасности.

### 2.1 Технологический стек

| Слой | Технология |
|------|-----------|
| Язык | Go 1.25 |
| HTTP | Gin (`gin-gonic/gin`) |
| БД | PostgreSQL 16 |
| Драйвер | `jackc/pgx/v5` + `pgxpool` |
| Запросы | sqlc (генерация типобезопасных запросов) |
| Миграции | golang-migrate |
| Парсеры SOQL/DML | `alecthomas/participle/v2` |
| CEL | `google/cel-go` (бэк), `@marcbachmann/cel-js` (фронт) |
| JWT | `golang-jwt/jwt/v5` |
| Пароли | `golang.org/x/crypto` bcrypt (cost=12) |
| OpenAPI | `getkin/kin-openapi`, `oapi-codegen` |
| Тесты БД | pgTAP (`pg_prove`) |
| Логи | `log/slog` (JSON в stdout) |
| Фронтенд | Vue 3 + TS + Pinia + Vue Router + Tailwind v4 + Radix/Reka UI + CodeMirror 6 |
| Сборка | Docker, docker-compose |

### 2.2 Структура репозитория

```
cmd/api/              — точка входа (main.go, setupRouter)
internal/
  pkg/                — инфраструктура: config, database, apperror, identity, logger, pagination, validator
  platform/           — доменные движки: metadata, security(+ols/fls/rls), soql, dml, cel,
                        procedure, automation, credential, templates
  modules/            — bounded contexts: auth (+ зарезервированные accounts/contacts/deals/…)
  handler/            — HTTP-обработчики
  middleware/         — RequestID, Logger, Recovery, JWTAuth, DevAuth
  service/            — прикладные сервисы (RecordService — generic CRUD)
  repository/         — sqlc-сгенерированный код доступа к данным
  api/                — oapi-codegen-сгенерированные типы/спека
  dto/                — DTO запросов/ответов
  testutil/           — тест-хелперы (jwt, fixtures, pgtest)
migrations/           — golang-migrate (схемы metadata/iam/security/public)
ee/                   — Enterprise Edition (территории, проприетарная лицензия)
sqlc/                 — sqlc.yaml + queries/
api/openapi.yaml      — единый источник истины по контракту
docs/adr/             — 37 ADR
docs/roadmap.md       — дорожная карта
web/                  — Vue 3 SPA
tests/pgtap/          — pgTAP-тесты схемы
```

### 2.3 Конфигурация (env)

Загружается один раз при старте через `config.Load()`.

| Переменная | Default | Назначение |
|-----------|---------|-----------|
| `PORT` | 8080 | HTTP-порт |
| `DB_HOST`/`DB_PORT`/`DB_USER`/`DB_PASSWORD`/`DB_NAME`/`DB_SSLMODE` | localhost/5432/crm/crm_secret/crm/disable | подключение к PostgreSQL |
| `LOG_LEVEL` | info | debug/info/warn/error |
| `JWT_SECRET` | — | **обязателен в проде**, HMAC-ключ |
| `JWT_ACCESS_TTL` | 15m | TTL access-токена |
| `JWT_REFRESH_TTL` | 168h | TTL refresh-токена |
| `ADMIN_INITIAL_PASSWORD` | — | сид пароля админа при первом запуске |
| `CREDENTIAL_ENCRYPTION_KEY` | — | hex-ключ 32 байта (AES-256-GCM для Named Credentials) |
| `RESET_PASSWORD_URL` | http://localhost:5173/reset-password | базовый URL для ссылки сброса пароля |

### 2.4 Загрузка приложения (bootstrap)

1. Инициализация slog (JSON → stdout).
2. `config.Load()`.
3. `database.NewPool(ctx, dsn)` → `*pgxpool.Pool`.
4. Создание `MetadataCache`, загрузка из БД (`cache.Load`).
5. `setupRouter(pool, cache, cfg)` — ручная сборка (DI) всех сервисов и middleware.
6. Запуск Outbox-воркера (пересчёт effective-кэшей безопасности).
7. Холодный пересчёт всех effective-кэшей (`RecomputeAll`).
8. HTTP-сервер на `:PORT`; graceful shutdown по SIGINT/SIGTERM.

---

## 3. Подсистема метаданных

### 3.1 Объекты (`metadata.object_definitions`)

Каждый объект описывается строкой метаданных и отражается в реальную таблицу.

Ключевые поля: `id (UUID)`, `api_name (UNIQUE, ^[a-z][a-z0-9_]*$)`, `label`, `plural_label`, `description`, `schema_name (default public)`, `table_name (UNIQUE)`, `object_type (standard|custom)`, `visibility (private|public_read|public_read_write|controlled_by_parent)`.

Поведенческие флаги: `is_platform_managed`, `is_visible_in_setup`, `is_custom_fields_allowed`, `is_deleteable_object`.
Флаги операций: `is_createable`, `is_updateable`, `is_deleteable`, `is_queryable`, `is_searchable`.
Фичи: `has_activities`, `has_notes`, `has_history_tracking`, `has_sharing_rules`.
Аудит: `created_at`, `updated_at`.

### 3.2 Таблица данных объекта (`obj_{api_name}`)

Создаётся/изменяется DDL-генератором в рантайме. Системные колонки каждой таблицы:

```sql
id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
owner_id    UUID NOT NULL,                 -- владелец записи (RLS)
created_by  UUID NOT NULL,
created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
updated_by  UUID NOT NULL,
updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
-- + пользовательские поля (через миграции/DDL)
```

Для объектов с `visibility != public_read_write` создаётся таблица шаринга `obj_{name}__share` (см. §5.5).

### 3.3 Поля (`metadata.field_definitions`)

Двухуровневая иерархия `field_type` (хранение) + `field_subtype` (семантика), ADR-0004.

Ключевые поля: `id`, `object_id (FK)`, `api_name (UNIQUE per object)`, `label`, `description`, `help_text`, `field_type`, `field_subtype`, `referenced_object_id (FK, nullable)`, `is_required`, `is_unique`, `config (JSONB)`, `is_system_field`, `is_custom`, `is_platform_managed`, `sort_order`, `created_at`, `updated_at`.

**Полный реестр типов и подтипов:**

| field_type | подтипы | хранение |
|-----------|---------|----------|
| `text` | plain, area, rich, email, phone, url | VARCHAR(n)/TEXT |
| `number` | integer, decimal, currency, percent, auto_number | NUMERIC(p,s)/sequence |
| `boolean` | — | BOOLEAN |
| `datetime` | date, datetime, time | DATE/TIMESTAMPTZ/TIME |
| `picklist` | single, multi | VARCHAR / VARCHAR[] |
| `reference` | association, composition, polymorphic | UUID / (VARCHAR+UUID) |

**Семантика reference-подтипов (ADR-0005):**

| подтип | FK | on_delete | NOT NULL | self-ref | особенность |
|--------|----|-----------|----------|----------|-------------|
| association | да | set_null \| restrict | нет | разрешён | мягкая связь, объекты независимы |
| composition | да | cascade \| restrict | да | запрещён | жёсткая связь, наследование владельца, контроль reparent |
| polymorphic | нет | в коде | — | разрешён | две колонки (object_type, record_id); цели в `polymorphic_targets`; FK отсутствует, проверка в DML |

**Примеры `config` (JSONB):**
```jsonc
// text/plain
{ "max_length": 255, "default_value": "" }
// number/decimal
{ "precision": 18, "scale": 2, "default_value": "0" }
// picklist
{ "picklist_id": "uuid|null", "values": [{"value":"new","label":"New","sort_order":1,"is_default":true,"is_active":true}], "default_value": "new" }
// reference/association
{ "relationship_name": "Contacts", "on_delete": "set_null" }
// reference/composition
{ "relationship_name": "LineItems", "on_delete": "cascade", "is_reparentable": false }
// дефолты (любой тип, для DML Stage 3)
{ "default_expr": "user.id", "default_on": "create" }   // CEL
```

### 3.4 Сопутствующие таблицы метаданных

- `metadata.picklist_definitions` / `metadata.picklist_values` — глобальные пиклисты.
- `metadata.polymorphic_targets (field_id, object_id)` — допустимые цели полиморфных ссылок.
- `metadata.translations (resource_type, resource_id, field_name, locale, value)` — i18n меток/описаний (ADR-0002: дефолт inline + таблица переводов).
- `metadata.relationship_registry` — материализованное представление/кэш связей, **производное** от `field_definitions` + `polymorphic_targets` (ADR-0006). Отдельной таблицы связей нет.

### 3.5 In-memory кэш метаданных (`MetadataCache`)

Потокобезопасный (RWMutex) кэш, реализует интерфейс `MetadataReader` (ADR-0030), от которого зависят все потребители (SOQL, DML, handlers).

Содержит карты: объекты по id/api_name, поля по id/object_id, прямые/обратные связи (`forwardRels`/`reverseRels`), правила валидации, функции, порталы, layout-ы, shared_layouts, процедуры по code, правила автоматизации.

`RelationshipInfo` (кэшируемая запись связи): `FieldID`, `FieldAPIName`, `RelationshipName`, `Child/ParentObjectID(+APIName)`, `ReferenceSubtype`, `OnDelete`.

**Инвалидация:** изменение полей/полиморфных целей → полная перестройка; изменение порталов/layout-ов/правил — частичная перезагрузка соответствующих карт.

---

## 4. Безопасность — трёхслойная модель (OLS → FLS → RLS)

Слои проверяются последовательно; отказ OLS прекращает оценку FLS/RLS.

### 4.1 Идентичности и права

- `iam.users`: id, username (UNIQUE), email (UNIQUE), first_name, last_name, `profile_id (FK, RESTRICT)`, `role_id (FK, SET NULL, nullable)`, `password_hash`, `is_active`.
- `iam.profiles`: id, api_name (UNIQUE), label, description, `base_permission_set_id (FK, RESTRICT)`. Профиль — особый permission set: его базовый PS держит OLS/FLS по умолчанию.
- `iam.permission_sets`: id, api_name (UNIQUE), label, description, `ps_type (grant|deny)`.
- `iam.permission_set_to_users (permission_set_id, user_id, UNIQUE)` — прямые назначения PS пользователю.
- `iam.user_roles`: id, api_name (UNIQUE), label, `parent_id (self-FK, SET NULL)` — иерархия ролей.
- `iam.groups`: id, api_name, label, `group_type (personal|role|role_and_subordinates|public|territory)`, `related_role_id`, `related_user_id`, `related_territory_id (ee)`.
- `iam.group_members (group_id, member_user_id XOR member_group_id)` — прямое членство, поддержка вложенности.
- `iam.refresh_tokens`, `iam.password_reset_tokens` — токены (хранятся как SHA-256 hash).

### 4.2 OLS — Object-Level Security

Формула эффективных прав:
```
effective = (base_ps | grant_ps1 | … ) & ~(deny_ps1 | … )
```
Битовая маска CRUD: `Read=1, Create=2, Update=4, Delete=8`.

Таблицы: `security.object_permissions (permission_set_id, object_id, permissions INT, UNIQUE)`.
Кэш: `security.effective_ols (user_id, object_id, permissions, computed_at)`.

Энфорсер OLS: `CanRead/CanCreate/CanUpdate/CanDelete(ctx, userID, objectID)` — читает из `effective_ols`. Применяется в Describe (фильтрация объектов), в навигации (пересечение с OLS), в SOQL/DML.

### 4.3 FLS — Field-Level Security

Та же формула grant/deny. Маска: `Read=1, Write=2`. Поле без grant-записи → доступ 0 (скрыто).

Таблицы: `security.field_permissions (permission_set_id, field_id, permissions INT, UNIQUE)`.
Кэши: `security.effective_fls (user_id, field_id, permissions)` и `security.effective_field_lists (user_id, object_id, mask, field_names TEXT[])` — готовые списки читаемых/редактируемых полей.

Энфорсер FLS: `CanReadField/CanWriteField`, `GetReadableFields/GetWritableFields`. Применяется в Describe (фильтрация полей), при формировании SELECT, при проверке записи. Системные поля (Id, OwnerId, CreatedAt, …) всегда доступны.

### 4.4 RLS — Row-Level Security

OWD (`object_definitions.visibility`):

| visibility | чтение | запись |
|-----------|--------|--------|
| private | владелец + иерархия ролей (вниз) + шаринг | владелец + шаринг |
| public_read | все с OLS Read | владелец + шаринг |
| public_read_write | все | все с OLS Update |
| controlled_by_parent | наследуется от родителя | наследуется |

**Иерархия ролей:** менеджер (роль-предок) видит записи подчинённых (только Read). Реализуется через closure-таблицу.

**Группы (ADR-0013):** 4 типа, автогенерация:
- `personal` — при создании пользователя (member = он сам).
- `role` и `role_and_subordinates` — при создании роли.
- `public` — вручную, может включать пользователей/роли/вложенные группы.

Единый «грантополучатель» (grantee) всегда `group_id` (не полиморфный). Ручной шаринг пользователю → грант его personal-группе.

### 4.5 Таблицы шаринга (`obj_{name}__share`)

Создаются для объектов с `visibility != public_read_write`:
```sql
id UUID PK,
record_id UUID NOT NULL REFERENCES obj_{name}(id) ON DELETE CASCADE,
grantee_id UUID NOT NULL REFERENCES iam.groups(id) ON DELETE CASCADE,
access_level INT NOT NULL,            -- 1=read, 5=read+update
reason VARCHAR(30) CHECK (reason IN ('owner','sharing_rule','territory','manual')),
created_at TIMESTAMPTZ,
UNIQUE (record_id, grantee_id, reason)
```

### 4.6 Sharing Rules (`security.sharing_rules`)

`rule_type (owner_based|criteria_based)`, `source_group_id`, `target_group_id`, `access_level (read|read_write)`, для criteria — `criteria_field(_id)`, `criteria_operator (eq|neq|in|gt|lt)`, `criteria_value`. Генерация share-записей — асинхронно через outbox.

### 4.7 Effective-кэши и инвалидация (ADR-0012)

Closure-таблицы (предвычисленные иерархии):
- `security.effective_role_hierarchy (ancestor_role_id, descendant_role_id, depth)`.
- `security.effective_object_hierarchy (ancestor_object_id, descendant_object_id, depth)`.
- `security.effective_visible_owner (user_id, visible_owner_id)` — каких владельцев видит пользователь по иерархии (read-only).
- `security.effective_group_members (group_id, user_id)` — плоское членство (раскрытие вложенности).

**Outbox-паттерн:** `security.security_outbox (id BIGSERIAL, event_type, entity_type, entity_id, payload JSONB, created_at, processed_at)`. Триггер `AFTER INSERT` вызывает `pg_notify('security_outbox', id)`. Go-воркер слушает (LISTEN/NOTIFY + fallback-свип каждые 30 c), пересчитывает только затронутые кэши.

Матрица событий → инвалидируемые кэши: смена профиля/роли пользователя → ols/fls/visible_owner; смена parent роли → role_hierarchy + visible_owner; смена членства группы → group_members; смена PS → ols/fls/field_lists; смена visibility/parent объекта → object_hierarchy.

### 4.8 Инжекция RLS в запросы

Для `private`/`controlled_by_parent` (и для записи при `public_read`) генерируется WHERE:
```sql
WHERE (
  t.owner_id = :user_id
  OR t.owner_id IN (SELECT visible_owner_id FROM security.effective_visible_owner WHERE user_id = :user_id)
  OR t.id IN (SELECT record_id FROM obj_{name}__share
              WHERE grantee_id IN (SELECT group_id FROM security.effective_group_members WHERE user_id = :user_id))
)
```
- **SOQL (чтение):** `soql/executor.go` строит WHERE, добавляет к существующему через AND либо вставляет перед ORDER BY/LIMIT, перенумеровывает плейсхолдеры.
- **DML (UPDATE/DELETE):** `dml/executor.go` инжектирует тот же фильтр (для INSERT — нет; owner_id ставится в сервисе).

### 4.9 Холодный старт кэшей

При запуске: `EffectiveComputer.RecomputeAll()` (OLS/FLS всех пользователей) → `RecomputeRoleHierarchy` → `RecomputeVisibleOwnersAll` → `RecomputeGroupMembersAll` → `RecomputeObjectHierarchy`.

---

## 5. SOQL — движок запросов

Единственная точка чтения данных; безопасность инжектируется автоматически. Парсер на `participle/v2`.

### 5.1 Грамматика

```
SELECT [ROW] <select-list>
FROM <object>
[WHERE <expr>] [GROUP BY <field>] [HAVING <expr>]
[ORDER BY <field> [ASC|DESC] [NULLS FIRST|LAST]]
[LIMIT n] [OFFSET n] [FOR UPDATE] [WITH SECURITY_ENFORCED]
```

Элементы SELECT: простое поле; dot-нотация до 5 уровней (`Account.Owner.Manager.Email` → JOIN-ы); агрегаты `COUNT, COUNT_DISTINCT, SUM, AVG, MIN, MAX`; скалярные функции `COALESCE, UPPER, LOWER, TRIM, LENGTH, SUBSTRING, ABS, ROUND, FLOOR, CEIL, NULLIF, CONCAT`; `TYPEOF … WHEN … THEN … ELSE … END` (полиморфизм); подзапрос parent→child `(SELECT … FROM Contacts WHERE …)`.

`SELECT ROW` — гарантированно ≤1 запись (используется как scalar-источник в порталах; LIMIT 2, >1 строки — ошибка).

WHERE: приоритет `OR > AND > NOT > сравнения`. Операторы `=, ==, !=, <>, >, <, >=, <=, IN/NOT IN (values|subquery), LIKE/NOT LIKE, IS NULL/IS NOT NULL`, арифметика `+ - * / % ||`. Динамические даты: `TODAY(), YESTERDAY(), TOMORROW(), THIS_WEEK, LAST_WEEK, THIS_MONTH, LAST_N_DAYS:N`; статические ISO-8601.

### 5.2 Конвейер исполнения

`Parse → Validate (по метаданным + AccessController) → Resolve relationships → Compile SQL (параметризация $1,$2…) → Execute (pgx) → inject RLS`.

`CompiledQuery`: `SQL`, `Params`, `DateParams` (резолв на момент выполнения), `Shape` (маппинг SOQL-имя → SQL-алиас + типы), `Pagination`, `Dependencies` (для инвалидации), `ForUpdate`, `WithSecurityEnforced`, `IsRow`.

`WITH SECURITY_ENFORCED` → строгий режим (ошибка при отсутствии доступа); без него — мягкая фильтрация.

### 5.3 Адаптеры и лимиты

`MetadataAdapter` (MetadataReader → engine.MetadataProvider, маппинг типов, добавление системных полей), `AccessControllerAdapter` (OLS/FLS → engine.AccessController).

Лимиты: длина запроса ≤99 999, ≤1000 полей SELECT, ≤500 сравнений WHERE, ≤10 ORDER BY, LIMIT ≤2000, dot-traversal ≤5 уровней.

---

## 6. DML — движок изменений

Единственная точка записи. Парсер на `participle/v2`. Поддержка `INSERT (multi-row до 10 000), UPDATE … SET … WHERE, DELETE … WHERE, UPSERT … ON <external_id>`.

### 6.1 Конвейер (типизированные стадии, Option-паттерн, ADR-0020)

1. **Parse** — DMLStatement AST.
2. **Resolve** — ObjectMeta + эффективный ruleset.
3. **Defaults** (опц.) — внедрение отсутствующих полей; каскад Layout > Portal > Metadata; CEL `default_expr` приоритетнее статического `default_value`; учёт `default_on (create|update|create,update)`.
4a. **Validate (constraints)** — required, типы, FLS на запись.
4b. **Validate (rules)** — CEL-правила (`metadata.validation_rules`), аддитивный AND по каскаду; severity `error` блокирует, `warning` — нет.
5. **Compute** (опц.) — пересчёт формульных полей (план на будущее).
6. **Compile** — CompiledDML (SQL + params, RETURNING).
7. **Execute** — выполнение + инжекция RLS для UPDATE/DELETE.
8. **PostExecute** (опц.) — хук автоматизаций (CompositePostExecuteHook).

Интерфейсы стадий: `MetadataProvider`, `WriteAccessController (CanWriteObject, CheckWritableFields)`, `DefaultResolver`, `RuleValidator`, `Executor`. Подключение через опции `WithMetadata/WithWriteAccessController/WithDefaultResolver/WithRuleValidator/WithExecutor`.

Сборка движка:
```go
engine := dml.NewEngine(
  dml.WithMetadata(metaAdapter),
  dml.WithWriteAccessController(flsEnforcer),
  dml.WithDefaultResolver(celDefaultResolver),
  dml.WithRuleValidator(celRuleValidator),
  dml.WithExecutor(rlsExecutor),
)
```

Лимиты: INSERT ≤10 000 строк, UPDATE ≤100 000, ≤1000 полей.

### 6.2 Правила валидации (`metadata.validation_rules`)

Поля: `object_id`, `api_name`, `expression (CEL→bool)`, `when_expression (CEL, опц.)`, `applies_to (create|update|create,update)`, `error_code`, `error_message`, `severity`, `sort_order`, `is_active`.

---

## 7. CEL — единый язык выражений

Двойной стек: `cel-go` (бэк) + `cel-js` (фронт, `@marcbachmann/cel-js`). Семантика идентична; функции (`fn.*`) синхронизируются через Describe API.

### 7.1 Окружения (env) и переменные

| Env | Переменные |
|-----|-----------|
| StandardEnv (валидации) | `record`, `old (nil при INSERT)`, `user{id,profile_id,role_id}`, `now` |
| DefaultEnv (дефолты) | `record`, `user`, `now` |
| PortalEnv (аргументы портала) | `args` |
| GateEnv (гейты, ADR-0037) | `args`, `datasets`, `data`, `user` |

`ProgramCache` — потокобезопасный кэш скомпилированных программ (compile-once, eval-many): `GetOrCompile`, `EvaluateBool`, `EvaluateAny`, `Reset`. Ошибки: CompileError / EvalError.

### 7.2 Кастомные функции (`fn.*`, ADR-0026)

`metadata.functions`: `name (UNIQUE, ^[a-z][a-z0-9_]*$)`, `description`, `params (JSONB, ≤10, типы string|number|boolean|list|map|any)`, `return_type`, `body (CEL, ≤4096)`.

`FunctionRegistry` прекомпилирует все тела при старте, регистрирует в env как `fn.<name>(...)`. Композиция допускается (`fn.total(fn.discount(...), ...)`).

Ограничения безопасности: тело ≤4 КБ, исполнение ≤100 мс, вложенность ≤3, ≤10 параметров, ≤200 функций; запрет рекурсии (статанализ DetectCycles) и циклических зависимостей.

---

## 8. Метаданные UI: Describe, Portal, Layout, Form, Navigation

### 8.1 Describe API

- `GET /api/v1/describe` — список объектов для навигации (фильтр OLS).
- `GET /api/v1/describe/:objectName` — поля (фильтр FLS) + вычисленная форма (Portal+Layout merge). Заголовки `X-Form-Factor: desktop|tablet|mobile`, `X-Form-Mode: read|view`.

Алгоритм резолва формы: загрузка ObjectDefinition → проверка OLS → построение полей (системные + пользовательские) → фильтр FLS → резолв Portal (или автогенерация) → резолв Layout по fallback-цепочке → merge Portal+Layout → разрешение `layout_ref` shared-layout-ов → пересечение с FLS → ответ `{fields, form}`.

Принцип: Portal/Layout **сужают, но не расширяют** доступ; фронтенд видит только то, что разрешено OLS/FLS/RLS.

### 8.2 Portal — bounded context на профиль (ADR-0022/0035/0037)

`metadata.portals`: `id`, `profile_id (nullable)`, `api_name (UNIQUE)`, `label`, `config (JSONB)`. Портал **отвязан от объекта** (Portal Unbinding, 9b).

Модель — **граф гейтов** (ADR-0037). Config:
```
{ args[], arg_rules[], entry_gate, gates{ <key>: {
    label, args[], arg_rules[],
    body[ { name, type:"soql"|"dml", soql?, dml?, when?, page_size?, restrict_filters[] } ],
    layout{ fields[], root?, section_config{}, field_config{}, list_config? },
    outcomes[ { name, gate, label, icon, type, args_template{<key>:CEL} } ]
}}}
```

Тип запроса выводится из синтаксиса (`SELECT ROW` → scalar, `SELECT` → list). HTTP-семантика гейта выводится из состава body: только SOQL → view-гейт (GET); SOQL+DML с layout → form-гейт (GET показывает, POST исполняет); только DML/без layout → action-гейт (POST, авто-сворачивание).

Пагинация/фильтрация выводятся из SELECT + метаданных полей (operators по типу поля, значения пиклистов). Лимиты: ≤20 гейтов, ≤10 шагов/гейт, ≤10 outcomes/гейт, ≤20 args/гейт.

Вычисляемые поля портала образуют DAG; при сохранении — детекция циклов (алгоритм Кана).

### 8.3 Layout (ADR-0027)

`metadata.layouts`: `portal_id`, `form_factor (desktop|tablet|mobile)`, `mode (read|view)`, `config (JSONB)`, `UNIQUE(portal_id, form_factor, mode)`.

Config: дерево компонентов `root` (grid/group/highlights/field_section/related_list/query_widget/actions_bar/activity_feed/tabs), `section_config` (columns, collapsed, visibility_expr), `field_config` (col_span, ui_kind, required_expr, readonly_expr, reference{display_fields, search_fields, target}, visibility_expr), `list_config` (columns{field,label,width,align,sortable,filterable,ui_kind}, sort_by, search, row_actions).

`ui_kind` (короткая форма — строка, длинная — объект): ввод (text, textarea, number, select, multi_select, checkbox, toggle, radio, date, datetime, time, lookup, file_upload); отображение (badge, avatar, link, progress, rating, relative_time, currency, percent, color, template).

**Shared Layouts** (`metadata.shared_layouts`): `api_name (UNIQUE)`, `type (field|section|list)`, `config`. Подключение через `layout_ref` (мелкий merge; inline-переопределения побеждают). Удаление защищено RESTRICT.

Fallback-цепочка: запрошенный (form_factor, mode) → тот же form_factor любой mode → desktop тот же mode → desktop read → автогенерация.

### 8.4 Profile Navigation (ADR-0032)

`metadata.profile_navigation (profile_id UNIQUE, config JSONB)`. Config:
```
{ groups[ { key, label, icon, items[ {type:"object"|"page"|"link"|"divider", object_api_name?, portal_api_name?, label?, url?, icon?} ] } ] }
```
Резолв `GET /api/v1/navigation`: пересечение с OLS (скрытие недоступных объектов и пустых групп), fallback к плоскому OLS-списку. Лимиты: ≤20 групп, ≤50 элементов/группу. Тип `page` рендерит портал как самостоятельную страницу.

### 8.5 App Templates (ADR-0018)

Паттерн Registry + Applier, шаблоны — Go-код в бинаре (типобезопасность). Применение в два прохода (объекты → поля), guard `object_definitions.count == 0`, автоназначение полного OLS/FLS админ-профилю.

Встроенные: **Sales CRM** (Account, Contact, Opportunity, Task — 4 объекта, 36 полей), **Recruiting** (Position, Candidate, Application, Interview — 4 объекта, 28 полей).

---

## 9. Автоматизация

### 9.1 Procedure Engine (ADR-0024)

Декларативные процедуры (JSON DSL), хранение версионируемое (`metadata.procedures` + `metadata.procedure_versions`, статусы draft/published/superseded, авто-инкремент версии).

**6 типов команд:**

| тип | назначение | примеры |
|-----|-----------|---------|
| record | данные через DML | record.create/update/delete/get/query |
| compute | трансформации | compute.transform/validate/aggregate/fail |
| flow | управление | flow.call (процедура), flow.start (сценарий), flow.if, flow.match, flow.try |
| notification | коммуникации | notification.email/sms/push/in_app |
| integration | внешние API | integration.http (через Named Credentials) |
| wait | время/сигналы | wait.signal/timer/until |

DSL: массив `commands[]`, каждая `{ type, …поля, as?, when?, optional?, retry{max_attempts,delay_ms,backoff_mult}?, rollback{…}? }`, плюс `result{}`. Любое строковое значение, начинающееся с `$`, — CEL.

Контекст выражений: `$.input`, `$.user`, `$.now`, `$.<stepname>` (результат шага с `as`), `$.steps.<name>`, `$.warnings`, `$.error` (в rollback).

**Saga rollback:** LIFO-стек компенсаций; при ошибке на шаге N откатываются N-1…1 (с доступом к `$.error`); не все шаги имеют компенсацию.

Лимиты: ≤30 c, ≤50 команд, вложенность call ≤3, if/match ≤5, JSON ≤64 КБ, ≤10 HTTP, ≤10 нотификаций. Нет циклов/произвольного кода/ФС. Все `record.*` идут через DML (OLS/FLS/RLS).

Constructor UI: форменный билдер (CommandPicker, CommandEditor, KeyValueEditor), Expression Builder, dry-run без сайд-эффектов.

### 9.2 Automation Rules (ADR-0031)

`metadata.automation_rules`: `object_id`, `name`, `event_type (before/after × insert/update/delete)`, `condition (CEL|null)`, `procedure_code`, `execution_mode (per_record|per_batch)`, `sort_order`, `is_active`, `UNIQUE(object_id, name)`.

Переменные условия: `new`, `old`, `user`, `now`. Граница транзакции (гибрид): before_* + DML + after_* (data) — в одной TX; integration/notification — в асинхронный outbox. Рекурсия ограничена (depth ≤3), ≤20 правил/объект/событие, таймаут 30 c. Подключается как PostExecuteHook (Stage 8 DML).

### 9.3 Named Credentials (ADR-0028)

`metadata.credentials`: `code (UNIQUE)`, `name`, `type (api_key|basic|oauth2_client)`, `base_url (HTTPS only)`, `auth_data_encrypted (BYTEA)`, `auth_data_nonce`, `is_active`. `metadata.credential_tokens` — кэш OAuth2-токена (зашифрован). `metadata.credential_usage_log` — аудит вызовов.

Шифрование AES-256-GCM, мастер-ключ из `CREDENTIAL_ENCRYPTION_KEY`, уникальный nonce на запись. SSRF-защита: только HTTPS, совпадение хоста base_url+path, блок-лист внутренних подсетей. В процедуре допустима только ссылка `"credential": "code"`, инлайн-URL/заголовки запрещены.

### 9.4 Scenario Engine (ADR-0025, план — Phase 14a)

Durable async-оркестрация: `metadata.scenarios` + `scenario_versions`; рантайм `scenario_executions (status: pending/running/waiting/compensating/completed/failed/cancelled, context JSONB)` + `scenario_step_history`. Шаги вызывают процедуры/inline-команды; ожидание сигналов/таймеров (`POST /executions/{id}/signal`). Идемпотентность `{executionId}-{stepCode}`; пиннинг версий процедур в `scenario_run_snapshots`; восстановление после рестарта. Saga-компенсация LIFO.

### 9.5 Стратегия версионирования (ADR-0029)

Полное версионирование (draft/published/superseded, целочисленный инкремент) — только у Procedure и Scenario. Остальное (функции, валидации, автоматизации, порталы, layout-ы, credentials) — save=live с проверкой при сохранении. Retention superseded — последние 10 версий.

---

## 10. HTTP API

OpenAPI как единственный источник истины (`api/openapi.yaml`). Базовый префикс `/api/v1`. JWT Bearer на всех защищённых эндпойнтах.

### 10.1 Конвенции

- Успех: одиночная сущность `{ "data": <entity> }`; список `{ "data": [...], "pagination": {page, per_page, total, total_pages} }`; 204 без тела.
- Ошибка: `{ "error": { "code": "...", "message": "..." } }`.
- Коды: 400 BAD_REQUEST, 401 UNAUTHORIZED, 403 FORBIDDEN (OLS/FLS), 404 NOT_FOUND, 409 CONFLICT, 422 (ошибка DML с деталями), 500 INTERNAL.
- Пагинация: `page` (1), `per_page` (20, max 100); SOQL `pageSize` (100, max 2000); гейты `per_page` (20, max 200).
- Заголовки: `Authorization`, `X-Request-ID` (генерится при отсутствии), `X-Form-Factor`, `X-Form-Mode`, `X-Dev-User-Id` (dev).
- DTO: snake_case в JSON; обязательные поля помечены `required` в OpenAPI → value-типы в Go и non-optional в TS.

### 10.2 Каталог эндпойнтов (сводно)

**Система:** `GET /healthz`.

**Auth:** `POST /auth/login` (rate-limit 5/15 мин/IP), `/auth/refresh`, `/auth/logout`, `GET /auth/me`, `POST /auth/forgot-password`, `POST /auth/reset-password`, `PUT /admin/security/users/{id}/password`.
JWT claims: `sub (user_id)`, `pid (profile_id)`, `rid (role_id)`. HMAC-SHA256. Refresh — ротация, SHA-256 hash в БД. bcrypt cost=12.

**SOQL/DML:** `GET|POST /query`, `POST /data`; admin-инструменты дизайн-тайма: `/admin/soql/validate|test|objects|objects/{name}/fields`, `/admin/dml/validate|test|objects|…`, `POST /admin/cel/validate`.

**Generic CRUD:** `GET|POST /records/:objectName`, `GET|PUT|DELETE /records/:objectName/:recordId`.

**Describe:** `GET /describe`, `GET /describe/:objectName`.

**Metadata admin:** objects, fields (`/objects/{id}/fields`), validation rules (`/objects/{id}/rules`), functions, procedures (+ draft/publish/rollback/versions/execute/dry-run), automation-rules, portals/object-views, layouts, shared-layouts, navigation, credentials (+ test/activate/deactivate/usage), templates (list/apply).

**Security admin:** roles, permission-sets (+ object-permissions, field-permissions), profiles, users (+ permission-sets), groups (+ members), sharing-rules.

**Portal runtime:** `GET /portal/:apiName`, `GET /portal/:apiName/gate/:gateName` (SOQL, 405 при DML), `POST /portal/:apiName/gate/:gateName` (DML), `GET /portal/:apiName/query/:queryName`, `GET /view/:ovApiName`, `GET /navigation`.

### 10.3 Контрактное тестирование (ADR-0021)

Бэк: тесты handlers валидируют ответы против OpenAPI через `kin-openapi` (`openapi3filter.ValidateResponse`) middleware. Фронт: `openapi-typescript` генерирует `web/src/types/openapi.d.ts`, поверх — `CamelCaseKeys<T>`. Команда `make generate-api` регенерирует Go (oapi-codegen: gin,types,spec) и TS типы. Workflow: правка спеки → генерация → правка кода → `go test ./internal/handler/...` + `npm run type-check`.

---

## 11. Фронтенд (Vue 3 SPA)

### 11.1 Стек

Vue 3 (Composition API, `<script setup>`), TypeScript 5.9, Vite 7, Vue Router 5, Pinia 3, Tailwind v4 (`@tailwindcss/vite`), Radix Vue + Reka UI, Lucide icons, CodeMirror 6 (+ `vue-codemirror`), `@marcbachmann/cel-js`, `@vueuse/core`, vue-draggable-plus, vue-sonner. Тесты: Vitest (unit), Playwright (e2e, Chromium).

### 11.2 Структура `web/src`

`api/` (HttpClient + по-модульные клиенты, авто camel↔snake, авто-refresh на 401, токены в localStorage), `stores/` (Pinia: auth, metadata, records, functions, securityAdmin, permissionEditor, territoryAdmin), `router/`, `layouts/` (AppLayout, AdminLayout), `views/` (auth, app/*, admin/{metadata,security,territory}), `components/` (ui/, app/, admin/{soql-editor, expression-builder, layouts, procedures, portal, metadata, security}, records/{FieldRenderer, FieldDisplay}), `composables/`, `lib/` (case, cel, codemirror), `types/` (openapi.d.ts + домены).

### 11.3 Две UI-зоны

- **`/app/*`** (рабочая зона CRM): `/app` (Welcome), `/app/page/:portalApiName`, `/app/:objectName` (динамический список), `/app/:objectName/new` (динамическая форма), `/app/:objectName/:recordId` (детали + related lists). Рендер полностью из Describe API + Form.
- **`/admin/*`** (администрирование): metadata (objects/fields/rules/functions/procedures/credentials/automation-rules/portals/layouts/shared-layouts/navigation), security (roles/permission-sets/profiles/users/groups/sharing-rules), territory (models/territories), templates. Каждый раздел — List/Create/Detail.

### 11.4 Ключевые компоненты

`FieldRenderer` (выбор инпута по типу/подтипу), `FieldDisplay` (форматирование), `SoqlEditor` (CodeMirror + контекстный автокомплит + валидация + test), `ExpressionBuilder` (CEL: FieldPicker/FunctionPicker, live-preview через cel-js, autocomplete), визуальный Layout Constructor (Form Layout / List Config / JSON табы, канвас секций, DnD колонок), Procedure Constructor, Portal Constructor (General/Gates/Args), редактор прав (OLS/FLS).

### 11.5 Генерация типов и сборка

`npm run generate:types` (из `../api/openapi.yaml`), `npm run dev` (Vite + proxy `/api`→:8080), `npm run build` (type-check + сборка), `npm run test:unit`, `npm run test:e2e`.

---

## 12. База данных и инфраструктура

### 12.1 Схемы

`metadata` (объекты, поля, связи, пиклисты, переводы, валидации, функции, процедуры+версии, credentials, automation_rules, порталы, layouts, shared_layouts, profile_navigation), `iam` (users, profiles, permission_sets, user_roles, groups, group_members, токены), `security` (object/field_permissions, effective_*, sharing_rules, security_outbox), `public` (таблицы `obj_*` + `obj_*__share`), `ee` (территории).

### 12.2 Миграции и генерация

golang-migrate: `migrations/` (ядро) и `ee/migrations/` (через отдельную `x-migrations-table`). sqlc генерирует репозитории (`pgx/v5`, override uuid→google/uuid, timestamptz→time.Time) из `sqlc/queries/`. pgTAP-тесты схемы в `tests/pgtap/` (запуск `pg_prove`).

### 12.3 Docker

- `api` (Dockerfile: multi-stage golang:1.25-alpine → alpine:3.19, бинарь + миграции, EXPOSE 8080).
- `postgres` (кастомный образ `docker/postgres`, init + pgTAP-маунты, healthcheck `pg_isready`).
- `redis:7-alpine` (зарезервирован).
- env из `.env`.

### 12.4 Makefile (целевые команды)

`build, run, test (race), lint (golangci-lint), fmt`, `migrate-up/down/create`, `sqlc-generate`, `generate-api`, `web-generate-types`, `docker-up/down/build/reset`, `test-integration`, `test-pgtap`.

---

## 13. Enterprise Edition (`ee/`)

### 13.1 Модель Open Core (ADR-0014)

Ядро (AGPL v3): метаданные, SOQL/DML, полная безопасность OLS/FLS/RLS, JWT-auth, Vue-фронт, шаблоны, процедуры/автоматизации. `ee/` (коммерческая лицензия): Territory Management, Audit Trail, SSO/SAML, Field History, расширенная аналитика и пр.

Интеграция через интерфейсы в ядре + build-теги: `TerritoryResolver`/`TerritoryAssignmentEvaluator` с noop-реализацией (`//go:build !enterprise`) и полной в `ee/` (`//go:build enterprise`). Сборка: `go build ./cmd/api` (Community) / `go build -tags enterprise ./cmd/api` (Enterprise).

### 13.2 Territory Management (ADR-0015)

`ee.territory_models (status: planning|active|archived, ≤1 active)`, `ee.territories (model_id, parent_id, …)`, `ee.territory_object_defaults (territory_id, object_id, access_level)`, `ee.user_territory_assignments`, `ee.record_territory_assignments (record_id, object_id, territory_id, reason)`, `ee.territory_assignment_rules (criteria_field, criteria_op, criteria_value, rule_order)`.

Видимость через ancestor-walk: запись, назначенная территории, порождает share-записи для каждого предка территории, у которого есть object_default (с его access_level), reason='territory'. Кэши: `security.effective_territory_hierarchy`, `security.effective_user_territory`. Автогруппы типа `territory`.

PL/pgSQL-функции (производительность активации): `ee.rebuild_territory_hierarchy(model_id)` (recursive CTE closure), `ee.generate_record_share_entries(...)`, `ee.activate_territory_model(new_model_id)` (полная оркестрация в одной транзакции). Оценка правил назначения — синхронно в DML.

### 13.3 Industry Modules (ADR-0034) и Offline Sync (ADR-0033) — план

**Industry Modules:** компилируемые Go-модули, реализующие интерфейс `Module` (Name/Version/Dependencies/OnInstall/RegisterRoutes/PostExecuteHooks), реестр с топосортировкой, трекинг `metadata.module_installations`, маршруты `/api/v1/modules/{name}/...`, `CompositePostExecuteHook`.

**Offline Sync (FMCG):** операционная синхронизация — клиент копит DML-операции (IndexedDB/Dexie), `POST /sync/push` (реплей через DML-конвейер, идемпотентность по `op_id`, `sync.processed_operations`), `POST /sync/pull` (дельта через SOQL). Domain-aware разрешение конфликтов (заказ = заявка, не проводка). Без дублирования бизнес-логики: клиент — рендерер форм + очередь, сервер — источник истины.

---

## 14. Нефункциональные требования

- **Безопасность:** невозможность обхода OLS/FLS/RLS; параметризация всех SQL (нет конкатенации); bcrypt cost=12; хеш-хранение токенов; AES-256-GCM для секретов; SSRF-защита интеграций; rate-limit на auth.
- **Производительность:** table-per-object (нативные JOIN, FK, EXPLAIN); in-memory кэш метаданных; предвычисленные effective-кэши (без полной материализации матрицы доступа, размер < O(users×records)); ProgramCache для CEL; PL/pgSQL для тяжёлых операций (активация территорий).
- **Согласованность кэшей:** outbox + LISTEN/NOTIFY, инкрементальный пересчёт (<1 c типично), fallback-свип 30 c.
- **Наблюдаемость:** структурные логи slog с request_id; usage-log credentials; (план) audit log.
- **Тестируемость:** интерфейсы (MetadataReader, *Service, executor-ы); contract-тесты против OpenAPI; pgTAP для схемы; e2e Playwright; race-детектор в `make test`.
- **Однотенантность** (ADR-0016): одна организация на инстанс, простое развёртывание.
- **Лицензионная чистота:** заголовки в `ee/`-файлах, разделение AGPL/коммерческой.

---

## 15. Дорожная карта (контекст зрелости)

Завершено: Phase 0 (каркас), 1 (метаданные), 2a/2b/2c (identity+OLS/FLS, RLS, группы), 3 (SOQL), 4 (DML), 5 (auth), 6 (шаблоны), 7a/7b (generic CRUD + CEL/валидации/дефолты), 8 (Custom Functions + Expression Builder), 9a/9b/9c (Portal/Navigation/Layout+Form), 10a/10b (Procedure Engine + Named Credentials, Automation Rules), а также инфраструктура модульного монолита (ADR-0030).

В планах: 11 (related lists, reference lookup, поиск/сортировка, recycle bin, CSV-импорт), 12 (нотификации, email, saved views, activity, bulk), 13 (файлы, экспорт, глобальный поиск, audit log, формульные поля), 14 (Scenario Engine + approvals, record types + dynamic forms, расширенные объекты), 15 (SOSL), 16 (CDC/streaming/webhooks), 17 (отчёты/дашборды), N (enterprise: audit trail, SSO, шифрование at-rest и пр.).

---

## Приложение A. Реестр ADR (источник проектных решений)

0001 UUID v4 как PK · 0002 i18n · 0003 структура метаданных объекта · 0004 иерархия тип/подтип поля · 0005 reference-типы · 0006 relationship registry как кэш · 0007 table-per-object · 0008 admin в монорепо web/ · 0009 обзор безопасности · 0010 OLS/FLS grant+deny bitmask · 0011 RLS · 0012 кэширование безопасности (closure + outbox) · 0013 группы (4 типа) · 0014 Open Core · 0015 территории · 0016 однотенантность · 0017 auth · 0018 App Templates · 0019 декларативная бизнес-логика · 0020 расширение DML-конвейера · 0021 contract testing · 0022 Portal bounded context · 0023 терминология действий (Command→Procedure→Scenario) · 0024 Procedure Engine · 0025 Scenario Engine · 0026 Custom Functions · 0027 Layout+Form · 0028 Named Credentials · 0029 версионирование · 0030 модульный монолит · 0031 Automation Rules · 0032 Profile Navigation · 0033 offline sync · 0034 industry modules · 0035 Portal data binding · 0036 Portal action model · 0037 Portal gate graph.

*Полные тексты — в `docs/adr/`. Дорожная карта — `docs/roadmap.md`. Контракт — `api/openapi.yaml`.*
