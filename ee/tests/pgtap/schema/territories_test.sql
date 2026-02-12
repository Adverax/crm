-- Copyright 2026 Adverax. All rights reserved.
-- Licensed under the Adverax Commercial License.
-- See ee/LICENSE for details.
-- Unauthorized use, copying, or distribution is prohibited.

BEGIN;
SELECT plan(23);

-- ee.territories table
SELECT has_table('ee', 'territories', 'table ee.territories exists');

-- id
SELECT has_column('ee', 'territories', 'id', 'has id');
SELECT col_type_is('ee', 'territories', 'id', 'uuid', 'id is uuid');
SELECT col_has_default('ee', 'territories', 'id', 'id has default');
SELECT col_is_pk('ee', 'territories', 'id', 'id is PK');

-- model_id
SELECT has_column('ee', 'territories', 'model_id', 'has model_id');
SELECT col_not_null('ee', 'territories', 'model_id', 'model_id is NOT NULL');
SELECT fk_ok('ee', 'territories', 'model_id', 'ee', 'territory_models', 'id',
    'model_id FK to ee.territory_models');

-- parent_id
SELECT has_column('ee', 'territories', 'parent_id', 'has parent_id');
SELECT fk_ok('ee', 'territories', 'parent_id', 'ee', 'territories', 'id',
    'parent_id FK to ee.territories');

-- api_name
SELECT has_column('ee', 'territories', 'api_name', 'has api_name');
SELECT col_type_is('ee', 'territories', 'api_name', 'character varying(100)', 'api_name is varchar(100)');
SELECT col_not_null('ee', 'territories', 'api_name', 'api_name is NOT NULL');

-- label
SELECT has_column('ee', 'territories', 'label', 'has label');
SELECT col_not_null('ee', 'territories', 'label', 'label is NOT NULL');

-- description
SELECT has_column('ee', 'territories', 'description', 'has description');
SELECT col_not_null('ee', 'territories', 'description', 'description is NOT NULL');

-- created_at / updated_at
SELECT has_column('ee', 'territories', 'created_at', 'has created_at');
SELECT col_not_null('ee', 'territories', 'created_at', 'created_at is NOT NULL');

SELECT has_column('ee', 'territories', 'updated_at', 'has updated_at');
SELECT col_not_null('ee', 'territories', 'updated_at', 'updated_at is NOT NULL');

-- Indexes
SELECT has_index('ee', 'territories', 'idx_territories_model_id', 'index on model_id');
SELECT has_index('ee', 'territories', 'idx_territories_parent_id', 'index on parent_id');

SELECT finish();
ROLLBACK;
