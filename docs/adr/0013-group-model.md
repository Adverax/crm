# ADR-0013: Модель групп

**Статус:** Принято
**Дата:** 2026-02-08
**Участники:** @roman_myakotin

## Контекст

Группы — ключевой механизм sharing. Sharing rules и share tables ссылаются на получателей
доступа. Необходимо определить:
- Типы групп и правила их создания
- Модель grantee в share tables
- Lifecycle авто-генерируемых групп

## Рассмотренные варианты

### Вариант A — Auto-generated groups, единый grantee (выбран)

Все sharing идёт через groups. Share table содержит только `grantee_id` (всегда group).
Система автоматически создаёт группы для ролей и пользователей.

Плюсы: унифицированная модель — один grantee type, один кэш, один resolution path.
Минусы: auto-generation при изменении ролей/пользователей. Больше строк в group/group_member.

### Вариант B — Polymorphic grantee, без auto-generation

Groups — только ручные. Share tables имеют `grantee_type: 'user' | 'group' | 'role' | 'role_and_subordinates'`.
Для каждого типа — своя resolution logic.

Плюсы: нет auto-generation, меньше данных.
Минусы: несколько resolution paths в enforcement, сложнее WHERE clause.

## Решение

### Типы групп

| Type | Создание | `related_role_id` | Membership |
|------|----------|-------------------|------------|
| `personal` | Авто при создании User | NULL | Только этот user |
| `role` | Авто при создании UserRole | NOT NULL | Users с этой ролью |
| `role_and_subordinates` | Авто при создании UserRole | NOT NULL | Users с ролью + subordinate roles |
| `public` | Вручную админом | NULL | Вручную: users + nested groups |

Territory-based типы (`territory`, `territory_and_subordinates`) — Phase N.

### Таблица groups

```sql
CREATE TABLE iam.groups (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_name          VARCHAR(100) NOT NULL UNIQUE,
    label             VARCHAR(255) NOT NULL,
    group_type        VARCHAR(30) NOT NULL
                      CHECK (group_type IN ('personal', 'role', 'role_and_subordinates', 'public')),
    related_role_id   UUID REFERENCES iam.user_roles(id) ON DELETE CASCADE,
    related_user_id   UUID REFERENCES iam.users(id) ON DELETE CASCADE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

- `related_role_id` — NOT NULL для `role` и `role_and_subordinates`
- `related_user_id` — NOT NULL для `personal`
- Оба NULL для `public`

### Таблица group_members

```sql
CREATE TABLE iam.group_members (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id        UUID NOT NULL REFERENCES iam.groups(id) ON DELETE CASCADE,
    member_user_id  UUID REFERENCES iam.users(id) ON DELETE CASCADE,
    member_group_id UUID REFERENCES iam.groups(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (
        (member_user_id IS NOT NULL AND member_group_id IS NULL) OR
        (member_user_id IS NULL AND member_group_id IS NOT NULL)
    ),
    UNIQUE (group_id, member_user_id),
    UNIQUE (group_id, member_group_id)
);
```

Поддерживает nested groups: group может содержать users и другие groups.
`effective_group_members` (ADR-0012) раскрывает nested membership в плоский список.

### Auto-generation lifecycle

| Событие | Действие |
|---------|----------|
| Создан User | Авто-создать Group type=`personal` с `related_user_id`, добавить user как member |
| Удалён User | Каскадное удаление personal group (ON DELETE CASCADE) |
| Создана UserRole | Авто-создать Group type=`role` + Group type=`role_and_subordinates` с `related_role_id` |
| Удалена UserRole | Каскадное удаление связанных groups |
| User назначен на роль | Добавить в `role` group этой роли. Добавить во все `role_and_subordinates` groups ancestor-ролей |
| User снят с роли | Удалить из соответствующих auto-generated groups |
| Изменён parent_id роли | Пересчитать membership `role_and_subordinates` groups всех затронутых ролей |

Все изменения → outbox event → пересчёт `effective_group_members` (ADR-0012).

### Единый grantee в share tables

Share table содержит только `grantee_id` — всегда group_id:

```sql
CREATE TABLE obj_{name}__share (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    record_id       UUID NOT NULL REFERENCES obj_{name}(id) ON DELETE CASCADE,
    grantee_id      UUID NOT NULL REFERENCES iam.groups(id) ON DELETE CASCADE,
    access_level    INT NOT NULL DEFAULT 1,
    reason          VARCHAR(30) NOT NULL
                    CHECK (reason IN ('owner', 'sharing_rule', 'territory', 'manual')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (record_id, grantee_id, reason)
);
```

- Manual share конкретному user → grant на его personal group
- Sharing rule для роли → grant на role/role_and_subordinates group
- Всегда один resolution path через `effective_group_members`

### Единый WHERE clause в SOQL engine

```sql
-- RLS: share table check
t.id IN (
    SELECT s.record_id
    FROM obj_{name}__share s
    WHERE s.grantee_id IN (
        SELECT group_id
        FROM security.effective_group_members
        WHERE user_id = :user_id
    )
)
```

Один path, один кэш, один JOIN.

### Sharing rules — source/target через groups

```sql
-- Sharing rule ссылается на groups
source_group_id  UUID NOT NULL REFERENCES iam.groups(id),
target_group_id  UUID NOT NULL REFERENCES iam.groups(id),
```

Вместо `source_type + source_id` / `target_type + target_id` — прямые FK на groups.
"Записи пользователей роли Sales" → source = Group type=`role` where related_role_id = Sales.
"Доступны роли Support и подчинённым" → target = Group type=`role_and_subordinates` where related_role_id = Support.

## Последствия

- Каждый user и каждая роль автоматически получают связанные groups
- Share tables имеют единый тип grantee (group_id), без polymorphic dispatch
- `effective_group_members` — единственный кэш для RLS resolution
- Sharing rules ссылаются напрямую на groups (FK), без enum source_type/target_type
- Auto-generation управляется через outbox events
- Public groups создаются и наполняются вручную через admin UI
- Territory-based groups добавляются в Phase N без изменения модели
