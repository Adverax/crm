# Adverax CRM — Enterprise Edition

Проприетарные расширения платформы Adverax CRM.

> **Лицензия:** Adverax Commercial License. См. [LICENSE](LICENSE).
> Использование, копирование и распространение без лицензии запрещено.

---

## Возможности

### Territory Management

Механизм назначения записей на территории (регионы, направления, вертикали)
для управления видимостью и доступом. Территории ортогональны role hierarchy:
роли = «кто ты в организации», территории = «за какой участок отвечаешь».

**Статус:** Реализовано

**Что включено:**
- Территориальные модели с lifecycle (`planning` → `active` → `archived`)
- Иерархия территорий (дерево, closure table)
- Назначение пользователей на территории (M2M)
- Назначение записей на территории (правила + ручное)
- Object defaults — какие объекты участвуют в территориальном доступе
- Интеграция с RLS через share tables и группы типа `territory`
- Effective caches для территориальной видимости
- REST API для управления территориями

**Архитектурное решение:** [ADR-0015](../docs/adr/0015-territory-management.md)

---

### Audit Trail

Журнал изменений всех записей с полной историей «кто, что, когда изменил».

**Статус:** Запланировано

---

### SSO (Single Sign-On)

Интеграция с корпоративными identity-провайдерами (SAML 2.0, OIDC).

**Статус:** Запланировано

---

### Advanced Analytics

Расширенная аналитика, дашборды, отчёты.

**Статус:** Запланировано

---

## Активация

Enterprise-функции активируются через Go build tag:

```bash
go build -tags enterprise ./cmd/api
```

Без тега `enterprise` — сборка включает только core (AGPL v3).

Для Vue.js фронтенда — переменная окружения:

```bash
VITE_ENTERPRISE=true npm run build
```

---

## Структура

```
ee/
├── LICENSE                         ← Adverax Commercial License
├── internal/
│   ├── handler/                    ← HTTP-хендлеры enterprise API
│   │   └── territory_handler.go
│   ├── platform/
│   │   └── territory/              ← Territory Management (сервисы, репозитории)
│   └── setup/                      ← Инициализация enterprise-модулей
├── migrations/                     ← SQL-миграции (отдельная таблица ee_schema_migrations)
├── sqlc/queries/                   ← Enterprise-запросы для sqlc
├── web/src/                        ← Vue-компоненты enterprise
└── tests/
    └── pgtap/                      ← pgTAP-тесты enterprise-схемы
```

---

## Миграции

Enterprise-миграции используют отдельную таблицу `ee_schema_migrations`
и работают со схемой `ee`:

```bash
make migrate-ee-up      # Применить enterprise-миграции
make migrate-ee-down    # Откатить enterprise-миграции
make test-pgtap-ee      # Запустить pgTAP-тесты enterprise
```

---

## Контакты

По вопросам лицензирования: [adverax.com](https://adverax.com)
