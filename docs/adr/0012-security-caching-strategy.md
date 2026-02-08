# ADR-0012: Стратегия кэширования безопасности

**Статус:** Принято
**Дата:** 2026-02-08
**Участники:** @roman_myakotin

## Контекст

Security enforcement выполняется при каждом запросе к данным. Вычисление прав
на лету (recursive CTE по иерархиям, JOIN через все PermissionSet) создаёт
неприемлемую нагрузку. Необходим слой кэширования с гарантированной консистентностью.

Ключевые требования:
- Быстрый lookup при SOQL/DML (O(1) или один JOIN)
- Корректная инвалидация при изменении прав, ролей, групп
- Размер кэшей должен быть управляемым (не O(users × records))

## Решение

### Closure Tables — иерархии

Хранят все пары (ancestor, descendant) для быстрых иерархических запросов.

#### effective_role_hierarchy

```sql
CREATE TABLE security.effective_role_hierarchy (
    ancestor_role_id    UUID NOT NULL REFERENCES iam.user_role(id) ON DELETE CASCADE,
    descendant_role_id  UUID NOT NULL REFERENCES iam.user_role(id) ON DELETE CASCADE,
    computed_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (ancestor_role_id, descendant_role_id)
);
```

Размер: O(roles²) в worst case, реально O(roles × depth). Десятки-сотни строк.

#### effective_territory_hierarchy

```sql
CREATE TABLE security.effective_territory_hierarchy (
    ancestor_territory_id   UUID NOT NULL REFERENCES iam.territory(id) ON DELETE CASCADE,
    descendant_territory_id UUID NOT NULL REFERENCES iam.territory(id) ON DELETE CASCADE,
    computed_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (ancestor_territory_id, descendant_territory_id)
);
```

Размер: O(territories × depth). Сотни-тысячи строк.

#### effective_object_hierarchy

```sql
CREATE TABLE security.effective_object_hierarchy (
    ancestor_object_id   UUID NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    descendant_object_id UUID NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    computed_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (ancestor_object_id, descendant_object_id)
);
```

Для `controlled_by_parent` OWD — быстрый поиск parent-chain.
Размер: O(objects × depth). Десятки строк.

### Flattened Group Membership

```sql
CREATE TABLE security.effective_group_members (
    group_id    UUID NOT NULL REFERENCES iam.group(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES iam.user(id) ON DELETE CASCADE,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (group_id, user_id)
);
```

Раскрывает nested groups в плоский список `(group, user)`.
Размер: O(groups × avg_members). Тысячи-десятки тысяч строк.

### Effective Visible Owners

```sql
CREATE TABLE security.effective_visible_owner (
    user_id          UUID NOT NULL REFERENCES iam.user(id) ON DELETE CASCADE,
    visible_owner_id UUID NOT NULL REFERENCES iam.user(id) ON DELETE CASCADE,
    permissions      INT NOT NULL DEFAULT 1,  -- role hierarchy = Read only
    computed_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, visible_owner_id)
);

CREATE INDEX ix_evo_user ON security.effective_visible_owner (user_id)
INCLUDE (visible_owner_id, permissions);
CREATE INDEX ix_evo_owner ON security.effective_visible_owner (visible_owner_id);
```

Pre-materialized JOIN: `effective_role_hierarchy × users`.
Пользователь A видит записи пользователя B, если роль A — предок роли B.
Permissions = 1 (Read only, ADR-0011).

Размер: O(users × avg_subordinates). Десятки тысяч строк.

### Effective User Territories

```sql
CREATE TABLE security.effective_user_territory (
    user_id      UUID NOT NULL REFERENCES iam.user(id) ON DELETE CASCADE,
    territory_id UUID NOT NULL REFERENCES iam.territory(id) ON DELETE CASCADE,
    permissions  INT NOT NULL DEFAULT 0,
    computed_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, territory_id)
);

CREATE INDEX ix_eut_user ON security.effective_user_territory (user_id)
INCLUDE (territory_id, permissions);
CREATE INDEX ix_eut_territory ON security.effective_user_territory (territory_id);
```

Все территории пользователя (прямые + транзитивные по иерархии)
с агрегированными permissions от `territory_object_default`.

Размер: O(users × avg_territories). Тысячи-десятки тысяч строк.

### Effective OLS / FLS

```sql
CREATE TABLE security.effective_ols (
    user_id     UUID NOT NULL REFERENCES iam.user(id) ON DELETE CASCADE,
    object_id   UUID NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    permissions INT NOT NULL DEFAULT 0,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, object_id)
);

CREATE TABLE security.effective_fls (
    user_id     UUID NOT NULL REFERENCES iam.user(id) ON DELETE CASCADE,
    field_id    UUID NOT NULL REFERENCES metadata.field_definitions(id) ON DELETE CASCADE,
    permissions INT NOT NULL DEFAULT 0,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, field_id)
);

CREATE TABLE security.effective_field_lists (
    user_id     UUID NOT NULL REFERENCES iam.user(id) ON DELETE CASCADE,
    object_id   UUID NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    mask        INT NOT NULL,             -- 1=readable, 2=writable
    field_names TEXT[] NOT NULL DEFAULT '{}',
    computed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, object_id, mask)
);
```

`effective_ols`: `(OR all grant PS) & ~(OR all deny PS)` (ADR-0010).
Размер: O(users × objects). Тысячи строк.

`effective_fls`: `(OR all grant PS) & ~(OR all deny PS)` (ADR-0010).
Размер: O(users × fields). Десятки-сотни тысяч строк.

`effective_field_lists`: pre-computed списки полей для API.
Размер: O(users × objects × 2). Тысячи строк.

### Outbox Pattern — инвалидация кэшей

```sql
CREATE TABLE security.security_outbox (
    id            BIGSERIAL PRIMARY KEY,
    event_type    VARCHAR(50) NOT NULL,
    entity_type   VARCHAR(50) NOT NULL,
    entity_id     UUID NOT NULL,
    payload       JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at  TIMESTAMPTZ
);

CREATE INDEX ix_outbox_unprocessed
ON security.security_outbox (created_at)
WHERE processed_at IS NULL;
```

Triggers на source-таблицы пишут events в outbox. Worker обрабатывает:

```sql
SELECT * FROM security.security_outbox
WHERE processed_at IS NULL
ORDER BY created_at
FOR UPDATE SKIP LOCKED
LIMIT 1;
```

| Событие | Инвалидирует |
|---------|-------------|
| `user_changed` (profile/role) | effective_ols, effective_fls, effective_visible_owner |
| `role_changed` (parent) | effective_role_hierarchy, effective_visible_owner |
| `group_changed` (members) | effective_group_members |
| `permission_set_changed` | effective_ols, effective_fls, effective_field_lists |
| `territory_changed` (parent/model) | effective_territory_hierarchy, effective_user_territory |
| `object_changed` (visibility/parent) | effective_object_hierarchy |

### Сводка размеров кэшей

| Кэш | Размер | Инвалидация |
|------|--------|-------------|
| effective_role_hierarchy | O(roles × depth) | Редко (структура ролей) |
| effective_territory_hierarchy | O(territories × depth) | Редко (структура территорий) |
| effective_object_hierarchy | O(objects × depth) | Редко (metadata change) |
| effective_group_members | O(groups × members) | При изменении членства |
| effective_visible_owner | O(users × subordinates) | При изменении ролей/юзеров |
| effective_user_territory | O(users × territories) | При изменении territory assignments |
| effective_ols | O(users × objects) | При изменении PS/profiles |
| effective_fls | O(users × fields) | При изменении PS/profiles |
| effective_field_lists | O(users × objects × 2) | При изменении FLS |

Ни один кэш не имеет размер O(users × records) — это ключевое отличие
от отклонённого подхода `effective_rls` (ADR-0011).

## Рассмотренные варианты

### In-memory cache в Go (отклонено для permission caches)

OLS/FLS можно было бы кэшировать в памяти Go-процесса.
Но: multi-instance deployment требует distributed invalidation (Redis pub/sub, etc.).
PostgreSQL-таблицы — single source of truth, работают для любого deployment.

Closure tables и effective_* — в PostgreSQL. Горячие данные (текущий пользователь)
могут дополнительно кэшироваться в Redis или in-memory с коротким TTL.

### Materialized Views (отклонено)

PostgreSQL materialized views не поддерживают инкрементальное обновление.
`REFRESH MATERIALIZED VIEW` полностью пересоздаёт view. Outbox + таблицы
позволяют точечное обновление затронутых строк.

## Последствия

- Все кэши — PostgreSQL таблицы в схеме `security`
- Инвалидация через outbox pattern (eventual consistency, обычно < 1 сек)
- Worker обрабатывает events последовательно с `FOR UPDATE SKIP LOCKED`
- При cold start — полный пересчёт всех кэшей
- Горячий кэш в Redis/memory — опциональная оптимизация поверх PG-таблиц
- Мониторинг: алерт если outbox queue > threshold
