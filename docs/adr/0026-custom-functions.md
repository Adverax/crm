# ADR-0026: Custom Functions — именованные чистые вычисления

**Статус:** Принято
**Дата:** 2026-02-15
**Участники:** @roman_myakotin

## Контекст

### Проблема: дублирование CEL-выражений

Платформа использует CEL (Common Expression Language) как единый expression language (ADR-0019) во множестве подсистем:

| Подсистема | Где CEL | Пример |
|------------|---------|--------|
| Validation Rules (Phase 7b) | `expression` | `record.amount > 0 && record.amount < 1000000` |
| Default Expressions (Phase 7b) | `default_expr` | `record.tier == "gold" ? record.amount * 0.2 : 0` |
| Object View (ADR-0022) | `visibility_expr` | `record.status == "draft" && record.amount > 10000` |
| Procedure (ADR-0024) | `when`, `input.*` | `$.input.tier == "gold" ? $.input.amount * 0.2 : 0` |
| Scenario (ADR-0025) | `when`, `input.*` | `$.steps.check.tier == "gold"` |
| Automation Rules (ADR-0019) | `condition` | `new.status == "paid" && old.status != "paid"` |
| Dynamic Forms (Phase 9c) | field visibility | `record.type == "enterprise"` |

Когда одна и та же вычислительная логика нужна в нескольких местах, администратор **дублирует CEL-выражение**:

```
// Validation rule для Order:
record.tier == "gold" ? record.amount * 0.2 : record.tier == "silver" ? record.amount * 0.1 : 0

// Default expression для discount_amount:
record.tier == "gold" ? record.amount * 0.2 : record.tier == "silver" ? record.amount * 0.1 : 0

// Procedure input:
$.input.tier == "gold" ? $.input.amount * 0.2 : $.input.tier == "silver" ? $.input.amount * 0.1 : 0

// Object View visibility:
(record.tier == "gold" ? record.amount * 0.2 : record.tier == "silver" ? record.amount * 0.1 : 0) > 5000
```

Одно и то же выражение — 4 копии. При изменении бизнес-правила (добавление tier "platinum") нужно найти и обновить все копии. Это:
- **Ненадёжно** — легко пропустить одну из копий
- **Непрактично** — сложные выражения становятся нечитаемыми при встраивании
- **Противоречит DRY** — фундаментальный принцип разработки

### Procedure не решает проблему

Procedure (ADR-0024) — это набор Commands с side effects (CRUD, email, HTTP). Для чистых вычислений Procedure — overkill:

| Аспект | Function (нужно) | Procedure (есть) |
|--------|-----------------|------------------|
| Назначение | Вычислить значение | Выполнить действия |
| Side effects | Нет | Да (CRUD, email, HTTP) |
| Вызов | Inline из любого CEL | `flow.call` из Procedure |
| Возврат | Значение (any) | ProcedureResult |
| Rollback | Не нужен | Saga pattern |
| Где доступна | Везде, где есть CEL | Только как action |

Функция `fn.discount(tier, amount)` вызывается **внутри** CEL-выражения. Procedure вызывается **вместо** CEL-выражения. Это разные уровни абстракции.

### Dual-stack: cel-go + cel-js

CEL уже работает на двух сторонах (ADR-0019):
- **Backend**: cel-go — validation rules, defaults, procedure/scenario engine
- **Frontend**: cel-js — Object View `visibility_expr`, Dynamic Forms field visibility

Custom Functions должны быть доступны на **обеих сторонах**. Поскольку Functions — чистые выражения без side effects, они портабельны между cel-go и cel-js без адаптаций.

## Рассмотренные варианты

### Вариант A — Копипаст CEL-выражений (status quo)

Администратор дублирует одинаковые CEL-выражения во всех местах использования.

**Плюсы:**
- Нет новых абстракций
- Каждое выражение самодостаточно (видно всю логику в одном месте)

**Минусы:**
- Нарушение DRY: одно изменение → обновление N мест
- Ошибки: легко пропустить одну из копий
- Нечитаемость: сложные выражения inline — тяжело поддерживать
- Масштабирование: чем больше подсистем используют CEL, тем больше дублирования

### Вариант B — Custom Functions как именованные CEL-выражения (выбран)

Администратор определяет именованную функцию с типизированными параметрами. Тело функции — CEL-выражение. Функция доступна из любого CEL-контекста через namespace `fn.*`.

**Плюсы:**
- DRY: одно определение — множество использований
- Единый механизм: Functions работают везде, где есть CEL (backend + frontend)
- Декомпозиция: сложная логика разбивается на понятные именованные блоки
- Тестируемость: функцию можно протестировать изолированно
- Чистота: нет side effects — безопасно, предсказуемо, кэшируемо
- Минимальная реализация: расширение существующего CEL Environment, не новый engine

**Минусы:**
- Новая абстракция: администратору нужно понять концепцию "функция"
- Отладка: при ошибке в выражении нужно проверить и вызывающий код, и тело функции
- Зависимости: удаление/изменение функции может сломать использующие её выражения

### Вариант C — Macro expansion (шаблонные подстановки)

Вместо runtime-функций — шаблоны, которые подставляются (expand) в CEL-выражение при компиляции.

**Плюсы:**
- Прозрачность: администратор видит итоговое "развёрнутое" выражение
- Нет runtime overhead: всё раскрывается на этапе компиляции

**Минусы:**
- Нет параметров (или сложная подстановка): `${discount}` с заменой аргументов — по сути самодельный template engine
- Ошибки в развёрнутом выражении трудно соотнести с исходным шаблоном
- Нет type checking на этапе определения макроса
- Не работает на frontend (cel-js не знает о макросах)

### Вариант D — Computed Fields вместо Functions

Вычисляемые поля (Formula Fields, Phase 10) могут покрыть часть кейсов: `discount_amount = tier == "gold" ? amount * 0.2 : ...`.

**Плюсы:**
- Уже запланированы (Phase 10)
- Привязаны к объекту — естественная точка доступа

**Минусы:**
- Привязаны к конкретному объекту — нельзя переиспользовать между объектами
- Нельзя передать произвольные параметры (только поля текущей записи)
- Не работают в Procedure/Scenario (другой контекст: `$.input.*`, а не `record.*`)
- Дублирование: одна и та же формула на разных объектах

## Решение

**Выбран вариант B: Custom Functions как именованные CEL-выражения.**

### Определение Function

Function хранится в `metadata.functions` как JSONB:

```json
{
  "name": "discount",
  "description": "Рассчитать скидку по уровню клиента",
  "params": [
    { "name": "tier", "type": "string", "description": "Уровень клиента" },
    { "name": "amount", "type": "number", "description": "Сумма заказа" }
  ],
  "return_type": "number",
  "body": "tier == \"gold\" ? amount * 0.2 : tier == \"silver\" ? amount * 0.1 : 0"
}
```

- **name** — уникальный идентификатор (snake_case), вызывается как `fn.name()`
- **params** — типизированные параметры (string, number, boolean, list, map, any)
- **return_type** — тип возвращаемого значения (для type checking)
- **body** — CEL-выражение; параметры доступны как переменные по имени

### Вызов из CEL

Все Functions доступны через namespace `fn.*` в любом CEL-контексте:

```
// Validation rule
fn.discount(record.tier, record.amount) > 5000

// Default expression
fn.discount(record.tier, record.amount)

// Procedure command input
"discount": "fn.discount($.input.tier, $.input.amount)"

// Object View visibility_expr
fn.discount(record.tier, record.amount) > 5000

// Scenario when
fn.discount($.steps.order.tier, $.steps.order.amount) > 10000

// Composition: функция вызывает функцию
fn.total_with_tax(fn.discount(record.tier, record.amount), record.tax_rate)
```

Namespace `fn.*` отделяет пользовательские функции от встроенных (`size()`, `has()`, `matches()`), исключая конфликты имён.

### Dual-stack: загрузка в cel-go и cel-js

```
metadata.functions (PostgreSQL JSONB)
        │
        ├──→ Backend startup / cache invalidation
        │    └── cel-go: env.RegisterFunction("fn.discount", ...)
        │        → Validation Rules, Defaults, Procedure, Scenario
        │
        └──→ GET /api/v1/describe (Describe API)
             └── response.functions: [{ name, params, body }]
                 └── cel-js: env.registerFunction("fn.discount", ...)
                     → visibility_expr, Dynamic Forms
```

**Backend**: Functions загружаются в cel-go Environment при старте и при инвалидации кэша (outbox pattern, ADR-0012).

**Frontend**: Describe API отдаёт определения функций. cel-js регистрирует их как custom functions. `visibility_expr: "fn.discount(record.tier, record.amount) > 5000"` вычисляется **мгновенно в браузере** без round-trip на сервер.

### Constructor UI

В Expression Builder (ADR-0024) Functions появляются как категория:

1. **Function picker**: раздел "Пользовательские функции" в каталоге функций Expression Builder
   - Каждая функция с описанием, типами параметров, примером использования
   - Автоподстановка: выбрал `fn.discount` → шаблон `fn.discount(tier, amount)` с placeholder'ами

2. **Function Constructor**: отдельная admin-страница для создания/редактирования функций
   - Имя + описание
   - Параметры: name + type + description (drag-and-drop для порядка)
   - Тело: Expression Builder (тот же компонент) с параметрами в field picker
   - Live preview: тестовые значения параметров → результат в реальном времени
   - Валидация: type checking тела при сохранении

3. **Dependency view**: где используется функция (список validation rules, defaults, procedures, Object Views)

### Ограничения

| Параметр | Лимит | Обоснование |
|----------|-------|-------------|
| Размер тела | 4 KB | Функция — компактное выражение, не программа |
| Время выполнения | 100 ms | Вызывается inline, не должна блокировать |
| Вложенность | 3 уровня | `fn.a()` → `fn.b()` → `fn.c()` → stop |
| Параметры | 10 max | Больше — это уже Procedure |
| Рекурсия | Запрещена | Отслеживание call stack, ошибка `recursive_function_call` |
| Количество функций | 200 | Per-instance; предотвращение раздувания namespace |

### Safety

| Угроза | Защита |
|--------|--------|
| Бесконечная рекурсия | Запрещена: call stack tracking, статический анализ при сохранении |
| Circular dependencies | Граф зависимостей проверяется при сохранении (`fn.a` → `fn.b` → `fn.a` = ошибка) |
| Side effects | Невозможны: CEL — чистый expression language, нет I/O в grammar |
| Deletion с зависимостями | Запрет удаления: dependency view показывает использования; `DELETE` → 409 Conflict |
| Rename | Каскадное обновление: найти все CEL-выражения с `fn.old_name` → заменить |
| Resource exhaustion | Лимиты: 100ms timeout, 4KB body, 3 уровня вложенности |

### Хранение

Таблица `metadata.functions`:

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID PK | Уникальный ID |
| name | VARCHAR UNIQUE | Имя функции (snake_case) |
| description | TEXT | Описание назначения |
| params | JSONB | Массив параметров `[{name, type, description}]` |
| return_type | VARCHAR | Тип возвращаемого значения |
| body | TEXT | CEL-выражение |
| created_at | TIMESTAMPTZ | Время создания |
| updated_at | TIMESTAMPTZ | Время обновления |

### API

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/api/v1/admin/functions` | Список функций |
| POST | `/api/v1/admin/functions` | Создать функцию |
| GET | `/api/v1/admin/functions/:id` | Получить функцию |
| PUT | `/api/v1/admin/functions/:id` | Обновить функцию |
| DELETE | `/api/v1/admin/functions/:id` | Удалить (409 если есть зависимости) |
| POST | `/api/v1/admin/functions/:id/test` | Тест: input → result |
| GET | `/api/v1/admin/functions/:id/dependencies` | Где используется |

### Отношение к Formula Fields (Phase 10)

Formula Fields и Custom Functions — разные инструменты:

| Аспект | Formula Field | Custom Function |
|--------|--------------|-----------------|
| Привязка | К конкретному объекту и полю | Глобальная, без привязки |
| Контекст | `record.*` (поля текущей записи) | Произвольные параметры |
| Результат | Значение поля (хранится/вычисляется) | Значение, возвращённое в CEL |
| Reuse | Нет (per object) | Да (anywhere in CEL) |
| Где доступен | SOQL SELECT, record display | Любое CEL-выражение |

Formula Field **может вызывать** Custom Function: `fn.discount(tier, amount)` как часть формулы поля. Но Formula Field привязан к объекту, а Function — нет.

## Последствия

### Позитивные

- **DRY** — одно определение, множество использований; изменение в одном месте
- **Dual-stack** — одна и та же функция работает на backend (cel-go) и frontend (cel-js) без адаптаций
- **Минимальная реализация** — расширение существующего CEL Environment; не новый engine, не новый runtime
- **Декомпозиция** — сложная логика разбивается на именованные, тестируемые блоки
- **Безопасность** — чистые выражения, нет side effects, нет I/O; защита от рекурсии и circular deps
- **Мгновенная оценка на frontend** — `visibility_expr` с `fn.*` вычисляется в браузере без round-trip
- **Интеграция с Expression Builder** — функции появляются в каталоге; автоподстановка параметров
- **Dependency tracking** — платформа знает, где используется каждая функция; защита от удаления

### Негативные

- **Новая абстракция** — администратору нужно понять концепцию "функция" (Constructor UI снижает порог)
- **Два уровня отладки** — ошибка может быть в вызывающем выражении или в теле функции
- **Синхронизация frontend/backend** — при обновлении функции нужно инвалидировать кэш cel-js (Describe API refetch)
- **Каскадные ошибки** — изменение типа параметра может сломать вызывающие выражения (защита: type checking при сохранении)

## Связанные ADR

- **ADR-0019** — Декларативная бизнес-логика: CEL как единый expression language. Functions расширяют CEL Environment пользовательскими вычислениями
- **ADR-0022** — Object View: `visibility_expr` может вызывать Functions для сложной логики видимости
- **ADR-0024** — Procedure Engine: CEL expressions в `when`, `input.*` могут вызывать Functions; Expression Builder показывает Functions в каталоге
- **ADR-0025** — Scenario Engine: CEL expressions в `when`, `input.*` могут вызывать Functions
