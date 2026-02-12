-- Copyright 2026 Adverax. All rights reserved.
-- Licensed under the Adverax Commercial License.
-- See ee/LICENSE for details.
-- Unauthorized use, copying, or distribution is prohibited.

-- Territory Management schema (ADR-0015)
CREATE SCHEMA IF NOT EXISTS ee;

-- Territory Models with lifecycle: planning -> active -> archived
CREATE TABLE ee.territory_models (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    api_name      VARCHAR(100) NOT NULL UNIQUE,
    label         VARCHAR(255) NOT NULL,
    description   TEXT        NOT NULL DEFAULT '',
    status        VARCHAR(20) NOT NULL DEFAULT 'planning'
                  CHECK (status IN ('planning', 'active', 'archived')),
    activated_at  TIMESTAMPTZ,
    archived_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Only one active model at any time
CREATE UNIQUE INDEX uq_territory_models_active
ON ee.territory_models (status)
WHERE status = 'active';

-- Territories within a model (hierarchical)
CREATE TABLE ee.territories (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id    UUID        NOT NULL REFERENCES ee.territory_models(id) ON DELETE CASCADE,
    parent_id   UUID        REFERENCES ee.territories(id) ON DELETE CASCADE,
    api_name    VARCHAR(100) NOT NULL,
    label       VARCHAR(255) NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (model_id, api_name)
);

CREATE INDEX idx_territories_model_id ON ee.territories (model_id);
CREATE INDEX idx_territories_parent_id ON ee.territories (parent_id) WHERE parent_id IS NOT NULL;

-- Per-object access levels within territories
CREATE TABLE ee.territory_object_defaults (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    territory_id  UUID        NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    object_id     UUID        NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    access_level  VARCHAR(20) NOT NULL CHECK (access_level IN ('read', 'read_write')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (territory_id, object_id)
);

CREATE INDEX idx_territory_object_defaults_territory ON ee.territory_object_defaults (territory_id);

-- M2M: users assigned to territories
CREATE TABLE ee.user_territory_assignments (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    territory_id  UUID        NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, territory_id)
);

CREATE INDEX idx_user_territory_assignments_user ON ee.user_territory_assignments (user_id);
CREATE INDEX idx_user_territory_assignments_territory ON ee.user_territory_assignments (territory_id);

-- Records assigned to territories (record_id has no FK — records in different obj_ tables)
CREATE TABLE ee.record_territory_assignments (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    record_id     UUID        NOT NULL,
    object_id     UUID        NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    territory_id  UUID        NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    reason        VARCHAR(30) NOT NULL DEFAULT 'manual'
                  CHECK (reason IN ('manual', 'assignment_rule')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (record_id, object_id, territory_id)
);

CREATE INDEX idx_record_territory_record ON ee.record_territory_assignments (record_id, object_id);
CREATE INDEX idx_record_territory_territory ON ee.record_territory_assignments (territory_id);

-- Criteria-based assignment rules for automatic record-to-territory mapping
CREATE TABLE ee.territory_assignment_rules (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    territory_id    UUID        NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    object_id       UUID        NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    rule_order      INT         NOT NULL DEFAULT 0,
    criteria_field  VARCHAR(255) NOT NULL,
    criteria_op     VARCHAR(20) NOT NULL
                    CHECK (criteria_op IN ('eq', 'neq', 'in', 'gt', 'lt', 'contains')),
    criteria_value  TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_territory_assignment_rules_territory ON ee.territory_assignment_rules (territory_id);
CREATE INDEX idx_territory_assignment_rules_object ON ee.territory_assignment_rules (object_id);
