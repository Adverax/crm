BEGIN;
SELECT plan(27);

-- picklist_definitions
SELECT has_table('metadata', 'picklist_definitions', 'table metadata.picklist_definitions exists');

SELECT has_column('metadata', 'picklist_definitions', 'id');
SELECT col_type_is('metadata', 'picklist_definitions', 'id', 'uuid');
SELECT col_is_pk('metadata', 'picklist_definitions', 'id');

SELECT has_column('metadata', 'picklist_definitions', 'api_name');
SELECT col_not_null('metadata', 'picklist_definitions', 'api_name');
SELECT col_is_unique('metadata', 'picklist_definitions', 'api_name');

SELECT has_column('metadata', 'picklist_definitions', 'label');
SELECT col_not_null('metadata', 'picklist_definitions', 'label');

SELECT has_column('metadata', 'picklist_definitions', 'description');
SELECT col_not_null('metadata', 'picklist_definitions', 'created_at');
SELECT col_not_null('metadata', 'picklist_definitions', 'updated_at');

-- picklist_values
SELECT has_table('metadata', 'picklist_values', 'table metadata.picklist_values exists');

SELECT has_column('metadata', 'picklist_values', 'id');
SELECT col_is_pk('metadata', 'picklist_values', 'id');

SELECT has_column('metadata', 'picklist_values', 'picklist_definition_id');
SELECT col_not_null('metadata', 'picklist_values', 'picklist_definition_id');

SELECT has_column('metadata', 'picklist_values', 'value');
SELECT col_not_null('metadata', 'picklist_values', 'value');

SELECT has_column('metadata', 'picklist_values', 'label');
SELECT col_not_null('metadata', 'picklist_values', 'label');

SELECT has_column('metadata', 'picklist_values', 'sort_order');
SELECT has_column('metadata', 'picklist_values', 'is_default');
SELECT has_column('metadata', 'picklist_values', 'is_active');

-- FK
SELECT fk_ok('metadata', 'picklist_values', 'picklist_definition_id', 'metadata', 'picklist_definitions', 'id');

-- Индексы
SELECT has_index('metadata', 'picklist_values', 'idx_picklist_values_picklist_definition_id', 'index on picklist_definition_id exists');

-- Unique constraint (picklist_definition_id, value)
SELECT has_unique('metadata', 'picklist_values', 'picklist_values has UNIQUE constraint');

SELECT finish();
ROLLBACK;
