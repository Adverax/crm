# SOQL: Архитектура реализации

## Введение

Этот документ описывает архитектуру реализации языка SOQL (Salesforce Object Query Language) для нашей CRM-системы. Реализация основана на спецификации из [MANUAL.md](./MANUAL.md) и следует принципам Clean Architecture.

### Цели

1. **Независимость от реализации** — модуль SOQL не зависит от конкретной БД или системы метаданных
2. **Расширяемость** — возможность добавлять новые возможности без изменения ядра
3. **Безопасность** — встроенный контроль доступа на уровне объектов и полей
4. **Соответствие спецификации** — максимальная совместимость с SOQL Salesforce
5. **Производительность** — кэширование метаданных и скомпилированных запросов

### Ключевые решения

| Вопрос | Решение |
|--------|---------|
| Формат результата | Иерархический (через `JSON_AGG` в PostgreSQL) |
| Литералы дат | Параметризованные запросы с вычислением в runtime |
| TYPEOF (полиморфные поля) | Не реализуем |
| Ограничения SQL | Сохраняем ограничения SOQL (нет `SELECT *`, `JOIN`, `UNION`) |
| Именование | Маппинг через метаданные (SOQL CamelCase → SQL snake_case) |
| Пагинация | Keyset (cursor-based) вместо OFFSET |
| Кэширование | Двухуровневое: метаданные + скомпилированные запросы |

---

## Структура модуля (Clean Architecture)

```
internal/data/soql/
├── domain/                      # Domain layer
│   ├── executor.go              # QueryExecutor interface
│   ├── query.go                 # Query, QueryParams, QueryResult
│   └── invalidator.go           # CacheInvalidator interface
│
├── application/                 # Application layer
│   ├── engine/                  # SOQL Engine (core)
│   │   ├── ast.go               # AST structures (Grammar, Expression, etc.)
│   │   ├── lexer.go             # Lexer (tokens)
│   │   ├── parser.go            # Parser (participle v2)
│   │   ├── primitives.go        # FieldType, Operator, Aggregate, Function, DateLiteral
│   │   ├── validator.go         # AST validation & semantic analysis
│   │   ├── compiler.go          # AST → SQL compilation
│   │   ├── metadata.go          # MetadataProvider, ObjectMeta, FieldMeta
│   │   ├── access.go            # AccessController interface
│   │   ├── dates.go             # DateResolver, date literal resolution
│   │   ├── pagination.go        # Keyset pagination logic
│   │   ├── limits.go            # Query limits configuration
│   │   ├── errors.go            # Typed errors (Parse, Validation, Access, etc.)
│   │   └── engine.go            # Engine: public API, query cache
│   │
│   └── service/                 # Application service
│       └── service.go           # QueryService (orchestrates Engine + Executor)
│
├── infrastructure/              # Infrastructure layer
│   ├── access/                  # Access control implementation
│   │   └── controller.go        # SOQLAccessController (OLS/FLS)
│   │
│   ├── metadata/                # Metadata provider implementation
│   │   └── adapter.go           # MetadataAdapter (with cache)
│   │
│   ├── postgres/                # PostgreSQL implementation
│   │   ├── postgres.go          # PostgresExecutor implementation
│   │   ├── executor.go          # DB interface, Executor interface, Record types
│   │   └── cursor.go            # Cursor encryption/decryption
│   │
│   ├── cache/                   # Cache infrastructure
│   │   └── invalidator.go       # AggregatedInvalidator
│   │
│   └── outbox/                  # Event processing
│       └── worker.go            # MetadataOutboxWorker
│
└── interface/                   # Interface layer
    └── http/                    # HTTP handlers
        └── api.go               # SOQLApi (REST endpoint)
```

---

## Pipeline обработки запроса

```
SOQL Query String
       │
       ▼
┌──────────────────┐
│  Query Cache     │───── HIT ────▶ CompiledQuery
│  (Engine)        │
└────────┬─────────┘
         │ MISS
         ▼
   ┌────────┐
   │ Parser │ → AST (Grammar)
   └────────┘
       │
       ▼
   ┌───────────┐      ┌─────────────────────┐
   │ Validator │─────▶│ SOQLAccessController │ (OLS/FLS проверки)
   └───────────┘      └─────────────────────┘
       │              ┌────────────────────┐
       ├─────────────▶│ Metadata Cache     │
       │              │ (MetadataAdapter)  │
       │              └────────────────────┘
       ▼
   ┌──────────┐
   │ Compiler │ → CompiledQuery {SQL, Params, DateParams, Shape, Dependencies}
   └──────────┘
       │
       ├──────▶ Store in Query Cache
       │
       ▼
   ┌───────────────┐
   │ DateResolver  │ → Resolve date literals (TODAY → 2024-01-15)
   └───────────────┘
       │
       ▼
   ┌───────────────────┐      ┌─────────────────┐
   │ PostgresExecutor  │─────▶│ RLSGuard        │ (app.user_id → RLS policies)
   └───────────────────┘      └─────────────────┘
       │
       ▼
   QueryResult {Records, NextCursor, Done}
```

---

## Интерфейсы

### MetadataProvider

Предоставляет метаданные объектов. Реализуется `MetadataAdapter`.

```go
type MetadataProvider interface {
    // GetObject возвращает метаданные объекта по SOQL-имени
    GetObject(ctx context.Context, name string) (*ObjectMeta, error)

    // ListObjects возвращает список всех доступных объектов
    ListObjects(ctx context.Context) ([]string, error)

    // InvalidateObject инвалидирует кэш для объекта
    InvalidateObject(ctx context.Context, name string)

    // InvalidateAll инвалидирует весь кэш метаданных
    InvalidateAll(ctx context.Context)
}
```

### AccessController

Проверяет права доступа. Реализуется `SOQLAccessController`.

```go
type AccessController interface {
    // CanAccessObject проверяет доступ к объекту (OLS)
    CanAccessObject(ctx context.Context, object string) error

    // CanAccessField проверяет доступ к полю объекта (FLS)
    CanAccessField(ctx context.Context, object, field string) error
}
```

> **Примечание:** RLS реализован на уровне PostgreSQL через policies и `app.user_id`.

**SOQLAccessController** использует:
- `OLSRepository` — проверка Object-Level Security
- `FLSRepository` — проверка Field-Level Security
- `ObjectRepository` — получение ObjectID по API name

Super-админы (роль `SUPER_ADMIN`) обходят проверки OLS/FLS.

### QueryExecutor

Выполняет скомпилированные запросы. Реализуется `PostgresExecutor`.

```go
type QueryExecutor interface {
    Execute(ctx context.Context, query *Query, params *QueryParams) (*QueryResult, error)
}
```

### CacheInvalidator

Управляет инвалидацией кэшей. Реализуется `AggregatedInvalidator`.

```go
type CacheInvalidator interface {
    InvalidateObject(ctx context.Context, objectApiName string)
    InvalidateAll(ctx context.Context)
    ClearQueryCache(ctx context.Context)
    InvalidateQueriesByObject(ctx context.Context, objectApiName string)
}
```

---

## Метаданные

### ObjectMeta

```go
type ObjectMeta struct {
    Name          string                       // SOQL-имя: "Account", "Contact"
    Table         string                       // SQL-таблица: "data.accounts"
    Fields        map[string]*FieldMeta        // поля объекта
    Lookups       map[string]*LookupMeta       // Child-to-Parent связи
    Relationships map[string]*RelationshipMeta // Parent-to-Child связи
}
```

### FieldMeta

```go
type FieldMeta struct {
    Name       string    // SOQL-имя: "FirstName"
    Column     string    // SQL-колонка: "first_name"
    Type       FieldType // тип данных
    Nullable   bool      // допускает NULL
    Filterable bool      // можно использовать в WHERE
    Sortable   bool      // можно использовать в ORDER BY
    Groupable  bool      // можно использовать в GROUP BY
}

type FieldType int
const (
    FieldTypeString FieldType = iota
    FieldTypeInteger
    FieldTypeFloat
    FieldTypeBoolean
    FieldTypeDate
    FieldTypeDateTime
    FieldTypeID
    FieldTypeArray  // для подзапросов
)
```

### LookupMeta (Child-to-Parent)

```go
type LookupMeta struct {
    Name         string // SOQL-имя связи: "Account"
    Field        string // FK-поле: "account_id"
    TargetObject string // целевой объект: "Account"
    TargetField  string // PK-поле: "id"
}
```

### RelationshipMeta (Parent-to-Child)

```go
type RelationshipMeta struct {
    Name        string // SOQL-имя: "Contacts"
    ChildObject string // дочерний объект: "Contact"
    ChildField  string // FK-поле: "account_id"
    ParentField string // PK-поле: "id"
}
```

### ValidatedWhereSubquery

Результат валидации WHERE-подзапроса (semi-join):

```go
type ValidatedWhereSubquery struct {
    AST          *WhereSubquery          // AST-узел подзапроса
    Object       *ObjectMeta             // метаданные объекта подзапроса
    Field        *FieldMeta              // единственное выбранное поле
    ResolvedRefs map[string]*ResolvedRef // разрешённые ссылки в WHERE подзапроса
}
```

**Пример:**
```sql
SELECT Name FROM Account
WHERE Id IN (SELECT AccountId FROM Contact WHERE Status = 'Active')
```

Здесь `ValidatedWhereSubquery` будет содержать:
- `Object`: метаданные объекта `Contact`
- `Field`: метаданные поля `AccountId`
- `ResolvedRefs`: разрешённая ссылка на поле `Status`

---

## Результат компиляции

```go
type CompiledQuery struct {
    SQL          string          // SQL с плейсхолдерами ($1, $2...)
    Params       []any           // статические параметры
    DateParams   []*DateParam    // параметры-даты для runtime resolution
    Shape        *ResultShape    // структура результата
    Pagination   *PaginationInfo // keyset pagination metadata
    Dependencies []string        // объекты, от которых зависит запрос
}
```

---

## Пагинация (Keyset)

Вместо OFFSET используется cursor-based пагинация:

```go
type PaginationInfo struct {
    SortKeys    keyset.SortKeys // поля сортировки
    SortKeySOQL []string        // SOQL-имена полей
    TieBreaker  string          // гарантированно уникальное поле (record_id)
    PageSize    int             // размер страницы
    HasOrderBy  bool            // есть явный ORDER BY
    Object      string          // имя объекта
}
```

**Cursor** — зашифрованный JSON с позицией последней записи:

```json
{
    "v": [123, "2024-01-15"],  // значения ORDER BY полей
    "s": [{"f": "created_at", "d": "desc"}, {"f": "record_id", "d": "desc"}],
    "o": "Account"
}
```

---

## Ошибки

### Типы ошибок

| Тип | Описание |
|-----|----------|
| `ParseError` | Синтаксическая ошибка |
| `ValidationError` | Семантическая ошибка (неизвестное поле, несовместимые типы) |
| `AccessError` | Нет прав доступа |
| `LimitError` | Превышен лимит |
| `ExecutionError` | Ошибка выполнения SQL |

### ValidationErrorCode

```go
const (
    ErrCodeUnknownObject
    ErrCodeUnknownField
    ErrCodeUnknownLookup
    ErrCodeUnknownRelationship
    ErrCodeTypeMismatch
    ErrCodeFieldNotFilterable
    ErrCodeFieldNotSortable
    ErrCodeFieldNotGroupable
    ErrCodeFieldNotAggregatable
    ErrCodeNestedSubqueryNotAllowed
    ErrCodeTooManyLookupLevels
    ErrCodeInvalidExpression
    ErrCodeMissingRequiredClause
    ErrCodeInvalidDateLiteral
    ErrCodeInvalidPagination
    ErrCodeWhereSubquerySingleField    // WHERE subquery должен выбирать одно поле
    ErrCodeWhereSubqueryAggregateField // WHERE subquery не может использовать агрегаты
)
```

Все ошибки содержат `Position` (line, column) для точной локализации.

---

## Ограничения

```go
type Limits struct {
    MaxSelectFields    int // макс. полей в SELECT (0 = без ограничений)
    MaxRecords         int // макс. записей (LIMIT по умолчанию)
    MaxLookupDepth     int // макс. вложенность Child-to-Parent
    MaxSubqueries      int // макс. подзапросов в SELECT
    MaxSubqueryRecords int // макс. записей в подзапросе
    MaxQueryLength     int // макс. длина запроса
}
```

| Параметр | Значение по умолчанию |
|----------|----------------------|
| MaxRecords | 2000 |
| MaxLookupDepth | 5 |
| MaxSubqueries | 20 |
| MaxSubqueryRecords | 200 |

---

## Кэширование

См. [CACHE.md](./CACHE.md) для детальной документации.

### Metadata Cache

- Кэширует `ObjectMeta` по `api_name`
- TTL: 5 минут
- Инвалидируется через outbox worker

### Query Cache

- Кэширует `CompiledQuery` по тексту SOQL
- TTL: 5 минут
- Хранит `Dependencies` для точечной инвалидации
- Инвалидируется через `InvalidateQueriesByObject()`

---

## Статус реализации

### MVP1 (Реализовано)

- [x] **Синтаксис**: SELECT, FROM, WHERE, ORDER BY, LIMIT
- [x] **Операторы**: `=`, `!=`, `<>`, `<`, `>`, `<=`, `>=`, `AND`, `OR`, `NOT`, `IN`, `LIKE`, `IS NULL`
- [x] **Child-to-Parent**: точечная нотация до 5 уровней (`Account.Owner.Name`)
- [x] **Parent-to-Child**: подзапросы в SELECT (`(SELECT Id FROM Contacts)`)
- [x] **WHERE Subqueries (Semi-Join)**: `Id IN (SELECT AccountId FROM Contact WHERE ...)`, `Id NOT IN (SELECT ...)`
- [x] **Арифметика**: `+`, `-`, `*`, `/`, `%` в SELECT и WHERE (`Amount * 0.1`, `Price + Tax`)
- [x] **Конкатенация строк**: `||` оператор (`FirstName || ' ' || LastName`)
- [x] **Агрегаты**: `COUNT()`, `COUNT_DISTINCT()`, `SUM()`, `AVG()`, `MIN()`, `MAX()`
- [x] **Скалярные функции**:
  - Строковые: `COALESCE`, `NULLIF`, `CONCAT`, `UPPER`, `LOWER`, `TRIM`, `LENGTH`/`LEN`, `SUBSTRING`/`SUBSTR`
  - Математические: `ABS`, `ROUND`, `FLOOR`, `CEIL`/`CEILING`
- [x] **Вложенные функции**: `UPPER(TRIM(Name))`, `COALESCE(TRIM(Name), 'N/A')`
- [x] **GROUP BY**, **HAVING**
- [x] **ORDER BY**: ASC/DESC, `NULLS FIRST/LAST`
- [x] **Статические литералы дат**: TODAY, YESTERDAY, TOMORROW, THIS_WEEK, LAST_WEEK, NEXT_WEEK, THIS_MONTH, LAST_MONTH, NEXT_MONTH, THIS_QUARTER, LAST_QUARTER, NEXT_QUARTER, THIS_YEAR, LAST_YEAR, NEXT_YEAR, LAST_90_DAYS, NEXT_90_DAYS
- [x] **Динамические литералы**: LAST_N_DAYS:n, NEXT_N_DAYS:n, LAST_N_WEEKS:n, NEXT_N_WEEKS:n, LAST_N_MONTHS:n, NEXT_N_MONTHS:n, LAST_N_QUARTERS:n, NEXT_N_QUARTERS:n, LAST_N_YEARS:n, NEXT_N_YEARS:n
- [x] **Типизированные ошибки** с позициями
- [x] **Keyset пагинация** (cursor-based)
- [x] **Кэширование** метаданных и запросов
- [x] **Инвалидация кэша** через PostgreSQL outbox
- [x] **Access Control**: OLS/FLS/RLS интеграция

### Access Control (Реализовано)

| Уровень | Реализация | Описание |
|---------|-----------|----------|
| **OLS** | `SOQLAccessController` | Проверка прав на объект через `OLSRepository.CheckUserPermissionByProfile()` |
| **FLS** | `SOQLAccessController` | Проверка прав на поле через `FLSRepository.GetReadableFields()` |
| **RLS** | PostgreSQL + RLSGuard | `app.user_id` устанавливается через pgxpool BeforeAcquire (stdlib wrapper), RLS policies в PostgreSQL |

**Архитектура Access Control:**

```
┌─────────────────────────────────────────────────────────────────┐
│                        SOQL Engine                               │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                      Validator                               ││
│  │  ┌───────────────────────────────────────────────────────┐  ││
│  │  │              SOQLAccessController                      │  ││
│  │  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────────┐  │  ││
│  │  │  │ OLSRepository│ │ FLSRepository│ │ ObjectRepository│  │  ││
│  │  │  └─────────────┘ └─────────────┘ └─────────────────┘  │  ││
│  │  └───────────────────────────────────────────────────────┘  ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     PostgresExecutor                             │
│  Uses stdlib.OpenDBFromPool → pgxpool → BeforeAcquire(RLSGuard) │
│  RLSGuard sets: SET app.user_id = '<user_id>'                   │
│  PostgreSQL RLS policies filter rows automatically              │
└─────────────────────────────────────────────────────────────────┘
```

### MVP2 (Планируется)

- [ ] **Дополнительные функции**: FORMAT, CALENDAR_YEAR, CALENDAR_MONTH, etc.
- [ ] **Lookups в WHERE subquery SELECT**: `Id IN (SELECT Account.OwnerId FROM Contact)`

### Не реализуем

- `SELECT *` — необходимо явно перечислять поля
- `JOIN` — только Relationship Queries
- `UNION`, `INTERSECT`, `EXCEPT`
- `TYPEOF` для полиморфных полей
- Вложенные подзапросы (подзапрос в подзапросе)
- OFFSET пагинация (только keyset)

---

## Пример использования

```go
// Создание сервиса
service := soqlService.NewQueryService(engine, executor)

// Выполнение запроса
result, err := service.Execute(ctx, &soqlModel.Query{
    SOQL: `
        SELECT Name, Email, Account.Name,
               (SELECT Id, Subject FROM Tasks ORDER BY CreatedDate DESC LIMIT 5)
        FROM Contact
        WHERE CreatedDate = LAST_N_DAYS:30
          AND Account.Type = 'Customer'
        ORDER BY Name ASC
        LIMIT 100
    `,
})

if err != nil {
    var parseErr *engine.ParseError
    var validErr *engine.ValidationError

    switch {
    case errors.As(err, &parseErr):
        log.Printf("Syntax error at line %d: %s", parseErr.Pos.Line, parseErr.Message)
    case errors.As(err, &validErr):
        log.Printf("Validation error [%s]: %s", validErr.Code, validErr.Message)
    }
    return
}

// Использование результата
for _, record := range result.Records {
    fmt.Printf("Name: %s, Email: %s\n", record["Name"], record["Email"])

    if tasks, ok := record["Tasks"].([]any); ok {
        for _, task := range tasks {
            fmt.Printf("  Task: %v\n", task)
        }
    }
}

// Следующая страница
if result.NextCursor != "" {
    nextResult, _ := service.Execute(ctx, &soqlModel.Query{
        SOQL:   "SELECT Name FROM Contact",
        Cursor: result.NextCursor,
    })
}
```

---

## Связанная документация

- [MANUAL.md](./MANUAL.md) — спецификация SOQL
- [CACHE.md](./CACHE.md) — система кэширования
