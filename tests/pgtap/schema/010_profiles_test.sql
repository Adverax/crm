BEGIN;
SELECT plan(17);

-- Таблица существует
SELECT has_table('iam', 'profiles', 'table iam.profiles exists');

-- Колонки и типы
SELECT has_column('iam', 'profiles', 'id', 'has id');
SELECT col_type_is('iam', 'profiles', 'id', 'uuid', 'id is uuid');
SELECT col_has_default('iam', 'profiles', 'id', 'id has default');
SELECT col_is_pk('iam', 'profiles', 'id', 'id is PK');

SELECT has_column('iam', 'profiles', 'api_name', 'has api_name');
SELECT col_type_is('iam', 'profiles', 'api_name', 'character varying(100)', 'api_name is character varying(100)');
SELECT col_not_null('iam', 'profiles', 'api_name', 'api_name is NOT NULL');
SELECT col_is_unique('iam', 'profiles', 'api_name', 'api_name is unique');

SELECT has_column('iam', 'profiles', 'label', 'has label');
SELECT col_not_null('iam', 'profiles', 'label', 'label is NOT NULL');

SELECT has_column('iam', 'profiles', 'description', 'has description');
SELECT col_has_default('iam', 'profiles', 'description', 'description has default');

SELECT has_column('iam', 'profiles', 'base_permission_set_id', 'has base_permission_set_id');
SELECT col_not_null('iam', 'profiles', 'base_permission_set_id', 'base_permission_set_id is NOT NULL');
SELECT fk_ok('iam', 'profiles', 'base_permission_set_id', 'iam', 'permission_sets', 'id', 'FK base_permission_set_id -> permission_sets.id');

SELECT has_index('iam', 'profiles', 'idx_profiles_base_permission_set_id', 'index on base_permission_set_id exists');

SELECT finish();
ROLLBACK;
