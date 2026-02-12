-- Copyright 2026 Adverax. All rights reserved.
-- Licensed under the Adverax Commercial License.
-- See ee/LICENSE for details.
-- Unauthorized use, copying, or distribution is prohibited.

BEGIN;
SELECT plan(15);

-- ee.user_territory_assignments table
SELECT has_table('ee', 'user_territory_assignments', 'table ee.user_territory_assignments exists');

-- id
SELECT has_column('ee', 'user_territory_assignments', 'id', 'has id');
SELECT col_type_is('ee', 'user_territory_assignments', 'id', 'uuid', 'id is uuid');
SELECT col_has_default('ee', 'user_territory_assignments', 'id', 'id has default');
SELECT col_is_pk('ee', 'user_territory_assignments', 'id', 'id is PK');

-- user_id
SELECT has_column('ee', 'user_territory_assignments', 'user_id', 'has user_id');
SELECT col_not_null('ee', 'user_territory_assignments', 'user_id', 'user_id is NOT NULL');
SELECT fk_ok('ee', 'user_territory_assignments', 'user_id', 'iam', 'users', 'id',
    'user_id FK to iam.users');

-- territory_id
SELECT has_column('ee', 'user_territory_assignments', 'territory_id', 'has territory_id');
SELECT col_not_null('ee', 'user_territory_assignments', 'territory_id', 'territory_id is NOT NULL');
SELECT fk_ok('ee', 'user_territory_assignments', 'territory_id', 'ee', 'territories', 'id',
    'territory_id FK to ee.territories');

-- created_at
SELECT has_column('ee', 'user_territory_assignments', 'created_at', 'has created_at');
SELECT col_not_null('ee', 'user_territory_assignments', 'created_at', 'created_at is NOT NULL');

-- Indexes
SELECT has_index('ee', 'user_territory_assignments', 'idx_user_territory_assignments_user', 'index on user_id');
SELECT has_index('ee', 'user_territory_assignments', 'idx_user_territory_assignments_territory', 'index on territory_id');

SELECT finish();
ROLLBACK;
