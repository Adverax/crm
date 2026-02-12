# DML: Архитектура реализации

## Введение

Этот документ описывает архитектуру реализации языка DML (Data Manipulation Language) для нашей CRM-системы. Реализация предоставляет унифицированный интерфейс для операций INSERT, UPDATE, DELETE и UPSERT, следуя принципам Clean Architecture.

### Цели

1. **Независимость от реализации** — модуль DML не зависит от конкретной БД или системы метаданных
2. **Безопасность** — встроенный контроль доступа на уровне объектов и полей (OLS/FLS)
3. **Типобезопасность** — проверка типов на этапе валидации, до выполнения SQL
4. **Защита от ошибок** — обязательный WHERE для DELETE, лимиты batch-операций
5. **Параметризованные запросы** — защита от SQL-инъекций

### Ключевые решения

| Вопрос | Решение |
|--------|---------|
| Формат синтаксиса | SQL-подобный, но упрощённый |
| UPSERT | PostgreSQL `ON CONFLICT` с внешним ID |
| Batch INSERT | Multi-row VALUES синтаксис |
| Безопасность DELETE | Обязательный WHERE по умолчанию |
| Параметризация | Все значения через плейсхолдеры ($1, $2...) |
| Именование | Маппинг через метаданные (DML CamelCase → SQL snake_case) |

---

## Структура модуля (Clean Architecture)

```
internal/data/dml/
├── application/                 # Application layer
│   └── engine/                  # DML Engine (core)
│       ├── ast.go               # AST structures (DMLStatement, Insert, Update, Delete, Upsert, Expr, FuncCall)
│       ├── lexer.go             # Lexer (tokens, включая ключевые слова функций)
│       ├── parser.go            # Parser (participle v2)
│       ├── primitives.go        # FieldType, Operator, Const, Function
│       ├── validator.go         # AST validation & semantic analysis (включая validateExpr, validateFuncCall)
│       ├── compiler.go          # AST → SQL compilation (включая compileExpr, compileFuncCall)
│       ├── deps.go              # Dependencies (MetadataProvider, WriteAccessController, etc.)
│       ├── limits.go            # Operation limits configuration
│       ├── errors.go            # Typed errors (Parse, Validation, Access, etc.)
│       └── engine.go            # Engine: public API
│
├── infrastructure/              # Infrastructure layer (TODO)
│   ├── access/                  # Access control implementation
│   └── metadata/                # Metadata provider implementation
│
└── interface/                   # Interface layer (TODO)
    └── http/                    # HTTP handlers
```

---

## Pipeline обработки DML

```
DML Statement String
       │
       ▼
┌──────────────────┐
│     Parser       │ → AST (DMLStatement)
└────────┬─────────┘
         │
         ▼
   ┌───────────┐      ┌─────────────────────────┐
   │ Validator │─────▶│ WriteAccessController   │ (OLS/FLS проверки)
   └───────────┘      └─────────────────────────┘
       │              ┌────────────────────┐
       ├─────────────▶│ MetadataProvider   │
       │              └────────────────────┘
       ▼
   ┌──────────┐
   │ Compiler │ → CompiledDML {SQL, Params, Operation, Object}
   └──────────┘
       │
       ▼
   ┌───────────────┐
   │   Executor    │ → Result {RowsAffected, InsertedIds, UpdatedIds, DeletedIds}
   └───────────────┘
```

---

## Интерфейсы

### MetadataProvider

Предоставляет метаданные объектов.

```go
type MetadataProvider interface {
    // GetObject возвращает метаданные объекта по DML-имени
    GetObject(ctx context.Context, name string) (*ObjectMeta, error)
}
```

### WriteAccessController

Проверяет права записи. Реализует OLS и FLS для операций модификации.

```go
type WriteAccessController interface {
    // CanWriteObject проверяет доступ к объекту (OLS)
    CanWriteObject(ctx context.Context, object string, op Operation) error

    // CheckWritableFields проверяет доступ к полям (FLS)
    CheckWritableFields(ctx context.Context, object string, fields []string) error
}
```

**Встроенные реализации:**
- `NoopWriteAccessController` — разрешает всё (для тестов)
- `DenyAllWriteAccessController` — запрещает всё (для тестов)
- `FuncWriteAccessController` — обёртка для функций

### Executor

Выполняет скомпилированные DML операции.

```go
type Executor interface {
    Execute(ctx context.Context, compiled *CompiledDML) (*Result, error)
}
```

**Result:**
```go
type Result struct {
    RowsAffected int64    // Количество затронутых строк
    InsertedIds  []string // ID вставленных записей (INSERT/UPSERT)
    UpdatedIds   []string // ID обновлённых записей (UPDATE)
    DeletedIds   []string // ID удалённых записей (DELETE)
}
```

---

## Метаданные

### ObjectMeta

```go
type ObjectMeta struct {
    Name       string                // DML-имя: "Account", "Contact"
    SchemaName string                // SQL-схема: "public", "data"
    TableName  string                // SQL-таблица (без схемы): "accounts"
    Fields     map[string]*FieldMeta // Поля объекта
    PrimaryKey string                // Имя первичного ключа (обычно "record_id")
}
```

**Методы:**
- `Table()` — возвращает полное имя таблицы (schema.table)
- `GetField(name)` — получить метаданные поля
- `GetFieldByColumn(column)` — найти поле по SQL-колонке
- `GetWritableFields()` — все записываемые поля
- `GetRequiredFields()` — обязательные поля (для INSERT)

### FieldMeta

```go
type FieldMeta struct {
    Name         string    // DML-имя: "FirstName"
    Column       string    // SQL-колонка: "first_name"
    Type         FieldType // Тип данных
    Nullable     bool      // Допускает NULL
    Required     bool      // Обязательно при INSERT
    ReadOnly     bool      // Только для чтения (системные поля)
    Calculated   bool      // Вычисляемое поле
    HasDefault   bool      // Имеет значение по умолчанию
    IsExternalId bool      // Можно использовать для UPSERT
    IsUnique     bool      // Уникальный индекс
}
```

### FieldType

```go
const (
    FieldTypeUnknown FieldType = iota
    FieldTypeNull
    FieldTypeString
    FieldTypeInteger
    FieldTypeFloat
    FieldTypeBoolean
    FieldTypeDate
    FieldTypeDateTime
    FieldTypeID
)
```

### Function

Скалярные функции для обработки значений в INSERT/UPDATE/UPSERT.

```go
const (
    FuncCoalesce  Function = iota // COALESCE(val1, val2, ...)
    FuncNullif                    // NULLIF(val1, val2)
    FuncConcat                    // CONCAT(str1, str2, ...)
    FuncUpper                     // UPPER(str)
    FuncLower                     // LOWER(str)
    FuncTrim                      // TRIM(str)
    FuncLength                    // LENGTH(str) / LEN(str)
    FuncSubstring                 // SUBSTRING(str, start, len) / SUBSTR
    FuncAbs                       // ABS(num)
    FuncRound                     // ROUND(num)
    FuncFloor                     // FLOOR(num)
    FuncCeil                      // CEIL(num) / CEILING(num)
)
```

**Методы Function:**
- `MinArgs()` — минимальное количество аргументов
- `MaxArgs()` — максимальное количество аргументов (-1 = неограничено)
- `ResultType(argTypes)` — тип результата функции

### Expr

Выражение значения (константа, функция или ссылка на поле).

```go
type Expr struct {
    FuncCall  *FuncCall // Вызов функции: UPPER('test')
    Const     *Const    // Константа: 'test', 42, NULL
    Field     *Field    // Ссылка на поле: FirstName
    FieldType FieldType // Вычисленный тип
}

type FuncCall struct {
    Name Function // Имя функции
    Args []*Expr  // Аргументы (могут быть вложенными)
}
```

---

## Результат компиляции

```go
type CompiledDML struct {
    SQL             string    // SQL с плейсхолдерами ($1, $2...)
    Params          []any     // Параметры в порядке следования
    Operation       Operation // INSERT, UPDATE, DELETE, UPSERT
    Object          string    // Имя объекта
    Table           string    // Полное имя таблицы
    RowCount        int       // Количество строк (для INSERT/UPSERT)
    ReturningColumn string    // Колонка для RETURNING (обычно record_id)
}
```

---

## Ошибки

### Типы ошибок

| Тип | Описание |
|-----|----------|
| `ParseError` | Синтаксическая ошибка |
| `ValidationError` | Семантическая ошибка (неизвестное поле, несовместимые типы) |
| `AccessError` | Нет прав доступа (OLS/FLS) |
| `LimitError` | Превышен лимит |
| `ExecutionError` | Ошибка выполнения SQL |

### ValidationErrorCode

```go
const (
    ErrCodeUnknownObject      // Неизвестный объект
    ErrCodeUnknownField       // Неизвестное поле
    ErrCodeTypeMismatch       // Несовместимые типы
    ErrCodeReadOnlyField      // Попытка записи в read-only поле
    ErrCodeMissingRequired    // Отсутствует обязательное поле
    ErrCodeInvalidExpression  // Некорректное выражение
    ErrCodeInvalidValue       // Некорректное значение
    ErrCodeExternalIdNotFound // Внешний ID не найден или не помечен
    ErrCodeExternalIdNotUnique // Внешний ID не уникален
    ErrCodeDeleteRequiresWhere // DELETE без WHERE
    ErrCodeDuplicateField     // Дублирующееся поле в списке
)
```

Все ошибки содержат `Position` (line, column) для точной локализации.

### Проверка типов ошибок

```go
if engine.IsParseError(err) { ... }
if engine.IsValidationError(err) { ... }
if engine.IsAccessError(err) { ... }
if engine.IsLimitError(err) { ... }
if engine.IsExecutionError(err) { ... }
```

---

## Ограничения

```go
type Limits struct {
    MaxBatchSize         int  // Макс. строк в INSERT/UPSERT (default: 10000)
    MaxFieldsPerRow      int  // Макс. полей на строку (default: без ограничения)
    MaxStatementLength   int  // Макс. длина запроса в символах (default: 100000)
    RequireWhereOnDelete bool // Требовать WHERE в DELETE (default: true)
    RequireWhereOnUpdate bool // Требовать WHERE в UPDATE (default: false)
}
```

**Предопределённые конфигурации:**

| Конфигурация | Описание |
|--------------|----------|
| `DefaultLimits` | Стандартные лимиты |
| `StrictLimits` | Строгие лимиты для production API |
| `NoLimits` | Без ограничений (для тестов) |

---

## Статус реализации

### MVP (Реализовано)

- [x] **INSERT**: `INSERT INTO Object (Field1, Field2) VALUES (val1, val2), (val3, val4)`
- [x] **UPDATE**: `UPDATE Object SET Field1 = val1, Field2 = val2 WHERE condition`
- [x] **DELETE**: `DELETE FROM Object WHERE condition`
- [x] **UPSERT**: `UPSERT Object (Field1, Field2) VALUES (val1, val2) ON ExternalIdField`
- [x] **WHERE**: операторы сравнения, AND, OR, NOT, IN, LIKE, IS NULL
- [x] **Типы данных**: String, Integer, Float, Boolean, Date, DateTime, NULL
- [x] **Функции**: UPPER, LOWER, TRIM, CONCAT, LENGTH, SUBSTRING, ABS, ROUND, FLOOR, CEIL, COALESCE, NULLIF
- [x] **Вложенные функции**: `UPPER(TRIM(' test '))`
- [x] **Типизированные ошибки** с позициями
- [x] **Access Control**: OLS/FLS интеграция
- [x] **Лимиты**: batch size, statement length, require WHERE

### Планируется

- [ ] Интеграция с существующей системой метаданных
- [ ] HTTP API endpoint
- [ ] Audit logging
- [ ] Транзакции (batch operations)
- [ ] Функции в WHERE условиях

### Не реализуем

- Вложенные подзапросы в WHERE
- Множественные таблицы (JOIN)
- RETURNING для произвольных полей (только PK)

---

## Пример использования

```go
// Создание Engine с зависимостями
engine := engine.NewEngineFromDependencies(&engine.Dependencies{
    MetadataProvider:      metadataAdapter,
    WriteAccessController: accessController,
    Executor:              dbExecutor,
})

// Выполнение INSERT
result, err := engine.Execute(ctx, `
    INSERT INTO Account (Name, Industry, AnnualRevenue)
    VALUES ('Acme Corp', 'Technology', 1000000),
           ('Globex Inc', 'Finance', 2500000)
`)
if err != nil {
    var parseErr *engine.ParseError
    var validErr *engine.ValidationError
    var accessErr *engine.AccessError

    switch {
    case errors.As(err, &parseErr):
        log.Printf("Syntax error at %s: %s", parseErr.Pos, parseErr.Message)
    case errors.As(err, &validErr):
        log.Printf("Validation error [%s]: %s", validErr.Code, validErr.Message)
    case errors.As(err, &accessErr):
        log.Printf("Access denied: %s", accessErr.Message)
    }
    return
}

fmt.Printf("Inserted %d records: %v\n", result.RowsAffected, result.InsertedIds)

// Выполнение UPDATE
result, err = engine.Execute(ctx, `
    UPDATE Account
    SET Industry = 'Software', UpdatedAt = 2024-01-15T10:30:00Z
    WHERE Name = 'Acme Corp'
`)

// Выполнение UPSERT
result, err = engine.Execute(ctx, `
    UPSERT Account (ExternalId, Name, Industry)
    VALUES ('ext-001', 'Acme Corp', 'Technology'),
           ('ext-002', 'Globex Inc', 'Finance')
    ON ExternalId
`)

// Выполнение DELETE
result, err = engine.Execute(ctx, `
    DELETE FROM Task
    WHERE Status = 'Completed' AND CreatedDate < 2023-01-01
`)
```

---

## Fluent API

```go
compiled, err := engine.Statement(`
    INSERT INTO Contact (FirstName, LastName, Email)
    VALUES ('John', 'Doe', 'john@example.com')
`).WithContext(ctx).Prepare()

// Или сразу выполнить
result, err := engine.Statement(`
    UPDATE Contact SET Status = 'Active' WHERE Email = 'john@example.com'
`).WithContext(ctx).Execute()
```

---

## Связанная документация

- [MANUAL.md](./MANUAL.md) — руководство по синтаксису DML
