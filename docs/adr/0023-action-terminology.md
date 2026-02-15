# ADR-0023: Терминология исполняемой логики — Action, Command, Procedure, Scenario, Function

**Статус:** Принято
**Дата:** 2026-02-15
**Участники:** @roman_myakotin

## Контекст

### Проблема: терминологический конфликт

Платформа развивает несколько подсистем, описывающих реакцию на действия пользователя:

1. **Object View** (ADR-0022) — кнопки на карточке записи ("Отгрузить", "Отправить предложение")
2. **Procedure Engine** (ADR-0024) — декларативный YAML DSL для бизнес-логики с CEL expressions
3. **Scenario Engine** (ADR-0025) — оркестрация долгоживущих бизнес-процессов с durability
4. **DML Pipeline** (ADR-0020) — стадии обработки данных при записи
5. **Automation Rules** (ADR-0019) — реактивная логика "когда X → сделай Y"

Между ними возникли терминологические конфликты:

| Термин | Где используется | Значение 1 | Значение 2 |
|--------|-----------------|------------|------------|
| **Action** | handler.md | Единица работы (`record.create`, `email send`) | — |
| **Action** | ADR-0022 | UI-кнопка на карточке записи | — |
| **Handler** | handler.md | Именованный набор actions (YAML DSL) | — |
| **Handler** | Go code | HTTP handler (Gin) | — |
| **Mutation** | ADR-0019 | DML-оркестрация в Object View | — |
| **Step** | scenario.md | Атомарный шаг сценария | — |

Три ключевых конфликта:

- **Action** перегружен: UI-кнопка vs атомарная операция
- **Handler** перегружен: декларативный YAML-блок vs HTTP handler
- **Mutation** изолирован в ADR-0019, не связан с общей моделью

Без единой терминологии:
- Разработчики путаются при обсуждении ("какой action имеешь в виду?")
- Документация противоречит сама себе
- Архитектурные решения не стыкуются между ADR

### Требования к терминологии

1. **Непротиворечивость** — один термин = одно значение во всех документах
2. **Иерархичность** — термины образуют понятную иерархию от абстрактного к конкретному
3. **Интуитивность** — новый разработчик понимает термин без заглядывания в глоссарий
4. **Совместимость** — не конфликтует с устоявшимися терминами Go (handler), SQL (procedure), HTTP (request)
5. **Расширяемость** — новые типы исполняемой логики встраиваются без ломки терминологии

### Индустриальный контекст

| Платформа | UI-кнопка | Атомарная операция | Набор операций | Долгий процесс |
|-----------|-----------|-------------------|---------------|----------------|
| **Salesforce** | Quick Action | — | Apex Trigger / Flow Element | Flow / Process |
| **Dynamics 365** | Command | Action Step | Action | Business Process Flow |
| **ServiceNow** | UI Action | Activity | Workflow Activity | Workflow |
| **Temporal** | — | Activity | — | Workflow |
| **n8n** | — | Node | — | Workflow |

Индустрия не имеет единого стандарта, но прослеживается паттерн: **иерархия от простого к сложному** с разным уровнем durability.

## Рассмотренные варианты

### Вариант A — Минимальный рефакторинг (переименовать только конфликты)

Переименовать только конфликтующие термины, оставив остальные как есть.

- handler.md "Action" → "Operation"
- handler.md "Handler" → "Logic Block"

**Плюсы:**
- Минимальные изменения в документации
- Быстро

**Минусы:**
- Не создаёт единой иерархии
- "Operation" и "Logic Block" — generic, не несут смысловой нагрузки
- Mutation из ADR-0019 остаётся изолированным

### Вариант B — Единая иерархия с зонтичным термином Action (выбран)

Action = зонтичный термин. Все виды исполняемой логики — подтипы Action. Конфликтующие термины получают новые имена.

**Плюсы:**
- Единая иерархия: Action → Command → Procedure → Scenario
- Каждый термин однозначен
- Mutation поглощается (action type: procedure)
- Расширяемо: новые типы action добавляются без ломки

**Минусы:**
- Требует рефакторинга docs/private/handler.md и docs/private/scenario.md
- "Procedure" может ассоциироваться с SQL stored procedure (но контекст различает)
- ADR-0022 config потребует обновления при реализации

### Вариант C — Salesforce-терминология

Перенять терминологию Salesforce: Quick Action, Flow, Apex Trigger.

**Плюсы:**
- Знакома пользователям Salesforce
- Проверена индустрией

**Минусы:**
- Привязка к чужому бренду
- "Apex Trigger" неприменим (у нас нет Apex)
- "Flow" конфликтует с Go control flow
- Не все концепции SF map 1:1 на нашу архитектуру

### Вариант D — Workflow-центричная терминология

Всё строится вокруг "Workflow": Workflow Action, Workflow Step, Workflow.

**Плюсы:**
- Единый корень
- Понятно

**Минусы:**
- "Workflow" перегружен в индустрии (GitHub Actions, n8n, Temporal)
- Не различает синхронное (procedure) и асинхронное (scenario)
- Слишком длинные составные термины

## Решение

**Выбран вариант B: Единая иерархия с зонтичным термином Action.**

### Иерархия терминов

Платформа разделяет **исполняемую логику** (действия с side effects) и **вычислительную логику** (чистые вычисления):

```
Исполняемая логика (Action hierarchy)          Вычислительная логика (CEL ecosystem)
─────────────────────────────────────          ─────────────────────────────────────
Action (зонтичный термин)                      CEL Expression (inline, одноразовое)
│                                              │
├── type: navigate    → URL transition         Function (именованное, reusable)
├── type: field_update → атомарный DML           fn.discount(tier, amount)
├── type: procedure   → Procedure (sync)         fn.is_high_value(amount)
└── type: scenario    → Scenario (async)         Вызывается из любого CEL-контекста
```

**Function** — ортогональна к Action hierarchy. Функции не выполняют действий — они вычисляют значения. Функции **вызываются внутри** CEL-выражений, которые используются на всех уровнях обеих иерархий.

### Глоссарий

| Термин | Определение | Аналог (Salesforce) | Уровень durability |
|--------|-------------|--------------------|--------------------|
| **Action** | Реакция системы на триггер. Зонтичный термин, определяемый типом (`navigate`, `field_update`, `procedure`, `scenario`). Может быть вызван из Object View (кнопка), Automation Rule (триггер), API (endpoint), или другого Action | Quick Action / Button | — |
| **Command** | Атомарная операция внутри Procedure: `record.create`, `notification.email`, `POST url`, `transform`, `validate`. Выполняется синхронно. Не имеет собственного state | Flow Element / Action Step | Нет (in-memory) |
| **Procedure** | Именованный набор Commands, описанный декларативно (JSON + CEL). Собирается через Constructor UI или редактируется как JSON. Хранится как JSONB. Выполняется синхронно в рамках одного запроса. Поддерживает условную логику (`when`, `if/else`, `match`), rollback (Saga pattern), вызов других Procedures (`call`). Аналог хранимой процедуры, но безопасной (sandbox, лимиты) | Invocable Action / Autolaunched Flow | Нет (транзакция) |
| **Scenario** | Долгоживущий бизнес-процесс, координирующий последовательность Steps с гарантиями durability (состояние переживает рестарт), consistency (откат при ошибках) и observability (полная история). Выполняется асинхронно. Поддерживает Signals (ожидание внешних событий), Timers (отложенные действия), Checkpoints | Screen Flow / Record-Triggered Flow | Да (PostgreSQL) |
| **Step** | Единица работы внутри Scenario. Вызывает Procedure, inline Command, или встроенную операцию (`wait signal`, `wait timer`). Имеет input/output mapping, retry policy, rollback | Flow Step | Да (persisted) |
| **Function** | Именованное чистое CEL-выражение с типизированными параметрами. Вызывается через `fn.*` namespace из любого CEL-контекста (validation rules, defaults, visibility, procedure input, scenario when). Нет side effects. Dual-stack: cel-go (backend) + cel-js (frontend). Не является Action — это вычислительная единица, ортогональная к Action hierarchy | Custom Formula Function | — |

### Mapping старых терминов

| Старый термин | Документ | Новый термин | Обоснование |
|--------------|----------|-------------|-------------|
| Action (unit of work) | handler.md | **Command** | Императивное название для атомарной операции |
| Action (UI button) | ADR-0022 | **Action** | Остаётся — это зонтичный термин |
| Handler (YAML DSL block) | handler.md | **Procedure** | Именованный набор commands, аналог stored procedure |
| Handler (HTTP) | Go code | **Handler** | Остаётся — это Go/HTTP термин, не конфликтует в контексте |
| Mutation (DML orchestration) | ADR-0019 | **Action type: procedure** | Поглощается — mutation = procedure с DML commands |
| Action Type (record, notification) | handler.md | **Command Type** | Категория command: `record.*`, `notification.*`, `integration.*` |

### Связь между уровнями

```
┌─────────────────────────────────────────────────────────────────┐
│                        Object View                               │
│   actions: [                                                     │
│     { key: "ship", type: "field_update", ... }                  │
│     { key: "send", type: "procedure", procedure: "send_prop" }  │
│     { key: "fulfill", type: "scenario", scenario: "order_ful" } │
│   ]                                                              │
└──────────┬──────────────────┬──────────────────┬────────────────┘
           │                  │                  │
           ▼                  ▼                  ▼
     ┌───────────┐    ┌──────────────┐    ┌──────────────┐
     │   DML     │    │  Procedure   │    │  Scenario    │
     │  Engine   │    │  Engine      │    │  Engine      │
     │           │    │              │    │              │
     │  UPDATE   │    │  commands:   │    │  steps:      │
     │  SET ...  │    │   - record.* │    │   - proc     │
     │           │    │   - email    │    │   - wait     │
     │           │    │   - POST     │    │   - signal   │
     └───────────┘    └──────────────┘    └──────────────┘
     Синхронно         Синхронно           Асинхронно
     Транзакция        Транзакция          Durable
```

### Типы Action: детализация

#### navigate

Клиентская навигация. Не вызывает backend. Выполняется frontend-роутером.

```jsonc
{
  "key": "create_task",
  "label": "Создать задачу",
  "type": "navigate",
  "navigate_to": "/app/Task/new?related_to=:recordId"
}
```

Доступен: Phase 9a (Object View core).

#### field_update

Атомарное обновление полей через DML. Одна операция, одна транзакция. Не требует Procedure Engine.

```jsonc
{
  "key": "mark_shipped",
  "label": "Отгрузить",
  "type": "field_update",
  "updates": {
    "status": "shipped",
    "shipped_at": "now()"
  },
  "visibility_expr": "record.status == 'confirmed'"
}
```

Выполнение: `DML UPDATE obj_order SET status='shipped', shipped_at=NOW() WHERE id=:recordId` с OLS/FLS/RLS enforcement.

Доступен: Phase 9a (Object View core).

#### procedure

Вызов именованной Procedure (бывш. Handler). Синхронное выполнение цепочки Commands.

```jsonc
{
  "key": "send_proposal",
  "label": "Отправить предложение",
  "type": "procedure",
  "procedure": "send_proposal",
  "visibility_expr": "record.status == 'draft'"
}
```

Procedure `send_proposal` (JSON, ADR-0024):
```json
{
  "name": "send_proposal",
  "commands": [
    {
      "type": "record.update",
      "object": "Order",
      "id": "$.input.recordId",
      "data": { "status": "\"proposal_sent\"", "proposal_sent_at": "$.now" }
    },
    {
      "type": "notification.email",
      "to": "$.input.record.client_email",
      "template": "proposal",
      "data": {
        "order_number": "$.input.record.order_number",
        "total_amount": "$.input.record.total_amount"
      }
    }
  ],
  "result": { "status": "\"proposal_sent\"" }
}
```

Доступен: Phase 13a (Procedure Engine).

#### scenario

Запуск долгоживущего Scenario. Асинхронное выполнение (fire-and-forget). Возвращает `execution_id`.

```jsonc
{
  "key": "start_fulfillment",
  "label": "Запустить исполнение",
  "type": "scenario",
  "scenario": "order_fulfillment",
  "visibility_expr": "record.status == 'paid'"
}
```

Доступен: Phase 13b (Scenario Engine).

### Связь с Automation Rules (ADR-0019)

Automation Rules используют ту же иерархию Action:

```json
{
  "object": "Order",
  "trigger": "record.after_update",
  "condition": "new.status == 'paid' && old.status != 'paid'",
  "action": {
    "type": "scenario",
    "scenario": "order_fulfillment",
    "input": { "orderId": "record.id" }
  }
}
```

Automation Rule — это не отдельная концепция, а **триггер, вызывающий Action**. Триггер определяет *когда*, Action определяет *что*.

### Command Types (бывш. Action Types)

Категории атомарных операций внутри Procedure:

| Command Type | Prefix | Примеры |
|-------------|--------|---------|
| **record** | `record.*` | `record.create`, `record.update`, `record.delete`, `record.get`, `record.query` |
| **notification** | `notification.*` | `notification.email`, `notification.sms`, `notification.push` |
| **integration** | `integration.*` | `POST url`, `GET url`, `webhook` |
| **compute** | `compute.*` | `transform`, `validate`, `aggregate`, `fail` |
| **flow** | `flow.*` | `call` (procedure), `start` (scenario) |
| **wait** | `wait.*` | `wait signal`, `wait timer`, `wait until` |

### Инкрементальная реализация

| Фаза | Что доступно | Action types / CEL |
|------|-------------|-------------------|
| **Phase 9a** | Object View core | `navigate`, `field_update` |
| **Phase 10** | Custom Functions (ADR-0026) | `fn.*` в любом CEL-контексте |
| **Phase 13a** | Procedure Engine | + `procedure` |
| **Phase 13b** | Scenario Engine | + `scenario` |
| **Phase 13c** | Approval Processes | Scenario + built-in approval commands |

Phase 9a стартует с `navigate` и `field_update` — они не требуют Procedure/Scenario Engine. Custom Functions (Phase 10) устраняют дублирование CEL-выражений. Когда Engine появится, Object View получит новые типы actions **без изменения архитектуры**.

### CEL как сквозной expression language

Все уровни используют CEL (ADR-0019, Phase 7b). Custom Functions (ADR-0026) устраняют дублирование CEL-выражений:

| Уровень | Где CEL | Пример с Function |
|---------|---------|-------------------|
| **Object View** | `visibility_expr` — когда показывать кнопку | `fn.is_high_value(record.amount)` |
| **Validation Rule** | `expression` — проверка при сохранении | `fn.discount(record.tier, record.amount) < 10000` |
| **Default Expression** | `default_expr` — значение по умолчанию | `fn.discount(record.tier, record.amount)` |
| **Procedure** | `when`, `input.*` — условия и маппинг | `fn.discount($.input.tier, $.input.amount)` |
| **Scenario** | `when`, `input.*` — условия и маппинг | `fn.is_vip($.steps.order.tier)` |
| **Automation Rule** | `condition` — триггер condition | `fn.needs_approval(new.amount, new.tier)` |

Единый expression language от UI до backend — cel-go (backend) + cel-js (frontend). Functions доступны на обеих сторонах (dual-stack, ADR-0026).

### Терминология в коде

| Термин | Go package | Таблица БД | API endpoint |
|--------|-----------|-----------|-------------|
| Action (definition) | `internal/platform/action` | `metadata.action_definitions` | `/api/v1/admin/actions` |
| Procedure | `internal/platform/procedure` | `metadata.procedures` | `/api/v1/admin/procedures` |
| Command | `internal/platform/procedure/command` | (inline in procedure JSON) | — |
| Scenario | `internal/platform/scenario` | `metadata.scenarios` + `scenario_executions` | `/api/v1/admin/scenarios` |
| Step | `internal/platform/scenario/step` | `scenario_step_history` | — |
| Function | `internal/platform/function` | `metadata.functions` | `/api/v1/admin/functions` |

## Последствия

### Позитивные

- **Единый словарь** — один термин = одно значение во всех ADR, документах, коде и обсуждениях
- **Иерархия понятна интуитивно**: Action (что?) → Command (атомарное) → Procedure (цепочка) → Scenario (долгий процесс)
- **Handler больше не конфликтует** — в Go остаётся HTTP handler, декларативный блок = Procedure
- **Mutation поглощён** — не нужен отдельный термин, это action type: procedure
- **Расширяемость** — новые action types (e.g. `approval`, `batch`) добавляются в иерархию без ломки
- **Инкрементальность** — Phase 9a работает с navigate + field_update, Procedure/Scenario Engine добавляются позже

### Негативные

- Переходный период: в существующем коде (если появится) могут встречаться старые термины — нужна пометка "deprecated terminology"
- "Procedure" может ассоциироваться с SQL stored procedure — различается контекстом (metadata vs database)
- Переходный период: старые термины в документах пока не обновлены — нужна пометка "deprecated terminology"

### Связанные ADR

- **ADR-0019** — Declarative business logic: Automation Rules используют Action types; Object View → Action binding; термин "Mutation" заменён на "Action type: procedure"
- **ADR-0020** — DML Pipeline: field_update action type выполняется через DML Engine
- **ADR-0022** — Object View: actions config использует типизацию из этого ADR (navigate, field_update, procedure, scenario)

- **ADR-0024** — Procedure Engine: JSON DSL + Constructor UI. Терминологический маппинг: "Handler" → **Procedure**, "Action" → **Command**, "Action Type" → **Command Type**
- **ADR-0025** — Scenario Engine: JSON DSL + Constructor UI. Терминологический маппинг: "Scenario" → **Scenario** (без изменений), "Step" → **Step** (без изменений), "Handler" (в контексте step) → **Procedure**
- **ADR-0026** — Custom Functions: именованные чистые CEL-выражения. **Function** — ортогональна к Action hierarchy (вычисления, а не действия). `fn.*` namespace, dual-stack (cel-go + cel-js)
