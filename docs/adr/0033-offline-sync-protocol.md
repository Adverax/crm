# ADR-0033: Offline Sync Protocol

**Date:** 2026-02-25

**Status:** Proposed

**Participants:** @roman_myakotin

## Context

### Problem: field teams cannot work without internet

The CRM platform targets enterprise segments where field personnel operate
in environments with unreliable or no internet connectivity:

| Segment | Role | Offline scenario |
|---------|------|-----------------|
| **FMCG distribution** | Trade representative | Visiting 15–25 retail stores/day, many in basements or rural areas |
| **Warehouse** | Warehouse worker | Receiving/shipping in industrial zones with poor coverage |
| **Field service** | Service engineer | Maintenance on remote sites (telecom towers, elevators, medical equipment) |
| **Manufacturing** | Production operator | MES operations in factory floors ("dead zones" for WiFi/LTE) |
| **Agriculture / Mining** | Field worker | Remote locations with no connectivity at all |

Current state: **the platform requires constant internet**. All data reads go
through SOQL and all writes through DML — both server-side only.

### Primary target segment: FMCG distribution

FMCG distribution is the chosen beachhead segment because:

1. **High offline dependency** — 80% of a trade rep's work is offline-compatible.
2. **Clear daily cadence** — sync morning + evening, opportunistic mid-day.
3. **Low conflict rate** — each rep works with their own customer set.
4. **Willingness to pay** — market expects $10–30/user/month.
5. **Weak competition in our quadrant** — no open-source ERP has native offline;
   specialized SFA tools (ST Mobi, Ivy Mobility) are closed and SaaS-only.

### Why our architecture is uniquely suited for offline

| Platform capability | Offline benefit |
|---|---|
| **Metadata-driven objects** | Object/field definitions sync to client — forms render offline |
| **CEL dual-stack (cel-go + cel-js)** | Validations, defaults, conditions execute in browser without server |
| **DML pipeline (8 stages)** | Same pipeline runs on server after sync — no separate offline logic |
| **SOQL** | Subset can execute client-side against local storage |
| **Object Views** | UI config (sections, fields, actions) downloads fully — offline rendering |
| **Custom Functions (fn.*)** | cel-js executes pure functions locally |
| **Procedure Engine** | Automation rules fire server-side on sync — no duplicate logic |

Competitors (Salesforce, SAP, Oracle) build **separate offline products** (Mobile SDK,
Fiori Offline). Our architecture supports transferring business logic to the client
**natively** through the existing metadata + CEL infrastructure.

### Competitive landscape

| Solution | Offline | ERP capabilities | Open source | Price |
|----------|---------|-----------------|-------------|-------|
| **ST Mobi** | Yes (native Android) | No (SFA only) | No | $15–30/user/mo |
| **Ivy Mobility** | Yes | No (SFA only) | No | $30–50/user/mo |
| **Salesforce Field Service** | Limited | Full CRM, limited ERP | No | $150+/user/mo |
| **SAP Fiori Offline** | Partial (OData) | Full ERP | No | Enterprise pricing |
| **Odoo** | No | Full ERP | Yes (AGPL) | $15–25/user/mo |
| **ERPNext** | No | Full ERP | Yes (GPL) | $10–15/user/mo |
| **1C** | Partial (via EDT) | Full ERP | No | Custom |
| **Our platform** | **Architecture-ready** | ERP + CRM | Yes (AGPL) | Target $10–15/user/mo |

No player occupies the **open-source ERP + native offline** quadrant.

## Considered Options

### Data synchronization model

**Option A: Operation-based sync (Chosen)**
- Client queues DML operations (INSERT/UPDATE/DELETE) and pushes them to server.
- Server replays operations through the standard DML pipeline.
- Pros: Full server-side validation, security enforcement, audit trail; idempotent replay.
- Cons: Requires operation queue on client; slightly more complex than state sync.

**Option B: State-based sync (row diff)**
- Client sends changed rows; server merges via last-write-wins or CRDT.
- Pros: Simpler client implementation; no operation queue.
- Cons: Loses operation semantics (who did what); LWW can silently drop data;
  CRDTs are complex and don't map well to business documents; bypasses DML pipeline.

**Option C: Event sourcing (full event log)**
- All operations are events; server and client maintain event logs.
- Pros: Full history; replay capability.
- Cons: Massive over-engineering for FMCG use case; event storage grows unbounded;
  projection complexity; not justified at this stage.

### Conflict resolution strategy

**Option A: Domain-aware resolution (Chosen)**
- Orders are treated as **requests**, not ledger entries — server always accepts them
  but may attach warnings (low stock) or adjustments (price changes).
- Blocking conflicts (customer blocked, product discontinued) result in rejection
  with explicit error messages.
- Pros: Matches FMCG business reality; minimal data loss; clear UX.
- Cons: Requires per-domain conflict rules.

**Option B: Last-write-wins**
- Latest timestamp wins for all conflicts.
- Pros: Simple, universal.
- Cons: Silent data loss; inappropriate for transactional documents (orders, visits).

**Option C: Manual conflict resolution**
- All conflicts require user intervention.
- Pros: No data loss.
- Cons: Terrible UX for field reps; blocks workflow.

### Client-side storage

**Option A: IndexedDB via Dexie.js (Chosen)**
- Dexie.js wrapper over IndexedDB for structured storage.
- Pros: Works in PWA and any browser; no native dependencies; good query API;
  transaction support; ~5–20 MB per rep is well within IndexedDB limits.
- Cons: No SQL engine — complex queries require application-level logic.

**Option B: SQLite via wa-sqlite (WebAssembly)**
- Full SQLite compiled to WASM, running in browser.
- Pros: Real SQL engine; could run SOQL-to-SQL transpilation client-side.
- Cons: Larger bundle (~800KB WASM); more complex setup; browser compatibility concerns.

**Option C: Native SQLite (React Native / Capacitor)**
- SQLite via native bridge in a mobile app shell.
- Pros: Full SQL; better performance; larger storage limits.
- Cons: Requires native app development; two codebases (web + mobile).

IndexedDB (Option A) for MVP; SQLite WASM (Option B) for Phase 2 if client-side
SOQL execution is needed.

### Mobile delivery

**Option A: PWA (Progressive Web App) (Chosen for MVP)**
- Vue.js app with service worker, offline caching, and installable on Android/iOS.
- Pros: Single codebase (our existing Vue.js stack); no app store approval;
  instant updates; works on any device with a browser.
- Cons: Limited background sync on iOS; no native APIs beyond what PWA supports.

**Option B: React Native wrapper**
- Thin native shell around web views with native bridges.
- Pros: Better offline/background capabilities; app store presence.
- Cons: New technology; two build pipelines; React Native ≠ Vue.js.

**Option C: Native Android app**
- Dedicated Android app (Kotlin/Java).
- Pros: Best performance; full hardware access.
- Cons: Completely separate codebase; doubles development cost.

PWA for MVP. Native wrapper considered only if PWA limitations block
critical features (background GPS, camera in offline mode).

## Decision

### Data categories

All data is classified into three categories with distinct sync directions:

| Category | Direction | Frequency | Conflict model |
|----------|-----------|-----------|---------------|
| **Reference data** (Products, Customers, PriceLists, Routes) | Server → Client | 1–2×/day (morning sync) | No conflict (client readonly) |
| **Transactional data** (Orders, Visits, Returns, Payments) | Client → Server | On connectivity | Rare (domain-aware resolution) |
| **Computed data** (StockBalances, Receivables, KPIs) | Server → Client (snapshot) | 1–2×/day | No conflict (server = source of truth) |

This classification eliminates 90% of sync complexity: trade reps don't edit
reference data, and each rep works with their own customer set.

### Operation queue

The client maintains a persistent operation queue in IndexedDB:

```typescript
interface QueuedOperation {
  op_id: string              // UUID v4, generated on client — idempotency key
  type: 'INSERT' | 'UPDATE' | 'DELETE'
  object: string             // object API name
  data: Record<string, any>  // field values
  children?: QueuedOperation[]  // composition (e.g., Order + OrderItems)
  attachments?: string[]     // UUIDs of locally stored photos/files

  // Queue metadata
  status: 'pending' | 'sending' | 'applied' | 'rejected'
  created_at: string         // ISO 8601
  device_sequence: number    // monotonic counter per device — gap detection
  retry_count: number
  last_error?: string
}
```

Guarantees:
- **Ordering**: `device_sequence` ensures FIFO processing.
- **Idempotency**: server checks `op_id` against `processed_operations` table;
  duplicate = skip, return previous result.
- **Atomicity**: parent + children (e.g., Order + OrderItems) are submitted
  and processed as a single batch within one DB transaction.
- **Retry**: max 3 attempts; after that, operation requires manual intervention.

### Sync API

```
POST /api/v1/sync/push          — submit queued operations
POST /api/v1/sync/pull          — fetch reference + computed data delta
POST /api/v1/sync/metadata      — fetch object/field definitions delta
POST /api/v1/sync/attachments   — upload photos/files (multipart)
GET  /api/v1/sync/status        — last sync status for device
```

All endpoints require valid JWT. Device identification via `X-Device-ID` header.

#### Push protocol

```
Client                                  Server
  │                                       │
  │  POST /sync/push                      │
  │  {device_id, sequence, operations[]}  │
  │ ────────────────────────────────────► │
  │                                       │  For each operation:
  │                                       │  ├─ Check op_id in processed_operations
  │                                       │  │  ├─ EXISTS → skip, return cached result
  │                                       │  │  └─ NOT EXISTS → proceed
  │                                       │  ├─ Map to DML call:
  │                                       │  │  dml.Insert(object, data) / Update / Delete
  │                                       │  ├─ DML pipeline (8 stages):
  │                                       │  │  parse → resolve → defaults → validate
  │                                       │  │  → compute → compile → execute → post-exec
  │                                       │  ├─ On success: record in processed_operations
  │                                       │  └─ On failure: record rejection reason
  │                                       │
  │  {results: [{op_id, status, ...}]}    │
  │ ◄──────────────────────────────────── │
  │                                       │
  │  applied → remove from queue          │
  │  rejected → show error to user        │
  │  no response → retry next sync        │
```

#### Push request

```json
{
  "device_id": "uuid",
  "sequence": 47,
  "operations": [
    {
      "op_id": "uuid",
      "type": "INSERT",
      "object": "Order",
      "data": {
        "customer_id": "uuid",
        "order_date": "2026-02-25",
        "status": "draft"
      },
      "children": [
        {
          "op_id": "uuid",
          "type": "INSERT",
          "object": "OrderItem",
          "data": {
            "product_id": "uuid",
            "qty": 10,
            "unit_price": 25.50
          }
        }
      ],
      "created_at": "2026-02-25T09:32:15Z",
      "device_sequence": 12
    }
  ]
}
```

#### Push response

```json
{
  "results": [
    {
      "op_id": "uuid",
      "status": "applied",
      "server_id": "uuid",
      "server_sequence": 1042
    },
    {
      "op_id": "uuid",
      "status": "rejected",
      "error": {
        "code": "CREDIT_LIMIT_EXCEEDED",
        "message": "Customer credit limit exceeded",
        "details": {"current_debt": 50000, "credit_limit": 45000}
      }
    },
    {
      "op_id": "uuid",
      "status": "applied",
      "warnings": [
        {
          "code": "LOW_STOCK",
          "message": "Product CC05: ordered 60, available 30"
        }
      ]
    },
    {
      "op_id": "uuid",
      "status": "applied",
      "adjustments": [
        {
          "field": "unit_price",
          "submitted": 100.00,
          "applied": 110.00,
          "reason": "Price updated since last sync"
        }
      ]
    }
  ],
  "server_sequence": 1042
}
```

Operation result statuses:
- **applied** — operation processed successfully (may include `warnings` or `adjustments`).
- **rejected** — operation failed validation; not processed; client must notify user.

#### Pull protocol

```
Client                                  Server
  │                                       │
  │  POST /sync/pull                      │
  │  {sync_token, metadata_version}       │
  │ ────────────────────────────────────► │
  │                                       │  Delta queries:
  │                                       │  SELECT ... WHERE updated_at > :sync_token
  │                                       │  Filtered by user's route/territory
  │                                       │
  │  {sync_token, reference_data,         │
  │   computed_data, route, ...}          │
  │ ◄──────────────────────────────────── │
  │                                       │
  │  Merge into local IndexedDB           │
```

#### Pull response

```json
{
  "sync_token": "2026-02-25T06:00:00Z",
  "metadata_version": 42,
  "reference_data": {
    "Product": {
      "upsert": [
        {"id": "uuid", "name": "Coca-Cola 0.5L", "sku": "CC05", "price": 25.50}
      ],
      "delete": ["uuid-old"]
    },
    "Customer": {
      "upsert": [{"id": "uuid", "name": "Store ABC", "credit_limit": 45000}],
      "delete": []
    }
  },
  "computed_data": {
    "StockBalance": [
      {"product_id": "uuid", "warehouse_id": "uuid", "available_qty": 1500}
    ],
    "Receivable": [
      {"customer_id": "uuid", "total_debt": 50000, "overdue": 12000}
    ]
  },
  "route": {
    "date": "2026-02-25",
    "customers": ["uuid-1", "uuid-2", "uuid-3"]
  }
}
```

- **sync_token**: cursor — client sends it on next pull, server returns only
  changes after this timestamp.
- **reference_data**: delta (upsert + delete), not full snapshot.
- **computed_data**: always full snapshot (stock balances cannot be delta-synced
  because they are aggregated values).
- **metadata_version**: if changed, client must call `/sync/metadata` to fetch
  updated object/field definitions.

### Idempotency table

```sql
CREATE TABLE sync.processed_operations (
    op_id        UUID PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES iam.users(id),
    device_id    UUID NOT NULL,
    object_type  TEXT NOT NULL,
    op_type      TEXT NOT NULL,       -- INSERT / UPDATE / DELETE
    status       TEXT NOT NULL,       -- applied / rejected
    result       JSONB,              -- warnings, adjustments, error details
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_processed_ops_user ON sync.processed_operations(user_id);
CREATE INDEX idx_processed_ops_created ON sync.processed_operations(created_at);
```

TTL: 30 days. A scheduled job prunes older entries.

### Device registration

```sql
CREATE TABLE sync.devices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES iam.users(id),
    device_name     TEXT NOT NULL,         -- e.g., "Samsung Galaxy A54"
    last_sync_at    TIMESTAMPTZ,
    last_sequence   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_devices_user_name ON sync.devices(user_id, device_name);
```

`last_sequence` tracks the last received `device_sequence` to detect gaps
(missing operations due to client-side data loss).

### Conflict resolution rules (FMCG-specific)

| Conflict | Server behavior | Client UX |
|----------|----------------|-----------|
| Order exceeds stock | **Accept** order, attach `LOW_STOCK` warning | Show warning badge on order |
| Customer blocked (credit limit) | **Reject** order | Show error, rep calls manager |
| Product discontinued | **Reject** order line | Show error with product name |
| Price changed since sync | **Accept** at server price, attach `adjustment` | Show "price updated" notification |
| Duplicate op_id | **Skip**, return cached result | No-op (idempotent) |
| Device sequence gap | **Accept** operations, log gap warning | No visible impact |

Key principle: **an order is a request, not a ledger entry**. The server accepts
whenever possible and attaches metadata for the warehouse to resolve at
fulfillment time.

### Client-side local storage schema

```
IndexedDB (Dexie.js)
├── meta                              — sync state
│   ├── sync_token
│   ├── metadata_version
│   └── device_sequence
├── object_definitions                — metadata cache
├── field_definitions                 — metadata cache
├── reference/Product[]               — server → client
├── reference/Customer[]              — server → client
├── reference/PriceList[]             — server → client
├── reference/Route[]                 — server → client
├── computed/StockBalance[]           — server snapshots
├── computed/Receivable[]             — server snapshots
├── operation_queue[]                 — pending DML operations
├── local/Order[]                     — locally created records
├── local/Visit[]                     — locally created records
├── local/Return[]                    — locally created records
└── attachments/{uuid → Blob}         — photos, files
```

Estimated storage per trade rep: **5–20 MB** (1,000–5,000 SKUs,
200–500 customers, photos for one day).

### Daily sync cycle (FMCG trade rep)

```
06:00  Morning sync (mandatory, requires connectivity)
       ├── Push: all pending operations from previous day
       ├── Pull: reference data delta (products, customers, prices)
       ├── Pull: computed data (stock balances, receivables)
       ├── Pull: today's route
       └── Pull: metadata delta (if version changed)

09:00  Store visit #1: order → operation queue
10:00  Store visit #2: visit + return → operation queue
       ...
11:00  Opportunistic sync (if 3G/WiFi available)
       ├── Push pending operations
       └── Pull fresh stock balances
       ...
14:00  Opportunistic sync
       ...
17:00  Store visit #15: last order → operation queue

18:00  Evening sync (mandatory)
       ├── Push all remaining operations
       ├── Result: 0 pending operations
       └── Pull updated stock balances
```

Mandatory syncs: **2 per day** (morning + evening).
Opportunistic syncs: **as connectivity allows**, background trigger.

### Integration with existing architecture

The sync layer is a **thin adapter**, not a parallel system:

| Existing component | Sync role |
|---|---|
| **DML pipeline** | Push handler maps sync operations to `dml.Insert/Update/Delete` calls — same 8-stage pipeline, same validation, same automation rules |
| **SOQL executor** | Pull handler uses SOQL queries to fetch delta data with full RLS/FLS |
| **Metadata engine** | Metadata delta endpoint returns changed object/field definitions |
| **CEL dual-stack** | cel-js on client validates inputs offline; cel-go re-validates on server |
| **Automation Rules** | Fire on server during DML execution (post-execute stage), not on client |
| **Named Credentials** | Procedure Engine pushes confirmed orders to external ERP (1C/SAP) |

No business logic is duplicated. The client is a **form renderer + operation queue**;
the server remains the single source of truth.

### Security considerations

| Concern | Mitigation |
|---------|-----------|
| Offline data on device | Reference data is filtered by user's route/territory. No access to other reps' data |
| Lost/stolen device | JWT expiry (15 min access, 7 day refresh). Server can revoke refresh token → device cannot sync |
| Tampered operations | DML pipeline validates all operations server-side. Client cannot bypass OLS/FLS/RLS |
| Replay attacks | `op_id` idempotency prevents duplicate processing |
| Man-in-the-middle | HTTPS only. Certificate pinning recommended for mobile |
| Offline data encryption | IndexedDB encrypted at rest by OS on modern Android/iOS. Application-level encryption deferred to Phase 2 |

### Platform limits

| Parameter | Default | Rationale |
|-----------|---------|-----------|
| Max operations per push | 200 | Prevents oversized payloads |
| Max attachment size | 5 MB | Photo upload limit |
| Max attachments per push | 50 | Daily photo budget |
| Max reference data per pull | 10,000 records | Prevents memory issues on client |
| Sync token max age | 7 days | Forces full re-sync if offline too long |
| Operation retry limit | 3 | After 3 failures → manual resolution |
| Processed operations TTL | 30 days | Idempotency table cleanup |
| Max devices per user | 3 | Prevents device sprawl |

### Schema: new `sync` schema

All sync-related tables live in a dedicated `sync` schema:

```sql
CREATE SCHEMA IF NOT EXISTS sync;
```

Tables: `sync.devices`, `sync.processed_operations`.

### Implementation phases

**Phase S1: Sync infrastructure (server-side)**
- `sync` schema + migrations
- `POST /sync/push` handler (operation → DML mapping, idempotency)
- `POST /sync/pull` handler (delta queries via SOQL)
- `POST /sync/metadata` handler (metadata delta)
- Device registration
- OpenAPI spec updates

**Phase S2: Offline client (PWA)**
- IndexedDB storage layer (Dexie.js)
- Operation queue (enqueue, retry, status tracking)
- SyncManager (morning/evening/opportunistic sync orchestration)
- Offline-aware UI components (sync status indicator, pending operations badge)

**Phase S3: FMCG app template**
- App template "FMCG Distribution" (Product, Customer, StockBalance, Order, Visit, Route)
- Mobile-optimized views (route map, order form, visit checklist)
- Photo capture (shelf audit)

**Phase S4: Advanced offline**
- Client-side SOQL executor (SQLite WASM)
- Background sync (service worker)
- Application-level encryption at rest
- Attachment sync with resume (chunked upload)

## Consequences

### Positive

- **Market differentiator** — only open-source ERP with native offline sync.
- **No duplicate logic** — client queues operations, server executes through existing
  DML pipeline with full validation and security.
- **Idempotent by design** — UUID-based `op_id` ensures exactly-once processing.
- **Domain-aware conflicts** — FMCG-specific rules (orders as requests, price adjustments)
  match real business workflows.
- **Minimal server changes** — sync endpoints are adapters over existing SOQL/DML;
  no new business logic layer.
- **Architecture reusable** — the same sync protocol works for warehouse, field service,
  and manufacturing segments.

### Negative

- **Stale data between syncs** — trade rep sees morning stock balances all day;
  acceptable for FMCG but may not be for other segments.
- **PWA offline limitations on iOS** — Safari restricts background sync and storage;
  may require native wrapper for iOS users.
- **New schema and API surface** — `sync` schema, 5 new endpoints, device management.
- **Client-side complexity** — operation queue, IndexedDB management, sync state machine
  are significant frontend additions.
- **No real-time collaboration** — offline model assumes isolated users; shared editing
  of the same records requires a different approach.

### Related ADRs

- ADR-0018: App Templates (FMCG Distribution template follows Registry+Applier pattern)
- ADR-0019: Declarative business logic (CEL dual-stack enables offline validation)
- ADR-0020: DML pipeline (sync push maps directly to DML operations)
- ADR-0026: Custom Functions (fn.* via cel-js runs locally)
- ADR-0022: Object Views (UI config syncs to client for offline form rendering)
- ADR-0028: Named Credentials (post-sync integration with external ERP)
- ADR-0031: Automation Rules (fire server-side during sync push execution)