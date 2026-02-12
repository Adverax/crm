# ADR-0015: Territory Management

**Статус:** Принято
**Дата:** 2026-02-11
**Участники:** @roman_myakotin

## Контекст

Territory Management — механизм назначения записей на территории (географические регионы,
продуктовые направления, вертикали) для управления видимостью и доступом. Территории
ортогональны role hierarchy: роли определяют "кто ты в организации", территории — "за
какой участок отвечаешь".

Необходимо определить:
- Архитектуру территориальных моделей (single model vs. multiple models)
- Назначение пользователей и записей на территории
- Механизм предоставления доступа через территории
- Интеграцию с существующей security-моделью (groups, share tables, effective caches)
- Границу между core (AGPL) и enterprise (ee/) кодом

Ключевые ограничения:
- Single-tenant архитектура (без `tenant_id`, ADR-0007)
- Территории — enterprise-функция (ADR-0014), весь код в `ee/`
- Минимальные изменения в core: только расширение group_type + интерфейсы
- Видимость через территории должна работать через existing share tables (ADR-0011, ADR-0013)

## Рассмотренные варианты

### Вариант A — Single Territory Model (отклонено)

Одна фиксированная иерархия территорий. Нет lifecycle, нет возможности подготовить
новую структуру без влияния на production.

Плюсы: простота реализации, нет complexity с активацией.
Минусы: нет draft/test workflow, невозможна сезонная реструктуризация, нет A/B тестирования
территориальных разбиений.

### Вариант B — Full Territory Models, ETM2-like (выбран)

Несколько именованных моделей с lifecycle (`planning` → `active` → `archived`).
Одна активная модель в любой момент. Модели в `planning` можно свободно редактировать.

Плюсы: draft/test/activate workflow; сезонная реструктуризация; подготовка при M&A;
A/B сравнение географического vs индустриального разбиения.
Минусы: complexity активации (тяжёлая транзакция). Complexity локализована в activation
service, не размазана по codebase.

### Вариант C — Territory groups с типом `territory_and_subordinates` (отклонено)

По аналогии с `role_and_subordinates` — создать два типа: `territory` и
`territory_and_subordinates`. Hierarchy propagation через group membership.

Плюсы: полная аналогия с ролевой моделью.
Минусы: невозможно обеспечить per-object access levels — группа предоставляет
одинаковый уровень доступа ко всем объектам. Territory Object Defaults требуют
разного access_level для разных объектов в одной территории. Share entries через
ancestor walk решают эту задачу с per-object гранулярностью.

## Решение

### Territory Models — lifecycle

Каждая модель имеет статус:

| Статус | Редактирование | Влияет на доступ | Переходы |
|--------|---------------|------------------|----------|
| `planning` | Полное (CRUD territories, rules, defaults) | Нет | → `active` |
| `active` | Только assignment rules, record assignments | Да | → `archived` |
| `archived` | Только чтение | Нет | — |

Инвариант: не более одной active модели в любой момент (enforced partial unique index).

### SQL-схема

Все territory-таблицы — в схеме `ee` (enterprise namespace).
Effective caches — в схеме `security` (общая конвенция, ADR-0012).

#### ee.territory_models

```sql
CREATE TABLE ee.territory_models (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    api_name      VARCHAR(100) NOT NULL UNIQUE,
    label         VARCHAR(255) NOT NULL,
    description   TEXT        NOT NULL DEFAULT '',
    status        VARCHAR(20) NOT NULL DEFAULT 'planning'
                  CHECK (status IN ('planning', 'active', 'archived')),
    activated_at  TIMESTAMPTZ,
    archived_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Не более одной active модели
CREATE UNIQUE INDEX uq_territory_models_active
ON ee.territory_models (status)
WHERE status = 'active';
```

#### ee.territories

```sql
CREATE TABLE ee.territories (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id    UUID        NOT NULL REFERENCES ee.territory_models(id) ON DELETE CASCADE,
    parent_id   UUID        REFERENCES ee.territories(id) ON DELETE CASCADE,
    api_name    VARCHAR(100) NOT NULL,
    label       VARCHAR(255) NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (model_id, api_name)
);

CREATE INDEX idx_territories_model_id ON ee.territories (model_id);
CREATE INDEX idx_territories_parent_id ON ee.territories (parent_id)
WHERE parent_id IS NOT NULL;
```

`parent_id` должен ссылаться на территорию в той же модели (enforced in service layer).

#### ee.territory_object_defaults

```sql
CREATE TABLE ee.territory_object_defaults (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    territory_id  UUID        NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    object_id     UUID        NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    access_level  VARCHAR(20) NOT NULL CHECK (access_level IN ('read', 'read_write')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (territory_id, object_id)
);

CREATE INDEX idx_territory_object_defaults_territory
ON ee.territory_object_defaults (territory_id);
```

Если у территории нет object_default для объекта — территория **не предоставляет** доступ
к записям этого объекта (даже если записи назначены на территорию).

#### ee.user_territory_assignments

```sql
CREATE TABLE ee.user_territory_assignments (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    territory_id  UUID        NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, territory_id)
);

CREATE INDEX idx_user_territory_assignments_user
ON ee.user_territory_assignments (user_id);
CREATE INDEX idx_user_territory_assignments_territory
ON ee.user_territory_assignments (territory_id);
```

M2M: пользователь может быть назначен на несколько территорий одновременно.

#### ee.record_territory_assignments

```sql
CREATE TABLE ee.record_territory_assignments (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    record_id     UUID        NOT NULL,
    object_id     UUID        NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    territory_id  UUID        NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    reason        VARCHAR(30) NOT NULL DEFAULT 'manual'
                  CHECK (reason IN ('manual', 'assignment_rule')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (record_id, object_id, territory_id)
);

CREATE INDEX idx_record_territory_record
ON ee.record_territory_assignments (record_id, object_id);
CREATE INDEX idx_record_territory_territory
ON ee.record_territory_assignments (territory_id);
```

`record_id` без FK — записи живут в разных `obj_{name}` таблицах (ADR-0007).
`object_id` необходим для определения, какой object_default применять.

#### ee.territory_assignment_rules

```sql
CREATE TABLE ee.territory_assignment_rules (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    territory_id    UUID        NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    object_id       UUID        NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    rule_order      INT         NOT NULL DEFAULT 0,
    criteria_field  VARCHAR(255) NOT NULL,
    criteria_op     VARCHAR(20) NOT NULL
                    CHECK (criteria_op IN ('eq', 'neq', 'in', 'gt', 'lt', 'contains')),
    criteria_value  TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_territory_assignment_rules_territory
ON ee.territory_assignment_rules (territory_id);
CREATE INDEX idx_territory_assignment_rules_object
ON ee.territory_assignment_rules (object_id);
```

Правила оцениваются по `rule_order` при create/update записи через DML engine.
Первое совпавшее правило на территорию побеждает.

### Механизм видимости — share entries через ancestor walk

Территориальная видимость реализуется через existing share tables (ADR-0011)
и existing `effective_group_members` (ADR-0013). Отдельный RLS path не нужен.

#### Алгоритм генерации share entries

При назначении записи R (`object_id = O`) на территорию T:

1. Построить цепочку предков: `[T, parent(T), grandparent(T), ..., root]`
2. Для каждой территории T' в цепочке:
   - Найти `territory_object_defaults` для `(T', O)` → `access_level`
   - Если object_default **существует**: создать share entry
     `(R, territory_group_T', access_level, reason='territory')`
   - Если object_default **не существует**: пропустить (нет доступа через эту территорию)

#### Пример

```
EMEA (object_default: Account → read)
└── France (object_default: Account → read_write)
    └── Paris (нет object_default для Account)
```

Запись Account #42 назначена на Paris:

| Share entry | grantee (territory group) | access_level |
|-------------|--------------------------|--------------|
| 1 | group(France) | read_write |
| 2 | group(EMEA) | read |

- Пользователь в Paris: **не видит** Account #42 через Paris (нет object_default).
  Видит только если также назначен на France или EMEA.
- Пользователь в France: видит с read_write (share entry 1).
- Пользователь в EMEA: видит с read (share entry 2).

#### RLS WHERE clause

Territory visibility проходит через existing share table path:

```sql
WHERE (
  t.owner_id = :user_id                                      -- 1. Owner
  OR t.owner_id IN (
    SELECT visible_owner_id
    FROM security.effective_visible_owner
    WHERE user_id = :user_id
  )                                                           -- 2. Role hierarchy
  OR t.id IN (
    SELECT s.record_id
    FROM obj_{name}__share s
    WHERE s.grantee_id IN (
      SELECT group_id
      FROM security.effective_group_members
      WHERE user_id = :user_id
    )
  )                                                           -- 3. Sharing (includes territory)
)
```

Share entries с `reason='territory'` участвуют в общем resolution через
`effective_group_members`. Не нужен отдельный JOIN для территорий.

### Territory groups

Один новый тип группы: `territory`. Одна группа на территорию.
Members — пользователи, напрямую назначенные на эту территорию.

Тип `territory_and_subordinates` **не нужен** — hierarchy propagation
обеспечивается share entries (ancestor walk), а не group membership.
Это даёт per-object гранулярность доступа, которую группа обеспечить не может.

#### Auto-generation

| Событие | Действие |
|---------|----------|
| Активация модели | Для каждой территории: создать Group `type='territory'`, `related_territory_id = T.id` |
| Назначение user на территорию | Добавить в `group_members` территориальной группы |
| Снятие user с территории | Удалить из `group_members` |
| Архивация модели | Удалить все territory groups (CASCADE удалит group_members, share entries) |

Все изменения → outbox event → пересчёт `effective_group_members` (ADR-0012).

### Effective caches

#### security.effective_territory_hierarchy

Closure table аналогичная `effective_role_hierarchy`:

```sql
CREATE TABLE security.effective_territory_hierarchy (
    ancestor_territory_id   UUID NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    descendant_territory_id UUID NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    depth                   INT  NOT NULL DEFAULT 0,
    computed_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (ancestor_territory_id, descendant_territory_id)
);

CREATE INDEX idx_eth_descendant
ON security.effective_territory_hierarchy (descendant_territory_id);
```

- Self entry: `(T, T, depth=0)` для каждой территории
- Ancestor entries: `(parent, T, 1)`, `(grandparent, T, 2)`, ...

Размер: O(territories * depth). Сотни-тысячи строк.

#### security.effective_user_territory

```sql
CREATE TABLE security.effective_user_territory (
    user_id        UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    territory_id   UUID NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    computed_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, territory_id)
);

CREATE INDEX idx_eut_user ON security.effective_user_territory (user_id);
CREATE INDEX idx_eut_territory ON security.effective_user_territory (territory_id);
```

Плоский список территорий пользователя (из `user_territory_assignments`).

Размер: O(users * avg_territories). Тысячи-десятки тысяч строк.

#### Инвалидация (outbox events)

| Событие | Инвалидирует |
|---------|-------------|
| Изменение иерархии территорий | `effective_territory_hierarchy` |
| Изменение user assignments | `effective_user_territory` + group memberships |
| Изменение record assignments | Share entries для записи |
| Изменение object_defaults | Share entries для всех записей территории |
| Активация/архивация модели | Полный пересчёт всех territory caches |

### Алгоритм активации модели

```
ActivateModel(newModelID):
  1. Verify: newModel.status == 'planning'
  2. Verify: newModel has at least one territory
  3. BEGIN TRANSACTION
     a. Find current active model (oldModel)
     b. If oldModel != nil:
        i.   UPDATE oldModel SET status='archived', archived_at=now()
        ii.  DELETE territory groups (CASCADE → group_members, share entries)
        iii. DELETE effective_territory_hierarchy for old territories
        iv.  DELETE effective_user_territory for old territories
        v.   DELETE share entries with reason='territory' for ALL share tables
     c. UPDATE newModel SET status='active', activated_at=now()
     d. For each territory T in newModel:
        i.   CREATE Group type='territory', related_territory_id=T.id
        ii.  For each user in user_territory_assignments(T):
             INSERT INTO group_members
     e. Rebuild effective_territory_hierarchy for newModel
     f. Rebuild effective_user_territory for all users
     g. Rebuild effective_group_members for new territory groups
     h. For each record_territory_assignment in newModel:
        i.   Ancestor walk → create share entries
  4. COMMIT
  5. Emit outbox events
```

Активация — тяжёлая операция. Для MVP: синхронно в одной транзакции.
Оптимизация: background job с progress tracking.

### Stored functions — гибридный подход Go + PL/pgSQL

#### Проблема: round-trip overhead

Текущий паттерн вычисления effective caches в проекте — pure Go:
рекурсивные вызовы в Go с отдельным запросом на каждый уровень вложенности,
INSERT в цикле по одной строке. Для security engine (десятки ролей, сотни
пользователей) это допустимо.

Territory management масштабируется иначе. Активация модели с 50 территориями,
200 пользователями и 10K записями порождает ~50K round-trips в одной транзакции:
создание групп (50), заполнение group_members (200), closure table (50 × depth),
effective_user_territory (200), share entries (10K × depth ancestor walk ×
проверка object_defaults × INSERT). При ~0.1ms/round-trip по loopback это 5 секунд
**только на latency**, не считая execution time.

#### Рассмотренные варианты

**Вариант A — Pure Go (как текущий security engine)**

Весь код — в Go. Для каждой операции отдельный SQL-запрос через sqlc.

Плюсы: единый стек, привычные unit-тесты, отладка в Go.
Минусы: ~50K round-trips при активации; O(records × depth) запросов при генерации
share entries; не масштабируется для production-объёмов.

**Вариант B — Pure PL/pgSQL (вся логика в stored procedures)**

Вся бизнес-логика, включая validation и lifecycle, — в хранимых функциях.

Плюсы: ноль round-trips, максимальная производительность.
Минусы: бизнес-логика в SQL сложна для тестирования, отладки и code review;
дублирование validation между Go (API) и PL/pgSQL; нарушение convention проекта
(handler → service → repository).

**Вариант C — Гибридный: PL/pgSQL для data-intensive ops, Go для бизнес-логики (выбран)**

Три stored functions для операций с высоким round-trip overhead.
Вся бизнес-логика, CRUD, validation, rule evaluation — в Go.

Плюсы: ~50K round-trips → ~3 вызова функций для активации; recursive CTE —
нативная сила PostgreSQL; бизнес-логика остаётся в Go (тестируемость, отладка);
чёткое разделение ответственности.
Минусы: два языка для territory engine; stored functions в миграциях (версионирование).

#### Stored functions (3 штуки)

##### 1. `ee.rebuild_territory_hierarchy(p_model_id UUID)`

Пересчитывает closure table `security.effective_territory_hierarchy` для модели.
Использует recursive CTE вместо рекурсивных Go-вызовов.

```sql
CREATE FUNCTION ee.rebuild_territory_hierarchy(p_model_id UUID)
RETURNS void AS $$
BEGIN
    DELETE FROM security.effective_territory_hierarchy
    WHERE ancestor_territory_id IN (
        SELECT id FROM ee.territories WHERE model_id = p_model_id
    );

    INSERT INTO security.effective_territory_hierarchy
        (ancestor_territory_id, descendant_territory_id, depth)
    WITH RECURSIVE closure AS (
        -- Self entries
        SELECT id AS ancestor, id AS descendant, 0 AS depth
        FROM ee.territories
        WHERE model_id = p_model_id
        UNION ALL
        -- Walk up: for each territory, add its parent as ancestor
        SELECT t.parent_id AS ancestor, c.descendant, c.depth + 1
        FROM closure c
        JOIN ee.territories t ON t.id = c.ancestor
        WHERE t.parent_id IS NOT NULL
          AND t.model_id = p_model_id
    )
    SELECT ancestor, descendant, depth FROM closure;
END;
$$ LANGUAGE plpgsql;
```

**Вместо:** Go-код, загружающий все территории, строящий parent_map в памяти,
обходящий вверх для каждого узла, и вставляющий строки по одной.
**Выигрыш:** N × depth INSERT → 1 INSERT...SELECT. Ноль round-trips.

##### 2. `ee.generate_record_share_entries(p_record_id UUID, p_object_id UUID, p_territory_id UUID, p_share_table TEXT)`

Генерирует share entries для одной записи при назначении на территорию.
Выполняет ancestor walk через closure table, проверяет object_defaults,
создаёт share entries в одном вызове.

```sql
CREATE FUNCTION ee.generate_record_share_entries(
    p_record_id    UUID,
    p_object_id    UUID,
    p_territory_id UUID,
    p_share_table  TEXT
) RETURNS void AS $$
DECLARE
    rec RECORD;
BEGIN
    FOR rec IN
        SELECT g.id AS group_id, tod.access_level
        FROM security.effective_territory_hierarchy eth
        JOIN ee.territory_object_defaults tod
            ON tod.territory_id = eth.ancestor_territory_id
            AND tod.object_id = p_object_id
        JOIN iam.groups g
            ON g.related_territory_id = eth.ancestor_territory_id
            AND g.group_type = 'territory'
        WHERE eth.descendant_territory_id = p_territory_id
    LOOP
        EXECUTE format(
            'INSERT INTO %I (record_id, group_id, access_level, reason)
             VALUES ($1, $2, $3, $4)
             ON CONFLICT (record_id, group_id, reason) DO UPDATE SET access_level = $3',
            p_share_table
        ) USING p_record_id, rec.group_id, rec.access_level, 'territory';
    END LOOP;
END;
$$ LANGUAGE plpgsql;
```

**Вместо:** Go-код с отдельными запросами: 1 SELECT ancestor chain, N SELECT
object_defaults, N INSERT share entries.
**Выигрыш:** 2 × depth запросов на запись → 1 вызов функции.
Для 10K записей: ~60K round-trips → ~10K вызовов.

##### 3. `ee.activate_territory_model(p_new_model_id UUID)`

Полная оркестрация активации модели: архивация старой, создание групп,
заполнение group_members, пересчёт всех caches, генерация share entries.

```sql
CREATE FUNCTION ee.activate_territory_model(p_new_model_id UUID)
RETURNS void AS $$
DECLARE
    v_old_model_id UUID;
    v_territory RECORD;
    v_assignment RECORD;
    v_group_id UUID;
    v_share_table TEXT;
BEGIN
    -- 1. Archive old active model (if exists)
    SELECT id INTO v_old_model_id
    FROM ee.territory_models WHERE status = 'active';

    IF v_old_model_id IS NOT NULL THEN
        UPDATE ee.territory_models
        SET status = 'archived', archived_at = now(), updated_at = now()
        WHERE id = v_old_model_id;

        -- CASCADE: delete territory groups → group_members → share entries
        DELETE FROM iam.groups
        WHERE related_territory_id IN (
            SELECT id FROM ee.territories WHERE model_id = v_old_model_id
        );

        -- Clean effective caches for old model
        DELETE FROM security.effective_territory_hierarchy
        WHERE ancestor_territory_id IN (
            SELECT id FROM ee.territories WHERE model_id = v_old_model_id
        );

        DELETE FROM security.effective_user_territory
        WHERE territory_id IN (
            SELECT id FROM ee.territories WHERE model_id = v_old_model_id
        );

        -- Clean territory share entries from all share tables
        FOR v_share_table IN
            SELECT table_name || '__share'
            FROM metadata.object_definitions
            WHERE visibility = 'private'
        LOOP
            EXECUTE format(
                'DELETE FROM %I WHERE reason = $1', v_share_table
            ) USING 'territory';
        END LOOP;
    END IF;

    -- 2. Activate new model
    UPDATE ee.territory_models
    SET status = 'active', activated_at = now(), updated_at = now()
    WHERE id = p_new_model_id;

    -- 3. Create territory groups + populate members (batch INSERT...SELECT)
    FOR v_territory IN
        SELECT id, api_name FROM ee.territories WHERE model_id = p_new_model_id
    LOOP
        INSERT INTO iam.groups (api_name, label, group_type, related_territory_id)
        VALUES (
            'territory_' || v_territory.api_name,
            (SELECT label FROM ee.territories WHERE id = v_territory.id),
            'territory',
            v_territory.id
        )
        RETURNING id INTO v_group_id;

        INSERT INTO iam.group_members (group_id, member_user_id)
        SELECT v_group_id, uta.user_id
        FROM ee.user_territory_assignments uta
        WHERE uta.territory_id = v_territory.id;
    END LOOP;

    -- 4. Rebuild effective_territory_hierarchy
    PERFORM ee.rebuild_territory_hierarchy(p_new_model_id);

    -- 5. Rebuild effective_user_territory
    INSERT INTO security.effective_user_territory (user_id, territory_id)
    SELECT uta.user_id, uta.territory_id
    FROM ee.user_territory_assignments uta
    JOIN ee.territories t ON t.id = uta.territory_id
    WHERE t.model_id = p_new_model_id;

    -- 6. Rebuild effective_group_members for territory groups
    INSERT INTO security.effective_group_members (group_id, user_id)
    SELECT gm.group_id, gm.member_user_id
    FROM iam.group_members gm
    JOIN iam.groups g ON g.id = gm.group_id
    WHERE g.group_type = 'territory'
      AND g.related_territory_id IN (
          SELECT id FROM ee.territories WHERE model_id = p_new_model_id
      );

    -- 7. Generate share entries for all record assignments
    FOR v_assignment IN
        SELECT rta.record_id, rta.object_id, rta.territory_id, od.table_name
        FROM ee.record_territory_assignments rta
        JOIN ee.territories t ON t.id = rta.territory_id
        JOIN metadata.object_definitions od ON od.id = rta.object_id
        WHERE t.model_id = p_new_model_id
    LOOP
        PERFORM ee.generate_record_share_entries(
            v_assignment.record_id,
            v_assignment.object_id,
            v_assignment.territory_id,
            v_assignment.table_name || '__share'
        );
    END LOOP;
END;
$$ LANGUAGE plpgsql;
```

**Вместо:** Go activation_service с ~50K round-trips.
**Выигрыш:** 1 вызов `SELECT ee.activate_territory_model(id)` заменяет
всю Go-оркестрацию. Вся работа — server-side, ноль network round-trips.

#### Что остаётся в Go

| Операция | Причина |
|----------|---------|
| CRUD для всех 6 таблиц | Тривиальные запросы, sqlc справляется |
| Validation (status transitions, parent same model) | Бизнес-логика, unit-тестируемая |
| Assignment rule evaluation | Интерпретация метаданных полей, сложная criteria matching |
| Outbox event dispatch | Существующий паттерн, orchestration |
| Service-level coordination | handler → service → repository |
| Вызов stored functions | Go-сервис вызывает `SELECT ee.activate_territory_model($1)` через repository |

#### Тестирование stored functions

- **pgTAP** (`ee/tests/pgtap/functions/`): unit-тесты для каждой stored function
  - `rebuild_territory_hierarchy_test.sql`: проверка closure table для дерева глубиной 1, 2, 3
  - `generate_record_share_entries_test.sql`: проверка ancestor walk с/без object_defaults
  - `activate_territory_model_test.sql`: full activation flow, архивация старой модели
- **Go integration tests** (`//go:build integration`): end-to-end через repository → function → assert DB state

### Assignment Rules — интеграция с DML

```
EvaluateAssignmentRules(objectID, recordID, recordFields):
  1. Get active model → territories
  2. For each territory T:
     a. Get rules for (T, objectID) WHERE is_active ORDER BY rule_order
     b. For each rule:
        i.   Extract field value from recordFields by criteria_field
        ii.  Apply criteria_op to value and criteria_value
        iii. If match:
             - INSERT record_territory_assignment (recordID, objectID, T, 'assignment_rule')
             - Generate share entries (ancestor walk + object_defaults)
             - break (first matching rule per territory)
```

Правила оцениваются синхронно на DML insert/update. Простые критерии для MVP.

### Минимальные изменения в core (AGPL)

1. **Migration**: расширить CHECK constraint в `iam.groups` — добавить `'territory'`
2. **Type constant**: `GroupTypeTerritory = "territory"` в `internal/platform/security/types.go`
3. **Validation**: обновить `ValidateCreateGroup` для приёма `territory`
4. **Interface**: `TerritoryResolver` в rls package (noop implementation `//go:build !enterprise`)
5. **Interface**: `TerritoryAssignmentEvaluator` в dml package (noop implementation)
6. **Outbox worker**: обработка `territory_changed` event type (делегирует в enterprise через interface)
7. **Share table DDL**: уже поддерживает `reason='territory'` — изменений не нужно

```go
// internal/platform/security/rls/territory.go (core, AGPL)
type TerritoryResolver interface {
    ResolveTerritoryGroups(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

// internal/platform/security/rls/territory_default.go (core, AGPL)
//go:build !enterprise

type noopTerritoryResolver struct{}

func (r *noopTerritoryResolver) ResolveTerritoryGroups(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
    return nil, nil
}
```

```go
// internal/platform/dml/territory.go (core, AGPL)
type TerritoryAssignmentEvaluator interface {
    EvaluateOnInsert(ctx context.Context, objectID, recordID uuid.UUID, fields map[string]interface{}) error
    EvaluateOnUpdate(ctx context.Context, objectID, recordID uuid.UUID, fields map[string]interface{}) error
}
```

### Расширение groups для enterprise

```sql
-- EE migration: добавить FK на территорию
ALTER TABLE iam.groups
  ADD COLUMN related_territory_id UUID REFERENCES ee.territories(id) ON DELETE CASCADE;

CREATE INDEX idx_iam_groups_related_territory
ON iam.groups (related_territory_id)
WHERE related_territory_id IS NOT NULL;
```

### Файловая структура ee/

```
ee/
├── internal/
│   └── platform/
│       └── territory/
│           ├── types.go                        ← типы: TerritoryModel, Territory, etc.
│           ├── inputs.go                       ← input structs
│           ├── validation.go                   ← валидация входных данных
│           ├── repository.go                   ← repository interfaces
│           ├── pg_model_repo.go                ← PG: TerritoryModelRepository
│           ├── pg_territory_repo.go            ← PG: TerritoryRepository
│           ├── pg_object_default_repo.go       ← PG: TerritoryObjectDefaultRepository
│           ├── pg_user_assignment_repo.go      ← PG: UserTerritoryAssignmentRepository
│           ├── pg_record_assignment_repo.go    ← PG: RecordTerritoryAssignmentRepository
│           ├── pg_assignment_rule_repo.go      ← PG: TerritoryAssignmentRuleRepository
│           ├── pg_effective_repo.go            ← PG: TerritoryEffectiveCacheRepository
│           ├── model_service.go                ← CRUD + lifecycle моделей
│           ├── territory_service.go            ← CRUD территорий, hierarchy ops
│           ├── object_default_service.go       ← управление object defaults
│           ├── user_assignment_service.go      ← назначение пользователей
│           ├── record_assignment_service.go    ← назначение записей
│           ├── assignment_rule_service.go      ← CRUD правил
│           ├── activation_service.go           ← логика активации (вызов stored function)
│           ├── share_generator.go              ← генерация share entries (вызов stored function)
│           ├── effective_computer.go           ← пересчёт closure table (вызов stored function)
│           ├── resolver.go                     ← TerritoryResolver implementation
│           ├── evaluator.go                    ← TerritoryAssignmentEvaluator implementation
│           ├── model_service_test.go
│           ├── territory_service_test.go
│           ├── activation_service_test.go
│           ├── share_generator_test.go
│           └── effective_computer_test.go
├── internal/
│   └── handler/
│       └── territory_handler.go                ← Gin HTTP handlers
├── migrations/
│   ├── 000001_create_territory_schema.up.sql
│   ├── 000001_create_territory_schema.down.sql
│   ├── 000002_create_effective_territory_caches.up.sql
│   ├── 000002_create_effective_territory_caches.down.sql
│   ├── 000003_alter_groups_add_territory.up.sql
│   ├── 000003_alter_groups_add_territory.down.sql
│   ├── 000004_create_territory_functions.up.sql   ← 3 stored functions
│   └── 000004_create_territory_functions.down.sql
├── sqlc/
│   └── queries/
│       ├── territory_models.sql
│       ├── territories.sql
│       ├── territory_object_defaults.sql
│       ├── user_territory_assignments.sql
│       ├── record_territory_assignments.sql
│       └── territory_assignment_rules.sql
├── tests/
│   └── pgtap/
│       ├── schema/
│       │   ├── territory_models_test.sql
│       │   ├── territories_test.sql
│       │   └── territory_effective_caches_test.sql
│       └── functions/
│           ├── rebuild_territory_hierarchy_test.sql
│           ├── generate_record_share_entries_test.sql
│           └── activate_territory_model_test.sql
└── web/
    └── src/
        ├── views/
        │   ├── TerritoryModelListView.vue
        │   ├── TerritoryModelCreateView.vue
        │   ├── TerritoryModelDetailView.vue
        │   ├── TerritoryTreeView.vue
        │   ├── TerritoryDetailView.vue
        │   └── TerritoryAssignmentRulesView.vue
        ├── components/
        │   ├── TerritoryTree.vue
        │   ├── TerritoryObjectDefaultsEditor.vue
        │   └── TerritoryAssignmentRuleForm.vue
        ├── stores/
        │   └── territory.ts
        └── router/
            └── territory-routes.ts
```

### API endpoints

```
POST   /api/v1/admin/territory/models              — создать модель
GET    /api/v1/admin/territory/models              — список моделей
GET    /api/v1/admin/territory/models/:id          — получить модель
PUT    /api/v1/admin/territory/models/:id          — обновить модель
DELETE /api/v1/admin/territory/models/:id          — удалить модель (только planning)
POST   /api/v1/admin/territory/models/:id/activate — активировать модель

POST   /api/v1/admin/territory/territories              — создать территорию
GET    /api/v1/admin/territory/territories?model_id=    — список по модели
GET    /api/v1/admin/territory/territories/:id          — получить территорию
PUT    /api/v1/admin/territory/territories/:id          — обновить территорию
DELETE /api/v1/admin/territory/territories/:id          — удалить территорию

POST   /api/v1/admin/territory/territories/:id/object-defaults        — задать object default
GET    /api/v1/admin/territory/territories/:id/object-defaults        — список object defaults
DELETE /api/v1/admin/territory/territories/:id/object-defaults/:objId — удалить

POST   /api/v1/admin/territory/territories/:id/users                  — назначить пользователя
GET    /api/v1/admin/territory/territories/:id/users                  — список пользователей
DELETE /api/v1/admin/territory/territories/:id/users/:userId          — снять пользователя

POST   /api/v1/admin/territory/territories/:id/records                — назначить запись
GET    /api/v1/admin/territory/territories/:id/records                — список записей
DELETE /api/v1/admin/territory/territories/:id/records/:recordId      — снять запись

POST   /api/v1/admin/territory/assignment-rules                       — создать правило
GET    /api/v1/admin/territory/assignment-rules?territory_id=         — список правил
PUT    /api/v1/admin/territory/assignment-rules/:id                   — обновить правило
DELETE /api/v1/admin/territory/assignment-rules/:id                   — удалить правило
```

### Фазы реализации

**Фаза 1: Core Integration Points (минимальные AGPL-изменения)**
- Core migration: `'territory'` в group_type CHECK
- `GroupTypeTerritory` constant
- Update `ValidateCreateGroup`
- `TerritoryResolver` interface (noop default)
- `TerritoryAssignmentEvaluator` interface (noop default)
- `territory_changed` handler в outbox worker

**Фаза 2: Enterprise Schema (ee/ migrations)**
- 6 territory таблиц
- Effective cache таблицы
- `related_territory_id` column на `iam.groups`
- 3 stored functions: `rebuild_territory_hierarchy`, `generate_record_share_entries`, `activate_territory_model`

**Фаза 3: Enterprise Backend**
- Types, inputs, validation
- Repository interfaces и PG implementations
- Services (Model, Territory, ObjectDefault, UserAssignment, RecordAssignment, AssignmentRule)
- Share generator, effective computer, activation service — вызовы stored functions через repository
- Assignment rule evaluator (Go, синхронно на DML)

**Фаза 4: Enterprise Handler + Routes**

**Фаза 5: Enterprise Frontend** (Pinia store, model views, tree view, detail views)

**Фаза 6: Tests** (pgTAP, Go unit, E2E)

## Последствия

- Все territory-таблицы в схеме `ee` (enterprise namespace)
- Effective caches в схеме `security` (общая конвенция, ADR-0012)
- Единственное изменение в core: `'territory'` в group_type CHECK + интерфейсы с noop defaults
- Share entries с `reason='territory'` используют unified enforcement через `effective_group_members` (ADR-0013)
- Гибридный подход Go + PL/pgSQL: 3 stored functions для data-intensive операций (closure table, share generation, activation), вся бизнес-логика в Go
- Активация модели выполняется одним вызовом `SELECT ee.activate_territory_model(id)` — ноль network round-trips
- Stored functions тестируются через pgTAP (`ee/tests/pgtap/functions/`), Go-сервисы — unit-тестами
- Assignment rules оцениваются синхронно на DML insert/update в Go; простые критерии для MVP
- Territory groups auto-generated при активации модели, удаляются при архивации
- Пересчёт share entries нужен при: назначении записи, изменении object_defaults, активации модели
- Frontend: enterprise views через dynamic imports + `VITE_ENTERPRISE` flag
- Build tag `//go:build enterprise` на всех Go файлах в `ee/`
- Один тип группы `territory` (без `territory_and_subordinates`) — hierarchy propagation через share entries
