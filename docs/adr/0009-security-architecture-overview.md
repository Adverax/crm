# ADR-0009: Архитектура безопасности — обзор

**Статус:** Принято
**Дата:** 2026-02-08
**Участники:** @roman_myakotin

## Контекст

CRM-платформа требует enterprise-grade систему безопасности, контролирующую доступ
на трёх уровнях: объект, поле, запись. Модель вдохновлена архитектурой Salesforce,
адаптированной под наш metadata-driven подход (ADR-0003, ADR-0007).

Ключевые требования:
- Каждая операция с данными проходит проверку безопасности
- Администратор настраивает доступ без изменения кода
- Модель расширяема (территории, группы, sharing rules) без ломающих изменений

## Решение

### Три слоя безопасности

| Слой | Вопрос | Гранулярность |
|------|--------|---------------|
| **OLS** (Object-Level Security) | Может ли пользователь выполнить CRUD над этим типом объекта? | user × object |
| **FLS** (Field-Level Security) | Может ли пользователь читать/писать это поле? | user × field |
| **RLS** (Row-Level Security) | Может ли пользователь видеть эту конкретную запись? | user × record |

Порядок проверки: OLS → FLS → RLS. Если OLS запрещает — FLS и RLS не проверяются.

### Модель permissions: Grant + Deny

Два типа PermissionSet:

- **Grant PS** (default) — расширяет права
- **Deny PS** — глобально подавляет права

Источники прав:

- Profile (базовый Grant PS) даёт baseline
- Grant PermissionSet **добавляет** права поверх Profile
- Deny PermissionSet **отнимает** права у результата
- Sharing Rules **открывают** доступ к записям сверх OWD
- Manual Share / Territory **открывают** доступ к конкретным записям

Вычисление OLS/FLS:

```
grants  = profile_base_ps | grant_ps1 | grant_ps2 | ...   (bitwise OR)
denies  = deny_ps1 | deny_ps2 | ...                       (bitwise OR)
effective = grants & ~denies
```

Deny всегда побеждает Grant. Порядок назначения PS не имеет значения.
Deny применяется только к OLS/FLS. RLS (sharing) остаётся строго аддитивным.

### Модель пользователя

```
User
  +id           UUID PK
  +profile_id   UUID NOT NULL → Profile
  +role_id      UUID → UserRole (nullable, один на пользователя)
  +...
```

- **Один профиль** на пользователя (обязателен) — определяет baseline OLS/FLS
- **Одна роль** на пользователя (опциональна) — определяет позицию в иерархии для RLS
- Роль хранится прямо в `User.role_id`, без junction-таблицы

### Enforcement flow

```
HTTP Request
  │ [JWT → UserContext (user_id, profile_id, role_id)]
  ▼
Handler
  │
  ▼
Service → SOQL (read) / DML (write)
  │
  ├─ OLS: profile + permission sets → может ли CRUD на объект?
  ├─ FLS: profile + permission sets → какие поля доступны?
  ├─ RLS: OWD + hierarchy + sharing → какие записи видны?
  │
  ▼
Repository (parameterized SQL)
  │
  ▼
PostgreSQL
```

SOQL/DML — единственная точка enforcement. Прямой доступ к БД из handlers/services запрещён.

### Фазирование

| Фаза | Компоненты |
|------|------------|
| Phase 2a | User, Profile, Grant/Deny PermissionSet, OLS, FLS, effective caches (OLS/FLS), outbox worker |
| Phase 2b | UserRole (hierarchy), OWD, share tables, effective_role_hierarchy, effective_visible_owner, sharing rules, manual sharing, SOQL WHERE injection |
| Phase 2c | Groups (personal, role, role_and_subordinates, public), auto-generation, effective_group_members, share → group resolution |
| Phase 3+ | PermissionSetGroup, PS/PSG → Group assignments |
| Phase N | Territory Management, Territory-based Groups, Audit Trail, Auth (JWT) |

Auth (JWT, login, register) — отдельная фаза. До интеграции Auth используется
dev middleware с `X-Dev-User-Id` header. Enforcement engine работает с абстракцией
`UserContext`, точка стыковки — одна middleware-замена.

Audit Trail — отдельная фаза. Интегрируется как consumer outbox events.
Существующий код не затрагивается.

## Рассмотренные варианты

### User.role — junction table vs прямой FK

**Вариант A — junction table `UserRoleAssignment`:**
Позволяет multiple roles на пользователя. Но создаёт неоднозначность:
какая роль определяет видимость записей при подъёме по иерархии?

**Вариант B — прямой FK `User.role_id` (выбран):**
Один пользователь = одна роль. Простая, однозначная модель.
Salesforce использует этот подход.

### Deny-правила: Muting PS vs Global Deny vs Трёхзначная логика

**Вариант A — Muting PS (Salesforce):** Deny только внутри PermissionSetGroup,
не влияет на другие источники. Ограниченный scope.

**Вариант B — Global Deny PS (выбран):** Deny PS глобально подавляет права
из любого источника. Формула: `effective = grants & ~denies`. Полный контроль.

**Вариант C — Трёхзначная логика (Grant/Deny/Unset):** Каждый permission — 2 бита.
Максимальная гибкость, но удваивает размер bitmask и усложняет диагностику.

## Последствия

- Каждый запрос к данным проходит три проверки (OLS → FLS → RLS)
- SOQL/DML engine — единственный путь доступа к данным
- Права вычисляются как `(OR all grants) & ~(OR all denies)`
- Один профиль и одна роль на пользователя
- Security engine развивается инкрементально без ломающих изменений
- Детали OLS/FLS — ADR-0010, RLS — ADR-0011, кэширование — ADR-0012
