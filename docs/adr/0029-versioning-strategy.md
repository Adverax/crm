# ADR-0029: Versioning Strategy for Scripted Entities

**Status:** Accepted
**Date:** 2026-02-15
**Participants:** @roman_myakotin

## Context

### Problem: Modifying Live Scripted Entities

The platform contains several entity types with executable logic:

| Entity | DSL | Called from | Long-running? |
|--------|-----|-------------|---------------|
| Procedure (ADR-0024) | JSON + CEL | Scenario, Automation Rules, UI actions | No (seconds) |
| Scenario (ADR-0025) | JSON + CEL | Triggers, manual start | **Yes** (minutes-days) |
| Custom Function (ADR-0026) | CEL expression | Any CEL context (inline) | No |
| Validation Rule | CEL expression | DML pipeline (inline) | No |
| Automation Rule | Trigger config | DML pipeline | No |
| Object View (ADR-0022) | Config JSONB | Describe API | No |
| Layout (ADR-0027) | Config JSONB | Describe API | No |
| Named Credential (ADR-0028) | Config | Procedure (integration.http) | No |

When an admin modifies a Procedure, three problems arise:

**1. Mid-flight consistency**
A Scenario is running and waiting for a signal (days). While waiting, the admin updated the Procedure. On the next step the Scenario will call the new version with a different contract, causing a failure.

**2. Unsafe deployment**
Save = live. An error in a Procedure goes straight to production. There is no way to test (dry-run) before publishing.

**3. No rollback**
After publishing an erroneous version — only manual editing. No one-button "rollback".

### Not All Entities Are Equal

The problems are **not equally critical** for different entity types:

| Problem | Procedure | Scenario | Function | VR / AR / OV / Layout / Credential |
|---------|-----------|----------|----------|-------------------------------------|
| Mid-flight | Low risk (synchronous, seconds) | **High** (async, days) | None (inline) | None |
| Unsafe deploy | **Yes** (complex JSON DSL) | **Yes** (complex JSON DSL) | Low (single CEL expression, test endpoint) | None (simple configuration) |
| No rollback | **Yes** | **Yes** | Low (editing a single line) | None |

**Conclusion:** Full versioning is needed only for Procedure and Scenario. Other entities are either too simple (CEL expression) or are configuration (OV, Layout, Credential).

## Considered Options

### Option A — No Versioning

Save = live for all entities. Scenario snapshots the definition (JSONB copy) at start.

**Pros:**
- Maximum simplicity: one table per entity
- No additional logic (draft/publish, version resolution)

**Cons:**
- No testing before publishing: an error in a Procedure goes straight to production
- No change history: who changed what is unknown
- No rollback: only manual editing back
- JSONB snapshot is heavy: duplicating the full definition in each Scenario run

### Option B — Draft/Published Without Semver (chosen)

Two states for Procedure and Scenario. One draft (editable, testable), one published (immutable, live). A simple auto-increment version counter (1, 2, 3...). Other entities — no versioning.

**Pros:**
- Testing before publishing (dry-run on draft)
- Rollback to previous published version — a single operation
- Change history (who, when, what)
- Scenario snapshot = FK to version_id (not JSONB copy)
- Salesforce-aligned: Flows use exactly this approach (active/inactive versions)
- Differentiation: complexity only where justified
- Cognitive load for admin is minimal: "Save draft" -> "Publish"

**Cons:**
- Additional `_versions` table for Procedure and Scenario
- Two pointers (draft_version_id, published_version_id) instead of inline definition
- Publish/rollback logic in the service layer

### Option C — Full Semver (MAJOR.MINOR.PATCH)

Semantic versioning with version constraints (`^2.0`, `~2.3`), backward compatibility validation, 4 statuses (draft/published/deprecated/archived), retention policy, snapshot tables.

**Pros:**
- Granular control: a Scenario can pin `^2.0` and receive only compatible updates
- Enterprise-grade: maximum safety

**Cons:**
- Enormous complexity: semver parser, constraint matcher, compatibility checker, 4 statuses, retention policy, snapshot tables
- YAGNI: one admin manages both Procedures and Scenarios — they know what they are changing
- Cognitive load: admin must understand semver, breaking changes, constraints
- No CRM platform uses semver for business logic (Salesforce Flows, Dynamics Power Automate, HubSpot Workflows — all use active/inactive)
- Version constraints (`^2.0`) assume independent evolution of consumers — in our single-tenant CRM the admin controls both sides
- Backward compatibility validation requires formalizing input/output schema — an additional layer of complexity

### Option D — Immutable + Latest

Each save creates a new immutable version. Current = latest. No draft.

**Pros:**
- Full history (every save is preserved)
- Simple model: write-only, no state transitions

**Cons:**
- No draft/dry-run: each save is immediately live
- Data bloat: dozens of versions per entity with frequent editing
- No explicit "publish moment" — everything is automatically live

## Decision

**Option B chosen: Draft/Published with differentiation by entity type.**

### Differentiation

| Entity | Versioning | Rationale |
|--------|------------|-----------|
| **Procedure** | Draft/Published | Complex JSON DSL; dry-run needed; called from Scenario |
| **Scenario** | Draft/Published | Complex JSON DSL; long-running; snapshot version at start |
| **Custom Function** | None | Single CEL expression; test endpoint on save; dependency check protects against breaking changes |
| **Validation Rule** | None | CEL expression, immediate apply; error -> DML returns error |
| **Automation Rule** | None | Trigger config; the Procedure it calls has its own draft/published |
| **Object View** | None | Presentation configuration, not logic |
| **Layout** | None | Presentation configuration |
| **Named Credential** | None | Connection configuration |

### Draft/Published Model

```
              Save draft             Publish
    +-------------------+    +------------------+
    |                   v    |                  v
    |              +----+----++           +----------+
    +--------------|  draft   |---------->|published |
     (re-save)     |(editable) |  Publish   |(immutable)|
                   +----------+           +----------+
                        |                      |
                        |                      | (on new Publish ->
                        |                      |  previous published
                   Delete draft                |  gets status
                        |                      |  "superseded")
                        v                      v
                   (deleted)              +-----------+
                                         |superseded |
                                         |(read-only) |
                                         +-----------+
```

**Three statuses:**

| Status | Description | Executable? | Editable? |
|--------|-------------|-------------|-----------|
| `draft` | Work in progress | Dry-run only | Yes |
| `published` | Active version | Yes | No |
| `superseded` | Previous version (replaced) | No (except running Scenario instances) | No |

No `deprecated` or `archived` from Option C — three statuses cover all needs.

### Storage

**Procedure:**

```sql
-- Procedure metadata (without definition)
CREATE TABLE metadata.procedures (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code                 VARCHAR(100) UNIQUE NOT NULL,
    name                 VARCHAR(255) NOT NULL,
    description          TEXT,
    draft_version_id     UUID,
    published_version_id UUID,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Procedure versions (definition lives here)
CREATE TABLE metadata.procedure_versions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    procedure_id   UUID NOT NULL REFERENCES metadata.procedures(id) ON DELETE CASCADE,
    version        INT NOT NULL,               -- auto-increment per procedure (1, 2, 3...)
    definition     JSONB NOT NULL,             -- JSON DSL
    status         VARCHAR(20) NOT NULL DEFAULT 'draft',  -- draft | published | superseded
    change_summary TEXT,                       -- what changed
    created_by     UUID,                       -- who created
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at   TIMESTAMPTZ,               -- when published (NULL for draft)
    CONSTRAINT procedure_versions_unique UNIQUE (procedure_id, version),
    CONSTRAINT procedure_versions_status_check CHECK (status IN ('draft', 'published', 'superseded'))
);
```

**Scenario — analogous:**

```sql
CREATE TABLE metadata.scenarios (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code                 VARCHAR(100) UNIQUE NOT NULL,
    name                 VARCHAR(255) NOT NULL,
    description          TEXT,
    draft_version_id     UUID,
    published_version_id UUID,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE metadata.scenario_versions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scenario_id   UUID NOT NULL REFERENCES metadata.scenarios(id) ON DELETE CASCADE,
    version       INT NOT NULL,
    definition    JSONB NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'draft',
    change_summary TEXT,
    created_by    UUID,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at  TIMESTAMPTZ,
    CONSTRAINT scenario_versions_unique UNIQUE (scenario_id, version),
    CONSTRAINT scenario_versions_status_check CHECK (status IN ('draft', 'published', 'superseded'))
);
```

**Scenario Run snapshot (for mid-flight consistency):**

```sql
-- Scenario run stores FK to specific Procedure versions
CREATE TABLE metadata.scenario_run_snapshots (
    scenario_run_id       UUID NOT NULL,
    procedure_id          UUID NOT NULL,
    procedure_version_id  UUID NOT NULL REFERENCES metadata.procedure_versions(id),
    PRIMARY KEY (scenario_run_id, procedure_id)
);
```

When a Scenario run starts:
1. For each `flow.call` in the definition — resolve the current Procedure's `published_version_id`
2. Write to `scenario_run_snapshots`
3. During step execution — use the pinned version, not the current published one

### Version — Auto-increment Integer

```
Version 1 -> draft -> published
Version 2 -> draft -> published (Version 1 -> superseded)
Version 3 -> draft -> published (Version 2 -> superseded)
```

No semver (MAJOR.MINOR.PATCH). No version constraints (`^2.0`). A simple increment, understandable to any admin.

### Workflow

#### Creating a New Procedure

```
1. Admin creates Procedure -> Version 1 (draft)
2. Admin edits draft (save N times — same draft, not a new version)
3. Admin tests: dry-run on draft
4. Admin publishes -> Version 1 (published)
   -> procedures.published_version_id = version_1.id
   -> procedures.draft_version_id = NULL
```

#### Updating an Existing One

```
1. Admin clicks "Edit" -> Version 2 (draft) is created as a copy of Version 1
2. Admin makes changes to the draft
3. Admin tests: dry-run on draft
4. Admin publishes -> Version 2 (published), Version 1 (superseded)
   -> procedures.published_version_id = version_2.id
   -> procedures.draft_version_id = NULL
```

#### Rollback

```
1. Admin clicks "Rollback" on Procedure
2. Current published (Version 3) -> superseded
3. Previous superseded (Version 2) -> published
   -> procedures.published_version_id = version_2.id
4. UI shows: "Rolled back to version 2"
```

#### Deleting a Draft

```
1. Admin clicks "Discard draft"
2. Draft version is deleted
3. procedures.draft_version_id = NULL
4. Published version remains active
```

### Entities WITHOUT Versioning

For entities without versioning (Function, Validation Rule, Automation Rule, OV, Layout, Credential) — the definition is stored **inline** in the main table:

```sql
-- Function: definition inline
CREATE TABLE metadata.functions (
    id          UUID PRIMARY KEY,
    name        VARCHAR(100) UNIQUE NOT NULL,
    params      JSONB NOT NULL,
    return_type VARCHAR(20) NOT NULL,
    body        TEXT NOT NULL,              -- CEL expression, directly in the table
    ...
);

-- Validation Rule: expression inline
CREATE TABLE metadata.validation_rules (
    id          UUID PRIMARY KEY,
    object_id   UUID NOT NULL,
    expression  TEXT NOT NULL,              -- CEL expression, directly in the table
    ...
);
```

Save = live. Protection against errors:
- **Function**: dependency check + type validation on save; test endpoint
- **Validation Rule**: CEL compilation check on save
- **OV / Layout**: preview in admin UI before saving
- **Credential**: test connection endpoint

### Constructor UI Integration

**For Procedure/Scenario (with versioning):**

```
+-------------------------------------------+
|  Procedure: create_order                   |
|                                           |
|  Published: Version 3 (2026-02-15)       |
|  Status: * Published                      |
|                                           |
|  [Edit]  [History]  [Rollback]            |
|                                           |
|  --- Draft (if exists) ---               |
|  Version 4 (draft)                       |
|  [Test]  [Publish]  [Delete]              |
|                                           |
|  --- Version History ---                  |
|  v3  published  2026-02-15  "Added webhook"       |
|  v2  superseded 2026-02-14  "New email field"     |
|  v1  superseded 2026-02-13  "First version"       |
+-------------------------------------------+
```

**For Function (without versioning):**

```
+-------------------------------------------+
|  Function: discount                       |
|                                           |
|  [Edit]  [Test]  [Delete]                 |
|                                           |
|  No version history — save = live         |
+-------------------------------------------+
```

### Retention

| Status | Retention |
|--------|-----------|
| draft | Until published or explicitly deleted |
| published | Indefinitely (current active version) |
| superseded | Last 10 versions; older ones — auto-delete |

Auto-deleting superseded versions beyond 10 prevents table bloat during frequent editing. 10 versions are sufficient for history analysis and rollback.

## Consequences

### Positive

- **Safe deployment** — draft + dry-run before publishing for complex DSLs (Procedure, Scenario)
- **Rollback** — a single operation to revert to the previous version
- **Mid-flight safety** — running Scenario instances use pinned Procedure versions
- **Change history** — who, when, what changed (change_summary)
- **Minimal complexity** — three statuses, auto-increment integer, no semver
- **Differentiation** — versioning only where justified; simple entities are not overcomplicated
- **Salesforce-aligned** — Flows use exactly this approach (active/inactive versions)
- **Admin-friendly** — "Save draft" -> "Publish" instead of "Choose MAJOR/MINOR/PATCH"

### Negative

- **Additional tables** — `procedure_versions`, `scenario_versions`, `scenario_run_snapshots`
- **Two pointers** — `draft_version_id` + `published_version_id` instead of inline definition
- **Publish workflow** — admin must explicitly publish; save != live (may be unfamiliar)
- **No granular constraints** — a Scenario cannot pin "any 2.x version of Procedure"; only latest published or snapshot at start

### What We Consciously Do NOT Implement

| Feature | Reason for Rejection |
|---------|----------------------|
| Semver (MAJOR.MINOR.PATCH) | YAGNI; one admin controls both sides; no CRM does this |
| Version constraints (`^2.0`) | Assumes independent evolution of consumers; irrelevant in single-tenant CRM |
| Backward compatibility validation | Requires formalizing input/output schema; dependency check on save is sufficient |
| 4+ statuses (deprecated, archived) | Three statuses cover all needs |
| Versioning Functions/VR/AR/OV/Layout | Simple entities; save = live + protection on save (type check, dependency check, test endpoint) |
| JSONB snapshot in Scenario run | Heavy; FK to version_id is sufficient + superseded versions are protected from deletion |

## Related ADRs

- **ADR-0024** — Procedure Engine: Procedure definition is stored in `procedure_versions.definition`, not directly in `procedures`
- **ADR-0025** — Scenario Engine: analogous; `scenario_run_snapshots` pins Procedure versions at start
- **ADR-0026** — Custom Functions: no versioning; dependency check + test endpoint on save
- **ADR-0019** — Declarative business logic: Validation Rules, Automation Rules — no versioning (CEL inline, immediate apply)
- **ADR-0022** — Object View: no versioning (configuration, not logic)
- **ADR-0027** — Layout: no versioning (presentation configuration)
