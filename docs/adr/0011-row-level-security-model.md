# ADR-0011: Row-Level Security — OWD и Sharing

**Статус:** Принято
**Дата:** 2026-02-08
**Участники:** @roman_myakotin

## Контекст

RLS определяет, какие **конкретные записи** видит пользователь. Это самый сложный
слой безопасности. Необходимо определить:
- Organization-Wide Defaults (OWD) — базовый уровень видимости записей объекта
- Механизмы расширения доступа (sharing rules, manual sharing)
- Роль иерархии ролей в видимости записей
- Физическое хранение row-level grants

## Решение

### OWD — Organization-Wide Defaults

Каждый объект имеет `visibility` — базовый уровень доступа к записям
для пользователей, имеющих OLS-доступ (Read) к объекту.

```sql
ALTER TABLE metadata.object_definitions
ADD COLUMN visibility VARCHAR(30) NOT NULL DEFAULT 'private'
CHECK (visibility IN ('private', 'public_read', 'public_read_write', 'controlled_by_parent'));
```

| Значение | Read | Edit | Семантика |
|----------|------|------|-----------|
| `private` | owner + hierarchy + sharing | owner + sharing | Максимально закрытый |
| `public_read` | все с OLS Read | owner + sharing | Все видят, edit ограничен |
| `public_read_write` | все с OLS Read | все с OLS Update | Полностью открытый |
| `controlled_by_parent` | наследуется от parent | наследуется от parent | Для composition (ADR-0005) |

`controlled_by_parent` применяется к child-объектам в composition-связи.
Доступ к child-записи определяется доступом к parent-записи.

### Role Hierarchy — только Read

Менеджер (parent role) видит записи, принадлежащие пользователям подчинённых ролей.
Доступ предоставляется **только на чтение** (permissions = 1).

Действует для OWD `private` и `public_read`:
- `private`: менеджер **видит** записи подчинённых (но не edit)
- `public_read`: все уже видят, hierarchy не добавляет ничего нового
- `public_read_write`: все уже видят и редактируют

Для edit по иерархии — используются sharing rules или manual sharing.

### Share Tables — per-object grant storage

Вместо полной материализации `(user × record)` — per-object таблицы
с компактными grants. Создаются DDL engine при создании объекта
с OWD != `public_read_write`.

```sql
CREATE TABLE obj_{name}__share (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    record_id       UUID NOT NULL REFERENCES obj_{name}(id) ON DELETE CASCADE,
    grantee_id      UUID NOT NULL REFERENCES iam.groups(id) ON DELETE CASCADE,
    access_level    INT NOT NULL DEFAULT 1,   -- bitmask: 1=R, 5=R+U, etc.
    reason          VARCHAR(30) NOT NULL
                    CHECK (reason IN ('owner', 'sharing_rule', 'territory', 'manual')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (record_id, grantee_id, reason)
);

CREATE INDEX ix_{name}__share_grantee
ON obj_{name}__share (grantee_id);
```

`grantee_id` — всегда group_id. Без polymorphic `grantee_type` (ADR-0013).
- Manual share конкретному user → grant на его personal group
- Sharing rule для роли → grant на role/role_and_subordinates group
- Единый resolution через `effective_group_members`

`reason` позволяет точечный revoke: удаление sharing rule удаляет только
записи с `reason = 'sharing_rule'`, не затрагивая manual shares.

### Sharing Rules

Хранятся в общей таблице с типом правила:

```sql
CREATE TABLE security.sharing_rules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    object_id       UUID NOT NULL REFERENCES metadata.object_definitions(id),
    rule_type       VARCHAR(20) NOT NULL
                    CHECK (rule_type IN ('owner_based', 'criteria_based')),

    -- Источник (чьи записи расшариваются) — всегда group (ADR-0013)
    source_group_id UUID NOT NULL REFERENCES iam.groups(id),

    -- Получатель (кому открывается доступ) — всегда group (ADR-0013)
    target_group_id UUID NOT NULL REFERENCES iam.groups(id),

    -- Уровень доступа
    access_level    INT NOT NULL DEFAULT 1,  -- 1=R, 5=R+U

    -- Критерий (только для criteria_based, NULL для owner_based)
    criteria_field_id   UUID REFERENCES metadata.field_definitions(id),
    criteria_operator   VARCHAR(10)
                        CHECK (criteria_operator IN ('eq', 'neq', 'in', 'gt', 'lt')),
    criteria_value      TEXT,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

Source и target — прямые FK на groups. Без polymorphic `source_type`/`target_type`.

**Owner-based:** "Записи Account, принадлежащие group(type=role, Sales), доступны group(type=role_and_subordinates, Support) с правом Read."

**Criteria-based:** "Записи Account, где `status = 'Active'`, доступны group(type=public, Partners) с правом Read."

При создании/изменении sharing rule — асинхронно генерируются grants в share tables
(через outbox pattern, ADR-0012).

### Manual Record Sharing

Прямые grants в share table с `reason = 'manual'`. Создаются через UI или API.
Не удаляются автоматически при изменении sharing rules.

### Query-time RLS

SOQL engine строит WHERE clause в зависимости от OWD объекта:

**OWD = `public_read_write`:** WHERE clause не нужен.

**OWD = `public_read`:** WHERE clause для write-операций.

**OWD = `private`:**
```sql
WHERE (
  -- 1. Owner
  t.owner_id = :user_id
  -- 2. Role hierarchy (read-down)
  OR t.owner_id IN (
    SELECT visible_owner_id
    FROM security.effective_visible_owner
    WHERE user_id = :user_id
  )
  -- 3. Share table grants (единый path через groups, ADR-0013)
  OR t.id IN (
    SELECT s.record_id
    FROM obj_{name}__share s
    WHERE s.grantee_id IN (
      SELECT group_id
      FROM security.effective_group_members
      WHERE user_id = :user_id
    )
  )
  -- 4. Territory access (Phase N)
  -- OR t.id IN (
  --   SELECT rta.record_id
  --   FROM iam.record_territory_assignment rta
  --   JOIN security.effective_user_territory eut
  --     ON eut.territory_id = rta.territory_id
  --   WHERE eut.user_id = :user_id AND rta.object_id = :object_id
  -- )
)
```

**OWD = `controlled_by_parent`:** рекурсивно проверяется доступ к parent-записи.

## Рассмотренные варианты

### Полная материализация effective_rls (отклонено)

Таблица `(user_id, object_id, record_id) → permissions`. Размер = O(users × records).
При 1000 пользователях и 1M записей = 1B строк. Не масштабируется.
Пересчёт при изменении sharing rule затрагивает миллионы строк.

### Query-time через closure tables (рассмотрено)

Без pre-materialization, только closure tables + runtime JOIN.
Работает, но `effective_visible_owner` даёт выигрыш в query-time
за счёт одного lookup вместо двух JOIN (ADR-0012).

### Share tables + effective caches (выбрано)

Компактные per-object share tables (grants, не полная матрица) +
pre-materialized helper caches. Лучший баланс между размером и скоростью.

## Последствия

- Каждый объект с OWD != `public_read_write` получает share table через DDL engine
- Sharing rules хранятся в общей таблице, grants генерируются асинхронно
- Role hierarchy даёт только Read — для edit нужен explicit grant
- SOQL engine строит WHERE clause динамически на основе OWD и кэшей
- Territory management — Phase N, но модель уже поддерживает его
- Criteria-based sharing rules с nullable полями в общей таблице (без отдельной таблицы критериев)
