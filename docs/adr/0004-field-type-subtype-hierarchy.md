# ADR-0004: Иерархия type/subtype для типизации полей

**Статус:** Принято
**Дата:** 2026-02-08
**Участники:** @roman_myakotin

## Контекст

Metadata-driven CRM поддерживает разнообразные типы полей: текст, числа, даты, ссылки,
списки выбора и т.д. Необходимо выбрать способ типизации, который обеспечит:
- Единообразную логику хранения в PostgreSQL
- Расширяемость без изменения базовой инфраструктуры
- Чёткое разделение storage-логики и семантики/валидации
- Простоту реализации UI-компонентов

## Рассмотренные варианты

### Вариант A: Плоский enum типов

```
field_type: text | textarea | rich_text | email | phone | url |
            number | currency | percent | auto_number |
            date | datetime | time | boolean |
            picklist | multipicklist | association | composition
```

**Плюсы:**
- Просто — одно поле, один enum
- Однозначный маппинг тип → поведение

**Минусы:**
- Каждый новый тип (например, `ip_address`, `rating`) расширяет enum
- Дублирование логики: `text`, `email`, `phone`, `url` хранятся одинаково (VARCHAR),
  но обрабатываются как разные типы — нужны отдельные ветки в каждом switch/case
- Нет группировки — валидатор, SOQL-оператор, UI-компонент не могут обработать
  "все строковые типы" одним блоком

### Вариант B: Иерархия type/subtype (выбран)

```
field_type:    text | number | boolean | datetime | picklist | reference
field_subtype: зависит от type (nullable для boolean)
```

`type` определяет storage (как хранить в PG), `subtype` определяет семантику
(как валидировать и рендерить).

**Плюсы:**
- Чёткое разделение: storage concern (type) vs semantic concern (subtype)
- Код организуется по типам: один handler для всех `text/*`, один для всех `number/*`
- Расширяемость: новый `text/ip_address` — это subtype + валидатор, без изменения storage
- UI: base-компонент по `type`, модификация поведения по `subtype`

**Минусы:**
- Два поля вместо одного в метаданных
- Нужна валидация допустимых комбинаций type+subtype

## Решение

Используем иерархию **type/subtype**. В таблице `field_definitions`:

```sql
field_type    VARCHAR(20) NOT NULL,  -- базовый тип: storage concern
field_subtype VARCHAR(20),           -- семантика: nullable (не обязателен для boolean)
```

### Полный реестр type/subtype

#### text → VARCHAR / TEXT

| subtype | PG storage | max_length | Валидация | UI |
|---------|-----------|-----------|-----------|-----|
| `plain` | VARCHAR(n) | 1–255 | — | text input |
| `area` | TEXT | — | — | textarea |
| `rich` | TEXT | — | HTML sanitize | rich editor |
| `email` | VARCHAR(255) | 255 | email format | mailto link |
| `phone` | VARCHAR(40) | 40 | phone format | tel link |
| `url` | VARCHAR(2048) | 2048 | URL format | clickable link |

#### number → NUMERIC

| subtype | PG storage | precision/scale | UI |
|---------|-----------|----------------|-----|
| `integer` | NUMERIC(18,0) | настраиваемый | number input |
| `decimal` | NUMERIC(p,s) | настраиваемый | number input |
| `currency` | NUMERIC(18,2) | фиксированный | с символом валюты |
| `percent` | NUMERIC(5,2) | фиксированный | с символом % |
| `auto_number` | sequence + format | — | display only, read-only |

#### boolean → BOOLEAN

| subtype | Описание |
|---------|----------|
| NULL | subtype не требуется |

#### datetime → DATE / TIMESTAMPTZ / TIME

| subtype | PG storage | UI |
|---------|-----------|-----|
| `date` | DATE | date picker |
| `datetime` | TIMESTAMPTZ | datetime picker |
| `time` | TIME | time picker |

#### picklist → VARCHAR / VARCHAR[]

| subtype | PG storage | UI |
|---------|-----------|-----|
| `single` | VARCHAR(255) | dropdown / radio |
| `multi` | VARCHAR(255)[] | multi-select / checkboxes |

#### reference → UUID FK

| subtype | nullable | ON DELETE | owner записи | UI |
|---------|----------|-----------|-------------|-----|
| `association` | yes | SET NULL | собственный | search/select, очищаемое |
| `composition` | no | CASCADE | наследуется от parent | search/select, обязательное |

Терминология взята из UML/DDD вместо Salesforce-специфичных `lookup`/`master_detail`:
- **association** — объекты связаны, но независимы. Удаление parent не уничтожает child.
- **composition** — lifecycle dependency. Child не существует без parent.

Hierarchical relationship (self-referencing, например User → User для org chart)
не выделяется в отдельный subtype — это `association` с `referenced_object = self`
и валидацией циклов в DML engine.

Детализация reference-типов (каскады, наследование, ограничения) — ADR-0005.

### Валидация допустимых комбинаций

Metadata engine при создании/обновлении поля проверяет, что пара (type, subtype)
входит в реестр допустимых комбинаций. Недопустимые комбинации отклоняются.

### Таблица `field_definitions`

```sql
CREATE TABLE field_definitions (
    -- Идентификация
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    object_id            UUID         NOT NULL REFERENCES object_definitions(id),
    api_name             VARCHAR(100) NOT NULL,
    label                VARCHAR(255) NOT NULL,
    description          TEXT         NOT NULL DEFAULT '',
    help_text            TEXT         NOT NULL DEFAULT '',

    -- Типизация
    field_type           VARCHAR(20)  NOT NULL,
    field_subtype        VARCHAR(20),

    -- Reference-связь (прямая колонка для FK constraint)
    referenced_object_id UUID         REFERENCES object_definitions(id),

    -- Структурные constraints
    is_required          BOOLEAN      NOT NULL DEFAULT false,
    is_unique            BOOLEAN      NOT NULL DEFAULT false,

    -- Type-specific параметры (JSONB вместо множества nullable-колонок)
    config               JSONB        NOT NULL DEFAULT '{}',

    -- Классификация
    is_system_field      BOOLEAN      NOT NULL DEFAULT false,
    is_custom            BOOLEAN      NOT NULL DEFAULT false,
    is_platform_managed  BOOLEAN      NOT NULL DEFAULT false,
    sort_order           INTEGER      NOT NULL DEFAULT 0,

    -- Timestamps
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),

    UNIQUE (object_id, api_name)
);
```

#### Хранение type-specific параметров: отдельные колонки vs JSONB config

**Проблема:** Type-specific атрибуты (`max_length`, `precision`, `scale`, `relationship_name`,
`on_delete`, `is_reparentable`, `auto_number_format` и т.д.) заполняются только для своего типа.
При ~8 nullable-колонках у поля `text/email` 7 из 9 будут NULL. Каждый новый атрибут = миграция.

**Решение:** Единая колонка `config JSONB`. Metadata engine при создании/обновлении поля
валидирует содержимое config по JSON-схеме, зависящей от `(field_type, field_subtype)`.

**Исключение:** `referenced_object_id` остаётся прямой колонкой — FK constraint на
`object_definitions` обеспечивает целостность на уровне БД.

#### Содержимое config по типам

| type/subtype | config |
|---|---|
| text/plain | `{"max_length": 100, "default_value": ""}` |
| text/email, phone, url | `{"default_value": ""}` |
| text/area, rich | `{"default_value": ""}` |
| number/integer | `{"precision": 18, "scale": 0, "default_value": "0"}` |
| number/decimal | `{"precision": 18, "scale": 2, "default_value": "0.00"}` |
| number/currency | `{"precision": 18, "scale": 2, "default_value": "0.00"}` |
| number/percent | `{"precision": 5, "scale": 2, "default_value": "0.00"}` |
| number/auto_number | `{"format": "INV-{0000}", "start_value": 1}` |
| boolean | `{"default_value": "false"}` |
| datetime/* | `{"default_value": ""}` |
| picklist/single | `{"values": [...], "default_value": ""}` — см. раздел Picklist values |
| picklist/multi | `{"values": [...], "default_value": []}` — см. раздел Picklist values |
| reference/association | `{"relationship_name": "Contacts", "on_delete": "set_null"}` |
| reference/composition | `{"relationship_name": "LineItems", "on_delete": "cascade", "is_reparentable": false}` |
| reference/polymorphic | `{"relationship_name": "Activities"}` |

### Picklist values

Picklist-значения хранятся в `config.values[]` — и для локальных, и для глобальных picklists.
Единообразный паттерн: metadata engine всегда читает значения из config, не из отдельной таблицы.

#### Формат values в config

```jsonc
{
  // Ссылка на глобальный picklist (null для локального)
  "picklist_id": "uuid-or-null",
  // Значения — всегда здесь, независимо от источника
  "values": [
    {"id": "uuid1", "value": "new", "label": "Новая", "sort_order": 1, "is_default": true, "is_active": true},
    {"id": "uuid2", "value": "in_progress", "label": "В работе", "sort_order": 2, "is_default": false, "is_active": true},
    {"id": "uuid3", "value": "closed", "label": "Закрыта", "sort_order": 3, "is_default": false, "is_active": true}
  ],
  "default_value": "new"
}
```

Каждое значение имеет `id` (UUID) — используется как `resource_id` в таблице `translations`
(`resource_type = 'PicklistValue'`) для i18n.

`is_active` необходим для picklist-значений: деактивированное значение не показывается
в dropdown для новых записей, но существующие записи сохраняют его. Удаление значения
сломало бы данные.

#### Глобальные picklists (Global Value Sets)

Переиспользуемые наборы значений. Хранятся в отдельных таблицах:

```sql
CREATE TABLE picklist_definitions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_name    VARCHAR(100) NOT NULL UNIQUE,
    label       VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE picklist_values (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    picklist_definition_id UUID         NOT NULL REFERENCES picklist_definitions(id) ON DELETE CASCADE,
    value                  VARCHAR(255) NOT NULL,
    label                  VARCHAR(255) NOT NULL,
    sort_order             INTEGER      NOT NULL DEFAULT 0,
    is_default             BOOLEAN      NOT NULL DEFAULT false,
    is_active              BOOLEAN      NOT NULL DEFAULT true,
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (picklist_definition_id, value)
);
```

#### Синхронизация глобальных picklists в config

При обновлении глобального picklist:
1. Обновить `picklist_values` в таблице
2. Найти все `field_definitions` где `config->>'picklist_id' = :id`
3. Перезаписать `config.values` актуальными данными из `picklist_values`

Синхронизация происходит при admin-операциях (редко). Runtime-чтение всегда из config.

Если `config.picklist_id` заполнен — поле привязано к глобальному picklist, значения
полностью синхронизируются. Локальные отклонения не допускаются. Чтобы отклониться —
отвязать от глобального (обнулить `picklist_id`), дальше управлять локально.

#### Отложено

- Зависимые picklists (dependent picklists) — не нужны для MVP
- Цвет/иконка значения — не нужны для MVP

## Последствия

- `field_definitions` содержит `field_type` + `field_subtype` (nullable) + `config` (JSONB)
- `referenced_object_id` — прямая колонка с FK для целостности
- Type-specific параметры в `config` — без миграций при добавлении новых атрибутов
- Metadata engine валидирует config по JSON-схеме для каждой комбинации type+subtype
- Структурные constraints (`is_required`, `is_unique`) — прямые колонки (влияют на DDL)
- Бизнес-валидация (CEL-выражения) — отдельная сущность Validation Rules (отложена)
- Валидаторы, SOQL-операторы, DML-обработчики организованы по `field_type`
- Добавление нового subtype — регистрация + валидатор, без миграций и изменения core
- UI-компоненты: base по `type`, поведение по `subtype`
- `boolean` не имеет subtype (`field_subtype = NULL`)
- Reference subtypes: `association`, `composition`, `polymorphic` (детали в ADR-0005)
- Hierarchical — не отдельный subtype, а `association` с self-reference
- i18n для `label`, `description`, `help_text` — через таблицу `translations` (ADR-0002)
- Picklist values всегда в `config.values[]` — единообразное чтение
- Глобальные picklists: таблицы `picklist_definitions` + `picklist_values`, sync в config
- Зависимые picklists и цвет — отложены
