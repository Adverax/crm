-- Copyright 2026 Adverax. All rights reserved.
-- Licensed under the Adverax Commercial License.
-- See ee/LICENSE for details.
-- Unauthorized use, copying, or distribution is prohibited.

-- Add territory FK to groups (ADR-0015)
ALTER TABLE iam.groups
    ADD COLUMN related_territory_id UUID REFERENCES ee.territories(id) ON DELETE CASCADE;

CREATE INDEX idx_iam_groups_related_territory
ON iam.groups (related_territory_id)
WHERE related_territory_id IS NOT NULL;
