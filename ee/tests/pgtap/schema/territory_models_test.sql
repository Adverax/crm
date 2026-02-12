-- Copyright 2026 Adverax. All rights reserved.
-- Licensed under the Adverax Commercial License.
-- See ee/LICENSE for details.
-- Unauthorized use, copying, or distribution is prohibited.

BEGIN;
SELECT plan(23);

-- ee.territory_models
SELECT has_table('ee', 'territory_models', 'table ee.territory_models exists');

SELECT has_column('ee', 'territory_models', 'id', 'has id');
SELECT col_type_is('ee', 'territory_models', 'id', 'uuid', 'id is uuid');
SELECT col_has_default('ee', 'territory_models', 'id', 'id has default');
SELECT col_is_pk('ee', 'territory_models', 'id', 'id is PK');

SELECT has_column('ee', 'territory_models', 'api_name', 'has api_name');
SELECT col_type_is('ee', 'territory_models', 'api_name', 'character varying(100)', 'api_name is varchar(100)');
SELECT col_not_null('ee', 'territory_models', 'api_name', 'api_name is NOT NULL');
SELECT col_is_unique('ee', 'territory_models', 'api_name', 'api_name is unique');

SELECT has_column('ee', 'territory_models', 'label', 'has label');
SELECT col_not_null('ee', 'territory_models', 'label', 'label is NOT NULL');

SELECT has_column('ee', 'territory_models', 'description', 'has description');
SELECT col_not_null('ee', 'territory_models', 'description', 'description is NOT NULL');

SELECT has_column('ee', 'territory_models', 'status', 'has status');
SELECT col_not_null('ee', 'territory_models', 'status', 'status is NOT NULL');
SELECT col_has_default('ee', 'territory_models', 'status', 'status has default');
SELECT has_check('ee', 'territory_models', 'ee.territory_models has CHECK constraint');

SELECT has_column('ee', 'territory_models', 'activated_at', 'has activated_at');
SELECT has_column('ee', 'territory_models', 'archived_at', 'has archived_at');

SELECT has_column('ee', 'territory_models', 'created_at', 'has created_at');
SELECT col_not_null('ee', 'territory_models', 'created_at', 'created_at is NOT NULL');

SELECT has_column('ee', 'territory_models', 'updated_at', 'has updated_at');
SELECT col_not_null('ee', 'territory_models', 'updated_at', 'updated_at is NOT NULL');

SELECT finish();
ROLLBACK;
