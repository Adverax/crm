# ADR-0010: Модель permissions — OLS/FLS

**Статус:** Принято
**Дата:** 2026-02-08
**Участники:** @roman_myakotin

## Контекст

Необходимо определить как хранятся и вычисляются Object-Level Security (OLS) и
Field-Level Security (FLS). Ключевой вопрос: Profile и PermissionSet имеют одинаковую
структуру permissions — хранить их раздельно или унифицировать?

## Рассмотренные варианты

### Вариант A — Profile = special PermissionSet (выбран)

Profile содержит `base_permission_set_id` — ссылка на PermissionSet.
OLS/FLS хранятся **только** в таблицах PermissionSet.
Одна точка правды, один enforcement path.

Плюсы: нет дублирования логики, единый механизм вычисления.
Минусы: чуть менее очевидная ментальная модель для админа.

### Вариант B — раздельное хранение

Profile имеет свои `ProfileObjectPermissions` / `ProfileFieldPermissions`.
PermissionSet имеет свои `ObjectPermissions` / `FieldPermissions`.

Плюсы: явное разделение baseline и additive.
Минусы: две таблицы с одинаковой структурой, два пути в коде, двойная поддержка.

## Решение

### Profile как special PermissionSet

```
Profile
  +id                      UUID PK
  +api_name                VARCHAR(100) UNIQUE
  +label                   VARCHAR(255)
  +description             TEXT
  +base_permission_set_id  UUID NOT NULL → PermissionSet
  +created_at              TIMESTAMPTZ
  +updated_at              TIMESTAMPTZ

PermissionSet
  +id           UUID PK
  +api_name     VARCHAR(100) UNIQUE
  +label        VARCHAR(255)
  +description  TEXT
  +ps_type      VARCHAR(10) NOT NULL DEFAULT 'grant'
                CHECK (ps_type IN ('grant', 'deny'))
  +created_at   TIMESTAMPTZ
  +updated_at   TIMESTAMPTZ
```

- `ps_type = 'grant'` — расширяет права (default)
- `ps_type = 'deny'` — глобально подавляет права

Profile — отдельная сущность (назначается пользователю, обязателен),
его permissions живут в привязанном PermissionSet (всегда `ps_type = 'grant'`).

### OLS — Object Permissions

```
ObjectPermissions
  +id                  UUID PK
  +permission_set_id   UUID NOT NULL → PermissionSet
  +object_id           UUID NOT NULL → object_definitions
  +permissions         INT NOT NULL DEFAULT 0
  +UNIQUE (permission_set_id, object_id)
```

Bitmask `permissions`:

| Бит | Значение | Операция |
|-----|----------|----------|
| 1   | 0x01     | Read     |
| 2   | 0x02     | Create   |
| 4   | 0x04     | Update   |
| 8   | 0x08     | Delete   |

Примеры: `1` = Read only, `3` = Read + Create, `15` = Full CRUD.

### FLS — Field Permissions

```
FieldPermissions
  +id                  UUID PK
  +permission_set_id   UUID NOT NULL → PermissionSet
  +field_id            UUID NOT NULL → field_definitions
  +permissions         INT NOT NULL DEFAULT 0
  +UNIQUE (permission_set_id, field_id)
```

Bitmask `permissions`:

| Бит | Значение | Операция |
|-----|----------|----------|
| 1   | 0x01     | Read     |
| 2   | 0x02     | Write    |

Примеры: `0` = Hidden, `1` = Read only, `3` = Read + Write.

### Назначение PermissionSet пользователям

```
PermissionSetToUser
  +id                  UUID PK
  +permission_set_id   UUID NOT NULL → PermissionSet
  +user_id             UUID NOT NULL → User
  +UNIQUE (permission_set_id, user_id)
```

### Вычисление effective permissions

**Effective OLS** для пользователя на объект:

```
-- Шаг 1: собрать все grant PS (profile base + назначенные grant PS)
grants = profile.base_ps.permissions[object]
       | grant_ps1.permissions[object]
       | grant_ps2.permissions[object]
       | ...

-- Шаг 2: собрать все deny PS
denies = deny_ps1.permissions[object]
       | deny_ps2.permissions[object]
       | ...

-- Шаг 3: deny побеждает grant
effective_ols(user, object) = grants & ~denies
```

**Effective FLS** — аналогично:

```
grants = OR(all grant PS field_permissions[field])
denies = OR(all deny PS field_permissions[field])
effective_fls(user, field) = grants & ~denies
```

Если поле не упомянуто ни в одном grant PS — доступ `0` (Hidden).
Deny PS на поле, которое не в grant — не имеет эффекта.

**Пример:**

```
Profile base PS:       Account = 15 (CRUD)
Grant PS "Sales":      Account = 15 (CRUD)
Deny PS "No Delete":   Account = 8  (Delete)

grants  = 15 | 15 = 15  (0b1111)
denies  = 8             (0b1000)
effective = 15 & ~8 = 7 (0b0111) → Read + Create + Update, NO Delete
```

### Кэширование

Effective permissions кэшируются в таблицах `effective_ols`, `effective_fls`,
`effective_field_lists` (см. ADR-0012). Инвалидация через outbox pattern.

## Последствия

- Profile и PermissionSet используют общие таблицы `ObjectPermissions` / `FieldPermissions`
- Единый enforcement path: собрать grant PS → OR, собрать deny PS → OR, результат = `grants & ~denies`
- Bitmask-кодирование позволяет эффективное вычисление и хранение
- Deny PS глобально подавляет права из любого источника (ADR-0009)
- Deny применяется только к OLS/FLS; RLS (sharing) остаётся строго аддитивным
- PermissionSetGroup (контейнер PS) — отложен до Phase 2b
- Назначение PS на группы (`PermissionSetToGroup`) — отложено до Phase 2b
