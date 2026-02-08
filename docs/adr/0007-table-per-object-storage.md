# ADR-0007: Таблица на каждый объект (table-per-object storage)

**Статус:** Принято
**Дата:** 2026-02-08
**Участники:** @roman_myakotin

## Контекст

Metadata-driven CRM позволяет создавать произвольные объекты (custom objects) с произвольными
полями. Необходимо определить, как данные записей хранятся в PostgreSQL.

Ключевые факторы:
- Custom objects создаются редко (admin-операции), запросы к данным — постоянно
- SOQL/DML — единый API для доступа к данным, транслируется в SQL
- Важны нативные constraints: FK, UNIQUE, NOT NULL, индексы
- PostgreSQL 16 поддерживает транзакционный DDL (CREATE TABLE, ALTER TABLE — атомарны)

## Рассмотренные варианты

### Вариант A: Таблица на каждый объект (выбран)

Каждый объект получает собственную таблицу. Создание объекта = `CREATE TABLE`,
добавление поля = `ALTER TABLE ADD COLUMN`.

**Плюсы:**
- Нативная производительность PostgreSQL — индексы, query planner, JOINs
- Сильная типизация на уровне БД (VARCHAR, NUMERIC, UUID, BOOLEAN, TIMESTAMPTZ)
- FK constraints работают нативно для reference-полей
- UNIQUE, NOT NULL constraints на уровне БД
- SOQL → SQL трансляция тривиальна (прямой маппинг object → table, field → column)
- Проверено на практике

**Минусы:**
- DDL в runtime (CREATE TABLE, ALTER TABLE)
- Нужны DDL-привилегии для приложения
- Количество таблиц растёт с количеством объектов

### Вариант B: EAV (Entity-Attribute-Value)

Одна таблица `record_values` с колонками `(record_id, field_id, value_text, value_number, ...)`.

**Плюсы:** никакого DDL в runtime, полностью динамическая схема.
**Минусы:** ужасная производительность для сложных запросов (self-JOIN на каждое поле),
невозможны FK/UNIQUE constraints, SOQL → SQL трансляция крайне сложная.

### Вариант C: Wide table (подход Salesforce)

Одна таблица `data_rows` с generic-колонками `val0..val500`, все значения как TEXT.

**Плюсы:** никакого DDL, единая таблица.
**Минусы:** потеря типизации, лимит на количество полей, sparse storage, сложный кастинг.

### Вариант D: JSONB-документ

Одна таблица `records` с колонкой `data JSONB`.

**Плюсы:** никакого DDL, JSONB хорошо оптимизирован в PG, GIN-индексы.
**Минусы:** нет FK constraints внутри JSONB, UNIQUE constraints только через partial index,
агрегации медленнее нативных колонок.

## Решение

Принимаем **Вариант A** — таблица на каждый объект.

### Расположение таблиц

Физическое расположение таблицы определяется полями `schema_name` и `table_name`
в `object_definitions` (ADR-0003). Это позволяет размещать данные в разных PG-схемах.

Конвенция по умолчанию:

| object_type | schema_name | table_name | Пример |
|-------------|-------------|------------|--------|
| standard | `public` | `obj_{api_name}` | `public.obj_account` |
| custom | `public` | `obj_{api_name}` | `public.obj_invoice` |

Администратор может переопределить схему при необходимости.

DDL и SOQL/DML engine обращаются к таблице через `{schema_name}.{table_name}` из метаданных,
а не через вычисление имени из `api_name`.

### Структура таблицы объекта

При создании объекта metadata engine генерирует DDL:

```sql
CREATE TABLE public.obj_invoice (
    -- Системные поля (обязательные для каждого объекта)
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    UUID        NOT NULL REFERENCES users(id),
    created_by  UUID        NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by  UUID        NOT NULL REFERENCES users(id),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Добавление поля

При создании поля metadata engine генерирует ALTER TABLE:

```sql
-- text/plain
ALTER TABLE obj_invoice ADD COLUMN number VARCHAR(20);

-- number/currency
ALTER TABLE obj_invoice ADD COLUMN amount NUMERIC(18,2);

-- reference/association
ALTER TABLE obj_invoice ADD COLUMN account_id UUID REFERENCES obj_account(id) ON DELETE SET NULL;

-- boolean
ALTER TABLE obj_invoice ADD COLUMN is_paid BOOLEAN NOT NULL DEFAULT false;

-- picklist/single
ALTER TABLE obj_invoice ADD COLUMN status VARCHAR(255);
```

### Маппинг field_type → DDL

| field_type | field_subtype | DDL column type |
|-----------|---------------|----------------|
| text | plain | `VARCHAR(n)` — n из config.max_length |
| text | area, rich | `TEXT` |
| text | email | `VARCHAR(255)` |
| text | phone | `VARCHAR(40)` |
| text | url | `VARCHAR(2048)` |
| number | integer | `NUMERIC(p,0)` — p из config.precision |
| number | decimal, currency, percent | `NUMERIC(p,s)` — из config |
| number | auto_number | `INTEGER GENERATED ALWAYS AS IDENTITY` |
| boolean | — | `BOOLEAN` |
| datetime | date | `DATE` |
| datetime | datetime | `TIMESTAMPTZ` |
| datetime | time | `TIME` |
| picklist | single | `VARCHAR(255)` |
| picklist | multi | `TEXT[]` |
| reference | association | `UUID REFERENCES obj_{target}(id) ON DELETE SET NULL` |
| reference | composition | `UUID NOT NULL REFERENCES obj_{target}(id) ON DELETE CASCADE` |
| reference | polymorphic | два столбца: `{name}_object_type VARCHAR(100) NOT NULL` + `{name}_record_id UUID NOT NULL` |

### Constraints из метаданных

```sql
-- is_required = true
ALTER TABLE obj_invoice ALTER COLUMN number SET NOT NULL;

-- is_unique = true
ALTER TABLE obj_invoice ADD CONSTRAINT uq_invoice_number UNIQUE (number);
```

### Индексы

Metadata engine автоматически создаёт индексы:
- FK-колонки (reference-поля)
- Поля с `is_unique = true`
- Composite index для polymorphic reference: `(object_type, record_id)`
- `owner_id` (для RLS-запросов)

### Удаление поля

```sql
ALTER TABLE obj_invoice DROP COLUMN amount;
```

### Удаление объекта

```sql
DROP TABLE obj_invoice;
```

Hard delete с подтверждением (ADR-0003).

## Последствия

- Metadata engine выполняет DDL при создании/изменении объектов и полей
- Приложение требует DDL-привилегий на схему данных
- SOQL → SQL трансляция: object → `{schema_name}.{table_name}` из метаданных, field → column name
- Нативные PG constraints обеспечивают целостность данных
- Schema migration для custom objects управляется платформой, не файлами миграций
- Standard objects (Account, Contact и т.д.) создаются seed-скриптом при инициализации
