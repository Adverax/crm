BEGIN;
SELECT plan(33);

-- Схема существует
SELECT has_schema('metadata');

-- Таблица существует
SELECT has_table('metadata', 'object_definitions', 'table metadata.object_definitions exists');

-- Колонки и типы
SELECT has_column('metadata', 'object_definitions', 'id', 'has id');
SELECT col_type_is('metadata', 'object_definitions', 'id', 'uuid', 'id is uuid');
SELECT col_has_default('metadata', 'object_definitions', 'id', 'id has default');
SELECT col_is_pk('metadata', 'object_definitions', 'id', 'id is PK');

SELECT has_column('metadata', 'object_definitions', 'api_name', 'has api_name');
SELECT col_type_is('metadata', 'object_definitions', 'api_name', 'character varying(100)', 'api_name is character varying(100)');
SELECT col_not_null('metadata', 'object_definitions', 'api_name', 'api_name is NOT NULL');

SELECT has_column('metadata', 'object_definitions', 'label', 'has label');
SELECT col_not_null('metadata', 'object_definitions', 'label', 'label is NOT NULL');

SELECT has_column('metadata', 'object_definitions', 'plural_label', 'has plural_label');
SELECT col_not_null('metadata', 'object_definitions', 'plural_label', 'plural_label is NOT NULL');

SELECT has_column('metadata', 'object_definitions', 'description', 'has description');
SELECT col_has_default('metadata', 'object_definitions', 'description', 'description has default');

SELECT has_column('metadata', 'object_definitions', 'table_name', 'has table_name');
SELECT col_not_null('metadata', 'object_definitions', 'table_name', 'table_name is NOT NULL');

SELECT has_column('metadata', 'object_definitions', 'object_type', 'has object_type');
SELECT col_not_null('metadata', 'object_definitions', 'object_type', 'object_type is NOT NULL');

-- Поведенческие флаги
SELECT col_not_null('metadata', 'object_definitions', 'is_platform_managed', 'is_platform_managed is NOT NULL');
SELECT col_not_null('metadata', 'object_definitions', 'is_visible_in_setup', 'is_visible_in_setup is NOT NULL');
SELECT col_not_null('metadata', 'object_definitions', 'is_custom_fields_allowed', 'is_custom_fields_allowed is NOT NULL');
SELECT col_not_null('metadata', 'object_definitions', 'is_deleteable_object', 'is_deleteable_object is NOT NULL');

-- Возможности записей
SELECT col_not_null('metadata', 'object_definitions', 'is_createable', 'is_createable is NOT NULL');
SELECT col_not_null('metadata', 'object_definitions', 'is_updateable', 'is_updateable is NOT NULL');
SELECT col_not_null('metadata', 'object_definitions', 'is_deleteable', 'is_deleteable is NOT NULL');
SELECT col_not_null('metadata', 'object_definitions', 'is_queryable', 'is_queryable is NOT NULL');
SELECT col_not_null('metadata', 'object_definitions', 'is_searchable', 'is_searchable is NOT NULL');

-- Timestamps
SELECT col_not_null('metadata', 'object_definitions', 'created_at', 'created_at is NOT NULL');
SELECT col_not_null('metadata', 'object_definitions', 'updated_at', 'updated_at is NOT NULL');

-- Constraints
SELECT has_check('metadata', 'object_definitions', 'object_definitions has CHECK constraint');

-- Unique constraints
SELECT col_is_unique('metadata', 'object_definitions', 'api_name', 'api_name is unique');

-- Индексы
SELECT has_index('metadata', 'object_definitions', 'idx_object_definitions_table_name', 'index on table_name exists');

SELECT finish();
ROLLBACK;
