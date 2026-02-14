# ADR-0021: Контрактное тестирование — OpenAPI validation + генерация TS типов

**Статус:** Принято
**Дата:** 2026-02-14
**Участники:** @roman_myakotin

## Контекст

Контракт между backend и frontend определён в `api/openapi.yaml` (OpenAPI 3.0.3, 1700+ строк). До этого решения существовали три точки рассинхронизации:

1. **OpenAPI spec ↔ Go handlers** — spec обновлялся вручную; ничто не проверяло, что реальный JSON-ответ handler'а соответствует описанной схеме. Handler-тесты проверяли только status code.
2. **OpenAPI spec ↔ TypeScript types** — типы в `web/src/types/` писались вручную и дрифтовали (уже ловили snake_case/camelCase баги).
3. **Go generated types ↔ OpenAPI spec** — `oapi-codegen` генерировал Go-типы, но без `required` arrays в spec все поля были pointer types, маскируя реальные ошибки.

Цена рассинхронизации росла: 20+ endpoints, 15+ entity schemas, 215 e2e тестов. Ручная синхронизация перестала масштабироваться.

Требования:
1. Автоматическая проверка соответствия HTTP-ответов OpenAPI-схеме в Go handler-тестах
2. Единый source of truth для TypeScript типов (без ручного дублирования)
3. Одна команда для регенерации Go и TS типов: `make generate-api`
4. Нулевые новые runtime-зависимости (только dev/test)
5. Обратная совместимость с существующими тестами

## Рассмотренные варианты

### Вариант A — Ручная синхронизация (status quo)

Продолжать вручную поддерживать соответствие spec ↔ код ↔ типы.

**Плюсы:**
- Нет дополнительной сложности
- Нет новых инструментов

**Минусы:**
- Дрифт неизбежен при росте API (20+ endpoints)
- Ошибки обнаруживаются только в runtime или e2e
- Двойная работа: описать в spec + написать TypeScript интерфейс
- Нет гарантии, что handler отдаёт то, что описано в spec

### Вариант B — Генерация OpenAPI из кода (code-first)

Генерировать OpenAPI spec из Go-структур или handler'ов (swaggo, go-swagger).

**Плюсы:**
- Spec всегда соответствует коду
- Не нужно поддерживать spec вручную

**Минусы:**
- Потеря контроля над дизайном API (spec как побочный эффект)
- Spec привязан к реализации, а не к контракту
- Сложная настройка аннотаций в коде
- Противоречит подходу spec-first, принятому в проекте

### Вариант C — Spec-first с контрактной валидацией в тестах (выбран)

OpenAPI spec остаётся source of truth. Go handler-тесты валидируют ответы против spec через `kin-openapi`. TypeScript типы генерируются из spec через `openapi-typescript`.

**Плюсы:**
- Spec = единый контракт, контролируемый разработчиком
- Автоматическое обнаружение дрифта в обе стороны (Go и TS)
- Нулевые runtime-зависимости (`kin-openapi` уже в go.mod как транзитивная)
- Существующие тесты получают контрактную проверку без изменений
- TypeScript типы всегда актуальны — ручной дрифт исключён

**Минусы:**
- Spec нужно обновлять при каждом изменении API (но это и есть цель)
- Добавление `required` arrays в spec меняет Go generated types (pointer → value), требуя одноразовых правок handler'ов
- `openapi-typescript` — новая devDependency

### Вариант D — Сторонний контрактный фреймворк (Pact, Dredd)

Использовать специализированные инструменты контрактного тестирования.

**Плюсы:**
- Развитый функционал (consumer-driven contracts, broker)
- Стандарт индустрии для микросервисов

**Минусы:**
- Overkill для monolith с единым frontend
- Дополнительная инфраструктура (broker, CI интеграция)
- Дублирует то, что `kin-openapi` даёт бесплатно
- Кривая обучения

## Решение

**Выбран вариант C: Spec-first с контрактной валидацией в тестах.**

### Архитектура

```
api/openapi.yaml                    ← Source of truth (единый контракт)
    │
    ├──► oapi-codegen               → internal/api/openapi_gen.go (Go types + routes)
    │
    ├──► openapi-typescript         → web/src/types/openapi.d.ts (TS types)
    │    └──► CamelCaseKeys<T>      → web/src/types/{metadata,auth,...}.ts (derived types)
    │
    └──► kin-openapi (в тестах)     → contractValidationMiddleware (response validation)
```

### Backend: Response validation middleware

Файл `internal/handler/testutil_contract_test.go` — общий middleware для всех handler-тестов:

- `loadSpec()` загружает OpenAPI spec один раз через `sync.Once`
- `responseCapture` оборачивает `gin.ResponseWriter` для перехвата response body
- `contractValidationMiddleware(t)` — Gin middleware, валидирующий каждый response:
  1. Находит route в spec через `gorillamux.Router`
  2. Строит `RequestValidationInput` + `ResponseValidationInput`
  3. Вызывает `openapi3filter.ValidateResponse()`
  4. При несоответствии — `t.Errorf()` с деталями

Каждый `setup*Router()` в тестах принимает `t *testing.T` и добавляет middleware:
```go
func setupRouter(t *testing.T, h *MetadataHandler) *gin.Engine {
    r := gin.New()
    r.Use(contractValidationMiddleware(t))
    // ...
}
```

### Frontend: Генерация типов из OpenAPI

1. `openapi-typescript` генерирует `web/src/types/openapi.d.ts` из spec
2. `CamelCaseKeys<T>` (`web/src/types/camelcase.ts`) конвертирует snake_case ключи в camelCase (HTTP-клиент делает это в runtime)
3. Каждый файл типов (`metadata.ts`, `auth.ts`, `validationRules.ts`, `records.ts`) экспортирует derived types:

```typescript
import type { components } from './openapi'
import type { CamelCaseKeys } from './camelcase'

export type ObjectDefinition = CamelCaseKeys<components['schemas']['ObjectDefinition']>
```

### Makefile: единая точка генерации

```makefile
generate-api:
    oapi-codegen -generate gin,types,spec -package api -o internal/api/openapi_gen.go api/openapi.yaml
    cd web && npx openapi-typescript ../api/openapi.yaml -o src/types/openapi.d.ts
```

### Жёсткость spec: required arrays

Добавлены `required` arrays в response entity schemas (`ObjectDefinition`, `FieldDefinitionSchema`, `ValidationRule`, `TokenPair`, `UserInfo`, `PaginationMeta`, `ObjectNavItem`, `ObjectDescribe`, `FieldDescribe`). Это обеспечивает:
- Go: value types вместо pointer types для обязательных полей
- TS: non-optional properties вместо `field?: type`

### Затронутые файлы

| Файл | Роль |
|------|------|
| `internal/handler/testutil_contract_test.go` | Новый — loadSpec, middleware, responseCapture |
| `internal/handler/*_test.go` (5 файлов) | setup*Router теперь принимает `t` + middleware |
| `web/src/types/camelcase.ts` | Новый — `CamelCaseKeys<T>` utility type |
| `web/src/types/openapi.d.ts` | Новый — автогенерированные типы из OpenAPI |
| `web/src/types/{metadata,auth,validationRules,records}.ts` | Derived types вместо ручных интерфейсов |
| `api/openapi.yaml` | Добавлены `required` arrays, `nullable`, `enum` |
| `Makefile` | Обновлён `generate-api`, добавлен `web-generate-types` |

## Последствия

### Позитивные
- **Дрифт обнаруживается мгновенно**: изменение spec без обновления handler'а → тест падает; изменение spec без `make generate-api` → TypeScript compilation error
- **Нулевые runtime-зависимости**: вся валидация — только в тестах
- **28 handler-тестов** автоматически получили контрактную проверку без изменений в тестовой логике
- **TypeScript типы** больше не пишутся вручную — один `make generate-api` обновляет всё
- **Smoke-тест контракта**: изменить поле в spec → и Go тесты, и TS type-check ломаются

### Негативные
- При изменении API нужно обновлять spec первым (spec-first дисциплина)
- Добавление `required` arrays в существующий spec потребовало одноразового рефакторинга Go handler'ов (pointer → value types)
- `openapi-typescript` — ещё одна devDependency во frontend

### Workflow для разработчика

1. Изменить `api/openapi.yaml` (добавить/изменить endpoint или schema)
2. `make generate-api` — регенерировать Go и TS типы
3. Обновить handler/frontend код под новые типы (компилятор подскажет)
4. `go test ./internal/handler/...` — контрактная валидация
5. `cd web && npm run type-check` — TypeScript проверка

### Будущие расширения
- Request validation middleware (валидация входящих запросов в тестах)
- CI pipeline step: `make generate-api && git diff --exit-code` — проверка, что spec и сгенерированный код синхронизированы
- Автоматическая генерация mock-данных из OpenAPI schemas для e2e тестов
