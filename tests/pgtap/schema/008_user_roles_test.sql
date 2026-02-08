BEGIN;
SELECT plan(18);

-- Схема существует
SELECT has_schema('iam');

-- Таблица существует
SELECT has_table('iam', 'user_roles', 'table iam.user_roles exists');

-- Колонки и типы
SELECT has_column('iam', 'user_roles', 'id');
SELECT col_type_is('iam', 'user_roles', 'id', 'uuid');
SELECT col_has_default('iam', 'user_roles', 'id');
SELECT col_is_pk('iam', 'user_roles', 'id');

SELECT has_column('iam', 'user_roles', 'api_name');
SELECT col_type_is('iam', 'user_roles', 'api_name', 'character varying(100)');
SELECT col_not_null('iam', 'user_roles', 'api_name');
SELECT col_is_unique('iam', 'user_roles', 'api_name');

SELECT has_column('iam', 'user_roles', 'label');
SELECT col_not_null('iam', 'user_roles', 'label');

SELECT has_column('iam', 'user_roles', 'description');
SELECT col_has_default('iam', 'user_roles', 'description');

SELECT has_column('iam', 'user_roles', 'parent_id');
SELECT fk_ok('iam', 'user_roles', 'parent_id', 'iam', 'user_roles', 'id');

-- Timestamps
SELECT col_not_null('iam', 'user_roles', 'created_at');
SELECT col_not_null('iam', 'user_roles', 'updated_at');

SELECT finish();
ROLLBACK;
