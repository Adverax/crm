BEGIN;
SELECT plan(16);

-- Таблица существует
SELECT has_table('iam', 'profiles', 'table iam.profiles exists');

-- Колонки и типы
SELECT has_column('iam', 'profiles', 'id');
SELECT col_type_is('iam', 'profiles', 'id', 'uuid');
SELECT col_has_default('iam', 'profiles', 'id');
SELECT col_is_pk('iam', 'profiles', 'id');

SELECT has_column('iam', 'profiles', 'api_name');
SELECT col_type_is('iam', 'profiles', 'api_name', 'character varying(100)');
SELECT col_not_null('iam', 'profiles', 'api_name');
SELECT col_is_unique('iam', 'profiles', 'api_name');

SELECT has_column('iam', 'profiles', 'label');
SELECT col_not_null('iam', 'profiles', 'label');

SELECT has_column('iam', 'profiles', 'description');
SELECT col_has_default('iam', 'profiles', 'description');

SELECT has_column('iam', 'profiles', 'base_permission_set_id');
SELECT col_not_null('iam', 'profiles', 'base_permission_set_id');
SELECT fk_ok('iam', 'profiles', 'base_permission_set_id', 'iam', 'permission_sets', 'id');

SELECT has_index('iam', 'profiles', 'idx_profiles_base_permission_set_id', 'index on base_permission_set_id exists');

SELECT finish();
ROLLBACK;
