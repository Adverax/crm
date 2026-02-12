-- Copyright 2026 Adverax. All rights reserved.
-- Licensed under the Adverax Commercial License.
-- See ee/LICENSE for details.
-- Unauthorized use, copying, or distribution is prohibited.

BEGIN;
SELECT plan(19);

-- security.effective_territory_hierarchy
SELECT has_table('security', 'effective_territory_hierarchy', 'table security.effective_territory_hierarchy exists');

SELECT has_column('security', 'effective_territory_hierarchy', 'ancestor_territory_id', 'has ancestor_territory_id');
SELECT col_not_null('security', 'effective_territory_hierarchy', 'ancestor_territory_id', 'ancestor_territory_id is NOT NULL');
SELECT fk_ok('security', 'effective_territory_hierarchy', 'ancestor_territory_id', 'ee', 'territories', 'id',
    'ancestor_territory_id FK to ee.territories');

SELECT has_column('security', 'effective_territory_hierarchy', 'descendant_territory_id', 'has descendant_territory_id');
SELECT col_not_null('security', 'effective_territory_hierarchy', 'descendant_territory_id', 'descendant_territory_id is NOT NULL');
SELECT fk_ok('security', 'effective_territory_hierarchy', 'descendant_territory_id', 'ee', 'territories', 'id',
    'descendant_territory_id FK to ee.territories');

SELECT has_column('security', 'effective_territory_hierarchy', 'depth', 'has depth');
SELECT col_not_null('security', 'effective_territory_hierarchy', 'depth', 'depth is NOT NULL');

SELECT has_index('security', 'effective_territory_hierarchy', 'idx_eth_descendant', 'index on descendant');

-- security.effective_user_territory
SELECT has_table('security', 'effective_user_territory', 'table security.effective_user_territory exists');

SELECT has_column('security', 'effective_user_territory', 'user_id', 'has user_id');
SELECT col_not_null('security', 'effective_user_territory', 'user_id', 'user_id is NOT NULL');
SELECT fk_ok('security', 'effective_user_territory', 'user_id', 'iam', 'users', 'id',
    'user_id FK to iam.users');

SELECT has_column('security', 'effective_user_territory', 'territory_id', 'has territory_id');
SELECT col_not_null('security', 'effective_user_territory', 'territory_id', 'territory_id is NOT NULL');
SELECT fk_ok('security', 'effective_user_territory', 'territory_id', 'ee', 'territories', 'id',
    'territory_id FK to ee.territories');

SELECT has_index('security', 'effective_user_territory', 'idx_eut_user', 'index on user_id');
SELECT has_index('security', 'effective_user_territory', 'idx_eut_territory', 'index on territory_id');

SELECT finish();
ROLLBACK;
