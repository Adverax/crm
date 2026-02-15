# ADR-0022: Object View — адаптер bounded context через role-based UI

**Статус:** Принято
**Дата:** 2026-02-15
**Участники:** @roman_myakotin

## Контекст

### Проблема: единые данные — разные контексты

Платформа metadata-driven: один объект (например `Order`) обслуживает разные бизнес-роли. Каждая роль работает в своём **bounded context** (термин DDD) — со своей ментальной моделью, набором полей, действий и связанных объектов:

| Аспект | Sales Rep | Кладовщик | Руководитель |
|--------|-----------|-----------|--------------|
| Фокус | Клиент, сделка, продажа | Отгрузка, склад, перемещение | Выручка, маржа, конверсия |
| Поля | client_name, products, discount, total_amount | warehouse, shipping_status, tracking, packages | margin, cost_price, revenue, conversion |
| Действия | send_proposal, create_task | mark_shipped, print_label | export_report, reassign |
| Related | Activities, Files | InventoryMovements | Reports, Subordinate deals |
| Sidebar | Accounts, Contacts, Deals, Tasks | Orders, Warehouses, Shipments | Dashboard, Reports, Users |

Сегодня OLS/FLS/RLS (ADR-0009..0012) контролируют **доступ к данным**, но не **представление**:

| Слой | Что решает | Чего не хватает |
|------|-----------|-----------------|
| OLS | Какие объекты видит профиль | Какие секции/действия показать |
| FLS | Какие поля доступны | В каком порядке, как сгруппировать |
| RLS | Какие записи видимы | Какие related lists показать |

Результат: все пользователи видят одинаковую форму со всеми доступными (по FLS) полями в порядке `sort_order` из metadata. Это создаёт:

1. **Когнитивную перегрузку** — 30+ полей на форме, когда роли нужно 8-10
2. **Невозможность role-specific workflow** — кладовщик не может иметь кнопку "Отгрузить" без кастомного кода
3. **Монолитный UI** — один layout для всех, вместо bounded context per role
4. **Барьер для CRM+ERP сценариев** — невозможно дать складской роли "ERP-подобный" интерфейс на тех же объектах

### Связь с ADR-0019

ADR-0019 определил Object View как 4-ю подсистему декларативной бизнес-логики с трёхуровневым каскадом `Metadata → Object View → Layout`. Однако ADR-0019 фокусировался на семантике каскада (additive validation, replace defaults) и не детализировал:

- Привязку Object View к профилю/роли (bounded context adapter)
- Структуру секций, действий, related lists
- Sidebar per profile
- Dashboard per role
- Механизм fallback при отсутствии view

Данный ADR детализирует Object View как **полноценный адаптер bounded context**, превращающий единые данные в role-specific UI без дублирования кода.

### Индустриальный контекст

| Платформа | Механизм | Привязка |
|-----------|----------|----------|
| **Salesforce** | Page Layouts + Record Types + App | Profile + Record Type |
| **Dynamics 365** | Forms + Views + App Modules | Security Role + App |
| **ServiceNow** | UI Policies + UI Actions + Views | Role |
| **HubSpot** | Record Customization + Views | Team + Pipeline |

Все enterprise-платформы решают задачу bounded context через привязку представления к роли/профилю. Это не опциональная фича — это фундамент масштабируемого UI.

## Рассмотренные варианты

### Вариант A — Хардкод views per role в frontend

Отдельные Vue-компоненты для каждой роли: `OrderSalesView.vue`, `OrderWarehouseView.vue`.

**Плюсы:**
- Быстрая реализация для 2-3 ролей
- Полная свобода в вёрстке

**Минусы:**
- Не масштабируется: N объектов × M ролей = N×M компонентов
- Противоречит metadata-driven архитектуре
- Custom objects не получают role-based views
- Код дублируется между views

### Вариант B — Object View как metadata-driven configuration (выбран)

Object View — JSON-конфигурация в metadata, привязанная к `(object, profile)`. Frontend рендерит UI по конфигурации. Fallback: нет Object View → все FLS-доступные поля в порядке `sort_order`.

**Плюсы:**
- Масштабируется на любое количество объектов и ролей
- Администратор настраивает через UI без кода
- Custom objects получают role-based views автоматически
- Единый рендерер — один компонент обрабатывает любой Object View
- Вписывается в трёхуровневый каскад ADR-0019
- Наследует security: Object View не может показать поле, запрещённое FLS

**Минусы:**
- Требует нового metadata storage + Admin UI
- Декларативная конфигурация ограничена — сложные кастомизации невозможны без кода
- Дополнительный запрос к metadata при загрузке формы

### Вариант C — Layout Builder (drag-and-drop)

Визуальный конструктор форм с drag-and-drop (как Salesforce Lightning App Builder).

**Плюсы:**
- Максимальная гибкость для администратора
- Визуальная настройка без знания JSON/конфигов

**Минусы:**
- Огромная сложность реализации (6+ месяцев)
- Не нужен на текущем этапе (80/20: JSON config покрывает 90% кейсов)
- Может быть добавлен поверх варианта B позже (Builder = visual editor для того же JSON)

### Вариант D — Profile-specific CSS/visibility rules

Скрывать/показывать элементы через CSS-классы или visibility rules на frontend.

**Плюсы:**
- Минимальные изменения backend
- Быстрая реализация

**Минусы:**
- Security through obscurity — данные всё равно приходят в payload
- Нет группировки полей в секции
- Нет role-specific действий и related lists
- Не масштабируется

## Решение

**Выбран вариант B: Object View как metadata-driven configuration с привязкой к профилю.**

### Концептуальная модель

```
┌─────────────────────────────────────────────────────────────┐
│                     Пользователь логинится                    │
│                           │                                  │
│                     JWT → Profile + Role                      │
│                           │                                  │
│              ┌────────────┼────────────────┐                 │
│              ▼            ▼                ▼                  │
│         ┌─────────┐  ┌─────────┐   ┌──────────────┐         │
│         │   OLS   │  │ Object  │   │   Dashboard  │         │
│         │ фильтр  │  │  View   │   │  per profile │         │
│         │ sidebar │  │ resolve │   │              │         │
│         └────┬────┘  └────┬────┘   └──────┬───────┘         │
│              │            │               │                  │
│              ▼            ▼               ▼                  │
│         Sidebar      Record Form      Home Page              │
│        (только       (секции,         (виджеты               │
│         доступные    порядок,          роли)                  │
│         объекты)     действия)                                │
└─────────────────────────────────────────────────────────────┘
```

### Структура Object View

Object View — запись в metadata, описывающая как отображать объект для конкретного профиля:

```
metadata.object_views
├── id               UUID PK
├── object_id        FK → object_definitions.id
├── profile_id       FK → iam.profiles.id (nullable — default view)
├── api_name         VARCHAR UNIQUE (e.g. "order_sales", "order_warehouse")
├── label            VARCHAR
├── description      TEXT
├── is_default       BOOLEAN (fallback view для профиля без specific view)
├── config           JSONB (см. ниже)
├── created_at       TIMESTAMPTZ
└── updated_at       TIMESTAMPTZ

UNIQUE(object_id, profile_id)  — один view на пару (объект, профиль)
```

### Config JSON Schema

```jsonc
{
  // Секции формы — группировка полей
  "sections": [
    {
      "key": "client_info",
      "label": "Информация о клиенте",
      "columns": 2,                    // 1 или 2 колонки
      "collapsed": false,              // свёрнута по умолчанию
      "fields": [
        "client_name",                 // api_name поля
        "contact_phone",
        "deal"                         // reference field
      ]
    },
    {
      "key": "products",
      "label": "Товары",
      "columns": 1,
      "fields": ["products", "total_amount", "discount"]
    }
  ],

  // Highlight panel — ключевые поля вверху карточки (Compact Layout)
  "highlight_fields": ["order_number", "status", "total_amount"],

  // Действия (кнопки) на карточке записи
  "actions": [
    {
      "key": "send_proposal",
      "label": "Отправить предложение",
      "type": "primary",                // primary | secondary | danger
      "icon": "mail",
      "visibility_expr": "record.status == 'draft'"  // CEL — когда показывать
    },
    {
      "key": "mark_shipped",
      "label": "Отгрузить",
      "type": "primary",
      "icon": "truck",
      "visibility_expr": "record.status == 'confirmed'"
    }
  ],

  // Related Lists — дочерние объекты внизу карточки
  "related_lists": [
    {
      "object": "Activity",
      "label": "Активности",
      "fields": ["subject", "type", "due_date", "status"],
      "filter": "WhatId = :recordId",
      "sort": "due_date DESC",
      "limit": 10
    }
  ],

  // List View — какие колонки показывать в таблице списка
  "list_fields": ["order_number", "client_name", "status", "total_amount", "created_at"],

  // Сортировка по умолчанию в списке
  "list_default_sort": "created_at DESC",

  // Фильтры по умолчанию в списке
  "list_default_filter": "owner_id = :currentUserId"
}
```

### Правила разрешения (Resolution Rules)

При открытии записи объекта `X` пользователем с профилем `P`:

```
1. Ищем object_views WHERE object_id = X AND profile_id = P
   → Найден? Используем.

2. Ищем object_views WHERE object_id = X AND is_default = true
   → Найден? Используем.

3. Fallback: авто-генерация из metadata
   → sections: одна секция "Детали" со всеми FLS-доступными полями
   → highlight_fields: первые 3 поля
   → actions: стандартные (Save, Delete)
   → related_lists: все дочерние объекты (composition/association)
   → list_fields: первые 5 полей
```

Fallback гарантирует, что **система работает без единого Object View** — текущее поведение сохраняется. Object View — необязательное улучшение.

### Взаимодействие с Security

Object View **не расширяет** доступ — он только **сужает представление**:

```
Видимые поля = Object View fields ∩ FLS-доступные поля
```

Если Object View включает поле, запрещённое FLS — поле не отображается (FLS побеждает).
Если FLS разрешает поле, но Object View его не включает — поле не отображается (View сужает).

```
┌──────────────────────────────────────────┐
│             FLS-доступные поля            │
│  ┌────────────────────────────────────┐  │
│  │     Object View fields            │  │
│  │  ┌──────────────────────────┐     │  │
│  │  │  Отображаемые поля      │     │  │
│  │  │  (пересечение)          │     │  │
│  │  └──────────────────────────┘     │  │
│  └────────────────────────────────────┘  │
└──────────────────────────────────────────┘
```

Действия (actions) проходят аналогичную проверку:
- `send_proposal` требует OLS Update на Order → если нет — кнопка скрыта
- `delete` требует OLS Delete → если нет — кнопка скрыта

### Интеграция с каскадом ADR-0019

Object View занимает второй уровень каскада:

```
Metadata (base)
   ↓ additive validation, inherit defaults
Object View (bounded context)        ← ЭТОТ ADR
   ↓ additive validation, replace defaults, override visibility
Layout (presentation, future)
```

| Аспект | Metadata → Object View | Механизм |
|--------|------------------------|----------|
| **Validation Rules** | Additive (AND) | OV добавляет правила, не удаляет metadata-уровень |
| **Default Expressions** | Replace | OV может переопределить default для поля |
| **Field visibility** | Restrict | OV показывает subset полей из metadata |
| **Actions** | Define | OV определяет доступные действия |
| **Related Lists** | Define | OV определяет дочерние объекты для отображения |

### Sidebar per Profile

OLS уже фильтрует объекты по профилю. Object View дополняет:

```
metadata.profile_navigation
├── id               UUID PK
├── profile_id       FK → iam.profiles.id
├── config           JSONB
├── created_at       TIMESTAMPTZ
└── updated_at       TIMESTAMPTZ

config = {
  "groups": [
    {
      "label": "Продажи",
      "items": ["Account", "Contact", "Opportunity"]  // api_name объектов
    },
    {
      "label": "Документы",
      "items": ["Order", "Contract", "Quote"]
    }
  ]
}
```

Fallback: нет записи в `profile_navigation` → sidebar из OLS-доступных объектов в алфавитном порядке (текущее поведение).

### Dashboard per Profile

Home page адаптируется к профилю:

```
metadata.profile_dashboards
├── id               UUID PK
├── profile_id       FK → iam.profiles.id
├── config           JSONB
├── created_at       TIMESTAMPTZ
└── updated_at       TIMESTAMPTZ

config = {
  "widgets": [
    {
      "type": "list",                      // list | chart | metric | calendar
      "label": "Мои открытые задачи",
      "query": "SELECT Id, Subject, DueDate FROM Task WHERE OwnerId = :currentUserId AND Status != 'Completed' ORDER BY DueDate LIMIT 10",
      "size": "half"                        // full | half | third
    },
    {
      "type": "metric",
      "label": "Сделки в этом месяце",
      "query": "SELECT COUNT(Id) FROM Opportunity WHERE CreatedDate = THIS_MONTH",
      "size": "third"
    }
  ]
}
```

Fallback: нет dashboard config → стандартный dashboard с recent items и tasks.

### Пример: один объект — три bounded context

**Order для Sales Rep (Profile: "Sales"):**
```jsonc
{
  "sections": [
    { "key": "client", "label": "Клиент", "fields": ["client_name", "contact_phone", "deal"] },
    { "key": "products", "label": "Товары", "fields": ["products", "total_amount", "discount"] },
    { "key": "delivery", "label": "Доставка", "fields": ["shipping_status", "delivery_date"] }
  ],
  "highlight_fields": ["order_number", "client_name", "total_amount"],
  "actions": [
    { "key": "send_proposal", "label": "Отправить предложение", "type": "primary" }
  ],
  "related_lists": [
    { "object": "Activity", "label": "Активности" }
  ],
  "list_fields": ["order_number", "client_name", "status", "total_amount"]
}
```

**Order для кладовщика (Profile: "Warehouse"):**
```jsonc
{
  "sections": [
    { "key": "order", "label": "Заказ", "fields": ["order_number", "client_name"] },
    { "key": "shipping", "label": "Отгрузка", "fields": ["warehouse", "products", "shipping_status", "tracking"] },
    { "key": "dimensions", "label": "Вес и габариты", "fields": ["total_weight", "packages_count"] }
  ],
  "highlight_fields": ["order_number", "shipping_status", "warehouse"],
  "actions": [
    { "key": "mark_shipped", "label": "Отгрузить", "type": "primary", "visibility_expr": "record.status == 'confirmed'" },
    { "key": "print_label", "label": "Печать этикетки", "type": "secondary" }
  ],
  "related_lists": [
    { "object": "InventoryMovement", "label": "Движения по складу" }
  ],
  "list_fields": ["order_number", "shipping_status", "warehouse", "created_at"]
}
```

**Order для руководителя (Profile: "Manager"):**
```jsonc
{
  "sections": [
    { "key": "overview", "label": "Обзор", "fields": ["order_number", "client_name", "status", "total_amount"] },
    { "key": "financials", "label": "Финансы", "fields": ["cost_price", "margin", "revenue", "discount"] },
    { "key": "execution", "label": "Исполнение", "fields": ["warehouse", "shipping_status", "delivery_date"] }
  ],
  "highlight_fields": ["order_number", "total_amount", "margin"],
  "actions": [
    { "key": "reassign", "label": "Переназначить", "type": "secondary" },
    { "key": "export", "label": "Экспорт", "type": "secondary" }
  ],
  "related_lists": [
    { "object": "Activity", "label": "Активности" },
    { "object": "AuditLog", "label": "История изменений" }
  ],
  "list_fields": ["order_number", "client_name", "total_amount", "margin", "status"]
}
```

Три профиля, один URL `/app/Order/123` — три разных интерфейса. Без единой строчки хардкода.

### API

```
GET  /api/v1/describe/:objectName          — включает resolved Object View для текущего профиля
GET  /api/v1/admin/object-views            — список всех Object View (admin)
POST /api/v1/admin/object-views            — создать Object View
GET  /api/v1/admin/object-views/:id        — получить Object View
PUT  /api/v1/admin/object-views/:id        — обновить Object View
DELETE /api/v1/admin/object-views/:id      — удалить Object View
GET  /api/v1/admin/profile-navigation/:id  — навигация профиля
PUT  /api/v1/admin/profile-navigation/:id  — обновить навигацию
GET  /api/v1/admin/profile-dashboards/:id  — dashboard профиля
PUT  /api/v1/admin/profile-dashboards/:id  — обновить dashboard
```

Describe API расширяется: если для текущего профиля есть Object View — response включает `view` с секциями, действиями, related lists. Frontend использует `view` для рендеринга вместо плоского списка полей.

### Хранение

Три таблицы в `metadata` schema:

- `metadata.object_views` — конфигурация формы/списка per (object, profile)
- `metadata.profile_navigation` — sidebar per profile
- `metadata.profile_dashboards` — home page per profile

Все конфигурации в JSONB — гибкость без миграций при расширении схемы.

### Дорожная карта реализации

```
Phase 9a: Object View Core                    Phase 9b: Navigation + Dashboard
────────────────────────────                   ──────────────────────────────────
- metadata.object_views table                  - metadata.profile_navigation table
- Admin CRUD API + UI                          - metadata.profile_dashboards table
- Describe API extension                       - Admin UI для navigation/dashboard
- Frontend: render by Object View              - Sidebar per profile
- Fallback logic                               - Home dashboard per profile
- FLS intersection                             - Widget types: list, metric
- Actions with visibility_expr                 - Chart widgets (Phase 15 dependency)
```

## Последствия

### Позитивные

- **Bounded context без дублирования** — один объект, N представлений, zero code per view
- **Graceful degradation** — система работает без Object Views (fallback = текущее поведение)
- **Security-first** — Object View сужает, но не расширяет доступ (FLS intersection)
- **Администратор настраивает, не разработчик** — Admin CRUD UI для Object Views
- **CRM+ERP без ERP** — складская роль получает "ERP-подобный" интерфейс через Object View
- **Вписывается в каскад ADR-0019** — validation additive, defaults replace, visibility restrict
- **Расширяемость** — Layout Builder (drag-and-drop) добавляется поверх как visual editor
- **App Templates** могут включать Object Views per profile — out-of-the-box role-specific UI

### Негативные

- Дополнительный metadata-запрос при загрузке записи (кэшируется на frontend)
- Complexity: конфигурация Object View может быть нетривиальной для неопытного админа
- Действия (actions) пока только декларативные — реальная логика требует Automation Rules (Phase 13)
- Dashboard widgets с SOQL — потенциальный performance concern при сложных запросах (решается через SOQL query limits)

### Связанные ADR

- **ADR-0009..0012** — Security layers (OLS/FLS/RLS): Object View строится поверх, не обходит
- **ADR-0019** — Declarative business logic: Object View = второй уровень каскада
- **ADR-0020** — DML Pipeline: Object View может добавлять validation rules (additive) и override defaults (replace)
- **ADR-0010** — Permission model: Profile = ключ привязки Object View
- **ADR-0018** — App Templates: могут включать Object View definitions
- **ADR-0027** — Layout + Form: Layout определяет presentation (HOW) поверх Object View (WHAT). Form = computed merge OV + Layout для фронтенда
