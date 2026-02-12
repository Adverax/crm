-- Copyright 2026 Adverax. All rights reserved.
-- Licensed under the Adverax Commercial License.
-- See ee/LICENSE for details.
-- Unauthorized use, copying, or distribution is prohibited.

BEGIN;
SELECT plan(4);

-- Migration 000003: related_territory_id column on iam.groups
SELECT has_column('iam', 'groups', 'related_territory_id',
    'iam.groups has related_territory_id column');
SELECT col_type_is('iam', 'groups', 'related_territory_id', 'uuid',
    'related_territory_id is uuid');
SELECT fk_ok('iam', 'groups', 'related_territory_id', 'ee', 'territories', 'id',
    'related_territory_id FK to ee.territories');
SELECT has_index('iam', 'groups', 'idx_iam_groups_related_territory',
    'partial index on related_territory_id');

SELECT finish();
ROLLBACK;
