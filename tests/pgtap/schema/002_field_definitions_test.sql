BEGIN;
SELECT plan(32);

-- Таблица существует
SELECT has_table('metadata', 'field_definitions', 'table metadata.field_definitions exists');

-- Колонки и типы
SELECT has_column('metadata', 'field_definitions', 'id', 'has id');
SELECT col_type_is('metadata', 'field_definitions', 'id', 'uuid', 'id is uuid');
SELECT col_has_default('metadata', 'field_definitions', 'id', 'id has default');
SELECT col_is_pk('metadata', 'field_definitions', 'id', 'id is PK');

SELECT has_column('metadata', 'field_definitions', 'object_id', 'has object_id');
SELECT col_type_is('metadata', 'field_definitions', 'object_id', 'uuid', 'object_id is uuid');
SELECT col_not_null('metadata', 'field_definitions', 'object_id', 'object_id is NOT NULL');

SELECT has_column('metadata', 'field_definitions', 'api_name', 'has api_name');
SELECT col_not_null('metadata', 'field_definitions', 'api_name', 'api_name is NOT NULL');

SELECT has_column('metadata', 'field_definitions', 'label', 'has label');
SELECT col_not_null('metadata', 'field_definitions', 'label', 'label is NOT NULL');

SELECT has_column('metadata', 'field_definitions', 'field_type', 'has field_type');
SELECT col_not_null('metadata', 'field_definitions', 'field_type', 'field_type is NOT NULL');

SELECT has_column('metadata', 'field_definitions', 'field_subtype', 'has field_subtype');

SELECT has_column('metadata', 'field_definitions', 'referenced_object_id', 'has referenced_object_id');

SELECT has_column('metadata', 'field_definitions', 'is_required', 'has is_required');
SELECT col_not_null('metadata', 'field_definitions', 'is_required', 'is_required is NOT NULL');

SELECT has_column('metadata', 'field_definitions', 'is_unique', 'has is_unique');
SELECT col_not_null('metadata', 'field_definitions', 'is_unique', 'is_unique is NOT NULL');

SELECT has_column('metadata', 'field_definitions', 'config', 'has config');
SELECT col_type_is('metadata', 'field_definitions', 'config', 'jsonb', 'config is jsonb');
SELECT col_has_default('metadata', 'field_definitions', 'config', 'config has default');

SELECT has_column('metadata', 'field_definitions', 'is_system_field', 'has is_system_field');
SELECT has_column('metadata', 'field_definitions', 'is_custom', 'has is_custom');
SELECT has_column('metadata', 'field_definitions', 'is_platform_managed', 'has is_platform_managed');
SELECT has_column('metadata', 'field_definitions', 'sort_order', 'has sort_order');

-- FK
SELECT fk_ok('metadata', 'field_definitions', 'object_id', 'metadata', 'object_definitions', 'id', 'FK object_id -> object_definitions.id');
SELECT fk_ok('metadata', 'field_definitions', 'referenced_object_id', 'metadata', 'object_definitions', 'id', 'FK referenced_object_id -> object_definitions.id');

-- Индексы
SELECT has_index('metadata', 'field_definitions', 'idx_field_definitions_object_id', 'index on object_id exists');
SELECT has_index('metadata', 'field_definitions', 'idx_field_definitions_referenced_object_id', 'index on referenced_object_id exists');
SELECT has_index('metadata', 'field_definitions', 'idx_field_definitions_field_type', 'index on field_type exists');

SELECT finish();
ROLLBACK;
