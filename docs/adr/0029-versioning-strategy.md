# ADR-0029: Стратегия версирования скриптовых сущностей

**Статус:** Принято
**Дата:** 2026-02-15
**Участники:** @roman_myakotin

## Контекст

### Проблема: изменение живых скриптовых сущностей

Платформа содержит несколько типов сущностей с исполняемой логикой:

| Сущность | DSL | Вызывается из | Long-running? |
|----------|-----|---------------|---------------|
| Procedure (ADR-0024) | JSON + CEL | Scenario, Automation Rules, UI actions | Нет (секунды) |
| Scenario (ADR-0025) | JSON + CEL | Triggers, manual start | **Да** (минуты-дни) |
| Custom Function (ADR-0026) | CEL-выражение | Любой CEL-контекст (inline) | Нет |
| Validation Rule | CEL-выражение | DML pipeline (inline) | Нет |
| Automation Rule | Trigger config | DML pipeline | Нет |
| Object View (ADR-0022) | Config JSONB | Describe API | Нет |
| Layout (ADR-0027) | Config JSONB | Describe API | Нет |
| Named Credential (ADR-0028) | Config | Procedure (integration.http) | Нет |

Когда admin изменяет Procedure, возникают три проблемы:

**1. Mid-flight consistency**
Scenario запущен и ждёт сигнала (дни). Пока ждёт — admin обновил Procedure. На следующем шаге Scenario вызовет новую версию с другим контрактом → сбой.

**2. Unsafe deployment**
Save = live. Ошибка в Procedure → сразу в production. Нет возможности протестировать (dry-run) перед публикацией.

**3. No rollback**
После публикации ошибочной версии — только ручное редактирование. Нет одной кнопки "откатить".

### Не все сущности равны

Проблемы **не одинаково критичны** для разных типов сущностей:

| Проблема | Procedure | Scenario | Function | VR / AR / OV / Layout / Credential |
|----------|-----------|----------|----------|-------------------------------------|
| Mid-flight | Низкий риск (синхронный, секунды) | **Высокий** (async, дни) | Нет (inline) | Нет |
| Unsafe deploy | **Да** (сложный JSON DSL) | **Да** (сложный JSON DSL) | Низкий (одно CEL-выражение, test endpoint) | Нет (простая конфигурация) |
| No rollback | **Да** | **Да** | Низкий (правка одной строки) | Нет |

**Вывод:** Полное версирование нужно только Procedure и Scenario. Остальные сущности — либо слишком просты (CEL-выражение), либо являются конфигурацией (OV, Layout, Credential).

## Рассмотренные варианты

### Вариант A — Без версирования

Save = live для всех сущностей. Scenario snapshots definition (JSONB copy) при старте.

**Плюсы:**
- Максимальная простота: одна таблица per entity
- Нет дополнительной логики (draft/publish, version resolution)

**Минусы:**
- Нет тестирования перед публикацией: ошибка в Procedure → сразу production
- Нет истории изменений: кто что менял — неизвестно
- Нет rollback: только ручное редактирование назад
- JSONB snapshot тяжёлый: дублирование полного definition в каждом Scenario run

### Вариант B — Draft/Published без semver (выбран)

Два состояния для Procedure и Scenario. Один draft (editable, testable), один published (immutable, live). Простой auto-increment version counter (1, 2, 3...). Остальные сущности — без версирования.

**Плюсы:**
- Тестирование перед публикацией (dry-run на draft)
- Rollback к предыдущей published версии — одна операция
- История изменений (кто, когда, что)
- Scenario snapshot = FK на version_id (не JSONB copy)
- Salesforce-aligned: Flows используют exactly этот подход (active/inactive versions)
- Дифференциация: сложность только там, где оправдана
- Cognitive load для admin минимален: "Сохранить черновик" → "Опубликовать"

**Минусы:**
- Дополнительная таблица `_versions` для Procedure и Scenario
- Два указателя (draft_version_id, published_version_id) вместо inline definition
- Логика publish/rollback в service layer

### Вариант C — Full Semver (MAJOR.MINOR.PATCH)

Семантическое версирование с version constraints (`^2.0`, `~2.3`), backward compatibility validation, 4 статуса (draft/published/deprecated/archived), retention policy, snapshot tables.

**Плюсы:**
- Granular control: Scenario может зафиксировать `^2.0` и получать только совместимые обновления
- Enterprise-grade: maximum safety

**Минусы:**
- Огромная сложность: semver parser, constraint matcher, compatibility checker, 4 статуса, retention policy, snapshot tables
- YAGNI: один admin управляет и Procedures, и Scenarios — он знает, что меняет
- Cognitive load: admin должен понимать semver, breaking changes, constraints
- Ни одна CRM-платформа не использует semver для бизнес-логики (Salesforce Flows, Dynamics Power Automate, HubSpot Workflows — все используют active/inactive)
- Version constraints (`^2.0`) предполагают независимую эволюцию потребителей — в нашей single-tenant CRM admin контролирует обе стороны
- Backward compatibility validation требует формализации input/output schema — дополнительный слой сложности

### Вариант D — Immutable + Latest

Каждое сохранение создаёт новую immutable версию. Текущая = latest. Нет draft.

**Плюсы:**
- Полная история (каждое сохранение сохранено)
- Простая модель: write-only, no state transitions

**Минусы:**
- Нет draft/dry-run: каждое сохранение — сразу live
- Раздувание данных: десятки versions per entity при частом редактировании
- Нет явного "момента публикации" — всё автоматически live

## Решение

**Выбран вариант B: Draft/Published с дифференциацией по типу сущности.**

### Дифференциация

| Сущность | Версирование | Обоснование |
|----------|-------------|-------------|
| **Procedure** | Draft/Published | Сложный JSON DSL; dry-run нужен; вызывается из Scenario |
| **Scenario** | Draft/Published | Сложный JSON DSL; long-running; snapshot version при старте |
| **Custom Function** | Нет | Одно CEL-выражение; test endpoint при сохранении; dependency check защищает от breaking changes |
| **Validation Rule** | Нет | CEL-выражение, immediate apply; ошибка → DML возвращает ошибку |
| **Automation Rule** | Нет | Trigger config; Procedure, которую вызывает, имеет свой draft/published |
| **Object View** | Нет | Конфигурация представления, не логика |
| **Layout** | Нет | Конфигурация презентации |
| **Named Credential** | Нет | Конфигурация подключения |

### Модель Draft/Published

```
              Save draft             Publish
    ┌───────────────────┐    ┌──────────────────┐
    │                   ▼    │                  ▼
    │              ┌─────────┴┐           ┌──────────┐
    └──────────────│  draft   │──────────▶│published │
     (re-save)     │(editable) │  Publish   │(immutable)│
                   └──────────┘           └──────────┘
                        │                      │
                        │                      │ (при новом Publish →
                        │                      │  предыдущий published
                   Delete draft                │  получает статус
                        │                      │  "superseded")
                        ▼                      ▼
                   (удалён)              ┌───────────┐
                                        │superseded │
                                        │(read-only) │
                                        └───────────┘
```

**Три статуса:**

| Статус | Описание | Исполняемый? | Редактируемый? |
|--------|----------|-------------|----------------|
| `draft` | Черновик, в работе | Только dry-run | Да |
| `published` | Активная версия | Да | Нет |
| `superseded` | Предыдущая версия (заменена) | Нет (кроме running Scenario instances) | Нет |

Нет `deprecated` и `archived` из варианта C — три статуса покрывают все потребности.

### Хранение

**Procedure:**

```sql
-- Метаданные Procedure (без definition)
CREATE TABLE metadata.procedures (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code                 VARCHAR(100) UNIQUE NOT NULL,
    name                 VARCHAR(255) NOT NULL,
    description          TEXT,
    draft_version_id     UUID,
    published_version_id UUID,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Версии Procedure (definition живёт здесь)
CREATE TABLE metadata.procedure_versions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    procedure_id   UUID NOT NULL REFERENCES metadata.procedures(id) ON DELETE CASCADE,
    version        INT NOT NULL,               -- auto-increment per procedure (1, 2, 3...)
    definition     JSONB NOT NULL,             -- JSON DSL
    status         VARCHAR(20) NOT NULL DEFAULT 'draft',  -- draft | published | superseded
    change_summary TEXT,                       -- что изменилось
    created_by     UUID,                       -- кто создал
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at   TIMESTAMPTZ,               -- когда опубликовано (NULL для draft)
    CONSTRAINT procedure_versions_unique UNIQUE (procedure_id, version),
    CONSTRAINT procedure_versions_status_check CHECK (status IN ('draft', 'published', 'superseded'))
);
```

**Scenario — аналогично:**

```sql
CREATE TABLE metadata.scenarios (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code                 VARCHAR(100) UNIQUE NOT NULL,
    name                 VARCHAR(255) NOT NULL,
    description          TEXT,
    draft_version_id     UUID,
    published_version_id UUID,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE metadata.scenario_versions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scenario_id   UUID NOT NULL REFERENCES metadata.scenarios(id) ON DELETE CASCADE,
    version       INT NOT NULL,
    definition    JSONB NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'draft',
    change_summary TEXT,
    created_by    UUID,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at  TIMESTAMPTZ,
    CONSTRAINT scenario_versions_unique UNIQUE (scenario_id, version),
    CONSTRAINT scenario_versions_status_check CHECK (status IN ('draft', 'published', 'superseded'))
);
```

**Scenario Run snapshot (для mid-flight consistency):**

```sql
-- Scenario run хранит FK на конкретные версии Procedures
CREATE TABLE metadata.scenario_run_snapshots (
    scenario_run_id       UUID NOT NULL,
    procedure_id          UUID NOT NULL,
    procedure_version_id  UUID NOT NULL REFERENCES metadata.procedure_versions(id),
    PRIMARY KEY (scenario_run_id, procedure_id)
);
```

При старте Scenario run:
1. Для каждого `flow.call` в definition — resolve `published_version_id` текущей Procedure
2. Записать в `scenario_run_snapshots`
3. При исполнении шага — использовать зафиксированную version, не текущую published

### Версия — auto-increment integer

```
Version 1 → draft → published
Version 2 → draft → published (Version 1 → superseded)
Version 3 → draft → published (Version 2 → superseded)
```

Нет semver (MAJOR.MINOR.PATCH). Нет version constraints (`^2.0`). Простой инкремент, понятный любому admin.

### Workflow

#### Создание новой Procedure

```
1. Admin создаёт Procedure → Version 1 (draft)
2. Admin редактирует draft (save N раз — тот же draft, не новая version)
3. Admin тестирует: dry-run на draft
4. Admin публикует → Version 1 (published)
   → procedures.published_version_id = version_1.id
   → procedures.draft_version_id = NULL
```

#### Обновление существующей

```
1. Admin нажимает "Редактировать" → Version 2 (draft) создаётся как копия Version 1
2. Admin вносит изменения в draft
3. Admin тестирует: dry-run на draft
4. Admin публикует → Version 2 (published), Version 1 (superseded)
   → procedures.published_version_id = version_2.id
   → procedures.draft_version_id = NULL
```

#### Rollback

```
1. Admin нажимает "Откатить" на Procedure
2. Текущий published (Version 3) → superseded
3. Предыдущий superseded (Version 2) → published
   → procedures.published_version_id = version_2.id
4. UI показывает: "Откачено к версии 2"
```

#### Удаление draft

```
1. Admin нажимает "Отменить черновик"
2. Draft version удаляется
3. procedures.draft_version_id = NULL
4. Published version остаётся активной
```

### Сущности БЕЗ версирования

Для сущностей без версирования (Function, Validation Rule, Automation Rule, OV, Layout, Credential) — definition хранится **inline** в основной таблице:

```sql
-- Function: definition inline
CREATE TABLE metadata.functions (
    id          UUID PRIMARY KEY,
    name        VARCHAR(100) UNIQUE NOT NULL,
    params      JSONB NOT NULL,
    return_type VARCHAR(20) NOT NULL,
    body        TEXT NOT NULL,              -- CEL expression, прямо в таблице
    ...
);

-- Validation Rule: expression inline
CREATE TABLE metadata.validation_rules (
    id          UUID PRIMARY KEY,
    object_id   UUID NOT NULL,
    expression  TEXT NOT NULL,              -- CEL expression, прямо в таблице
    ...
);
```

Save = live. Защита от ошибок:
- **Function**: dependency check + type validation при сохранении; test endpoint
- **Validation Rule**: CEL compilation check при сохранении
- **OV / Layout**: preview в admin UI перед сохранением
- **Credential**: test connection endpoint

### Constructor UI интеграция

**Для Procedure/Scenario (с версирование):**

```
┌───────────────────────────────────────────┐
│  Procedure: create_order                   │
│                                           │
│  Published: Version 3 (2026-02-15)       │
│  Status: ● Published                      │
│                                           │
│  [Редактировать]  [История]  [Откатить]  │
│                                           │
│  ─── Draft (если есть) ───               │
│  Version 4 (draft)                       │
│  [Тестировать]  [Опубликовать]  [Удалить] │
│                                           │
│  ─── История версий ───                  │
│  v3  published  2026-02-15  "Добавлен webhook"  │
│  v2  superseded 2026-02-14  "Новое поле email"  │
│  v1  superseded 2026-02-13  "Первая версия"     │
└───────────────────────────────────────────┘
```

**Для Function (без версирования):**

```
┌───────────────────────────────────────────┐
│  Function: discount                       │
│                                           │
│  [Редактировать]  [Тестировать]  [Удалить]│
│                                           │
│  Нет истории версий — save = live         │
└───────────────────────────────────────────┘
```

### Retention

| Статус | Хранение |
|--------|----------|
| draft | До публикации или явного удаления |
| published | Бессрочно (текущая active version) |
| superseded | Последние 10 версий; старше — auto-delete |

Auto-delete superseded versions старше 10 — предотвращает раздувание таблицы при частом редактировании. 10 версий достаточно для анализа истории и rollback.

## Последствия

### Позитивные

- **Safe deployment** — draft + dry-run перед публикацией для сложных DSL (Procedure, Scenario)
- **Rollback** — одна операция для возврата к предыдущей версии
- **Mid-flight safety** — running Scenario instances используют зафиксированные версии Procedures
- **История изменений** — кто, когда, что изменил (change_summary)
- **Минимальная сложность** — три статуса, auto-increment integer, нет semver
- **Дифференциация** — версирование только там, где оправдано; простые сущности не усложняются
- **Salesforce-aligned** — Flows используют exactly этот подход (active/inactive versions)
- **Admin-friendly** — "Сохранить черновик" → "Опубликовать" вместо "Выберите MAJOR/MINOR/PATCH"

### Негативные

- **Дополнительные таблицы** — `procedure_versions`, `scenario_versions`, `scenario_run_snapshots`
- **Два указателя** — `draft_version_id` + `published_version_id` вместо inline definition
- **Publish workflow** — admin должен явно опубликовать; save ≠ live (может быть непривычно)
- **Нет granular constraints** — Scenario не может зафиксировать "любая 2.x версия Procedure"; только latest published или snapshot at start

### Что сознательно НЕ реализуем

| Фича | Причина отказа |
|------|----------------|
| Semver (MAJOR.MINOR.PATCH) | YAGNI; один admin контролирует обе стороны; ни одна CRM так не делает |
| Version constraints (`^2.0`) | Предполагают независимую эволюцию потребителей; в single-tenant CRM нерелевантно |
| Backward compatibility validation | Требует формализации input/output schema; dependency check при сохранении достаточен |
| 4+ статуса (deprecated, archived) | Три статуса покрывают все потребности |
| Версирование Functions/VR/AR/OV/Layout | Простые сущности; save = live + защита при сохранении (type check, dependency check, test endpoint) |
| JSONB snapshot в Scenario run | Тяжело; FK на version_id достаточен + superseded versions защищены от удаления |

## Связанные ADR

- **ADR-0024** — Procedure Engine: Procedure definition хранится в `procedure_versions.definition`, не в `procedures` напрямую
- **ADR-0025** — Scenario Engine: аналогично; `scenario_run_snapshots` фиксирует versions Procedures при старте
- **ADR-0026** — Custom Functions: без версирования; dependency check + test endpoint при сохранении
- **ADR-0019** — Декларативная бизнес-логика: Validation Rules, Automation Rules — без версирования (CEL inline, immediate apply)
- **ADR-0022** — Object View: без версирования (конфигурация, не логика)
- **ADR-0027** — Layout: без версирования (конфигурация презентации)
