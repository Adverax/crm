BEGIN;
SELECT plan(34);

-- Схема существует
SELECT has_schema('metadata');

-- Таблица существует
SELECT has_table('metadata', 'object_definitions', 'table metadata.object_definitions exists');

-- Колонки и типы
SELECT has_column('metadata', 'object_definitions', 'id');
SELECT col_type_is('metadata', 'object_definitions', 'id', 'uuid');
SELECT col_has_default('metadata', 'object_definitions', 'id');
SELECT col_is_pk('metadata', 'object_definitions', 'id');

SELECT has_column('metadata', 'object_definitions', 'api_name');
SELECT col_type_is('metadata', 'object_definitions', 'api_name', 'character varying(100)');
SELECT col_not_null('metadata', 'object_definitions', 'api_name');

SELECT has_column('metadata', 'object_definitions', 'label');
SELECT col_not_null('metadata', 'object_definitions', 'label');

SELECT has_column('metadata', 'object_definitions', 'plural_label');
SELECT col_not_null('metadata', 'object_definitions', 'plural_label');

SELECT has_column('metadata', 'object_definitions', 'description');
SELECT col_has_default('metadata', 'object_definitions', 'description');

SELECT has_column('metadata', 'object_definitions', 'table_name');
SELECT col_not_null('metadata', 'object_definitions', 'table_name');

SELECT has_column('metadata', 'object_definitions', 'object_type');
SELECT col_not_null('metadata', 'object_definitions', 'object_type');

-- Поведенческие флаги
SELECT col_not_null('metadata', 'object_definitions', 'is_platform_managed');
SELECT col_not_null('metadata', 'object_definitions', 'is_visible_in_setup');
SELECT col_not_null('metadata', 'object_definitions', 'is_custom_fields_allowed');
SELECT col_not_null('metadata', 'object_definitions', 'is_deleteable_object');

-- Возможности записей
SELECT col_not_null('metadata', 'object_definitions', 'is_createable');
SELECT col_not_null('metadata', 'object_definitions', 'is_updateable');
SELECT col_not_null('metadata', 'object_definitions', 'is_deleteable');
SELECT col_not_null('metadata', 'object_definitions', 'is_queryable');
SELECT col_not_null('metadata', 'object_definitions', 'is_searchable');

-- Timestamps
SELECT col_not_null('metadata', 'object_definitions', 'created_at');
SELECT col_not_null('metadata', 'object_definitions', 'updated_at');

-- Constraints
SELECT has_check('metadata', 'object_definitions', 'object_definitions has CHECK constraint');

-- Unique constraints
SELECT col_is_unique('metadata', 'object_definitions', 'api_name');

-- Индексы
SELECT has_index('metadata', 'object_definitions', 'idx_object_definitions_table_name', 'index on table_name exists');

SELECT finish();
ROLLBACK;
