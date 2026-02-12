BEGIN;
SELECT plan(25);

-- Таблица существует
SELECT has_table('iam', 'users', 'table iam.users exists');

-- Колонки и типы
SELECT has_column('iam', 'users', 'id', 'has id');
SELECT col_type_is('iam', 'users', 'id', 'uuid', 'id is uuid');
SELECT col_has_default('iam', 'users', 'id', 'id has default');
SELECT col_is_pk('iam', 'users', 'id', 'id is PK');

SELECT has_column('iam', 'users', 'username', 'has username');
SELECT col_type_is('iam', 'users', 'username', 'character varying(100)', 'username is character varying(100)');
SELECT col_not_null('iam', 'users', 'username', 'username is NOT NULL');
SELECT col_is_unique('iam', 'users', 'username', 'username is unique');

SELECT has_column('iam', 'users', 'email', 'has email');
SELECT col_type_is('iam', 'users', 'email', 'character varying(255)', 'email is character varying(255)');
SELECT col_not_null('iam', 'users', 'email', 'email is NOT NULL');
SELECT col_is_unique('iam', 'users', 'email', 'email is unique');

SELECT has_column('iam', 'users', 'first_name', 'has first_name');
SELECT has_column('iam', 'users', 'last_name', 'has last_name');

SELECT has_column('iam', 'users', 'profile_id', 'has profile_id');
SELECT col_not_null('iam', 'users', 'profile_id', 'profile_id is NOT NULL');
SELECT fk_ok('iam', 'users', 'profile_id', 'iam', 'profiles', 'id', 'FK profile_id -> profiles.id');

SELECT has_column('iam', 'users', 'role_id', 'has role_id');
SELECT fk_ok('iam', 'users', 'role_id', 'iam', 'user_roles', 'id', 'FK role_id -> user_roles.id');

SELECT has_column('iam', 'users', 'is_active', 'has is_active');
SELECT col_not_null('iam', 'users', 'is_active', 'is_active is NOT NULL');
SELECT col_has_default('iam', 'users', 'is_active', 'is_active has default');

-- Timestamps
SELECT col_not_null('iam', 'users', 'created_at', 'created_at is NOT NULL');
SELECT col_not_null('iam', 'users', 'updated_at', 'updated_at is NOT NULL');

SELECT finish();
ROLLBACK;
