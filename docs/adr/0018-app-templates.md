# ADR-0018: App Templates вместо стандартных объектов

**Статус:** Принято
**Дата:** 2026-02-14
**Участники:** @roman_myakotin

## Контекст

Phase 6 предполагала создание hardcoded "стандартных объектов" (Account, Contact, Opportunity, Task) через seed-миграции. Однако CRM — metadata-driven платформа (ADR-0003, ADR-0007), и разные предметные области требуют разных наборов объектов:

- **Sales CRM**: Account, Contact, Opportunity, Task
- **Recruiting**: Position, Candidate, Application, Interview
- **Real Estate**: Property, Client, Showing, Deal
- **IT Service Desk**: Ticket, Asset, SLA, Knowledge Article

Жёсткое кодирование одного набора доменных entity противоречит горизонтальной архитектуре платформы. Нужен механизм, позволяющий администратору выбрать подходящий набор объектов при первом запуске.

Требования:
1. Шаблон применяется через Admin UI, не при bootstrap
2. Одноразовое применение (если `object_definitions` не пуста — блокировка)
3. MVP: 2 шаблона (Sales CRM, Recruiting)
4. Шаблоны встроены в бинарник (не внешние файлы)
5. Создание объектов/полей — через существующие ObjectService/FieldService (с DDL, constraints, share tables)

## Рассмотренные варианты

### Вариант A — Hardcoded seed-миграции

SQL-миграция, создающая стандартные объекты при `migrate up`.

**Плюсы:**
- Просто реализовать
- Объекты есть сразу после миграции

**Минусы:**
- Навязывает один домен всем пользователям
- Невозможно выбрать набор объектов
- Обход ObjectService/FieldService → нет DDL, share tables, OLS/FLS
- Миграция необратима без ручного вмешательства

### Вариант B — JSON/YAML файлы шаблонов

Шаблоны хранятся как JSON/YAML файлы, читаются при применении.

**Плюсы:**
- Легко добавлять новые шаблоны
- Можно редактировать без перекомпиляции

**Минусы:**
- Нет compile-time валидации структуры
- Нужна десериализация с обработкой ошибок
- Файлы нужно поставлять вместе с бинарником
- Сложнее тестировать

### Вариант C — Go-код embedded в бинарник (выбран)

Шаблоны определены как Go-структуры, компилируются в бинарник.

**Плюсы:**
- Compile-time type safety
- Нет внешних зависимостей (single binary)
- Легко тестировать (unit tests на структуры)
- Автодополнение в IDE

**Минусы:**
- Нужна перекомпиляция для добавления шаблонов
- Не подходит для user-defined шаблонов (но это не требование MVP)

## Решение

**Выбран вариант C: Go-код embedded в бинарник.**

### Архитектура

```
internal/platform/templates/
├── types.go         — Template, ObjectTemplate, FieldTemplate
├── registry.go      — Registry (map[string]Template)
├── all.go           — BuildRegistry() → регистрация всех шаблонов
├── applier.go       — Applier: двухпроходное создание через services
├── sales_crm.go     — SalesCRM() → Template
└── recruiting.go    — Recruiting() → Template
```

### Registry + Applier pattern

- `Registry` хранит все доступные шаблоны в `map[string]Template`
- `Applier` применяет шаблон, используя существующие `ObjectService.Create()` и `FieldService.Create()`
- Двухпроходное создание: сначала все объекты (собираем `map[apiName]UUID`), потом все поля (резолвим reference → UUID)
- После создания: OLS (full CRUD) + FLS (full RW) для SystemAdmin PS на все новые объекты/поля
- Guard: `objectRepo.Count(ctx) > 0` → `apperror.Conflict`

### API

- `GET /api/v1/admin/templates` — список шаблонов со статусом (available/applied/blocked)
- `POST /api/v1/admin/templates/:templateId/apply` — применить шаблон

## Последствия

### Позитивные
- Phase 6 scope меняется с "стандартные объекты" на "App Templates" — платформа остаётся горизонтальной
- Администратор выбирает домен при первом запуске
- Легко добавлять новые шаблоны (HR, Real Estate, IT Service Desk)
- Все объекты создаются через platform services → автоматический DDL, share tables, constraints

### Негативные
- Одноразовое применение (MVP ограничение) — нельзя применить второй шаблон
- Нет UI для кастомизации шаблона перед применением
- Нет "стандартных" объектов в классическом Salesforce-смысле (is_platform_managed=true)

### Будущие расширения
- User-defined шаблоны (JSON export/import)
- Частичное применение (выбор объектов из шаблона)
- Template marketplace
