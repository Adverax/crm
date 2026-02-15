# ADR-0025: Scenario Engine — оркестрация долгоживущих бизнес-процессов

**Статус:** Принято
**Дата:** 2026-02-15
**Участники:** @roman_myakotin

## Контекст

### Проблема: координация многошаговых бизнес-процессов

Платформа поддерживает синхронную бизнес-логику через DML Pipeline (ADR-0020) и Procedure Engine (ADR-0024): валидация, defaults, computed fields, цепочки Commands. Однако реальные бизнес-процессы часто выходят за рамки одного HTTP-запроса:

| Процесс | Длительность | Особенности |
|---------|-------------|-------------|
| Onboarding клиента | Часы-дни | Ожидание email-подтверждения, настройка интеграций |
| Согласование скидки | Дни | Human-in-the-loop, эскалация при timeout |
| Обработка заказа | Минуты-часы | Saga: резервирование, оплата, отгрузка, rollback при ошибке |
| Подписание контракта | Дни-недели | Внешние сигналы (DocuSign), напоминания |

Текущая архитектура не решает:

1. **Durability** -- состояние процесса хранится в памяти; при рестарте сервера контекст теряется, процесс "зависает"
2. **Consistency** -- при ошибке на шаге 4 из 7 нет механизма автоматического отката завершённых шагов 1-3
3. **Ожидание внешних событий** -- процесс не может "заснуть" и проснуться при получении webhook или действии пользователя
4. **Observability** -- нет единой истории выполнения; логи разбросаны, невозможно ответить "на каком шаге зависла заявка?"
5. **Idempotency** -- при retry после сбоя возможны дублирования (двойные списания, повторные email)

### Операционные и финансовые риски без оркестрации

| Риск | Последствие |
|------|-------------|
| "Зависшие" операции | Клиент оплатил, но заказ не создан -- нет автоматического rollback |
| Потеря контекста при сбое | После рестарта непонятно, какие операции завершены |
| Двойные списания | Повторное выполнение без проверки идемпотентности |
| Упущенные сделки | Процесс "завис", менеджер забыл, клиент ушёл к конкуренту |

### Связь с терминологией ADR-0023

ADR-0023 определил иерархию исполняемой логики: Action (зонтичный) -> Command (атомарная операция) -> Procedure (синхронная цепочка) -> **Scenario** (асинхронный долгоживущий процесс). Scenario -- верхний уровень иерархии, обеспечивающий durability и координацию. Каждый Step сценария вызывает Procedure, inline Command или встроенную операцию (`wait signal`, `wait timer`).

### Связь с Object View (ADR-0022) и Automation Rules (ADR-0019)

- **Object View**: action type `scenario` запускает Scenario из UI-кнопки на карточке записи
- **Automation Rules**: триггер "после обновления записи" может запустить Scenario как реакцию (post-execute stage, ADR-0020)

## Рассмотренные варианты

### Вариант A -- Прямые сервисные вызовы (status quo)

Каждый процесс реализуется как цепочка вызовов в Go-сервисе: `createAccount() -> sendEmail() -> setupIntegration()`. Состояние -- в переменных текущего запроса.

**Плюсы:**
- Нет новых абстракций
- Простая реализация для 2-3 процессов
- Нет overhead на сериализацию состояния

**Минусы:**
- Нет durability: рестарт = потеря контекста
- Нет rollback: ошибка на шаге N оставляет систему в неконсистентном состоянии
- Невозможно ждать внешние события (approval, webhook)
- Boilerplate: retry, timeout, state management -- в каждом сервисе заново
- Не масштабируется: 10+ процессов = технический долг
- "Только Вася знает, как работает этот процесс"

### Вариант B -- Event-driven хореография (pub/sub)

Сервисы общаются через события: `AccountCreated -> EmailService.SendWelcome -> IntegrationService.Setup`. Каждый сервис реагирует на события других.

**Плюсы:**
- Слабая связанность между сервисами
- Естественная масштабируемость
- Каждый сервис развивается независимо

**Минусы:**
- Control flow распределён по сервисам -- нет единого места для понимания процесса
- Отладка через трассировку событий -- сложнее, чем одна точка наблюдения
- Failure handling в каждом сервисе отдельно -- нет централизованного rollback
- Циклические зависимости между событиями -- труднообнаружимые баги
- Сложно добавить условную логику ("если сумма > 100k, другой approver")
- Требует отдельной инфраструктуры (message broker)

### Вариант C -- JSON DSL + Constructor UI (выбран)

Центральный координатор выполняет сценарии, описанные декларативно (JSON, JSONB в PostgreSQL). Администратор собирает сценарий через Constructor UI (аналогично Procedure Constructor из ADR-0024). Saga pattern для rollback.

**Плюсы:**
- Constructor-first: администратор собирает сценарий через формы, не пишет JSON
- JSON нативен для стека: `encoding/json` (Go), JSONB (PostgreSQL), TypeScript
- Явный control flow: весь процесс виден в одном месте
- Централизованная обработка ошибок и rollback
- Durability из коробки (PostgreSQL)
- Встроенные signals, timers, retry policies
- Простота отладки: одна точка наблюдения, полная история execution
- CEL как сквозной expression language (ADR-0019, Phase 7b)

**Минусы:**
- Overhead для простых операций (одношаговые -- не нужен Scenario)
- PostgreSQL как единственный backend для durability
- Constructor UI -- дополнительные инвестиции во frontend
- Декларативность ограничивает: для сложной логики нужна Procedure (ADR-0024)

### Вариант D -- Temporal/Cadence (внешний workflow engine)

Использование production-grade платформы (Temporal.io) для оркестрации.

**Плюсы:**
- Production-ready, проверен в масштабе (Uber, Netflix, Stripe)
- Детерминистический replay
- Fork/join, child workflows, versioning
- Активное сообщество и документация

**Минусы:**
- Внешняя зависимость: отдельный сервис, кластер, мониторинг
- Противоречит self-hosted фокусу платформы (ADR-0016) -- пользователь должен разворачивать Temporal
- Learning curve значительно выше (Temporal SDK, worker concept, activity vs workflow)
- Код workflow на Go, не декларативный -- администратор не может настраивать
- Overkill для 80% сценариев CRM (линейные approval flows)
- Привязка к конкретному вендору

## Решение

**Выбран вариант C: JSON DSL + Constructor UI, персистентность в PostgreSQL (JSONB) и Saga pattern для rollback.**

Администратор собирает сценарий через Constructor UI (Scenario Constructor). JSON — внутреннее представление (IR). Power users могут редактировать JSON напрямую. При росте потребностей (fork/join, детерминистический replay) возможна миграция на Temporal — декларативный DSL может быть скомпилирован в Temporal workflow.

### Архитектурный принцип: Orchestration over Choreography

Для критичных бизнес-процессов выбрана оркестрация:
- Центральный координатор (Orchestrator) знает все шаги
- Явный control flow виден в одном месте
- Централизованная обработка ошибок и rollback
- Одна точка наблюдения для мониторинга и отладки

### Модель выполнения: Workflow

**Sequential Workflow** -- основная модель. Шаги выполняются последовательно с условным пропуском (`when`).

Расширения workflow:
- **`goto`** -- переход к произвольному шагу (создание циклов, возвратов)
- **`loop`** -- повторение группы шагов пока выполняется условие (`while`)

**State Machine отложен.** Workflow + `goto` + `wait signal` покрывает 95% бизнес-сценариев. State Machine запланирован для будущих версий (документооборот со статусами, подписки с lifecycle).

```
Workflow с расширениями:

  +---------+
  | Step 1  |
  +----+----+
       |
       v
  +---------+     +---------+
  | Step 2  |---->| Step 4  |  (goto)
  +----+----+     +---------+
       | when
       v
  +---------+
  | Step 3  |<---+
  +----+----+    | (loop)
       +---------+
```

### Структура Scenario

```json
{
  "code": "order_fulfillment",
  "name": "Исполнение заказа",
  "version": 1,
  "description": "Полный цикл: резервирование → оплата → отгрузка",
  "input": [
    { "name": "orderId", "type": "uuid", "required": true },
    { "name": "amount", "type": "number", "required": true }
  ],
  "steps": [],
  "procedures": {},
  "onError": "compensate",
  "settings": {
    "timeout": "30d",
    "retryPolicy": { "maxAttempts": 3, "delay": "5s", "backoff": 2 }
  },
  "meta": {}
}
```

### Структура Step

Каждый Step -- единица работы внутри сценария. В соответствии с ADR-0023, поле `procedure` (ранее `handler`) указывает на Procedure или inline Command.

```json
{
  "code": "charge_payment",
  "name": "Списание оплаты",
  "procedure": "process_payment",
  "input": {
    "orderId": "$.input.orderId",
    "amount": "$.input.amount"
  },
  "rollback": {
    "procedure": "refund_payment",
    "input": { "paymentId": "$.steps.charge_payment.paymentId" }
  },
  "retry": { "maxAttempts": 3, "delay": "5s", "backoff": 2 },
  "timeout": "30s",
  "when": "$.steps.reserve.success",
  "goto": null,
  "meta": {}
}
```

Форматы указания `procedure` в step:

| Формат | Интерпретация | Пример |
|--------|---------------|--------|
| Строка-идентификатор | Ссылка на Procedure | `"procedure": "create_customer"` |
| Строка с namespace | Внешняя Procedure | `"procedure": "integrations.sync_1c"` |
| Inline Command (JSON) | Одна команда | `"procedure": { "type": "notification.email", ... }` |
| Массив commands (JSON) | Несколько команд | `"procedure": [{ "type": "record.create", ... }]` |

Порядок разрешения имени: сначала локальные `procedures` сценария, затем глобальный реестр Procedure (ADR-0024).

### Execution Lifecycle

```
  pending ---------> running ----------> completed
     |                  |
     |                  +---> waiting ---> running
     |                  |
     |                  +---> compensating ---> failed
     |
     +---> cancelled
```

| Статус | Описание |
|--------|----------|
| `pending` | Создан, ожидает запуска |
| `running` | Выполняется (активный шаг) |
| `waiting` | Ожидает signal или timer |
| `compensating` | Выполняется откат (Saga) |
| `completed` | Успешно завершён |
| `failed` | Ошибка (после rollback или fail_fast) |
| `cancelled` | Отменён вручную или по API |

### Signals

**Signal** -- внешнее событие, влияющее на выполнение. Execution приостанавливается (`status=waiting`) до получения сигнала.

```json
{
  "code": "wait_approval",
  "procedure": { "type": "wait.signal", "signalType": "approval_decision", "timeout": "24h" },
  "input": {}
}
```

Signal API:

```
POST /api/v1/executions/{executionId}/signal
{
  "type": "approval_decision",
  "payload": { "approved": true, "comment": "OK" }
}
```

Типичные сигналы: `approval_decision`, `email_confirmed`, `payment_completed`, `document_signed`.

### Timers

Четыре типа таймеров:

| Тип | Назначение | Пример |
|-----|------------|--------|
| `delay` | Пауза на N времени | Подождать 1 час перед follow-up |
| `until` | Ждать до timestamp | Активировать в дату начала |
| `timeout` | Ограничение ожидания signal | Отменить если не ответили за 7 дней |
| `reminder` | Периодическое напоминание | Напоминать каждые 24 часа |

### Rollback (Saga Pattern)

При ошибке на шаге N автоматически откатываются все завершённые шаги в обратном порядке (LIFO):

```
Forward:  Step1 --> Step2 --> Step3 --> Step4 (X Error!)
Rollback:          Comp3 <-- Comp2 <-- Comp1
```

Rollback выполняется по принципу best-effort: если откатное действие падает, ошибка логируется и продолжается откат следующего шага. Не все шаги обязаны иметь rollback (email отправлен -- не откатить).

Стратегии обработки ошибок:

| Стратегия | Поведение |
|-----------|-----------|
| `fail_fast` | При первой ошибке -- сразу `failed`, без rollback |
| `retry` | Retry по политике, затем fail_fast |
| `compensate` | Retry, затем rollback всех выполненных шагов |

### Durability и Recovery

Состояние execution персистится в PostgreSQL после каждого шага. При рестарте приложения:

1. Найти executions со статусом `running`, `waiting`, `compensating`
2. Для `running` -- определить последний завершённый step, retry текущий с тем же idempotency key
3. Для `waiting` -- проверить поступившие сигналы/таймеры, возобновить если есть
4. Для `compensating` -- продолжить rollback с текущего шага

Step записывается как завершённый только после успешного выполнения. При recovery -- retry текущего step, не повторение завершённых.

### Idempotency

Платформа автоматически генерирует idempotency key для каждого шага:

```
idempotencyKey = {executionId}-{stepCode}
```

Этот ключ передаётся в Procedure и далее во внешние вызовы. Повторный вызов с тем же ключом должен давать тот же результат. Все встроенные Procedures платформы (`record.*`, `notification.*`) -- idempotent by design. Для HTTP-интеграций ключ передаётся в заголовке запроса.

### Context (модель контекста)

Контекст накапливается по мере выполнения и доступен через CEL expressions:

| Путь | Описание | Мутабельность |
|------|----------|---------------|
| `$.input` | Входные параметры сценария | Immutable |
| `$.steps.<code>` | Результат шага (null если пропущен по `when`) | Append-only |
| `$.steps.<code>.meta` | Метаданные шага из определения | Immutable |
| `$.signals` | Полученные сигналы | Append-only |
| `$.meta` | Метаданные сценария | Immutable |
| `$.execution` | Системные данные (startedAt, attempt) | Read-only |
| `$.user` | Текущий пользователь | Immutable |
| `$.now` | Текущее время | Computed |

Пропущенные шаги (условие `when` = false): `$.steps.<code>` = `null`. Безопасный доступ: `$.steps.x != null && $.steps.x.field == "value"`.

### Архитектурные компоненты

| Компонент | Ответственность |
|-----------|-----------------|
| **Scenario Registry** | Хранит определения сценариев (JSON/JSONB в DB, Go-embedded для built-in) |
| **Orchestrator** | Запускает и координирует executions, управляет lifecycle |
| **Step Executor** | Выполняет конкретный шаг: резолвит Procedure, передаёт input, сохраняет output |
| **Compensator** | Выполняет откатные действия в обратном порядке (LIFO) |
| **Signal Handler** | Принимает внешние сигналы через API, пробуждает waiting executions |
| **Timer Scheduler** | Планирует и запускает отложенные события, проверяет timeout |
| **Execution Repository** | Персистирует состояние execution и историю шагов в PostgreSQL |

### Storage

Две таблицы в PostgreSQL:

**`scenario_executions`** -- состояние execution:

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | Уникальный ID execution |
| scenario_code | VARCHAR | Код сценария |
| scenario_version | INT | Версия на момент запуска |
| status | VARCHAR | pending/running/waiting/compensating/completed/failed/cancelled |
| input | JSONB | Входные параметры |
| context | JSONB | Накопленный контекст (steps results, signals) |
| current_step | VARCHAR | Код текущего шага |
| error | JSONB | Информация об ошибке (если есть) |
| started_at | TIMESTAMPTZ | Время запуска |
| completed_at | TIMESTAMPTZ | Время завершения |
| created_at | TIMESTAMPTZ | Время создания |
| updated_at | TIMESTAMPTZ | Время обновления |

**`scenario_step_history`** -- история выполнения шагов:

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | Уникальный ID записи |
| execution_id | UUID FK | Ссылка на execution |
| step_code | VARCHAR | Код шага |
| status | VARCHAR | completed/failed/skipped/compensated |
| input | JSONB | Входные данные шага |
| output | JSONB | Результат выполнения |
| error | JSONB | Ошибка (если есть) |
| attempt | INT | Номер попытки |
| started_at | TIMESTAMPTZ | Время начала |
| completed_at | TIMESTAMPTZ | Время завершения |
| created_at | TIMESTAMPTZ | Время создания |

### Лимиты

| Параметр | Ограничение | Обоснование |
|----------|-------------|-------------|
| Максимум шагов в сценарии | 50 | Предотвращение чрезмерно сложных процессов |
| Максимальная глубина goto | 100 итераций | Защита от бесконечных циклов |
| Timeout execution по умолчанию | 30 дней | Защита от "забытых" executions |
| Retry maxAttempts по умолчанию | 3 | Баланс надёжности и ресурсов |
| Размер context (JSONB) | 1 MB | Предотвращение разрастания состояния |

### State Machine (не в MVP)

Запланирован для будущих версий. Workflow + goto + wait signal покрывает 95% бизнес-сценариев CRM. State Machine понадобится для:
- Документооборот со статусами (draft -> review -> approved -> active)
- Подписки с lifecycle (trial -> active -> paused -> cancelled)
- Процессы, где состояние важнее последовательности шагов

Планируемый синтаксис: `mode: state_machine`, блок `states` вместо `steps`, переходы по событиям (`on`). Альтернатива в MVP: Workflow + goto + wait signal для event-driven логики.

## Последствия

### Позитивные

- **Durability** -- состояние execution переживает рестарт; recovery с последнего checkpoint; нет "зависших" операций
- **Saga-гарантии** -- автоматический rollback всех завершённых шагов при ошибке; система возвращается в консистентное состояние
- **Observability** -- полная история выполнения каждого шага с input/output/error; одна точка для диагностики
- **Self-service для администраторов** -- новые процессы собираются через Constructor UI без разработки; время внедрения: дни вместо недель
- **Встроенные примитивы** -- signals, timers, retry policies, idempotency -- из коробки, без boilerplate в каждом сервисе
- **Вписывается в иерархию ADR-0023** -- Scenario = верхний уровень (Action -> Command -> Procedure -> Scenario); Step вызывает Procedure
- **CEL как сквозной expression language** -- единый язык от Object View до Scenario (`when`, `input.*`)
- **Инкрементальная реализация** -- Phase 13b; не блокирует Phase 9a (Object View) и Phase 13a (Procedure Engine)
- **Путь миграции на Temporal** -- при росте потребностей декларативный DSL может быть скомпилирован в Temporal workflow

### Негативные

- **Overhead для простых операций** -- одношаговые процессы без ожиданий не должны быть Scenarios (использовать Procedure или прямой сервисный вызов)
- **PostgreSQL dependency** -- durability привязана к PostgreSQL; при высоких нагрузках на execution может потребоваться отдельная БД
- **Learning curve** -- lifecycle, signals, retry policies, Saga pattern -- новые концепции для команды (Constructor UI снижает порог входа)
- **Декларативные ограничения** -- для сложной вычислительной логики внутри шага нужна Procedure (ADR-0024), а не inline expression

## Связанные ADR

- **ADR-0019** -- Declarative business logic: Automation Rules (post-execute trigger) запускают Scenario; CEL как expression language
- **ADR-0020** -- DML Pipeline: post-execute stage может запустить Scenario через Automation Rules
- **ADR-0022** -- Object View: action type `scenario` запускает Scenario из UI-кнопки на карточке записи
- **ADR-0023** -- Action terminology: Scenario в иерархии Action -> Command -> Procedure -> Scenario; Step вызывает Procedure
- **ADR-0024** -- Procedure Engine: Steps выполняют Procedures; Procedure = синхронная цепочка Commands
