# ADR-0006: Реестр связей как кэш, а не отдельная таблица

**Статус:** Принято
**Дата:** 2026-02-08
**Участники:** @roman_myakotin

## Контекст

Reference-поля в `field_definitions` (ADR-0004, ADR-0005) неявно определяют связи
между объектами. Например, поле `Contact.account_id` типа `reference/association`
создаёт связь Account → Contacts.

Для работы платформы необходима **навигация по связям**:
- SOQL: `SELECT (SELECT Name FROM Contacts) FROM Account` — обратная связь
- Admin UI: "Покажи все связи объекта Account"
- Каскадный анализ: "Что удалится при удалении Account?"
- Junction-обнаружение: объект с 2+ composition = junction

Вопрос: нужна ли отдельная таблица `relationship_definitions` или связи
можно получить из существующих метаданных?

## Рассмотренные варианты

### Вариант A: Отдельная таблица relationship_definitions

```sql
CREATE TABLE relationship_definitions (
    id                UUID PRIMARY KEY,
    parent_object_id  UUID REFERENCES object_definitions(id),
    child_object_id   UUID REFERENCES object_definitions(id),
    child_field_id    UUID REFERENCES field_definitions(id),
    relationship_name VARCHAR(100),
    relationship_type VARCHAR(20),
    created_at        TIMESTAMPTZ
);
```

**Плюсы:**
- Быстрые прямые запросы по parent/child
- Удобная навигация по графу связей

**Минусы:**
- Денормализация: вся информация уже есть в `field_definitions` + `polymorphic_targets`
- Два источника правды → необходимость синхронизации при каждом изменении поля
- Рассинхронизация = баги в SOQL-резолве и каскадах

### Вариант B: Без реестра — всё из field_definitions

**Плюсы:**
- Один источник правды, нет синхронизации

**Минусы:**
- Обратные запросы ("все дети Account") требуют scan по `field_definitions` всех объектов
- SOQL-резолв обратных связей медленнее

### Вариант C: Кэш в памяти, построенный из field_definitions (выбран)

Relationship graph **вычисляется** из `field_definitions` + `polymorphic_targets`
и хранится в in-memory кэше. Инвалидируется при изменении метаданных.

**Плюсы:**
- Один источник правды (field_definitions) — нет рассинхронизации
- Быстрый доступ: O(1) lookup по (object_id, relationship_name)
- Метаданные меняются редко (admin-операции), читаются на каждый SOQL/DML — кэш идеален
- Кэширование метаданных неизбежно для производительности — relationship graph становится
  частью общего metadata cache, а не отдельной подсистемой

**Минусы:**
- Нужен механизм инвалидации кэша при изменении метаданных
- При холодном старте — построение графа из БД (одноразовая операция)

## Решение

Принимаем **Вариант C**. Никакой отдельной таблицы. Relationship graph — часть
in-memory metadata cache.

### Структура кэша

```
MetadataCache
├── objects: map[api_name] → ObjectDefinition
├── fields:  map[object_id] → []FieldDefinition
└── relationships:
    ├── forward:  map[object_id][field_api_name] → RelationshipInfo
    └── reverse:  map[object_id][relationship_name] → RelationshipInfo
```

`RelationshipInfo` содержит:
- `parent_object_id` — родительский объект
- `child_object_id` — дочерний объект
- `child_field_id` — поле-ссылка
- `relationship_name` — имя связи (для SOQL)
- `relationship_type` — association / composition / polymorphic
- `on_delete` — set_null / cascade / restrict

### Построение кэша

При старте приложения (или инвалидации):

1. Загрузить все `field_definitions` с `field_type = 'reference'`
2. Загрузить все `polymorphic_targets`
3. Построить forward map: `(child_object, field) → parent_object`
4. Построить reverse map: `(parent_object, relationship_name) → child_object + field`
5. Для polymorphic: одна запись в reverse map для каждого target-объекта

### Инвалидация

- Любое изменение в `field_definitions` (CREATE/UPDATE/DELETE reference-поля) → rebuild
- Любое изменение в `polymorphic_targets` → rebuild
- Изменение `object_definitions` (DELETE объекта) → rebuild
- Метаданные меняются редко, полный rebuild допустим (десятки/сотни объектов)

### Использование

```
// SOQL: SELECT (SELECT Name FROM Contacts) FROM Account
// Резолв "Contacts":
rel := cache.relationships.reverse["Account"]["Contacts"]
// → {child_object: "Contact", child_field: "account_id", type: "association"}

// Каскадный анализ при удалении Account:
children := cache.relationships.reverse["Account"]
// → filter by on_delete == "cascade"

// Все связи объекта (для Admin UI):
forward := cache.relationships.forward["Contact"]   // куда ссылается
reverse := cache.relationships.reverse["Contact"]   // кто ссылается на него
```

## Последствия

- Нет таблицы `relationship_definitions` — один источник правды
- Metadata cache — обязательный компонент платформы (не опциональная оптимизация)
- SOQL engine, DML engine, Admin UI работают с кэшем, не с прямыми запросами к field_definitions
- Инвалидация кэша при любом изменении метаданных
- В кластерной конфигурации — инвалидация через pub/sub (Redis, PG NOTIFY) — проектируется позже
