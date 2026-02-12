BEGIN;
SELECT plan(27);

-- picklist_definitions
SELECT has_table('metadata', 'picklist_definitions', 'table metadata.picklist_definitions exists');

SELECT has_column('metadata', 'picklist_definitions', 'id', 'has id');
SELECT col_type_is('metadata', 'picklist_definitions', 'id', 'uuid', 'id is uuid');
SELECT col_is_pk('metadata', 'picklist_definitions', 'id', 'id is PK');

SELECT has_column('metadata', 'picklist_definitions', 'api_name', 'has api_name');
SELECT col_not_null('metadata', 'picklist_definitions', 'api_name', 'api_name is NOT NULL');
SELECT col_is_unique('metadata', 'picklist_definitions', 'api_name', 'api_name is unique');

SELECT has_column('metadata', 'picklist_definitions', 'label', 'has label');
SELECT col_not_null('metadata', 'picklist_definitions', 'label', 'label is NOT NULL');

SELECT has_column('metadata', 'picklist_definitions', 'description', 'has description');
SELECT col_not_null('metadata', 'picklist_definitions', 'created_at', 'created_at is NOT NULL');
SELECT col_not_null('metadata', 'picklist_definitions', 'updated_at', 'updated_at is NOT NULL');

-- picklist_values
SELECT has_table('metadata', 'picklist_values', 'table metadata.picklist_values exists');

SELECT has_column('metadata', 'picklist_values', 'id', 'has id');
SELECT col_is_pk('metadata', 'picklist_values', 'id', 'id is PK');

SELECT has_column('metadata', 'picklist_values', 'picklist_definition_id', 'has picklist_definition_id');
SELECT col_not_null('metadata', 'picklist_values', 'picklist_definition_id', 'picklist_definition_id is NOT NULL');

SELECT has_column('metadata', 'picklist_values', 'value', 'has value');
SELECT col_not_null('metadata', 'picklist_values', 'value', 'value is NOT NULL');

SELECT has_column('metadata', 'picklist_values', 'label', 'has label');
SELECT col_not_null('metadata', 'picklist_values', 'label', 'label is NOT NULL');

SELECT has_column('metadata', 'picklist_values', 'sort_order', 'has sort_order');
SELECT has_column('metadata', 'picklist_values', 'is_default', 'has is_default');
SELECT has_column('metadata', 'picklist_values', 'is_active', 'has is_active');

-- FK
SELECT fk_ok('metadata', 'picklist_values', 'picklist_definition_id', 'metadata', 'picklist_definitions', 'id', 'FK picklist_definition_id -> picklist_definitions.id');

-- Индексы
SELECT has_index('metadata', 'picklist_values', 'idx_picklist_values_picklist_definition_id', 'index on picklist_definition_id exists');

-- Unique constraint (picklist_definition_id, value)
SELECT has_unique('metadata', 'picklist_values', 'picklist_values has UNIQUE constraint');

SELECT finish();
ROLLBACK;
