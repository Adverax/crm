BEGIN;
SELECT plan(16);

-- Таблица существует
SELECT has_table('metadata', 'translations', 'table metadata.translations exists');

-- Колонки
SELECT has_column('metadata', 'translations', 'id', 'has id');
SELECT col_type_is('metadata', 'translations', 'id', 'uuid', 'id is uuid');
SELECT col_is_pk('metadata', 'translations', 'id', 'id is PK');

SELECT has_column('metadata', 'translations', 'resource_type', 'has resource_type');
SELECT col_not_null('metadata', 'translations', 'resource_type', 'resource_type is NOT NULL');

SELECT has_column('metadata', 'translations', 'resource_id', 'has resource_id');
SELECT col_not_null('metadata', 'translations', 'resource_id', 'resource_id is NOT NULL');

SELECT has_column('metadata', 'translations', 'field_name', 'has field_name');
SELECT col_not_null('metadata', 'translations', 'field_name', 'field_name is NOT NULL');

SELECT has_column('metadata', 'translations', 'locale', 'has locale');
SELECT col_not_null('metadata', 'translations', 'locale', 'locale is NOT NULL');

SELECT has_column('metadata', 'translations', 'value', 'has value');
SELECT col_not_null('metadata', 'translations', 'value', 'value is NOT NULL');

-- Unique constraint (resource_type, resource_id, field_name, locale)
SELECT has_unique('metadata', 'translations', 'translations has UNIQUE constraint');

-- Индекс
SELECT has_index('metadata', 'translations', 'idx_translations_resource', 'index on resource exists');

SELECT finish();
ROLLBACK;
