# ADR-0024: Procedure Engine — декларативный DSL для бизнес-логики

**Статус:** Принято
**Дата:** 2026-02-15
**Участники:** @roman_myakotin

## Контекст

### Проблема: бизнес-логика требует разработчиков

Платформа metadata-driven (ADR-0003, ADR-0007): объекты, поля, валидация, defaults определяются декларативно. Security (OLS/FLS/RLS — ADR-0009..0012) и DML Pipeline (ADR-0020) обеспечивают безопасный доступ к данным. Однако **процедурная бизнес-логика** — цепочки действий, интеграции, условные уведомления — по-прежнему требует Go-кода:

| Что есть (декларативно) | Чего нет |
|---|---|
| Валидация полей (Validation Rules, ADR-0019) | Цепочка операций: создать запись → отправить email → вызвать API |
| Динамические дефолты (Default Expressions) | Условная логика: если сумма > 10K → запрос одобрения |
| Кнопки на карточке (Object View, ADR-0022) | Компенсационные действия: откат при ошибке (Saga) |
| Типы Action (ADR-0023: navigate, field_update) | Синхронные и асинхронные процессы |

Каждое изменение бизнес-процесса (новое уведомление, интеграция с внешней системой, условие маршрутизации) требует цикла разработки: код → review → deploy. Типовые изменения занимают 2-5 дней вместо 15 минут.

### Терминология (ADR-0023)

Данный ADR использует терминологию, установленную в ADR-0023:

| Термин | Определение |
|--------|-------------|
| **Procedure** | Именованный набор Commands, описанный декларативно (JSON + CEL). Выполняется синхронно |
| **Command** | Атомарная операция внутри Procedure: `record.create`, `notification.email`, `integration.http` |
| **Command Type** | Категория command: `record.*`, `notification.*`, `integration.*`, `compute.*`, `flow.*`, `wait.*` |
| **Action** | Зонтичный термин; Procedure — один из action types (`type: procedure`) |

### Индустриальный контекст

| Платформа | Декларативная логика | Expression Language | Sandbox | UI |
|-----------|---------------------|-----------------------|---------|-----|
| **Salesforce** | Flow Builder (visual) + Apex (code) | Formula | Partial (Apex limits) | Drag-and-drop builder |
| **Dynamics 365** | Power Automate (low-code) | Power Fx | Да | Visual + form builder |
| **ServiceNow** | Workflow Designer | JavaScript (scoped) | Partial | Form-based constructor |
| **Zapier/n8n** | Visual workflow | JS / expressions | Нет | Visual node editor |

Все платформы предоставляют визуальный интерфейс для построения бизнес-логики. Администраторы не пишут код — они собирают процедуры из готовых блоков через UI. Текстовый формат (JSON/YAML) — внутреннее представление, а не пользовательский интерфейс.

### Требования

1. Администраторы создают и изменяют процедурную логику через **визуальный конструктор** без знания DSL
2. Безопасность by design: нет циклов, произвольного кода, файлового I/O
3. CEL как единый expression language (уже используется в Phase 7b для validation rules и defaults)
4. **JSON** как формат хранения — нативная поддержка Go, PostgreSQL (JSONB), TypeScript
5. Компенсационные действия (Saga pattern) для распределённых операций
6. Вызов одних процедур из других с контролем глубины
7. Полная observability: логирование, метрики, tracing каждого command
8. Dry-run и декларативные тесты без побочных эффектов

## Рассмотренные варианты

### Вариант A — Императивный Go-код per business rule (status quo)

Каждое бизнес-правило реализуется как Go-функция в service layer. Новые правила добавляются через цикл разработки.

**Плюсы:**
- Полный контроль и максимальная производительность
- Compile-time type safety
- Знакомый подход для Go-разработчиков

**Минусы:**
- Каждое изменение требует разработчика (2-5 дней вместо 15 минут)
- Администраторы полностью зависят от команды разработки
- Нет стандартизации: каждый разработчик пишет по-своему
- Сложность аудита: нужно читать Go-код, чтобы понять бизнес-логику

### Вариант B — Встроенный скриптовый движок (JavaScript/Starlark sandbox)

Предоставить администраторам sandbox-среду для написания скриптов на знакомом языке.

**Плюсы:**
- Максимальная гибкость выражений
- JavaScript знаком многим администраторам
- Starlark (Google) — sandbox by design, детерминистичный

**Минусы:**
- Два expression language: CEL (validation, defaults) + JS/Starlark (procedures)
- JavaScript sandbox (V8/QuickJS) тяжёлый, сложно ограничить полностью
- Скрипты сложнее анализировать статически, чем JSON-структуры
- Безопасность: даже в sandbox возможны timing attacks, resource exhaustion

### Вариант C — JSON DSL + Constructor UI + CEL (выбран)

Процедурная логика хранится как JSON (JSONB в PostgreSQL). Администратор собирает процедуру через **form-based Constructor** — выбирает command type из списка, заполняет параметры, строит выражения через Expression Builder. JSON генерируется конструктором автоматически. Power users могут редактировать JSON напрямую.

**Плюсы:**
- **Constructor-first**: нулевой порог входа для администраторов — не нужно знать JSON/DSL
- **JSON нативен** для всего стека: `encoding/json` (Go), JSONB (PostgreSQL), TypeScript — без дополнительных парсеров
- **JSONB в PostgreSQL**: индексация, GIN-индексы, jsonpath-запросы, partial updates
- Переиспользование CEL — единый expression language от UI до backend
- Безопасность by design: фиксированный набор command types, нет циклов
- Статический анализ: JSON-schema позволяет валидировать procedure при сохранении
- Расширяемость: новые command types добавляются в Go, сразу доступны в Constructor
- Dry-run и декларативные тесты из коробки

**Минусы:**
- Ограниченная гибкость: невозможно выразить произвольный алгоритм
- JSON менее читаем для человека, чем YAML — но администратор работает через Constructor, а не читает JSON
- Для нестандартных сценариев нужен новый command type в Go

### Вариант D — Visual Flow Builder (drag-and-drop граф)

Графический редактор потоков, аналог Salesforce Flow Builder или n8n.

**Плюсы:**
- Наглядность: визуальная диаграмма процесса
- Популярный подход в enterprise CRM

**Минусы:**
- Сложность версионирования: визуальная модель хранится как JSON-граф, diff нечитаем
- Merge conflicts практически неразрешимы
- Значительные инвестиции во frontend (canvas editor, node rendering, edge routing)
- Overkill для линейных процедур (80% случаев — последовательные commands)
- Form-based Constructor проще реализовать и покрывает те же кейсы
- Visual Builder может быть добавлен поверх JSON DSL как альтернативный UI в будущем

## Решение

**Выбран вариант C: JSON DSL + Constructor UI + CEL.**

Procedure Engine — runtime для выполнения именованных Procedures, описанных в JSON. Администратор собирает процедуры через **Constructor UI**. JSON — внутреннее представление (IR), хранится как JSONB в PostgreSQL. CEL — единый expression language.

### Архитектура: Constructor → JSON → Engine

```
Администратор → Constructor UI → JSON (JSONB) → Procedure Engine
                                      ↑
Power user → Raw JSON editor ─────────┘
```

Constructor UI — **основной** интерфейс. Raw JSON editor — для power users и отладки. Администратору не нужно знать ни JSON-структуру, ни CEL-синтаксис — конструктор предоставляет формы, dropdown'ы и expression builder.

### Структура Procedure

```json
{
  "name": "create_customer",
  "commands": [
    {
      "type": "record.create",
      "object": "Account",
      "data": {
        "email": "$.input.email",
        "name": "$.input.name"
      },
      "as": "account"
    },
    {
      "type": "notification.email",
      "to": "$.input.email",
      "template": "welcome",
      "data": {
        "name": "$.input.name"
      }
    }
  ],
  "result": {
    "accountId": "$.account.id"
  }
}
```

Procedure состоит из:
- **name** — уникальный идентификатор (snake_case)
- **commands** — упорядоченный список commands
- **result** — CEL-маппинг, формирующий возвращаемое значение

### Constructor UI

#### Procedure Constructor

Form-based интерфейс для построения процедур:

1. **Добавление command**: кнопка "+" → выбор command type из categorized dropdown:
   - Данные: Создать запись, Обновить запись, Удалить запись, Найти записи
   - Уведомления: Email, SMS, Push
   - Интеграции: HTTP запрос, Webhook
   - Логика: Вычисление, Валидация, Ошибка
   - Поток: Вызвать процедуру, Запустить сценарий

2. **Настройка command**: форма с полями, соответствующими command type:
   - Dropdown для выбора объекта (для `record.*`)
   - Field mapping table: поле → значение/выражение
   - CEL expression через Expression Builder (кнопка `fx`)
   - Toggle для `optional`, `as` (сохранить результат)

3. **Условная логика**: toggle "Условие" → Expression Builder для `when`
   - `if/else`: визуальный блок с двумя ветками
   - `match`: список case → команды

4. **Rollback**: toggle "Откат при ошибке" → форма для компенсационного command

5. **Порядок**: drag-and-drop для переупорядочивания commands

6. **Preview**: сгенерированный JSON (read-only) + dry-run

#### Expression Builder (конструктор выражений)

Визуальный конструктор CEL-выражений, применимый во всех подсистемах платформы (Procedure, Validation Rules, Default Expressions, Object View `visibility_expr`, Scenario `when`):

1. **Field picker**: дерево доступных переменных
   - `$.input.*` — входные параметры
   - `$.user.*` — текущий пользователь (id, role, profile)
   - `$.<step>.*` — результаты предыдущих commands
   - `$.now` — текущее время
   - Контекстные поля: `record.*` (для validation rules), `old.*`/`new.*` (для automation)

2. **Operator picker**: типизированные операторы
   - Сравнение: `=`, `!=`, `>`, `<`, `>=`, `<=`
   - Логика: `И`, `ИЛИ`, `НЕ`
   - Строки: `содержит`, `начинается с`, `заканчивается на`
   - Списки: `в списке`, `не в списке`
   - Null: `пусто`, `не пусто`

3. **Function picker**: каталог функций с описаниями
   - Строковые: `UPPER()`, `LOWER()`, `TRIM()`, `CONCAT()`
   - Числовые: `ABS()`, `ROUND()`, `CEIL()`, `FLOOR()`
   - Дата/время: `now()`, `duration()`
   - Коллекции: `size()`, `has()`
   - Каждая функция с описанием, типами параметров и примером

4. **Live preview**: результат выражения на sample data в реальном времени

5. **Валидация**: синтаксическая проверка CEL при вводе, подсветка ошибок

6. **Двойной режим**: визуальный конструктор ↔ текстовый CEL (toggle для power users)

### Command Types

| Command Type | Prefix | Commands | Описание |
|-------------|--------|----------|----------|
| **record** | `record.*` | `create`, `update`, `delete`, `get`, `query` | CRUD через DML Engine с OLS/FLS/RLS |
| **notification** | `notification.*` | `email`, `sms`, `push` | Уведомления через шаблоны |
| **integration** | `integration.*` | `http` | HTTP-вызовы внешних API (method в параметрах) |
| **compute** | `compute.*` | `transform`, `validate`, `aggregate`, `fail` | Вычисления и валидация |
| **flow** | `flow.*` | `call` (procedure), `start` (scenario) | Вызов procedure / запуск scenario |
| **wait** | `wait.*` | `signal`, `timer`, `until` | Ожидание сигнала, паузы, времени |

Все `record.*` commands выполняются через DML Engine (ADR-0020) с полным enforcement OLS/FLS/RLS. Procedure не может обойти security-слои.

### JSON-схема Command

Каждый command — JSON-объект с обязательным полем `type` и параметрами, специфичными для command type:

```json
{
  "type": "record.create",
  "object": "Account",
  "data": {
    "email": "$.input.email",
    "name": "$.input.name"
  },
  "as": "account",
  "optional": false,
  "when": "$.input.createAccount",
  "rollback": {
    "type": "record.delete",
    "object": "Account",
    "id": "$.account.id"
  }
}
```

Общие поля (для всех command types):
- `type` — command type (обязательный)
- `as` — имя переменной для сохранения результата
- `optional` — не прерывать при ошибке (ошибка → `$.warnings`)
- `when` — CEL-условие выполнения
- `rollback` — компенсационный command (Saga)
- `retry` — политика повторов

### Контекст и переменные

Все значения в procedure — CEL expressions (строки, начинающиеся с `$`). Контекст накапливается по мере выполнения commands:

| Переменная | Описание | Пример |
|------------|----------|--------|
| `$.input` | Входные параметры procedure | `$.input.email` |
| `$.user` | Текущий пользователь | `$.user.id`, `$.user.role` |
| `$.now` | Текущее время UTC | `$.now` |
| `$.secrets` | Секреты (API-ключи), только runtime | `$.secrets.stripe_key` |
| `$.<name>` | Результат command с `as: name` | `$.account.id` |
| `$.warnings` | Массив ошибок optional commands | `$.warnings` |
| `$.error` | Текущая ошибка (в rollback-блоке) | `$.error.code` |

Результат command сохраняется через `as` и доступен всем последующим commands до конца procedure.

### Условная логика

Три формы условного выполнения:

```json
// when — условное выполнение одного command
{
  "type": "notification.email",
  "to": "$.input.email",
  "template": "welcome",
  "when": "$.input.sendWelcome"
}

// if/else — ветвление
{
  "type": "flow.if",
  "condition": "$.input.amount > 10000",
  "then": [
    {
      "type": "notification.email",
      "to": "\"manager@company.com\"",
      "template": "approval_needed"
    }
  ],
  "else": [
    {
      "type": "compute.transform",
      "data": { "approved": true },
      "as": "decision"
    }
  ]
}

// match — множественный выбор
{
  "type": "flow.match",
  "expression": "$.input.priority",
  "cases": {
    "critical": [
      { "type": "notification.email", "to": "\"oncall@company.com\"", "template": "critical_alert" }
    ],
    "high": [
      { "type": "notification.email", "to": "\"team@company.com\"", "template": "high_priority" }
    ]
  },
  "default": [
    { "type": "notification.email", "to": "$.input.assignee_email", "template": "standard_notification" }
  ]
}
```

В Constructor UI: `if/else` — визуальный блок с двумя ветками; `match` — список вариантов с кнопкой "Добавить вариант".

### Обработка ошибок

#### Структурированная ошибка

```json
{
  "code": "string",
  "message": "string",
  "details": {},
  "retryable": false,
  "source": "procedure"
}
```

Категории: `validation_*`, `not_found_*`, `permission_*`, `external_*`, `timeout_*`, `limit_*`, `internal_*`.

#### Rollback (Saga pattern)

Компенсационные actions выполняются в порядке LIFO при ошибке в последующих commands:

```json
{
  "commands": [
    {
      "type": "record.create",
      "object": "Order",
      "data": { "customerId": "$.input.customerId" },
      "as": "order",
      "rollback": {
        "type": "record.delete",
        "object": "Order",
        "id": "$.order.id"
      }
    },
    {
      "type": "integration.http",
      "method": "POST",
      "url": "https://payment.com/charge",
      "body": { "amount": "$.input.amount" },
      "as": "payment",
      "rollback": {
        "type": "integration.http",
        "method": "POST",
        "url": "https://payment.com/refund",
        "body": { "paymentId": "$.payment.id" }
      }
    },
    {
      "type": "notification.email",
      "to": "$.input.email",
      "template": "order_confirmed",
      "data": { "orderId": "$.order.id" }
    }
  ]
}
```

Ошибка в email → rollback payment → rollback order (LIFO). Rollback регистрируется только для успешно выполненных commands. В rollback-блоке доступен `$.error`.

#### try/catch, retry

```json
// try/catch
{
  "type": "flow.try",
  "commands": [
    {
      "type": "integration.http",
      "method": "POST",
      "url": "https://api.payment.com/charge",
      "body": { "amount": "$.input.amount" },
      "as": "payment"
    }
  ],
  "catch": [
    {
      "type": "compute.fail",
      "code": "payment_error",
      "message": "$.error.message"
    }
  ]
}

// retry
{
  "type": "integration.http",
  "method": "POST",
  "url": "https://api.flaky.com/data",
  "body": { "payload": "$.input.data" },
  "retry": {
    "attempts": 3,
    "delay": "1s",
    "backoff": 2,
    "on": ["timeout", "5xx"]
  }
}
```

### Семантика call и start

| Аспект | `call` (procedure) | `start` (scenario) |
|--------|-------------------|-------------------|
| Выполнение | Синхронное | Асинхронное (fire-and-forget) |
| Результат | Полный result вызванной procedure | Только `executionId` |
| Ошибки | Всплывают в вызывающую procedure | Не влияют на вызывающую procedure |
| Rollback | Каскадный | Независимый |
| Контекст | `$.user`, `$.secrets` наследуются | `$.user` копируется |
| Глубина | Максимум 3 уровня вложенности | Без ограничений (независимый процесс) |

Защита от циклов: платформа отслеживает стек вызовов и предотвращает циклические зависимости (`proc_a → proc_b → proc_a` = ошибка `circular_procedure_call`).

### Лимиты

| Параметр | Лимит | Описание |
|----------|-------|----------|
| Время выполнения | 30 секунд | Максимальное время на всю procedure (без wait) |
| Количество commands | 50 | Суммарно с вызванными procedures |
| Вложенность call | 3 уровня | procedure → procedure → procedure |
| Вложенность if/match | 5 уровней | Максимальная глубина условий |
| Размер JSON | 64 KB | Максимальный размер определения procedure |
| Размер input | 1 MB | Входные данные |
| Размер context | 10 MB | Накопленный контекст |
| HTTP timeout | 10 секунд | На один HTTP-запрос |
| HTTP запросов | 10 | Максимум в одной procedure |
| Уведомлений | 10 | Максимум email/sms/push в одной procedure |
| Retry attempts | 3 | Максимум попыток |

При превышении лимита procedure завершается с типизированной ошибкой (`limit_exceeded_*`).

### Security sandbox

| Угроза | Защита |
|--------|--------|
| Бесконечные циклы | Циклы отсутствуют в DSL by design |
| Произвольный код | Только фиксированные command types |
| Доступ к файловой системе | Нет I/O-операций |
| Неконтролируемые HTTP | Только через `integration.http`, с лимитами |
| Обход RLS/FLS/OLS | Все `record.*` commands выполняются через DML Engine |
| Утечка секретов | `$.secrets` доступен только в runtime, маскируется в логах |
| Ресурсоёмкие операции | Жёсткие лимиты (таймаут, количество commands, размер данных) |
| Циклические вызовы | Отслеживание стека вызовов, защита от рекурсии |

### Хранение и версионирование

Procedures хранятся в PostgreSQL как **JSONB** (таблица `metadata.procedures`):

| Решение | Обоснование |
|---------|-------------|
| JSONB | Нативный тип PostgreSQL: индексация, jsonpath-запросы, partial updates, валидация |
| БД вместо файлов | Hot-reload без перезапуска; inline procedures в scenario steps; целостность ссылок |
| Snapshot-версионирование | При старте Scenario фиксируются версии всех используемых procedures |

Snapshot-подход гарантирует предсказуемость: нет изменений mid-execution. Новый запуск scenario использует актуальные версии.

### Observability

Каждый command — явная единица работы, логируется автоматически:

| Компонент | Реализация |
|-----------|------------|
| **Logging** | Structured JSON; маскирование sensitive-полей (password, token, secret, key) |
| **Metrics** | Prometheus: `procedure_executions_total`, `procedure_duration_seconds`, `command_executions_total`, `command_duration_seconds` |
| **Tracing** | OpenTelemetry: span per procedure + child spans per command; `trace_id` propagation в HTTP headers |

### Тестирование

| Метод | Описание |
|-------|----------|
| **Dry-run** | Выполнение без побочных эффектов; `record.create` возвращает fake ID, `notification.*` логирует без отправки |
| **Декларативные тесты** | Input, mocks, expected result и expected commands в JSON-файле |
| **Snapshot testing** | Сравнение результата с сохранённым snapshot |

```json
{
  "name": "create_order создаёт заказ и отправляет email",
  "procedure": "create_order",
  "input": {
    "customerId": "cust_123",
    "amount": 1500
  },
  "mocks": {
    "http": [
      { "url": "https://payment.com/charge", "response": { "status": 200, "body": { "paymentId": "pay_789" } } }
    ]
  },
  "expect": {
    "success": true,
    "result": { "orderId": { "type": "uuid" } },
    "commands": [
      { "type": "record.create", "object": "Order" },
      { "type": "notification.email", "template": "order_confirmed" }
    ]
  }
}
```

### Результат выполнения

Procedure возвращает `ProcedureResult`:

```json
// Успех
{
  "success": true,
  "result": { "accountId": "acc_123" },
  "error": null
}

// Ошибка
{
  "success": false,
  "result": null,
  "error": {
    "code": "external_payment_declined",
    "message": "Payment was declined",
    "retryable": false
  }
}
```

### Эволюционный путь

```
Этап 1: Procedure Engine (MVP)
  ├── JSON DSL + Constructor UI + Expression Builder
  └── Базовые command types: record, notification, integration, compute, flow

Этап 2: Расширенные command types
  └── Новые типы по мере потребностей (batch, aggregate, approval)

Этап 3: Visual Flow Builder (опционально)
  └── Drag-and-drop граф поверх JSON для сложных ветвлений

Этап 4: Marketplace commands
  └── Готовые интеграции (Slack, Stripe, 1C, Telegram)
```

## Последствия

### Позитивные

- **Constructor-first** — нулевой порог входа: администратор собирает процедуру через формы и dropdown'ы, не изучая DSL
- **Expression Builder** — единый визуальный конструктор CEL-выражений для всех подсистем платформы (procedures, validation rules, defaults, visibility)
- **JSON нативен для стека** — `encoding/json` (Go), JSONB (PostgreSQL), TypeScript — без дополнительных парсеров и зависимостей
- **JSONB в PostgreSQL** — индексация, jsonpath-запросы, partial updates; возможность анализировать procedures SQL-запросами
- **Безопасность by design** — sandbox без циклов, произвольного кода и файлового I/O; жёсткие лимиты; OLS/FLS/RLS enforcement
- **Тестируемость** — dry-run, декларативные JSON-тесты, snapshot testing
- **Observability** — structured logging, Prometheus metrics, OpenTelemetry tracing для каждого command
- **Saga pattern** — LIFO rollback для компенсационных действий при распределённых операциях
- **Расширяемость** — новые command types добавляются в Go один раз, сразу доступны в Constructor

### Негативные

- **Ограниченная гибкость** — невозможно выразить произвольный алгоритм; для нестандартных сценариев нужен новый command type в Go
- **Constructor — инвестиция во frontend** — form-based UI для каждого command type требует разработки Vue-компонентов
- **Expression Builder — сложность реализации** — визуальный конструктор CEL с live preview, autocomplete, валидацией
- **JSON менее читаем** — для power users, работающих с raw JSON, менее наглядно чем YAML — компенсируется Constructor UI
- **Зависимость от CEL** — cel-go/cel-js становятся критической зависимостью платформы (но уже приняты в ADR-0019)

## Связанные ADR

- **ADR-0019** — Декларативная бизнес-логика: CEL как expression language, validation rules, подсистемы поведенческой логики. Procedure Engine реализует процедурный слой, дополняющий декларативный. Expression Builder переиспользуется во всех подсистемах
- **ADR-0020** — DML Pipeline Extension: все `record.*` commands выполняются через DML Engine с typed stages (defaults → validate → compute → execute)
- **ADR-0022** — Object View: action type `procedure` в конфигурации кнопок вызывает Procedure Engine
- **ADR-0023** — Action terminology: устанавливает терминологию (Procedure, Command, Command Type), используемую в данном ADR
