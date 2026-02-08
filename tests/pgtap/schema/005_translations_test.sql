BEGIN;
SELECT plan(16);

-- Таблица существует
SELECT has_table('metadata', 'translations', 'table metadata.translations exists');

-- Колонки
SELECT has_column('metadata', 'translations', 'id');
SELECT col_type_is('metadata', 'translations', 'id', 'uuid');
SELECT col_is_pk('metadata', 'translations', 'id');

SELECT has_column('metadata', 'translations', 'resource_type');
SELECT col_not_null('metadata', 'translations', 'resource_type');

SELECT has_column('metadata', 'translations', 'resource_id');
SELECT col_not_null('metadata', 'translations', 'resource_id');

SELECT has_column('metadata', 'translations', 'field_name');
SELECT col_not_null('metadata', 'translations', 'field_name');

SELECT has_column('metadata', 'translations', 'locale');
SELECT col_not_null('metadata', 'translations', 'locale');

SELECT has_column('metadata', 'translations', 'value');
SELECT col_not_null('metadata', 'translations', 'value');

-- Unique constraint (resource_type, resource_id, field_name, locale)
SELECT has_unique('metadata', 'translations', 'translations has UNIQUE constraint');

-- Индекс
SELECT has_index('metadata', 'translations', 'idx_translations_resource', 'index on resource exists');

SELECT finish();
ROLLBACK;
