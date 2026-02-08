BEGIN;
SELECT plan(32);

-- Таблица существует
SELECT has_table('metadata', 'field_definitions', 'table metadata.field_definitions exists');

-- Колонки и типы
SELECT has_column('metadata', 'field_definitions', 'id');
SELECT col_type_is('metadata', 'field_definitions', 'id', 'uuid');
SELECT col_has_default('metadata', 'field_definitions', 'id');
SELECT col_is_pk('metadata', 'field_definitions', 'id');

SELECT has_column('metadata', 'field_definitions', 'object_id');
SELECT col_type_is('metadata', 'field_definitions', 'object_id', 'uuid');
SELECT col_not_null('metadata', 'field_definitions', 'object_id');

SELECT has_column('metadata', 'field_definitions', 'api_name');
SELECT col_not_null('metadata', 'field_definitions', 'api_name');

SELECT has_column('metadata', 'field_definitions', 'label');
SELECT col_not_null('metadata', 'field_definitions', 'label');

SELECT has_column('metadata', 'field_definitions', 'field_type');
SELECT col_not_null('metadata', 'field_definitions', 'field_type');

SELECT has_column('metadata', 'field_definitions', 'field_subtype');

SELECT has_column('metadata', 'field_definitions', 'referenced_object_id');

SELECT has_column('metadata', 'field_definitions', 'is_required');
SELECT col_not_null('metadata', 'field_definitions', 'is_required');

SELECT has_column('metadata', 'field_definitions', 'is_unique');
SELECT col_not_null('metadata', 'field_definitions', 'is_unique');

SELECT has_column('metadata', 'field_definitions', 'config');
SELECT col_type_is('metadata', 'field_definitions', 'config', 'jsonb');
SELECT col_has_default('metadata', 'field_definitions', 'config');

SELECT has_column('metadata', 'field_definitions', 'is_system_field');
SELECT has_column('metadata', 'field_definitions', 'is_custom');
SELECT has_column('metadata', 'field_definitions', 'is_platform_managed');
SELECT has_column('metadata', 'field_definitions', 'sort_order');

-- FK
SELECT fk_ok('metadata', 'field_definitions', 'object_id', 'metadata', 'object_definitions', 'id');
SELECT fk_ok('metadata', 'field_definitions', 'referenced_object_id', 'metadata', 'object_definitions', 'id');

-- Индексы
SELECT has_index('metadata', 'field_definitions', 'idx_field_definitions_object_id', 'index on object_id exists');
SELECT has_index('metadata', 'field_definitions', 'idx_field_definitions_referenced_object_id', 'index on referenced_object_id exists');
SELECT has_index('metadata', 'field_definitions', 'idx_field_definitions_field_type', 'index on field_type exists');

SELECT finish();
ROLLBACK;
