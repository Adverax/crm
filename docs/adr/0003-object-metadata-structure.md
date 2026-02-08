# ADR-0003: Структура метаданных объекта

**Статус:** Принято
**Дата:** 2026-02-08
**Участники:** @roman_myakotin

## Контекст

Metadata-driven CRM требует формального описания каждого объекта (Account, Contact,
пользовательские объекты). Описание определяет идентификацию, классификацию, допустимые
операции и подключённые подсистемы.

Ключевые решения:
- Классификация: минимальный enum `standard | custom` + поведенческие флаги (вместо
  богатого enum с типами `system`, `junction` и т.д.)
- Безопасность (`default_sharing_model`) вынесена в Phase 2 (Security engine)
- Soft delete объектов (`is_active`) отложен — не нужен для MVP
- Soft delete записей (`is_deleted`, `deleted_at`, `deleted_by`) отложен — каскадное
  удаление/восстановление создаёт неоднозначности, требует отдельной подсистемы
- i18n: `label`, `plural_label`, `description` хранятся как default-значения,
  переводы — через таблицу `translations` (ADR-0002)

## Рассмотренные варианты

### Классификация: богатый enum vs минимальный enum + флаги

**Вариант A — богатый enum:** `standard | custom | system | junction`

Плюсы: явная фильтрация по типу.
Минусы: жёсткий, при появлении нового типа надо менять enum. Junction выводится
из связей, system — это поведение, а не тип.

**Вариант B — минимальный enum + флаги (выбран):**
`object_type: standard | custom`, а `system`-поведение определяется через
`is_platform_managed` и другие флаги.

Плюсы: гибкий, расширяемый, не требует предугадывать все будущие типы.

### Soft delete объектов (is_active)

Отложен. Цена: каскад поведения (что с полями, связями, записями при деактивации),
фильтрация `WHERE is_active = true` в каждом запросе к метаданным.
Для MVP: standard-объекты всегда активны, custom-объекты удаляются через hard delete.
Добавить `is_active` позже — одна миграция.

### Soft delete записей (is_deleted / deleted_at / deleted_by)

Отложен. Основная проблема — неоднозначность каскадного удаления и восстановления:

**Сценарий 1:** Удалили деталь вручную → удалили мастера каскадно.
При восстановлении мастера деталь, удалённая ранее вручную, не должна восстанавливаться,
но система не может отличить её от каскадно удалённых деталей без дополнительных метаданных
(`delete_reason`, `delete_operation_id`).

**Сценарий 2:** Удалили мастера → каскадно удалились все детали.
Восстановление одной детали без мастера — сломанный FK. Восстановление мастера — нужно решить,
какие детали восстанавливать.

Корректная реализация требует:
- `delete_reason`: `user_action` | `cascade`
- `delete_operation_id`: UUID для группировки каскадных удалений
- Логика restore с учётом дерева зависимостей
- UI: отображение деревьев удаления, конфликт-резолвер, bulk restore с превью

Это отдельная подсистема (Recycle Bin / Archive), которая будет спроектирована отдельным ADR.
Для MVP: hard delete с подтверждением в UI.

## Решение

### Таблица `object_definitions`

```sql
CREATE TABLE object_definitions (
    -- Идентификация
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_name                 VARCHAR(100)  NOT NULL UNIQUE,
    label                    VARCHAR(255)  NOT NULL,
    plural_label             VARCHAR(255)  NOT NULL,
    description              TEXT          NOT NULL DEFAULT '',

    -- Физическое хранение (ADR-0007)
    schema_name              VARCHAR(63)   NOT NULL DEFAULT 'public',
    table_name               VARCHAR(63)   NOT NULL,

    -- Классификация
    object_type              VARCHAR(20)   NOT NULL CHECK (object_type IN ('standard', 'custom')),

    -- Поведенческие флаги (уровень схемы — что можно делать с самим объектом)
    is_platform_managed      BOOLEAN       NOT NULL DEFAULT false,
    is_visible_in_setup      BOOLEAN       NOT NULL DEFAULT true,
    is_custom_fields_allowed BOOLEAN       NOT NULL DEFAULT true,
    is_deleteable_object     BOOLEAN       NOT NULL DEFAULT true,

    -- Возможности записей (что можно делать с записями этого объекта)
    is_createable            BOOLEAN       NOT NULL DEFAULT true,
    is_updateable            BOOLEAN       NOT NULL DEFAULT true,
    is_deleteable            BOOLEAN       NOT NULL DEFAULT true,
    is_queryable             BOOLEAN       NOT NULL DEFAULT true,
    is_searchable            BOOLEAN       NOT NULL DEFAULT true,

    -- Фичи (подключаемые подсистемы)
    has_activities            BOOLEAN       NOT NULL DEFAULT false,
    has_notes                 BOOLEAN       NOT NULL DEFAULT false,
    has_history_tracking      BOOLEAN       NOT NULL DEFAULT false,
    has_sharing_rules         BOOLEAN       NOT NULL DEFAULT false,

    -- Системные timestamps
    created_at               TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ   NOT NULL DEFAULT now()
);
```

### Системные поля записей (в таблице данных каждого объекта, не в метаданных)

Каждая таблица данных объекта автоматически содержит:

```sql
id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
owner_id    UUID        NOT NULL REFERENCES users(id),
created_by  UUID        NOT NULL REFERENCES users(id),
created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
updated_by  UUID        NOT NULL REFERENCES users(id),
updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
```

Soft delete записей (`is_deleted`, `deleted_at`, `deleted_by`) отложен.
Для MVP — hard delete с подтверждением. Recycle Bin / Archive — отдельная фича
со своим ADR, включающим проектирование каскадного удаления и восстановления.

### Примеры конфигураций

| Объект | object_type | is_platform_managed | is_deleteable_object | is_custom_fields_allowed |
|--------|-------------|---------------------|----------------------|--------------------------|
| Account | standard | false | false | true |
| Contact | standard | false | false | true |
| User | standard | true | false | true |
| Profile | standard | true | false | false |
| Invoice__c | custom | false | true | true |

## Последствия

- Таблица `object_definitions` — центральный реестр всех объектов системы
- Metadata engine при SOQL/DML обращается к этим флагам для валидации операций
- Security (`default_sharing_model`, OLS, FLS, RLS) добавляется в Phase 2 отдельными таблицами
- `is_active` / soft delete объектов добавляется позже при необходимости
- Soft delete записей отложен — Recycle Bin / Archive будет отдельным ADR
- Для MVP: hard delete записей с подтверждением в UI
- i18n для `label`, `plural_label`, `description` — через таблицу `translations` (ADR-0002)
