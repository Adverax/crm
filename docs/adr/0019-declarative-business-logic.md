# ADR-0019: Декларативная бизнес-логика объектов

**Статус:** Принято
**Дата:** 2026-02-14
**Участники:** @roman_myakotin

## Контекст

Платформа metadata-driven (ADR-0003, ADR-0007): объекты определяются через метаданные, хранятся в реальных PG-таблицах `obj_{api_name}`, SOQL/DML обеспечивают чтение/запись с security (OLS/FLS/RLS — ADR-0009..0012).

Текущая метадатная модель покрывает **структуру хранения**: типы полей, обязательность, уникальность, ссылки (ADR-0004, ADR-0005). Однако поведенческая логика — бизнес-правила, вычисляемые поля, динамические дефолты — отсутствует:

| Что есть (metadata) | Чего нет |
|---|---|
| Тип поля (text, number, boolean) | Кросс-полевая валидация (`close_date > created_at`) |
| `is_required` flag | Условная обязательность (`feedback required when status=completed`) |
| `default_value` как статическая строка | Динамические дефолты (`owner_id = current_user.id`) |
| — | Вычисляемые поля для отображения (`full_name = first + last`) |
| — | Единые правила для frontend и backend |

### Ограничения текущего DML Engine

1. **`DefaultValue`** — `*string` в `FieldConfig` JSONB. На уровне DDL используется только для boolean. DML Engine не инжектит значения — только пропускает required-check когда `HasDefault=true`
2. **Pipeline** (parse → validate → compile → execute) не имеет хуков: нет pre-insert, post-update, нет trigger-системы
3. **Валидация** ограничена: required-fields + type-compatibility. Кросс-полевая валидация невозможна

### Потребности платформы

Для metadata-driven CRM недостаточно описывать только схему хранения. Платформа нуждается в **декларативном слое поведенческой логики**, который:

1. Определяет бизнес-правила (валидация) декларативно, без кода per-object
2. Вычисляет производные значения (formula fields) на основе выражений
3. Устанавливает динамические значения по умолчанию
4. Работает одинаково на backend и frontend (единый expression language)
5. Позволяет разным контекстам (формам, API-endpoint'ам) иметь разные наборы правил
6. Гарантирует минимальный уровень data integrity независимо от контекста вызова

### Индустриальный контекст

Все крупные CRM-платформы разделяют поведенческую логику на независимые подсистемы:

| Платформа | Validation | Computed Fields | Automation | Queries |
|-----------|-----------|-----------------|------------|---------|
| **Salesforce** | Validation Rules | Formula Fields | Flow / Apex Triggers | SOQL |
| **Dynamics 365** | Business Rules | Calculated Fields | Power Automate | FetchXML |
| **HubSpot** | Property Rules | Calculated Properties | Workflows | — |
| **Zoho CRM** | Validation Rules | Formula Fields | Workflows | — |

Ни одна не объединяет всё в монолитный объект. Каждая подсистема имеет свой жизненный цикл, хранение и модель расширения.

## Рассмотренные варианты

### Вариант A — Монолитный декларативный объект

Один YAML/JSON-документ на объект, описывающий всё: запросы, вычисляемые поля, валидацию, дефолты, мутации, автоматизацию.

**Плюсы:**
- Единая абстракция — вся логика объекта в одном месте
- Максимальная декларативность

**Минусы:**
- God object (7+ ответственностей): загрузка данных, трансформация, валидация, персистенция, автоматизация
- Смешивает per-object логику (validation — должна работать всегда) и per-view логику (queries — зависят от UI-контекста)
- Блокирует Phase 7a: требует CEL engine, executor, storage, YAML parser, dependency resolver — месяцы работы
- Не соответствует индустриальному паттерну

### Вариант B — Декомпозиция на подсистемы с трёхуровневым каскадом (выбран)

Разбить поведенческую логику на независимые подсистемы (Validation Rules, Default Expressions, Formula Fields, Object View, Automation Rules). Связать уровни (Metadata → Object View → Layout) каскадной моделью наследования.

**Плюсы:**
- Каждая подсистема независимо полезна и тестируема
- Инкрементальная реализация — не блокирует текущую фазу
- Переиспользование (validation rules работают при любом способе записи: API, import, integration)
- Каскад с наследованием — DRY + гибкость
- Соответствует индустриальному паттерну (Salesforce, Dynamics)

**Минусы:**
- Нет единого документа для всей логики объекта
- Больше отдельных ADR (каждая подсистема = отдельное решение)
- Композиционный слой (Object View) откладывается

### Вариант C — Отложить полностью

Построить Phase 7a без абстракций, хардкодить логику per-object.

**Плюсы:**
- Быстрая доставка Phase 7a

**Минусы:**
- Технический долг при росте количества объектов
- Дублирование validation на frontend/backend
- Рефакторинг позже будет дороже
- Противоречит metadata-driven архитектуре платформы

### Вариант D — Минимальные расширения метаданных

Расширить `FieldConfig` для static defaults и simple validation без expression engine.

**Плюсы:**
- Быстрая доставка, минимальные изменения

**Минусы:**
- Static defaults недостаточны для динамических значений (`owner_id = current_user.id`)
- Кросс-полевая валидация невозможна без expression engine
- Дублирование логики на frontend

## Решение

**Выбран вариант B: Декомпозиция на независимые подсистемы с трёхуровневым каскадом.**

### Трёхуровневый каскад

Правила и настройки определяются на трёх уровнях. Каждый последующий уровень **наследует** правила предыдущего:

```
Metadata (base)
   ↓ наследует
Object View (business context)
   ↓ наследует
Layout (presentation)
```

#### Семантика каскада по типам

| Аспект | Metadata → Object View | Object View → Layout | Механизм |
|--------|------------------------|----------------------|----------|
| **Validation** | Additive (AND) | Additive (AND) | Только добавление новых правил |
| **Defaults** | Replace | Replace | Последний уровень побеждает |
| **Formula Fields** | Inherit (read-only) | Inherit (read-only) | Не перекрываются |
| **Field visibility** | N/A | Override | Layout скрывает/показывает |

#### Validation: аддитивная модель (только ужесточение)

Validation rules на каждом уровне каскада могут только **добавляться**, но никогда не удаляться и не заменяться. Итоговый набор — конъюнкция (AND) всех правил:

```
effective_validation = metadata_rules AND object_view_rules AND layout_rules
```

Это математически гарантирует ужесточение: добавление любого нового условия через AND сужает множество допустимых значений.

**Почему не разрешаем ослабление?** Программная верификация того, что одно CEL-выражение «строже» другого, сводится к задаче theorem proving: ∀ input: A(input)=true → B(input)=true. Для произвольных выражений это неразрешимо. Даже для ограниченного подмножества потребовался бы SMT-солвер — overkill для CRM. Аддитивная модель устраняет проблему: верификация не нужна, AND гарантирует ужесточение автоматически.

**Пример каскада:**

```yaml
# Metadata (universal invariant):
- expr: 'discount <= 50'          # data integrity

# Object View "partner_portal" (adds business rule):
- expr: 'discount <= 20'          # business context

# Layout "mobile_form" (adds UI rule):
- expr: 'has(discount)'           # field required on this form

# Effective: discount <= 50 AND discount <= 20 AND has(discount)
# = discount обязателен И не более 20%
```

**Следствие для проектирования правил:** если валидация может отличаться в разных контекстах — она принадлежит **Object View**, а не metadata. В metadata — только универсальные инварианты, нарушение которых = повреждение данных.

| Где определять | Критерий | Примеры |
|---|---|---|
| **Metadata** | Универсальный инвариант, нарушение = повреждение данных | `amount >= 0`, FK integrity, type safety |
| **Object View** | Бизнес-контекст, может различаться между view | Формат телефона, условная обязательность, `discount <= N` |
| **Layout** | UI-специфичное ужесточение | Обязательность поля на конкретной форме |

#### Defaults: замена (последний уровень побеждает)

Default — «какое значение подставить, если не указано». Замена безопасна: итоговая валидация всё равно проверит корректность значения.

```yaml
# Metadata:       status default = "new"
# Object View:    status default = "draft"      ← заменяет
# Layout:         (не переопределяет)
# Effective:      "draft"
```

### Подсистемы

Поведенческая логика разбивается на **6 независимых подсистем**:

```
┌─────────────────────────────────────────────────────────────────┐
│  GLOBAL (cross-cutting)                                          │
│  Доступны на всех уровнях каскада, на backend и frontend         │
│                                                                  │
│  ┌───────────────────────────────────────────────────────────┐   │
│  │  Custom Functions (ADR-0026)                               │   │
│  │  Именованные чистые CEL-выражения: fn.discount(tier, amt) │   │
│  │  Dual-stack: cel-go + cel-js. Нет side effects.           │   │
│  │  Вызываются из любого CEL-контекста ниже.                 │   │
│  └───────────────────────────────────────────────────────────┘   │
└──────────────────────────────┬──────────────────────────────────┘
                               │ доступны как fn.*
                               ▼
┌─────────────────────────────────────────────────────────┐
│              PER-OBJECT (metadata level)                 │
│   Базовые правила, наследуемые всеми Object View/Layout│
│                                                         │
│  ┌───────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  Validation   │  │   Default    │  │   Formula    │  │
│  │  Rules        │  │   Expressions│  │   Fields     │  │
│  │               │  │              │  │              │  │
│  │  CEL expr     │  │  CEL expr    │  │  CEL expr    │  │
│  │  per-object   │  │  per-field   │  │  per-field   │  │
│  │  in metadata  │  │  in metadata │  │  in metadata │  │
│  └───────────────┘  └──────────────┘  └──────────────┘  │
└──────────────────────────┬──────────────────────────────┘
                           │ наследует (additive validation,
                           │            replace defaults)
                           ▼
┌─────────────────────────────────────────────────────────┐
│           PER-VIEW (Object View level)                   │
│   Бизнес-контекст: конкретный UI-экран или API-endpoint │
│                                                         │
│  ┌───────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  + Validation │  │  + Default   │  │  Queries     │  │
│  │  (additive)   │  │  (replace)   │  │  (SOQL)      │  │
│  └───────────────┘  └──────────────┘  └──────────────┘  │
│  ┌───────────────┐  ┌──────────────┐                    │
│  │  Virtual      │  │  Mutations   │                    │
│  │  Fields (CEL) │  │  (DML)       │                    │
│  └───────────────┘  └──────────────┘                    │
└──────────────────────────┬──────────────────────────────┘
                           │ наследует (additive validation,
                           │            replace defaults)
                           ▼
┌─────────────────────────────────────────────────────────┐
│           PER-LAYOUT (Layout level)                      │
│   Презентация: визуальное расположение и UI-правила     │
│                                                         │
│  ┌───────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  + Validation │  │  + Default   │  │  Field       │  │
│  │  (additive)   │  │  (replace)   │  │  visibility  │  │
│  └───────────────┘  └──────────────┘  │  & ordering  │  │
│                                       └──────────────┘  │
└─────────────────────────────────────────────────────────┘
```

#### 1. Validation Rules (per-object, metadata level)

CEL-выражения, проверяющие данные перед записью. Хранятся в metadata schema, применяются DML Engine при любой операции.

```
metadata.validation_rules (object_id, expr, message, code, severity, when_expr, sort_order)
```

- Интеграция: DML Engine (validate step)
- Оценка: CEL runtime (cel-go backend, cel-js frontend)
- Переменные: `record` (текущие значения), `old` (предыдущие при update), `user`, `now`

#### 2. Default Expressions (per-field, metadata level)

CEL-выражения для вычисления значений по умолчанию. Расширение существующего `FieldConfig.DefaultValue` до динамических выражений.

```
field_definitions.config.default_expr — CEL expression (nullable)
field_definitions.config.default_on   — "create" | "update" | "create,update"
```

- Интеграция: DML Engine (pre-validate step, инжект missing fields)
- Переменные: `record`, `user`, `now`
- Статические дефолты (`"draft"`, `true`) остаются как `default_value`

#### 3. Formula Fields (per-field, metadata level, read-only)

CEL-выражения для вычисляемых полей, не хранящихся в БД. Вычисляются при чтении. Аналог Salesforce Formula Fields.

```
field_definitions.config.formula_expr — CEL expression (nullable)
field_definitions.field_type = "formula"
field_definitions.field_subtype = "string" | "number" | "boolean" | "datetime"
```

- Интеграция: SOQL Executor (post-fetch computation)
- Переменные: `record` (все поля текущей записи)
- Frontend: вычисляются локально по тому же CEL

#### 4. Object View (per-view, будущее)

Композиционный слой для конкретного UI-экрана или API-endpoint. Наследует validation и defaults из metadata (additive / replace). Содержит собственные компоненты:

- Queries — именованные SOQL-запросы с cross-references
- Virtual Fields — view-specific вычисляемые поля (CEL)
- Mutations — оркестрация DML-операций (foreach, sync)
- Validation overrides — дополнительные правила (additive only)
- Default overrides — альтернативные дефолты (replace)

**Архитектурная роль: адаптер bounded context (DDD).** Один и тот же объект (например `Order`) обслуживает разные бизнес-роли: менеджер по продажам, кладовщик, руководитель. Каждая роль работает в своём bounded context — со своим набором полей, действий, related lists и sidebar. Object View, привязанный к профилю (`profile_id`), адаптирует единые данные к контексту конкретной роли без дублирования кода. OLS/FLS/RLS контролируют *доступ к данным*, Object View контролирует *представление данных*. При этом Object View только сужает видимость (FLS intersection), но не расширяет доступ.

Детализация: [ADR-0022](0022-object-view-bounded-context.md) — структура config, resolution logic, sidebar/dashboard per profile, примеры role-based UI.

#### 5. Automation Rules (per-object, будущее)

Реактивная логика: «когда произошло X, выполни Y». Аналог Salesforce Flow / Process Builder.

- Trigger condition — CEL-выражение (`new.Stage == "Closed Won" && old.Stage != "Closed Won"`)
- Action — ссылка на Procedure (синхронная, ADR-0024) или Scenario (асинхронный, ADR-0025)
- Терминология: ADR-0023

Отдельный ADR при необходимости.

#### 6. Custom Functions (global, cross-cutting, ADR-0026)

Именованные чистые CEL-выражения с типизированными параметрами. Устраняют дублирование CEL-логики между подсистемами.

```
metadata.functions (name, params JSONB, return_type, body TEXT)
```

- **Чистые**: нет side effects — только вычисления, без CRUD/IO
- **Глобальные**: не привязаны к объекту, вызываются из любого CEL-контекста через `fn.*` namespace
- **Dual-stack**: загружаются в cel-go (backend) и cel-js (frontend) — одинаковое поведение
- **Reusable**: одно определение → использование в validation rules, defaults, formulas, visibility, procedure/scenario input
- **Composable**: `fn.total(fn.discount(tier, amount), tax_rate)` — функции вызывают друг друга (max 3 уровня)

Отличие от Formula Fields: Formula Field привязан к объекту и полю; Function глобальна и принимает произвольные параметры.

Детализация: [ADR-0026](0026-custom-functions.md).

### Execution context: DML Engine

DML Engine при выполнении операции получает **effective ruleset** — результат каскадного слияния правил в зависимости от контекста вызова:

```
Вызов без Object View (raw DML, import, integration):
  effective_validation = metadata_rules
  effective_defaults   = metadata_defaults

Вызов через Object View (UI form, specific API endpoint):
  effective_validation = metadata_rules AND object_view_rules
  effective_defaults   = merge(metadata_defaults, object_view_defaults)

Вызов через Layout:
  effective_validation = metadata_rules AND object_view_rules AND layout_rules
  effective_defaults   = merge(metadata_defaults, object_view_defaults, layout_defaults)
```

«Голый» DML (без Object View) всегда применяет metadata-level правила — минимальный гарантированный уровень защиты data integrity.

### CEL как expression language

Для всех выражений (validation, defaults, formulas, virtual fields) используется CEL — Common Expression Language (Google).

| Критерий | CEL | Альтернативы |
|----------|-----|-------------|
| Go runtime | `cel-go` (official) | Expr, Govaluate |
| JS runtime | `cel-js` | Только CEL имеет оба |
| Безопасность | Sandboxed, no side effects | Expr — аналогично, Govaluate — нет |
| Типизация | Static type checking | Expr — runtime only |
| Стандарт | Google, K8s, Firebase | Нет |
| Синтаксис | C-like, понятный | Expr — похожий |

CEL-интеграция вводится при реализации Validation Rules (Phase 7b).

### Дорожная карта

```
Phase 7a                  Phase 7b                Phase 10                Phase 9a/9b
──────────────────    ──────────────────    ──────────────────────    ──────────────────
Generic CRUD          CEL engine (cel-go)   Custom Functions          Object Views
+ Metadata-driven UI  + Validation Rules    + fn.* namespace          + Query composition
+ Static defaults     + Dynamic defaults    + Function Constructor    + Actions
+ System fields       + DML pipeline ext.   + Expression Builder      + Automation Rules
                      + Frontend CEL eval   + Formula Fields          + Layout cascade
                                            + SOQL integration
```

**Phase 7a.** Generic metadata-driven REST endpoints: один набор handlers обслуживает все объекты через SOQL (чтение) и DML (запись). Frontend рендерит формы по метаданным. Валидация: required + type constraints (уже в DML Engine). Static defaults: инжект `FieldConfig.default_value` для отсутствующих полей. Системные поля (`owner_id`, `created_by_id`, `created_at`, `updated_at`). Без CEL.

**Phase 7b — CEL + Validation Rules + Dynamic Defaults.** Интеграция `cel-go`. Таблица `metadata.validation_rules`. Расширение `FieldConfig` для `default_expr`. Интеграция в DML pipeline (Stage 3 dynamic + Stage 4b). Frontend-библиотека для CEL-eval (`cel-js`).

**Phase 10 — Custom Functions + Formula Fields.** Custom Functions (ADR-0026): глобальные именованные CEL-выражения с `fn.*` namespace, dual-stack (cel-go + cel-js), Function Constructor + интеграция в Expression Builder. Formula Fields: `field_type = "formula"`, CEL-выражение в config, SOQL executor вычисляет после fetch, frontend вычисляет локально. Formula Fields могут вызывать Custom Functions.

**Phase 9a/9b — Object Views + Automation + Layout cascade.** Полная композиция с трёхуровневым каскадом. Каскадный мержинг (metadata + Object View + Layout). ADR-0022 (Object View), ADR-0023 (терминология Action), ADR-0024 (Procedure Engine), ADR-0025 (Scenario Engine).

## Последствия

### Позитивные

- Phase 7a не блокируется — generic CRUD endpoints строятся на существующей инфраструктуре (SOQL + DML + MetadataCache)
- Каждая подсистема независимо полезна и тестируема
- Трёхуровневый каскад (Metadata → Object View → Layout) обеспечивает DRY + гибкость
- Аддитивная модель validation гарантирует ужесточение без программной верификации выражений
- Validation Rules работают при любом способе записи (API, import, integration)
- «Голый» DML без Object View всё равно защищён metadata-level правилами
- Инкрементальная доставка ценности
- Соответствие индустриальному паттерну (Salesforce, Dynamics, Zoho)

### Негативные

- Нет единого документа для всей логики объекта (сознательный trade-off в пользу SoC)
- Больше ADR (каждая подсистема = отдельное архитектурное решение)
- Композиционный слой (Object View + Layout cascade) откладывается до Phase N+2
- Аддитивная модель validation не позволяет ослабить правило — если правило может различаться в контекстах, его нужно изначально размещать в Object View, а не в metadata

### Связанные ADR

- ADR-0003 — Object metadata structure (расширяется validation rules и default expressions)
- ADR-0004 — Field type/subtype hierarchy (расширяется formula type)
- ADR-0007 — Table-per-object storage (generic CRUD работает с `obj_{api_name}` таблицами)
- ADR-0009..0012 — Security layers (validation rules дополняют, но не заменяют OLS/FLS/RLS)
- ADR-0018 — App Templates (создают schema; подсистемы из этого ADR определяют поведение)
- ADR-0020 — DML Pipeline Extension (typed stages — точки интеграции подсистем в DML Engine)
- ADR-0022 — Object View как адаптер bounded context (детализация подсистемы 4: role-based UI, config schema, resolution logic)
- ADR-0023 — Action terminology: единая иерархия Action → Command → Procedure → Scenario + Function (ортогональна)
- ADR-0024 — Procedure Engine: JSON DSL + Constructor UI для синхронной бизнес-логики (Mutations → Action type: procedure)
- ADR-0025 — Scenario Engine: JSON DSL + Constructor UI для асинхронных долгоживущих процессов
- ADR-0026 — Custom Functions (детализация подсистемы 6: fn.* namespace, dual-stack, ограничения, Constructor UI)
