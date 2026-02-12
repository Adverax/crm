-- Copyright 2026 Adverax. All rights reserved.
-- Licensed under the Adverax Commercial License.
-- See ee/LICENSE for details.
-- Unauthorized use, copying, or distribution is prohibited.

DROP INDEX IF EXISTS iam.idx_iam_groups_related_territory;
ALTER TABLE iam.groups DROP COLUMN IF EXISTS related_territory_id;
