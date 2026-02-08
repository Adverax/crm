# ADR-0005: Референсные типы полей

**Статус:** Принято
**Дата:** 2026-02-08
**Участники:** @roman_myakotin

## Контекст

Metadata-driven CRM требует описания связей между объектами. Связи определяются
через reference-поля в метаданных. Необходимо спроектировать:
- Виды связей (subtypes в рамках type `reference`, см. ADR-0004)
- Поведение при удалении parent-записи
- Ограничения (глубина, циклы, self-reference)
- Полиморфные ссылки (поле, указывающее на записи разных объектов)

## Решение

### Три subtype для type `reference`

#### association — мягкая связь

Объекты связаны, но независимы.

| Аспект | Значение |
|--------|----------|
| PG storage | `UUID` (nullable) |
| FK constraint | да |
| on_delete | `set_null` или `restrict` |
| owner записи | собственный |
| sharing/security | собственный |
| reparenting | всегда |
| self-reference | разрешён (например, Account.parent_account_id) |
| max на объект | без ограничений |

#### composition — жёсткая связь (lifecycle dependency)

Child не существует без parent. Часть целого.

| Аспект | Значение |
|--------|----------|
| PG storage | `UUID NOT NULL` |
| FK constraint | да |
| on_delete | `cascade` или `restrict` |
| owner записи | наследуется от parent |
| sharing/security | наследуется от parent |
| reparenting | по флагу `is_reparentable` (default: false) |
| self-reference | **запрещён** (рекурсивный каскад) |
| max на объект | не ограничено, но глубина цепочки ≤ 2 |

#### polymorphic — ссылка на разные типы объектов

Поле может указывать на записи разных объектов. Список допустимых
объектов-целей хранится явно.

| Аспект | Значение |
|--------|----------|
| PG storage | два столбца: `VARCHAR(100)` + `UUID` |
| FK constraint | **нет** (валидация в DML engine) |
| on_delete | зависит от контекста, валидация в коде |
| owner записи | собственный |
| self-reference | разрешён |
| max на объект | без ограничений |

### Метаданные reference-поля

Дополнительные атрибуты в `field_definitions` (помимо общих):

```sql
-- Для association и composition:
referenced_object_id UUID     REFERENCES object_definitions(id),
relationship_name    VARCHAR(100),  -- имя обратной связи для SOQL
on_delete            VARCHAR(20) NOT NULL DEFAULT 'set_null'
                     CHECK (on_delete IN ('set_null', 'cascade', 'restrict')),
is_reparentable      BOOLEAN NOT NULL DEFAULT true,

-- Для polymorphic:
-- referenced_object_id = NULL (целей несколько)
-- relationship_name = имя обратной связи
-- on_delete = поведение определяется для каждой цели или по умолчанию
```

### Таблица polymorphic_targets

```sql
CREATE TABLE polymorphic_targets (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    field_id   UUID NOT NULL REFERENCES field_definitions(id) ON DELETE CASCADE,
    object_id  UUID NOT NULL REFERENCES object_definitions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (field_id, object_id)
);
```

Явный список допустимых объектов-целей для каждого polymorphic-поля.
Без специальных флагов (`is_target_any`, `is_platform_managed`) —
при необходимости добавляются позже.

### Поведение при удалении (on_delete)

| on_delete | Описание | Доступен для |
|-----------|----------|-------------|
| `set_null` | Обнулить ссылку у child-записей | association |
| `cascade` | Удалить child-записи | composition |
| `restrict` | Запретить удаление parent, если есть children | association, composition |

### Ограничения

| Ограничение | Значение | Обоснование |
|-------------|----------|-------------|
| Self-reference composition | **запрещён** | Рекурсивный каскад при удалении |
| Глубина composition цепочки | **≤ 2** | A→B→C допустимо, A→B→C→D — нет. Ограничивает каскадную сложность |
| Циклы в composition | **запрещены** | Metadata engine проверяет при создании поля |
| Max composition на объект | без ограничений | Глубина цепочки уже ограничена |

### Хранение данных в таблицах объектов

```sql
-- association (например, Contact.account_id):
account_id UUID REFERENCES obj_account(id) ON DELETE SET NULL

-- composition (например, DealLineItem.deal_id):
deal_id UUID NOT NULL REFERENCES obj_deal(id) ON DELETE CASCADE

-- polymorphic (например, Task.what):
what_object_type VARCHAR(100) NOT NULL
what_record_id   UUID         NOT NULL
-- + composite index (what_object_type, what_record_id)
-- + DML-валидация: object_type ∈ polymorphic_targets
```

## Последствия

- Reference type имеет три subtype: `association`, `composition`, `polymorphic`
- Metadata engine валидирует ограничения при создании reference-полей
- DML engine проверяет referential integrity для polymorphic (нет FK)
- Schema generator создаёт разные DDL в зависимости от subtype
- Polymorphic targets — отдельная нормализованная таблица
- Security inheritance (owner/sharing) для composition — Phase 2
