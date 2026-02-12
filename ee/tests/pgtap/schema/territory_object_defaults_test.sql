-- Copyright 2026 Adverax. All rights reserved.
-- Licensed under the Adverax Commercial License.
-- See ee/LICENSE for details.
-- Unauthorized use, copying, or distribution is prohibited.

BEGIN;
SELECT plan(19);

-- ee.territory_object_defaults table
SELECT has_table('ee', 'territory_object_defaults', 'table ee.territory_object_defaults exists');

-- id
SELECT has_column('ee', 'territory_object_defaults', 'id', 'has id');
SELECT col_type_is('ee', 'territory_object_defaults', 'id', 'uuid', 'id is uuid');
SELECT col_has_default('ee', 'territory_object_defaults', 'id', 'id has default');
SELECT col_is_pk('ee', 'territory_object_defaults', 'id', 'id is PK');

-- territory_id
SELECT has_column('ee', 'territory_object_defaults', 'territory_id', 'has territory_id');
SELECT col_not_null('ee', 'territory_object_defaults', 'territory_id', 'territory_id is NOT NULL');
SELECT fk_ok('ee', 'territory_object_defaults', 'territory_id', 'ee', 'territories', 'id',
    'territory_id FK to ee.territories');

-- object_id
SELECT has_column('ee', 'territory_object_defaults', 'object_id', 'has object_id');
SELECT col_not_null('ee', 'territory_object_defaults', 'object_id', 'object_id is NOT NULL');
SELECT fk_ok('ee', 'territory_object_defaults', 'object_id', 'metadata', 'object_definitions', 'id',
    'object_id FK to metadata.object_definitions');

-- access_level
SELECT has_column('ee', 'territory_object_defaults', 'access_level', 'has access_level');
SELECT col_not_null('ee', 'territory_object_defaults', 'access_level', 'access_level is NOT NULL');
SELECT has_check('ee', 'territory_object_defaults', 'access_level has CHECK');

-- created_at / updated_at
SELECT has_column('ee', 'territory_object_defaults', 'created_at', 'has created_at');
SELECT col_not_null('ee', 'territory_object_defaults', 'created_at', 'created_at is NOT NULL');

SELECT has_column('ee', 'territory_object_defaults', 'updated_at', 'has updated_at');
SELECT col_not_null('ee', 'territory_object_defaults', 'updated_at', 'updated_at is NOT NULL');

-- Index
SELECT has_index('ee', 'territory_object_defaults', 'idx_territory_object_defaults_territory', 'index on territory_id');

SELECT finish();
ROLLBACK;
