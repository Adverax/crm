BEGIN;
SELECT plan(24);

-- Таблица существует
SELECT has_table('iam', 'users', 'table iam.users exists');

-- Колонки и типы
SELECT has_column('iam', 'users', 'id');
SELECT col_type_is('iam', 'users', 'id', 'uuid');
SELECT col_has_default('iam', 'users', 'id');
SELECT col_is_pk('iam', 'users', 'id');

SELECT has_column('iam', 'users', 'username');
SELECT col_type_is('iam', 'users', 'username', 'character varying(100)');
SELECT col_not_null('iam', 'users', 'username');
SELECT col_is_unique('iam', 'users', 'username');

SELECT has_column('iam', 'users', 'email');
SELECT col_type_is('iam', 'users', 'email', 'character varying(255)');
SELECT col_not_null('iam', 'users', 'email');
SELECT col_is_unique('iam', 'users', 'email');

SELECT has_column('iam', 'users', 'first_name');
SELECT has_column('iam', 'users', 'last_name');

SELECT has_column('iam', 'users', 'profile_id');
SELECT col_not_null('iam', 'users', 'profile_id');
SELECT fk_ok('iam', 'users', 'profile_id', 'iam', 'profiles', 'id');

SELECT has_column('iam', 'users', 'role_id');
SELECT fk_ok('iam', 'users', 'role_id', 'iam', 'user_roles', 'id');

SELECT has_column('iam', 'users', 'is_active');
SELECT col_not_null('iam', 'users', 'is_active');
SELECT col_has_default('iam', 'users', 'is_active');

-- Timestamps
SELECT col_not_null('iam', 'users', 'created_at');
SELECT col_not_null('iam', 'users', 'updated_at');

SELECT finish();
ROLLBACK;
