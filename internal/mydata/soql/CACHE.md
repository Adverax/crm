# SOQL Caching Architecture

## Overview

SOQL engine использует двухуровневую систему кэширования:

1. **Metadata Cache** — кэш метаданных объектов (схема, поля, связи)
2. **Query Cache** — кэш скомпилированных запросов (AST → SQL)

При изменении метаданных через PostgreSQL триггеры автоматически инвалидируются соответствующие записи в обоих кэшах.

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              SOQL Engine                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐               │
│  │    Parse     │ ───▶ │   Validate   │ ───▶ │   Compile    │               │
│  │  (SOQL→AST)  │      │  (AST+Meta)  │      │  (AST→SQL)   │               │
│  └──────────────┘      └──────┬───────┘      └──────┬───────┘               │
│                               │                      │                       │
│                               ▼                      ▼                       │
│                    ┌──────────────────┐   ┌──────────────────┐              │
│                    │  Metadata Cache  │   │   Query Cache    │              │
│                    │  (object→meta)   │   │  (soql→compiled) │              │
│                    └────────┬─────────┘   └────────┬─────────┘              │
│                             │                      │                         │
│                             │    Dependencies      │                         │
│                             │    ┌─────────────┐   │                         │
│                             └───▶│ Account     │◀──┘                         │
│                                  │ Contact     │                             │
│                                  │ User        │                             │
│                                  └─────────────┘                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Cache Components

### Metadata Cache (`MetadataAdapter`)

**Location:** `infrastructure/metadata/adapter.go`

Кэширует метаданные объектов, загруженные из `metadata.object` и `metadata.field`.

```go
type MetadataAdapter struct {
    objectCache     cache.GenericCache[string, *ObjectMeta]  // object_api_name → metadata
    objectListCache cache.GenericCache[string, []string]     // "_list_" → all object names
    ttl             time.Duration                            // default: 5 min
}
```

**Операции:**
- `GetObject(name)` — получить метаданные объекта (с lazy loading)
- `ListObjects()` — список всех объектов
- `InvalidateObject(name)` — удалить объект из кэша
- `InvalidateAll()` — очистить весь кэш

### Query Cache (`Engine.queryCache`)

**Location:** `application/engine/engine.go`

Кэширует скомпилированные запросы для повторного использования.

```go
type Engine struct {
    queryCache    cache.GenericCache[string, *CompiledQuery]  // soql_text → compiled
    queryCacheTTL time.Duration                               // default: 5 min
}
```

**Операции:**
- `Prepare(soql)` — компилирует запрос (с кэшированием)
- `ClearQueryCache()` — очистить весь кэш
- `InvalidateQuery(soql)` — удалить конкретный запрос
- `InvalidateQueriesByObject(name)` — удалить запросы, зависящие от объекта

---

## Query Dependencies

Каждый `CompiledQuery` хранит список объектов, от которых зависит:

```go
type CompiledQuery struct {
    SQL          string
    Params       []any
    Shape        *ResultShape
    Dependencies []string  // ["Account", "Contact", "User"]
}
```

**Сбор зависимостей (`collectDependencies`):**

| Источник | Пример | Зависимость |
|----------|--------|-------------|
| FROM clause | `FROM Account` | `Account` |
| Lookup field | `Account.Owner.Name` | `Account`, `User` |
| Subquery | `(SELECT Id FROM Contacts)` | `Contact` |

**Пример:**

```sql
SELECT Name, Account.Name, (SELECT Id FROM Contacts)
FROM Opportunity
WHERE Owner.IsActive = true
```

Dependencies: `["Opportunity", "Account", "Contact", "User"]`

---

## Cache Invalidation Flow

### Trigger-based Invalidation

```
┌─────────────────────┐
│  metadata.object    │
│  metadata.field     │
└──────────┬──────────┘
           │ INSERT/UPDATE/DELETE
           ▼
┌─────────────────────┐
│ trg_*_metadata_outbox│
│ (PostgreSQL Trigger)│
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ metadata.metadata_  │
│ outbox (table)      │
│ + pg_notify()       │
└──────────┬──────────┘
           │ NOTIFY 'metadata_outbox_events'
           ▼
┌─────────────────────┐
│ MetadataOutboxWorker│
│ (Go daemon)         │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ AggregatedInvalidator│
└──────────┬──────────┘
           │
     ┌─────┴─────┐
     ▼           ▼
┌─────────┐ ┌─────────┐
│Metadata │ │ Query   │
│ Cache   │ │ Cache   │
└─────────┘ └─────────┘
```

### Targeted Invalidation

При изменении объекта `Account`:

```
Event: object_changed, object_api_name = "Account"
                    │
                    ▼
        ┌───────────────────────┐
        │ InvalidateObject      │
        │ ("Account")           │
        └───────────┬───────────┘
                    │
                    ▼
        ┌───────────────────────┐
        │ InvalidateQueriesByObject│
        │ ("Account")           │
        └───────────┬───────────┘
                    │
                    ▼
        ┌───────────────────────┐
        │ queryCache.Remove(    │
        │   q => "Account" in   │
        │        q.Dependencies │
        │ )                     │
        └───────────────────────┘
```

**Результат:**
- Запросы к `Account` — удалены из кэша
- Запросы к `Account` + `Contact` — удалены из кэша
- Запросы только к `Contact` — остаются в кэше ✓

---

## Outbox Table Schema

```sql
CREATE TABLE metadata.metadata_outbox (
    id                    BIGSERIAL PRIMARY KEY,
    event_type            TEXT NOT NULL,        -- 'object_changed' | 'field_changed'
    entity_type           TEXT NOT NULL,        -- 'object' | 'field'
    entity_id             BIGINT NOT NULL,
    payload               JSONB NOT NULL,       -- {"object_api_name": "Account"}
    created_at            TIMESTAMPTZ NOT NULL,
    processed_at          TIMESTAMPTZ,
    processing_started_at TIMESTAMPTZ
);
```

**Event Types:**

| Event | Trigger Source | Payload |
|-------|---------------|---------|
| `object_changed` | `metadata.object` | `{object_api_name: "Account"}` |
| `field_changed` | `metadata.field` | `{object_api_name: "Account"}` (parent object) |

---

## Configuration

### Cache TTL

```go
// Metadata cache TTL (default: 5 minutes)
const DefaultCacheTTL = 5 * time.Minute

// Query cache TTL (default: 5 minutes)
const DefaultQueryCacheTTL = 5 * time.Minute
```

### Engine Options

```go
engine := soqlEngine.NewEngine(
    soqlEngine.WithMetadata(metadataProvider),
    soqlEngine.WithQueryCache(10 * time.Minute),  // custom TTL
    // or
    soqlEngine.WithQueryCacheDisabled(),          // disable caching
)
```

---

## DI Components

```go
// bootstrap/components.go

ComponentSOQLMetadataAdapter      // *MetadataAdapter (concrete type)
ComponentSOQLMetadataProvider     // MetadataProvider interface
ComponentSOQLEngine               // *Engine
ComponentSOQLCacheInvalidator     // CacheInvalidator interface
ComponentMetadataOutboxWorker     // *Worker (daemon)
```

---

## File Structure

```
internal/data/soql/
├── domain/
│   └── invalidator.go           # CacheInvalidator interface
├── application/engine/
│   ├── engine.go                # Engine with query cache
│   └── compiler.go              # CompiledQuery with Dependencies
└── infrastructure/
    ├── metadata/
    │   └── adapter.go           # MetadataAdapter with object cache
    ├── cache/
    │   └── invalidator.go       # AggregatedInvalidator
    └── outbox/
        └── worker.go            # MetadataOutboxWorker daemon

database/migrations/
├── 000111_metadata_outbox.up.sql   # Outbox table + triggers
└── 000111_metadata_outbox.down.sql
```

---

## Performance Considerations

### Cache Hit Ratio

При высоком cache hit ratio (>90%) большинство запросов обслуживаются из кэша:

```
Request: SELECT Id, Name FROM Account WHERE Type = 'Customer'
         │
         ▼
    ┌────────────┐
    │ Query Cache │──── HIT ────▶ Return cached CompiledQuery
    └────────────┘
         │
        MISS
         │
         ▼
    Parse → Validate → Compile → Cache → Return
```

### Targeted vs Full Invalidation

| Сценарий | Метод | Эффект |
|----------|-------|--------|
| Изменён 1 объект | `InvalidateQueriesByObject` | Удаляются только зависимые запросы |
| Массовое изменение | `ClearQueryCache` | Полная очистка |
| Нет `object_api_name` | `ClearQueryCache` | Полная очистка (fallback) |

### Memory Usage

- Metadata Cache: ~1-10 KB per object (зависит от количества полей)
- Query Cache: ~0.5-2 KB per compiled query
- Рекомендуется мониторить `cache.GetStats()` для анализа использования
