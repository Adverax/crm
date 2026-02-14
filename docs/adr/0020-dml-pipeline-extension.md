# ADR-0020: Расширение DML Pipeline (typed stages)

**Статус:** Принято
**Дата:** 2026-02-14
**Участники:** @roman_myakotin

## Контекст

DML Engine (Phase 4) реализует 4-stage pipeline:

```
parse → validate → compile → execute
```

Где `validate` выполняет: проверку существования полей, required-fields check, type-compatibility, FLS (writable fields). Этого достаточно для структурной валидации, но недостаточно для поведенческой логики, определённой в ADR-0019:

| Подсистема (ADR-0019) | Что требуется от DML | Текущий статус |
|---|---|---|
| Default Expressions | Инжект недостающих полей до валидации | Отсутствует. `DefaultValue` используется только для boolean DDL; DML не инжектит |
| Validation Rules | CEL-проверки после defaults | Отсутствует. Только required + type |
| Computed Fields (stored) | Пересчёт производных значений до compile | Отсутствует |
| Automation Rules | Реактивная логика после execute | Отсутствует |

Pipeline необходимо расширить для интеграции этих подсистем.

### Требования к расширению

1. **Чёткий порядок выполнения** — каждый stage имеет определённое место в pipeline, нет неоднозначностей
2. **Typed interfaces** — каждый stage = interface с конкретной сигнатурой, не arbitrary callback
3. **Инкрементальность** — stages добавляются по дорожной карте ADR-0019, не все сразу
4. **Effective ruleset** — pipeline принимает контекст вызова (metadata / Object View / Layout) для каскадного слияния правил (ADR-0019)
5. **Тестируемость** — каждый stage тестируется изолированно через mock

## Рассмотренные варианты

### Вариант A — Generic hooks (middleware pattern)

Произвольные callback'и, регистрируемые на события (before insert, after update):

```go
engine.Before("insert", func(ctx context.Context, record Record) error { ... })
engine.After("update", func(ctx context.Context, old, new Record) error { ... })
```

**Плюсы:**
- Максимальная гибкость
- Знакомый паттерн (Express middleware, Gin handlers)

**Минусы:**
- Нет гарантий порядка выполнения между hooks
- Arbitrary code — трудно тестировать, отлаживать
- «Какой hook сломал запись?» — классическая проблема Salesforce (35+ шагов в order of execution)
- Нарушает декларативный принцип платформы
- Hooks могут конфликтовать друг с другом

### Вариант B — Typed pipeline stages (выбран)

Фиксированный набор stages с typed interfaces. Каждый stage отвечает за одну задачу, имеет определённое место в pipeline.

**Плюсы:**
- Чёткий, предсказуемый порядок выполнения
- Каждый stage = typed interface → compile-time safety, легко тестировать
- Декларативные подсистемы (CEL) подключаются через эти interfaces
- Нет проблемы «hook A конфликтует с hook B»
- Простота отладки: каждый stage можно логировать отдельно

**Минусы:**
- Менее гибко, чем arbitrary hooks
- Для нестандартных сценариев нужен custom handler на уровне Automation Rules, не внутри pipeline

### Вариант C — Salesforce-style order of execution

Фиксированный порядок из 10+ шагов с чётким описанием каждого:

**Плюсы:**
- Проверен в production (Salesforce)

**Минусы:**
- Salesforce order of execution — 35+ шагов, печально известная проблема
- Triggers могут вызывать рекурсию (trigger → DML → trigger)
- Чрезмерная сложность для нашей платформы

## Решение

**Выбран вариант B: Typed pipeline stages.**

### Расширенный pipeline

```
┌──────────────────────────────────────────────────────────┐
│                     DML Pipeline                          │
│                                                          │
│  1. PARSE            ← AST из DML-выражения             │
│     Существующий parser (Participle)                     │
│                                                          │
│  2. RESOLVE          ← загрузка метаданных + контекста   │
│     ├─ ObjectMeta + FieldMeta (MetadataProvider)         │
│     └─ Effective ruleset (каскад ADR-0019)               │
│                                                          │
│  3. DEFAULTS         ← инжект недостающих полей          │
│     ├─ static: FieldConfig.default_value                 │
│     └─ dynamic: FieldConfig.default_expr (CEL)           │
│     Только для INSERT. Только для полей, отсутствующих   │
│     в statement. Каскад: metadata → OV → Layout.         │
│                                                          │
│  4. VALIDATE                                             │
│     a) Metadata constraints: required, type, unique      │
│     b) Validation Rules: CEL expressions (AND)           │
│     c) FLS: writable fields check                        │
│     Каскад: metadata_rules AND ov_rules AND layout_rules │
│                                                          │
│  5. COMPUTE          ← пересчёт stored computed fields   │
│     CEL expressions из FieldConfig.formula_expr          │
│     Только для полей с field_type="formula" + stored     │
│     Добавляет вычисленные значения в statement           │
│                                                          │
│  6. COMPILE          ← генерация SQL                     │
│     Существующий compiler (параметризованный SQL)        │
│                                                          │
│  7. EXECUTE          ← pgx                               │
│     Существующий executor + RLS injection                │
│                                                          │
│  8. POST-EXECUTE     ← реактивная логика (будущее)       │
│     Automation Rules: trigger conditions (CEL)           │
│     → Handler (sync) / Scenario (async)                  │
│     Выполняется после успешного execute, до commit       │
│     или после commit (в зависимости от типа action)      │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

### Stages и операции

Не все stages применяются ко всем DML-операциям:

| Stage | INSERT | UPDATE | DELETE | UPSERT |
|---|---|---|---|---|
| 1. Parse | Да | Да | Да | Да |
| 2. Resolve | Да | Да | Да | Да |
| 3. Defaults | Да | Условно (default_on=update) | Нет | Да (insert-часть) |
| 4a. Metadata validate | Да | Да | Нет | Да |
| 4b. Validation Rules | Да | Да | Условно | Да |
| 4c. FLS | Да | Да | Да | Да |
| 5. Compute | Да | Да | Нет | Да |
| 6. Compile | Да | Да | Да | Да |
| 7. Execute | Да | Да | Да | Да |
| 8. Post-execute | Да | Да | Да | Да |

### Typed interfaces

Каждый новый stage — interface, подключаемый через Option pattern (как существующие `WithMetadata`, `WithExecutor`):

```go
// Stage 3: Default injection
type DefaultResolver interface {
    ResolveDefaults(ctx context.Context, object string, operation Operation, fields map[string]Value) (map[string]Value, error)
}

// Stage 4b: Validation Rules
type RuleValidator interface {
    ValidateRules(ctx context.Context, object string, operation Operation, record, old map[string]Value) []ValidationError
}

// Stage 5: Computed fields
type ComputeEngine interface {
    ComputeFields(ctx context.Context, object string, record map[string]Value) (map[string]Value, error)
}

// Stage 8: Post-execute reactions
type PostExecutor interface {
    AfterExecute(ctx context.Context, object string, operation Operation, result *Result) error
}
```

Подключение через Options:

```go
engine := dml.NewEngine(
    dml.WithMetadata(metadataAdapter),
    dml.WithWriteAccessController(flsEnforcer),
    dml.WithExecutor(rlsExecutor),
    // Новые stages:
    dml.WithDefaultResolver(celDefaultResolver),
    dml.WithRuleValidator(celRuleValidator),
    dml.WithComputeEngine(celComputeEngine),
    dml.WithPostExecutor(automationDispatcher),
)
```

Каждый stage опционален. Если interface не предоставлен, stage пропускается. Это обеспечивает инкрементальное добавление по дорожной карте ADR-0019.

### Execution context

DML Engine принимает контекст вызова, определяющий effective ruleset (каскад из ADR-0019):

```go
type ExecutionContext struct {
    ObjectViewID *uuid.UUID  // nil = raw DML (только metadata rules)
    LayoutID     *uuid.UUID  // nil = без layout-level rules
}
```

Resolve stage использует контекст для каскадного слияния:

```
// Validation: additive (AND)
effective_rules = metadata_rules
if ctx.ObjectViewID != nil {
    effective_rules = append(effective_rules, ov_rules...)
}
if ctx.LayoutID != nil {
    effective_rules = append(effective_rules, layout_rules...)
}

// Defaults: replace (последний побеждает)
effective_defaults = metadata_defaults
if ctx.ObjectViewID != nil {
    effective_defaults = merge(effective_defaults, ov_defaults)
}
if ctx.LayoutID != nil {
    effective_defaults = merge(effective_defaults, layout_defaults)
}
```

При вызове без контекста (raw DML, import, integration) применяются только metadata-level правила — минимальный гарантированный уровень защиты.

### Validation Rules: переменные в CEL-среде

| Переменная | Тип | Доступность | Описание |
|---|---|---|---|
| `record` | map | INSERT, UPDATE, UPSERT | Текущие значения полей (после defaults) |
| `old` | map | UPDATE | Предыдущие значения (до изменения) |
| `user` | map | Всегда | Текущий пользователь (`id`, `profile_id`, `role_id`) |
| `now` | timestamp | Всегда | Текущее время UTC |

Для INSERT: `old` = nil. Validation rules с `old` в выражении автоматически пропускаются при INSERT.

### Default Expressions: порядок применения

1. Определяются поля, отсутствующие в DML-statement
2. Для каждого отсутствующего поля проверяется наличие default:
   - Сначала `default_value` (статический) — приоритет ниже
   - Затем `default_expr` (CEL) — перекрывает статический
   - Каскад: Layout > Object View > Metadata
3. Если `default_on` не соответствует текущей операции — пропуск
4. CEL-выражение вычисляется с переменными `record`, `user`, `now`
5. Результат добавляется в `record` до этапа validate

### Ошибки

Каждый stage возвращает типизированные ошибки:

| Stage | Код ошибки | HTTP | Описание |
|---|---|---|---|
| Defaults | `default_eval_error` | 500 | Ошибка вычисления CEL-выражения default |
| Validation (metadata) | `missing_required_field` | 400 | Отсутствует обязательное поле |
| Validation (metadata) | `type_mismatch` | 400 | Несовместимый тип значения |
| Validation (rules) | `validation_rule_failed` | 400 | Не пройдена CEL-валидация (code из rule) |
| Validation (rules) | `rule_eval_error` | 500 | Ошибка вычисления CEL-выражения rule |
| Compute | `compute_eval_error` | 500 | Ошибка вычисления computed field |
| Post-execute | `automation_error` | 500 | Ошибка в automation rule |

Validation rules с `severity: warning` НЕ блокируют выполнение — собираются в `Result.Warnings`.

### Нерекурсивность

Automation Rules (post-execute) могут выполнять DML-операции над другими объектами. Для предотвращения рекурсии:

- Automation Rules НЕ могут модифицировать объект, вызвавший trigger
- Максимальная глубина вложенности DML-вызовов из automation: 2 (аналог Salesforce trigger depth limit)
- DML-вызовы из automation выполняются с `ExecutionContext = nil` (только metadata rules)

## Дорожная карта

Stages добавляются инкрементально, в соответствии с ADR-0019:

| Phase | Добавляемые stages | Зависимости |
|---|---|---|
| **7a** | 3. Defaults (только static `default_value`) | — |
| **7b** | 3. Defaults (dynamic `default_expr`) + 4b. Validation Rules | CEL engine (cel-go) |
| **N+1** | 5. Compute | CEL engine (уже есть после 7b) |
| **N+2** | 2. Resolve (каскад) + 8. Post-execute | Object View storage, Automation Rules |

**Phase 7a** — static defaults: инжект `FieldConfig.default_value` для отсутствующих полей + системные поля (`owner_id`, `created_by_id`, `created_at`, `updated_at`). Без CEL. Pipeline расширяется Stage 3 в минимальном варианте.

**Phase 7b** — CEL engine: `cel-go` интеграция, `default_expr`, таблица `metadata.validation_rules`, Stage 4b. Validation Rules и dynamic defaults используют общий CEL runtime.

## Последствия

### Позитивные

- Предсказуемый порядок выполнения — нет «какой hook сломал запись?»
- Каждый stage тестируется изолированно через interface mock
- Опциональность stages через Option pattern — инкрементальное добавление
- Нерекурсивность automation гарантирована (depth limit)
- Декларативный подход сохранён (CEL, не arbitrary Go-код в pipeline)
- Ошибки типизированы и привязаны к конкретному stage

### Негативные

- Менее гибко, чем generic hooks — edge cases решаются через Automation Rules (Go handler), а не через pipeline injection
- Фиксированный порядок stages — нельзя вставить кастомный stage «между validate и compute»
- CEL dependency для stages 3, 4b, 5 (но это решение из ADR-0019)

### Связанные ADR

- ADR-0019 — Декларативная бизнес-логика (определяет подсистемы и каскад; данный ADR определяет точки интеграции в DML)
- ADR-0004 — Field type/subtype (расширяется formula type для stage 5)
- ADR-0009 — Security architecture (FLS остаётся в stage 4c, не смешивается с validation rules)
