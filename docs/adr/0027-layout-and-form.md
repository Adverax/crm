# ADR-0027: Layout + Form — presentation layer для Object View

**Статус:** Принято
**Дата:** 2026-02-15
**Участники:** @roman_myakotin

## Контекст

### Проблема: Object View определяет ЧТО, но не КАК

Object View (ADR-0022) решает задачу bounded context: один объект — разные представления для разных профилей. OV определяет **какие** секции, поля, действия и related lists видит каждая роль.

Но OV не отвечает на вопросы презентации:

| Вопрос | OV отвечает? |
|--------|-------------|
| Какие поля видит Sales? | Да |
| В сколько колонок рендерить секцию? | Нет (только `columns: 2` — минимум) |
| Какой col_span у конкретного поля? | Нет |
| Как выглядит поле status — text или badge? | Нет |
| Обязательно ли поле discount при amount > 10000? | Нет |
| Как выглядит list view — ширина колонок, сортировка? | Нет |
| Как форма адаптируется под мобильное устройство? | Нет |

### Проблема: God Object при расширении OV

Если добавить все layout-атрибуты в OV config, Object View станет God Object с двумя несовместимыми ответственностями:

- **Bounded context** (ЧТО) — бизнес-решение, какие поля/действия доступны роли
- **Presentation** (КАК) — визуальное решение, как рендерить на разных устройствах

Эти ответственности изменяются независимо: bounded context меняется при изменении бизнес-процесса, presentation — при изменении UI/UX требований. Разные люди могут отвечать за каждую.

### Проблема: один OV — несколько устройств

Desktop, tablet и mobile требуют **структурно разных** представлений:

```
Desktop (2 колонки, 3 секции):
┌──────────────────┬──────────────────┐
│ Client Name      │ Phone            │
│ col_span: 1      │ col_span: 1      │
├──────────────────┴──────────────────┤
│ Email (col_span: 2)                 │
└─────────────────────────────────────┘

Mobile (1 колонка, компактно):
┌─────────────────────────────────────┐
│ Client Name                         │
├─────────────────────────────────────┤
│ Phone                               │
├─────────────────────────────────────┤
│ Email                               │
└─────────────────────────────────────┘
```

Grid физически другой. Это не CSS media query — это разные col_span, разные collapsed-секции, возможно разный набор видимых полей (mobile может скрывать второстепенные).

Один OV не может содержать несколько grid-конфигураций без превращения в God Object. Отсюда потребность в **отдельной сущности Layout** с привязкой к form factor.

### Проблема: Frontend нуждается в едином контракте

Frontend не должен знать про OV и Layout как отдельные концепции. Ему нужен **один объект** (Form), содержащий всё для рендеринга: структуру, presentation, conditional behavior. Describe API должен возвращать готовый Form, а не два фрагмента для склейки на клиенте.

### Связь с ADR-0019 и ADR-0022

ADR-0019 определил трёхуровневый каскад:

```
Metadata (base)
   ↓ additive validation, inherit defaults
Object View (bounded context)        ← ADR-0022
   ↓ additive validation, replace defaults, override visibility
Layout (presentation)                ← ЭТОТ ADR
   ↓ merge
Form (frontend contract)             ← ЭТОТ ADR
```

Layout — третий уровень каскада. Form — результат каскадного разрешения.

## Рассмотренные варианты

### Вариант A — Расширить Object View config (God Object)

Все layout-атрибуты (grid, col_span, ui_kind, conditional behavior, list columns) добавляются в OV config JSONB.

**Плюсы:**
- Одна сущность, одна таблица, один API
- Нет проблемы синхронизации

**Минусы:**
- God Object: OV отвечает и за bounded context, и за presentation
- Невозможность разных представлений для desktop/mobile (один config)
- Нарушение SRP: бизнес-решения и визуальные решения в одном месте
- Когнитивная сложность: администратор настраивает всё в одном экране

### Вариант B — Layout per object (shared across profiles)

Layout привязан к объекту, не к профилю. Один Layout для Order, все профили используют.

**Плюсы:**
- DRY: одна визуальная конфигурация для объекта
- Простота: один Layout на объект

**Минусы:**
- Conditional behavior (`required_expr`, `readonly_expr`) не может быть per-profile
- Пример: Sales — discount required при amount > 10000; Manager — discount всегда optional. Один Layout не покрывает оба кейса
- Section-level config (columns, collapsed) тоже per-profile: Sales — 2 колонки, Warehouse — 1 колонка

### Вариант C — Layout per Object View + Form factor (выбран)

Layout привязан к конкретному Object View и form factor. Один OV может иметь несколько Layout (desktop, tablet, mobile). Form — computed merge OV + Layout.

**Плюсы:**
- Чистое разделение: OV = ЧТО, Layout = КАК
- Per-profile conditional behavior: каждый OV имеет свои Layout с своими условиями
- Мультиплатформенность: разные Layout для разных устройств
- Единый контракт фронтенда: Form содержит всё для рендеринга
- OV работает без Layout (fallback = default presentation)
- Layout Builder (drag-and-drop) в будущем — чистая точка редактирования

**Минусы:**
- Дополнительная таблица и API
- Sync при изменении OV: добавление поля в OV должно отразиться в Layout
- Администратор работает с двумя экранами (OV + Layout)

### Вариант D — CSS-only responsive

Адаптивность через CSS media queries. Минимум конфигурации.

**Плюсы:**
- Простейшая реализация

**Минусы:**
- Администратор не может настроить grid per section
- Нет conditional field behavior
- Нет ui_kind overrides
- Нет list column configuration

## Решение

**Выбран вариант C: Layout per Object View + form factor, Form как computed контракт фронтенда.**

### Три сущности, три ответственности

```
┌──────────────────────────┐
│      Object View         │  Хранится в metadata.object_views
│   (bounded context)      │  Per (object, profile)
│                          │
│  ЧТО: секции, поля,     │
│  действия, related lists │
└────────────┬─────────────┘
             │ 1:N
             ▼
┌──────────────────────────┐
│        Layout            │  Хранится в metadata.layouts
│    (presentation)        │  Per (object_view, form_factor)
│                          │
│  КАК: grid, col_span,   │
│  ui_kind, conditions,    │
│  list columns            │
└────────────┬─────────────┘
             │ resolve + merge
             ▼
┌──────────────────────────┐
│         Form             │  Computed (не хранится)
│  (frontend contract)     │  Describe API response
│                          │
│  ВСЁ: структура +       │
│  presentation +          │
│  conditional exprs       │
└──────────────────────────┘
```

### Object View config (ЧТО — без изменений относительно ADR-0022)

```jsonc
{
  // Секции — группировка полей
  "sections": [
    {
      "key": "client_info",
      "label": "Информация о клиенте",
      "fields": ["client_name", "contact_phone", "email"]
    },
    {
      "key": "products",
      "label": "Товары",
      "fields": ["products", "total_amount", "discount"]
    }
  ],

  // Ключевые поля вверху карточки
  "highlight_fields": ["order_number", "status", "total_amount"],

  // Действия
  "actions": [
    {
      "key": "send_proposal",
      "label": "Отправить предложение",
      "type": "primary",
      "icon": "mail",
      "visibility_expr": "record.status == 'draft'"
    }
  ],

  // Дочерние объекты
  "related_lists": [
    {
      "object": "Activity",
      "label": "Активности",
      "fields": ["subject", "type", "due_date", "status"],
      "sort": "due_date DESC",
      "limit": 10
    }
  ],

  // Колонки в списке (какие — без визуальных деталей)
  "list_fields": ["order_number", "client_name", "status", "total_amount", "created_at"],
  "list_default_sort": "created_at DESC"
}
```

### Layout config (КАК)

```jsonc
{
  // Презентация секций
  "section_config": {
    "client_info": {
      "columns": 2,
      "collapsed": false
    },
    "products": {
      "columns": 2,
      "collapsed": false,
      "visibility_expr": "record.status != 'cancelled'"
    }
  },

  // Презентация полей
  "field_config": {
    "client_name": {
      "col_span": 2,
      "ui_kind": "lookup",
      "reference_config": {
        "display_fields": ["name", "email"],
        "search_fields": ["name", "email", "phone"],
        "target": "popup"
      }
    },
    "status": {
      "col_span": 1,
      "ui_kind": "badge"
    },
    "discount": {
      "col_span": 1,
      "required_expr": "record.amount > 10000",
      "readonly_expr": "record.status == 'closed'"
    },
    "email": {
      "col_span": 2,
      "ui_kind": "email"
    }
  },

  // Презентация колонок списка
  "list_columns": {
    "order_number": {"width": "15%", "sortable": true},
    "client_name": {"width": "30%", "sortable": true},
    "status": {"width": "100px", "align": "center", "ui_kind": "badge"},
    "total_amount": {"width": "15%", "align": "right", "sortable": true},
    "created_at": {"width": "15%", "sortable": true, "sort_dir": "desc"}
  }
}
```

### Layout для мобильной платформы

Тот же Object View, другой Layout:

```jsonc
// Layout (Order Sales, mobile)
{
  "section_config": {
    "client_info": {
      "columns": 1,            // 1 колонка вместо 2
      "collapsed": false
    },
    "products": {
      "columns": 1,
      "collapsed": true,       // свёрнута на мобильном
      "visibility_expr": "record.status != 'cancelled'"
    }
  },

  "field_config": {
    "client_name": {
      "col_span": 1,           // full width (1 колонка)
      "ui_kind": "lookup",
      "reference_config": {
        "display_fields": ["name"],   // меньше полей для мобильного
        "target": "link"              // ссылка вместо popup
      }
    },
    "email": {
      "col_span": 1,
      "visibility_expr": "false"      // скрыто на мобильном
    }
  },

  "list_columns": {
    "order_number": {"width": "30%"},
    "status": {"width": "30%", "ui_kind": "badge"},
    "total_amount": {"width": "40%", "align": "right"}
  }
}
```

### Пример: фазы бизнес-процесса (визит врача)

Один OV + один Layout — разное поведение по фазе через `visibility_expr`:

```jsonc
// OV (Visit, Doctor) — все поля для профиля Doctor
{
  "sections": [
    {"key": "patient", "label": "Пациент", "fields": ["patient_name", "age", "diagnosis"]},
    {"key": "recommendations", "label": "Рекомендации", "fields": ["notes", "medications"]},
    {"key": "review", "label": "Результат визита", "fields": ["outcome", "next_visit", "rating"]}
  ],
  "actions": [
    {"key": "start_visit", "label": "Начать приём", "visibility_expr": "record.status == 'scheduled'"},
    {"key": "complete_visit", "label": "Завершить приём", "visibility_expr": "record.status == 'in_progress'"}
  ]
}

// Layout (Visit, Doctor, desktop)
{
  "section_config": {
    "patient": {"columns": 2, "collapsed": false},
    "recommendations": {
      "columns": 1,
      "visibility_expr": "record.status == 'scheduled' || record.status == 'in_progress'"
    },
    "review": {
      "columns": 2,
      "visibility_expr": "record.status == 'completed'"
    }
  },
  "field_config": {
    "outcome": {"required_expr": "record.status == 'completed'"},
    "next_visit": {"ui_kind": "date"},
    "rating": {"ui_kind": "rating", "readonly_expr": "record.status != 'completed'"}
  }
}
```

До визита: секция "Рекомендации" видна, "Результат" скрыт.
После визита: "Результат" видна с обязательным outcome, "Рекомендации" скрыта.

Один OV, один Layout. Фронтенд вычисляет `visibility_expr` через cel-js и адаптирует форму мгновенно.

### Form — computed контракт фронтенда

Form строится сервером при Describe API запросе:

```
GET /api/v1/describe/Order
Authorization: Bearer <jwt>        → определяет profile
X-Form-Factor: desktop             → определяет platform

Response: {
  "object": { ... },
  "fields": [ ... ],
  "form": {                         ← merged OV + Layout
    "sections": [
      {
        "key": "client_info",
        "label": "Информация о клиенте",
        "columns": 2,
        "collapsed": false,
        "fields": [
          {"field": "client_name", "col_span": 2, "ui_kind": "lookup", "reference_config": {...}},
          {"field": "contact_phone", "col_span": 1},
          {"field": "email", "col_span": 2, "ui_kind": "email"}
        ]
      }
    ],
    "highlight_fields": ["order_number", "status", "total_amount"],
    "actions": [
      {"key": "send_proposal", "label": "Отправить", "type": "primary", "visibility_expr": "..."}
    ],
    "related_lists": [...],
    "list_columns": [
      {"field": "order_number", "width": "15%", "sortable": true},
      {"field": "total_amount", "width": "15%", "align": "right"}
    ]
  }
}
```

**Frontend получает Form и работает только с ним.** Не знает про OV и Layout как отдельные концепции.

### Хранение

```sql
-- Layouts — presentation per (object_view, form_factor)
CREATE TABLE metadata.layouts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    object_view_id  UUID NOT NULL REFERENCES metadata.object_views(id) ON DELETE CASCADE,
    form_factor     VARCHAR(20) NOT NULL DEFAULT 'desktop',  -- desktop | tablet | mobile
    config          JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT layouts_form_factor_check
        CHECK (form_factor IN ('desktop', 'tablet', 'mobile')),
    CONSTRAINT layouts_object_view_form_factor_unique
        UNIQUE (object_view_id, form_factor)
);

CREATE INDEX idx_layouts_object_view_id ON metadata.layouts(object_view_id);
```

### Resolution rules

```
1. Resolve Object View:
   a. object_views WHERE object_id = X AND profile_id = P → found? use
   b. object_views WHERE object_id = X AND is_default = true → found? use
   c. Fallback: auto-generate OV (все FLS-доступные поля, одна секция)

2. Resolve Layout:
   a. layouts WHERE object_view_id = OV.id AND form_factor = requested → found? use
   b. layouts WHERE object_view_id = OV.id AND form_factor = 'desktop' → fallback to desktop
   c. No Layout → auto-generate (col_span=1, ui_kind=auto, все поля видны)

3. Merge → Form:
   a. OV sections + Layout section_config → merged sections
   b. OV fields + Layout field_config → merged fields with presentation
   c. OV actions (as-is, already have visibility_expr)
   d. OV list_fields + Layout list_columns → merged list columns
   e. FLS intersection: убрать поля, запрещённые FLS
```

### Взаимодействие с Security

Layout **не расширяет** доступ:

```
Видимые поля = OV fields ∩ FLS-доступные поля ∩ Layout visibility
```

- Если FLS запрещает поле → поле не в Form (security wins)
- Если OV не включает поле → поле не в Form (bounded context wins)
- Если Layout скрывает поле (`visibility_expr: "false"`) → поле не рендерится (presentation)

Layout может только **сужать** видимость, не расширять.

### Lifecycle: синхронизация OV → Layout

При изменении Object View (добавление/удаление поля):

```
OV field added (discount added to section "products")
  → Layout не меняется
  → Form merge: поле discount появляется с default presentation (col_span=1, ui_kind=auto)
  → Admin может обогатить Layout для нового поля

OV field removed (discount removed from section)
  → Layout field_config.discount остаётся (orphan — не влияет)
  → Form merge: поле discount не в OV → не появляется в Form
  → Orphan cleanup: периодическая или при сохранении Layout

OV section added
  → Layout section_config не содержит новую секцию → default (columns=1, collapsed=false)
  → Admin может обогатить Layout для новой секции
```

**Принцип: OV — source of truth для структуры. Layout дополняет, но не может показать то, чего нет в OV.**

### API

```
-- Layout CRUD (Admin)
GET    /api/v1/admin/layouts?object_view_id=:ovId    — список layouts для OV
POST   /api/v1/admin/layouts                          — создать layout
GET    /api/v1/admin/layouts/:id                      — получить layout
PUT    /api/v1/admin/layouts/:id                      — обновить layout
DELETE /api/v1/admin/layouts/:id                      — удалить layout

-- Form (User-facing — через Describe API)
GET    /api/v1/describe/:objectName                   — включает resolved Form
       Header: X-Form-Factor: desktop|tablet|mobile
```

### Constructor UI

**Layout Constructor** — отдельный admin-экран, доступный из Object View detail:

1. **Section tab**: per-section config (columns slider, collapsed toggle, visibility_expr)
2. **Fields tab**: per-field config (col_span slider, ui_kind picker, required_expr, readonly_expr, reference_config)
3. **List tab**: column config (width, align, sortable, filterable)
4. **Preview**: live preview с тестовыми данными (переключение desktop/tablet/mobile)

Навигация: Object View detail → кнопка "Layout (desktop)" / "Layout (mobile)" → Layout Constructor.

### Ограничения

| Параметр | Лимит | Обоснование |
|----------|-------|-------------|
| Layouts per OV | 3 (desktop, tablet, mobile) | Фиксированные form factors |
| field_config entries | Не ограничено | По количеству полей в OV |
| visibility_expr size | 1 KB | CEL-выражение, не программа |
| Nesting (col_span) | 1-12 | CSS grid column span |

### Типы ui_kind

| ui_kind | Описание | Применение |
|---------|----------|------------|
| `auto` | Авто-определение по field type/subtype | По умолчанию |
| `text` | Текстовое поле | string |
| `textarea` | Многострочное текстовое поле | text/long_text |
| `number` | Числовое поле | number |
| `currency` | Числовое с символом валюты | number/currency |
| `percent` | Числовое с % | number/percent |
| `email` | Email с иконкой и кликабельной ссылкой | string/email |
| `phone` | Телефон с кликабельной ссылкой | string/phone |
| `url` | URL с кликабельной ссылкой | string/url |
| `date` | Date picker | datetime/date |
| `datetime` | DateTime picker | datetime |
| `checkbox` | Чекбокс | boolean |
| `toggle` | Toggle switch | boolean |
| `select` | Dropdown | picklist |
| `radio` | Radio buttons | picklist (≤ 5 options) |
| `badge` | Цветной бейдж | picklist/status |
| `lookup` | Поле с поиском и выбором | reference |
| `rating` | Звёзды/шкала | number (1-5) |
| `slider` | Ползунок | number (range) |
| `color` | Color picker | string/color |
| `rich_text` | Rich text editor | text/rich |

При `ui_kind: "auto"` тип компонента определяется из field type/subtype (metadata). Override через Layout позволяет администратору выбрать альтернативный компонент.

## Последствия

### Позитивные

- **Чистое разделение**: OV = ЧТО (bounded context), Layout = КАК (presentation), Form = единый контракт
- **Мультиплатформенность**: один OV — разные Layout для desktop/tablet/mobile
- **Per-profile conditional behavior**: Layout per OV → каждый профиль может иметь свои required_expr, readonly_expr
- **Фазы бизнес-процесса**: visibility_expr на секциях/полях — одна форма адаптируется к состоянию записи
- **Frontend simplicity**: получает Form, не знает про OV и Layout
- **Graceful degradation**: нет Layout → default presentation из OV; нет OV → auto-generate из metadata + FLS
- **Layout Builder**: в будущем — visual drag-and-drop editor для Layout (чистая точка редактирования)
- **Dual-stack CEL**: visibility_expr, required_expr, readonly_expr вычисляются через cel-js на фронте мгновенно
- **Custom Functions (ADR-0026)**: `fn.*` доступны в Layout expressions (`fn.is_premium(record.tier)`)

### Негативные

- **Дополнительная таблица + API**: metadata.layouts с CRUD endpoints
- **Admin workflow**: два экрана (OV editor + Layout editor) вместо одного
- **Orphan config**: при удалении поля из OV, Layout field_config содержит orphan entries (нужен cleanup)
- **Merge complexity**: Form resolution требует merge OV + Layout + FLS intersection (кэшируемо)
- **Три концепции**: администратор должен понимать OV, Layout, Form (Constructor UI снижает порог)

## Связанные ADR

- **ADR-0019** — Декларативная бизнес-логика: Layout = третий уровень каскада (Metadata → OV → Layout → Form)
- **ADR-0022** — Object View: Layout строится поверх OV, не заменяет. Form = merge OV + Layout
- **ADR-0026** — Custom Functions: fn.* доступны в visibility_expr, required_expr, readonly_expr
- **ADR-0009..0012** — Security: Layout не расширяет доступ. Form = OV ∩ FLS ∩ Layout visibility
