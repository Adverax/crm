# ADR-0014: Лицензирование и бизнес-модель

**Статус:** Принято
**Дата:** 2026-02-08
**Участники:** @roman_myakotin

## Контекст

Проект приближается к стадии публичного релиза. Необходимо определить:
- Модель распространения (SaaS, open source, open core, source available)
- Лицензию для кода
- Границу между бесплатной и платной частями
- Техническую организацию кода с разными лицензиями в одном репозитории

Ключевые ограничения:
- Минимальный бюджет — нет ресурсов на отдельную инфраструктуру для SaaS на старте
- Security engine (OLS/FLS/RLS) глубоко интегрирован в ядро (SOQL, DML) — его нельзя вынести без серьёзного усложнения архитектуры
- Целевая аудитория — B2B компании, где compliance и юридические риски критичны

## Рассмотренные варианты

### Вариант A — Pure SaaS (закрытый код)

Полностью закрытый продукт, доступный только как облачный сервис.

Плюсы: полный контроль, простая монетизация, защита IP.
Минусы: нужен бюджет на инфраструктуру с первого дня, нет community-эффекта, высокая конкуренция с Salesforce/Bitrix24/amoCRM без differentiation.

### Вариант B — Pure Open Source (AGPL)

Весь код под AGPL v3. Монетизация через поддержку и консалтинг.

Плюсы: максимальное доверие, community contributions, быстрое adoption.
Минусы: сложно монетизировать — support-модель масштабируется линейно с людьми. Конкурент может взять код и продавать hosted-версию (AGPL обязывает открыть код, но не запрещает коммерческое использование).

### Вариант C — Open Core: AGPL + проприетарный `ee/` в одном репозитории (выбран)

Ядро (включая полный security engine) — AGPL v3. Enterprise add-ons — проприетарная лицензия в директории `ee/`. Один публичный репозиторий.

Плюсы: полнофункциональный self-hosted CRM привлекает пользователей; enterprise add-ons монетизируются через лицензии; AGPL защищает от hosted-конкурентов; proven модель (GitLab, Mattermost, Grafana); один repo — простая разработка и CI.
Минусы: проприетарный код технически доступен (но защищён юридически); нужно аккуратно маркировать границу лицензий.

### Вариант D — Source Available (BSL / ELv2)

Весь код под Business Source License или Elastic License 2.0. Запрет на конкурирующий managed service.

Плюсы: простая защита от конкурентов, весь код виден.
Минусы: не OSI-approved — community воспринимает негативно, меньше contributions, меньше доверия.

## Решение

### Модель распространения: Open Core

Один публичный репозиторий с двумя лицензиями:

| Область | Лицензия | Директория |
|---------|----------|------------|
| Ядро платформы | AGPL v3 | Всё кроме `ee/` |
| Enterprise add-ons | Adverax Commercial License | `ee/` |

### Граница бесплатной и платной частей

**AGPL v3 (бесплатно, self-hosted):**

Platform:
- Metadata engine (custom objects ≤ 20, custom fields per object ≤ 50)
- SOQL parser и executor
- DML engine
- Standard objects (contacts, accounts, deals, tasks)
- REST API (≤ 1000 req/min)
- Vue.js frontend
- Self-hosted deployment (Docker)
- Webhooks (outbound, все события, retry 3x)
- Data export (CSV)

Security:
- OLS + FLS полностью
- RLS полностью (OWD, share tables, role hierarchy, sharing rules, manual sharing)
- Groups (все 4 типа: personal, role, role_and_subordinates, public)
- Security caching (closure tables, effective caches)

Auth:
- JWT (access + refresh tokens)
- Login, register, password reset
- MFA (TOTP, WebAuthn)
- Basic login history (лог входов, дата, IP, user-agent)

**Adverax Commercial License (платно):**

Security & Access Control:
- Territory management (территориальная иерархия, territory-based groups)
- PermissionSetGroups (группировка permission sets)
- Delegated administration (делегирование admin-прав по подразделениям)
- IP whitelist / login restrictions
- Advanced session management (force logout, session policies)

Auth & Identity:
- SSO / SAML 2.0
- LDAP / Active Directory sync
- OAuth2 provider (CRM как IdP)

Compliance & Audit:
- Audit Trail (полный лог всех изменений записей)
- Field History Tracking (история изменений отдельных полей)
- Data retention policies (автоочистка, GDPR compliance)
- Security analytics (geo-аналитика входов, anomaly detection)

Automation:
- Workflow rules (field update, email alert, создание записей)
- Approval processes (цепочки согласования)
- Scheduled jobs / batch processing (фоновая обработка)

Analytics & Reporting:
- Custom reports builder (визуальный конструктор отчётов)
- Dashboards (настраиваемые дашборды, drag-and-drop)
- Scheduled report delivery (отправка отчётов по email)

Platform:
- Multi-org / multi-tenant режим
- Sandbox environments (dev/staging копия организации)
- Увеличенные лимиты: custom objects > 20, custom fields > 50, API > 1000 req/min

Services:
- Managed cloud hosting (SaaS)
- Priority support + SLA
- Professional services / onboarding

**Обоснование границы:**

- Security engine (OLS/FLS/RLS) глубоко интегрирован в SOQL/DML — разделение потребовало бы сложной plugin-архитектуры. Полный security в core.
- MFA, webhooks, CSV export, basic login history — в core для доверия и привлечения пользователей. Security by default, no vendor lock-in.
- Enterprise фичи — то, что нужно крупным компаниям: compliance (audit), advanced auth (SSO/LDAP), автоматизация, аналитика, территории.
- Лимиты в free tier (20 objects, 50 fields, 1000 req/min) достаточны для малого/среднего бизнеса. Enterprise снимает ограничения.

### Структура репозитория

Директория `ee/` зеркалирует основную структуру проекта и содержит все слои enterprise-кода: Go-пакеты, Vue-компоненты, SQL-миграции, тесты.

```
crm/
├── LICENSE                            ← AGPL v3 (default для всего)
├── internal/                          ← AGPL v3: ядро платформы
│   ├── platform/
│   │   ├── security/                  ← полный RLS/OLS/FLS
│   │   ├── metadata/
│   │   ├── soql/
│   │   └── dml/
│   ├── modules/
│   ├── handler/
│   └── service/
├── migrations/                        ← AGPL v3: core schema
├── web/                               ← AGPL v3: core frontend
│   └── src/
├── ee/                                ← Adverax Commercial License
│   ├── LICENSE                        ← проприетарная лицензия
│   ├── internal/
│   │   ├── platform/
│   │   │   ├── territory/             ← Go: territory hierarchy, territory-based groups
│   │   │   └── audit/                 ← Go: audit trail engine
│   │   ├── modules/
│   │   │   └── sso/                   ← Go: SSO / SAML / LDAP
│   │   ├── handler/                   ← Go: enterprise API endpoints
│   │   └── service/                   ← Go: enterprise business logic
│   ├── migrations/                    ← SQL: enterprise-only таблицы
│   ├── sqlc/
│   │   └── queries/                   ← SQL: enterprise queries
│   ├── web/
│   │   └── src/
│   │       ├── views/                 ← Vue: enterprise страницы
│   │       ├── components/            ← Vue: enterprise компоненты
│   │       └── stores/                ← Vue: enterprise Pinia stores
│   └── tests/
│       └── pgtap/                     ← pgTAP: enterprise schema tests
└── ...
```

### Принцип интеграции: интерфейсы в ядре, реализации в `ee/`

Ядро определяет интерфейсы (extension points). Community edition использует default-реализации (no-op / заглушки). Enterprise edition подставляет полные реализации через build tags.

```go
// internal/platform/security/rls/territory.go (ядро, AGPL)
// Интерфейс для territory-based access resolution.
type TerritoryResolver interface {
    ResolveTerritoryGroups(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

// internal/platform/security/rls/territory_default.go (ядро, AGPL)
//go:build !enterprise

// Default: territory не используется, возвращаем nil.
type noopTerritoryResolver struct{}

func (r *noopTerritoryResolver) ResolveTerritoryGroups(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
    return nil, nil
}
```

```go
// ee/internal/platform/territory/resolver.go (enterprise, проприетарная)
//go:build enterprise

// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License. See ee/LICENSE for details.

type territoryResolver struct { ... }

func (r *territoryResolver) ResolveTerritoryGroups(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
    // Полная реализация: territory hierarchy, territory-based groups,
    // effective_user_territory cache lookup.
}
```

Vue-компоненты подключаются аналогично — через dynamic imports с проверкой feature flag:

```typescript
// web/src/router/index.ts (ядро, AGPL)
const routes = [
  ...coreRoutes,
  // Enterprise routes загружаются динамически, если доступны
  ...(import.meta.env.VITE_ENTERPRISE === 'true' ? enterpriseRoutes : []),
]
```

```typescript
// ee/web/src/router/enterprise-routes.ts (enterprise, проприетарная)
export const enterpriseRoutes = [
  { path: '/admin/territories', component: () => import('../views/TerritoryManager.vue') },
  { path: '/admin/audit-log', component: () => import('../views/AuditLog.vue') },
  { path: '/admin/sso', component: () => import('../views/SSOConfig.vue') },
]
```

Enterprise-миграции запускаются отдельным migration path:

```makefile
# Community Edition migrations
migrate-up:
	migrate -path migrations/ -database $(DB_URL) up

# Enterprise Edition migrations (core + enterprise)
migrate-up-ee:
	migrate -path migrations/ -database $(DB_URL) up
	migrate -path ee/migrations/ -database $(DB_URL) up
```

### Маркировка лицензий в коде

Файлы в `ee/` содержат заголовок:

```go
// Copyright 2026 Adverax. All rights reserved.
// Licensed under the Adverax Commercial License.
// See ee/LICENSE for details.
// Unauthorized use, copying, or distribution is prohibited.
```

Файлы вне `ee/` опционально содержат:

```go
// Copyright 2026 Adverax.
// Licensed under AGPL v3. See LICENSE for details.
```

### Сборка

Enterprise features подключаются через Go build tags:

```go
// ee/territory/manager.go
//go:build enterprise

package territory
```

Два варианта сборки:

```makefile
# Community Edition (по умолчанию)
build:
	go build -o crm ./cmd/api

# Enterprise Edition
build-ee:
	go build -tags enterprise -o crm-ee ./cmd/api
```

### Юридическая защита

- **AGPL на ядро** — конкурент, хостящий модифицированную версию, обязан открыть весь свой код
- **Проприетарная лицензия на `ee/`** — использование без оплаты = нарушение copyright
- **B2B контекст** — целевые клиенты (компании) соблюдают лицензии из-за юридических рисков
- **Прецеденты** — GitLab, Mattermost, Sourcegraph успешно используют эту модель годами

### Прецеденты в индустрии

| Проект | Модель | Лицензия ядра | Enterprise |
|--------|--------|---------------|------------|
| GitLab | Open Core, один repo | MIT | `ee/` — проприетарная |
| Mattermost | Open Core, один repo | MIT + Apache 2.0 | `enterprise/` — проприетарная |
| Grafana | Open Core, один repo | AGPL v3 | Enterprise plugins — проприетарные |
| Sourcegraph | Open Core, один repo | Apache 2.0 | `enterprise/` — проприетарная |

## Последствия

- Весь security engine (OLS, FLS, RLS, groups, caching) реализуется в ядре под AGPL — без разделения
- Директория `ee/` зеркалирует основную структуру: `ee/internal/`, `ee/migrations/`, `ee/sqlc/`, `ee/web/`, `ee/tests/`
- Ядро определяет интерфейсы (extension points), community edition использует no-op заглушки (`//go:build !enterprise`), enterprise подставляет полные реализации (`//go:build enterprise`)
- Vue enterprise-компоненты подключаются через dynamic imports и feature flag `VITE_ENTERPRISE`
- Enterprise-миграции — отдельный migration path (`ee/migrations/`), запускается после core-миграций
- Файл `LICENSE` (AGPL v3) — в корне репозитория
- Файл `ee/LICENSE` (Adverax Commercial License) — в директории `ee/`
- Build tag `enterprise` используется для условной компиляции Go enterprise-кода
- Makefile получает targets: `build-ee`, `migrate-up-ee`, `test-pgtap-ee`
- Текущая разработка (Phase 2: Security engine) не затрагивается — весь security идёт в ядро
- Публичный релиз планируется после Phase 5-6 (auth + standard objects)
- Первая enterprise-фича (territory management) реализуется на Phase N
