# ADR-0016: Single-tenant архитектура

**Статус:** Принято
**Дата:** 2026-02-11
**Участники:** @roman_myakotin

## Контекст

Платформа проектируется как self-hosted CRM для B2B-компаний (ADR-0014: Open Core, AGPL).
Необходимо зафиксировать модель tenancy, поскольку она фундаментально влияет на:
- схему базы данных (shared schema vs isolated DB)
- security engine (RLS-политики, изоляция данных)
- metadata engine (table-per-object, DDL при admin-операциях — ADR-0007)
- инфраструктуру и deployment-модель

Ключевые факторы:

1. **Self-hosted фокус.** Продукт позиционируется как self-hosted: каждый клиент разворачивает собственный инстанс. Multi-tenant архитектура избыточна для этой модели.

2. **Простота для MVP.** Multi-tenant добавляет сквозную сложность: tenant_id в каждой таблице, tenant-aware кеши, tenant isolation в SOQL/DML, tenant-specific миграции. Это существенно замедлит доставку Phase 3–8.

3. **Security isolation.** B2B-клиенты требуют полной изоляции данных (compliance, GDPR, регуляторные требования). Физическая изоляция (отдельная БД на инстанс) проще для аудита и сертификации, чем логическая изоляция через row-level tenant filtering.

4. **Архитектурная совместимость.** Table-per-object (ADR-0007) генерирует DDL при создании объектов. В multi-tenant shared-schema DDL одного тенанта влияет на всех. Это создаёт блокировки и усложняет миграции.

## Рассмотренные варианты

### Вариант A — Multi-tenant: shared database, shared schema

Все тенанты в одной БД, изоляция через `tenant_id` колонку в каждой таблице.

**Плюсы:**
- Экономия ресурсов при большом количестве мелких клиентов
- Единый deployment, одна БД для обслуживания

**Минусы:**
- Сквозная сложность: `tenant_id` в каждом запросе, каждом индексе, каждом кеше
- Риск data leakage при ошибке в фильтрации
- DDL от table-per-object (ADR-0007) блокирует всех тенантов
- Metadata engine должен быть tenant-aware (отдельные object_definitions на тенанта)
- Усложняет SOQL/DML: каждый запрос обязан фильтровать по tenant_id
- Невозможно дать клиенту superuser-доступ к БД
- Аудит и compliance значительно сложнее

### Вариант B — Multi-tenant: shared database, separate schemas

Каждый тенант — отдельная PostgreSQL schema (`tenant_123.contacts`).

**Плюсы:**
- Лучшая изоляция, чем shared schema
- Нативная поддержка PostgreSQL `search_path`

**Минусы:**
- DDL от table-per-object всё ещё рискованный (блокировки на уровне каталога)
- Миграции нужно прогонять по всем схемам — O(tenants) время
- Кеши (metadata, security) должны быть per-schema
- Ограничения PostgreSQL: тысячи схем с тысячами таблиц замедляют `pg_catalog`
- Не упрощает deployment — всё равно одна БД

### Вариант C — Single-tenant: один инстанс на клиента (выбран)

Каждый клиент получает полностью изолированный инстанс приложения + БД.

**Плюсы:**
- Полная изоляция данных — trivial для compliance и аудита
- Нет `tenant_id` нигде — код проще, меньше ошибок, выше производительность
- Table-per-object DDL безопасен — влияет только на одного клиента
- Metadata engine, SOQL/DML, security caches — всё работает без tenant-awareness
- Клиент может получить superuser-доступ к своей БД
- Независимые миграции, бэкапы, масштабирование
- Естественно ложится на self-hosted модель (ADR-0014)
- Проще для MVP — можно сфокусироваться на бизнес-логике

**Минусы:**
- Больше инфраструктуры при SaaS-модели (отдельная БД на клиента)
- Нет resource sharing между клиентами
- Для managed SaaS потребуется оркестратор (Kubernetes, Terraform)

## Решение

**Выбран Вариант C — single-tenant архитектура.**

Один инстанс приложения обслуживает одну организацию. Каждый deployment включает:
- Собственный PostgreSQL-экземпляр (или отдельную БД)
- Собственный Redis
- Собственный API-сервер

### Последствия для кода

- **Нет `tenant_id`** — ни в таблицах, ни в запросах, ни в кешах
- **Metadata engine** (ADR-0003, ADR-0007) работает без изменений
- **Security engine** (ADR-0009–0013) — RLS/OLS/FLS без tenant-фильтрации
- **SOQL/DML** — запросы не содержат tenant-предикатов
- **Миграции** — стандартный `golang-migrate`, без tenant-цикла
- **Конфигурация** — через переменные окружения (`.env`), уникальные для инстанса

### Путь к SaaS (если потребуется)

Single-tenant не закрывает путь к managed cloud offering:
- **Kubernetes + Helm chart** — каждый клиент = namespace с отдельным deployment
- **Database-per-tenant** паттерн (в отличие от schema-per-tenant) дёшев в оркестрации
- **Terraform/Pulumi** — автоматизация provisioning
- Многие enterprise SaaS (Atlassian Data Center, GitLab Dedicated) используют single-tenant

## Связанные решения

- [ADR-0007: Table-per-object](0007-table-per-object-storage.md) — DDL при admin-операциях, безопасен в single-tenant
- [ADR-0009: Security architecture](0009-security-architecture-overview.md) — 3-слойная безопасность без tenant-awareness
- [ADR-0014: Open Core](0014-licensing-and-business-model.md) — self-hosted фокус подтверждает single-tenant
