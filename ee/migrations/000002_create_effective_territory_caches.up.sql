-- Copyright 2026 Adverax. All rights reserved.
-- Licensed under the Adverax Commercial License.
-- See ee/LICENSE for details.
-- Unauthorized use, copying, or distribution is prohibited.

-- Territory hierarchy closure table (ADR-0012, ADR-0015)
CREATE TABLE security.effective_territory_hierarchy (
    ancestor_territory_id   UUID NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    descendant_territory_id UUID NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    depth                   INT  NOT NULL DEFAULT 0,
    computed_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (ancestor_territory_id, descendant_territory_id)
);

CREATE INDEX idx_eth_descendant ON security.effective_territory_hierarchy (descendant_territory_id);

-- Flat list of user territory assignments (direct assignments only)
CREATE TABLE security.effective_user_territory (
    user_id        UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    territory_id   UUID NOT NULL REFERENCES ee.territories(id) ON DELETE CASCADE,
    computed_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, territory_id)
);

CREATE INDEX idx_eut_user ON security.effective_user_territory (user_id);
CREATE INDEX idx_eut_territory ON security.effective_user_territory (territory_id);
