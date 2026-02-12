BEGIN;
SELECT plan(18);

-- Схема существует
SELECT has_schema('iam');

-- Таблица существует
SELECT has_table('iam', 'user_roles', 'table iam.user_roles exists');

-- Колонки и типы
SELECT has_column('iam', 'user_roles', 'id', 'has id');
SELECT col_type_is('iam', 'user_roles', 'id', 'uuid', 'id is uuid');
SELECT col_has_default('iam', 'user_roles', 'id', 'id has default');
SELECT col_is_pk('iam', 'user_roles', 'id', 'id is PK');

SELECT has_column('iam', 'user_roles', 'api_name', 'has api_name');
SELECT col_type_is('iam', 'user_roles', 'api_name', 'character varying(100)', 'api_name is character varying(100)');
SELECT col_not_null('iam', 'user_roles', 'api_name', 'api_name is NOT NULL');
SELECT col_is_unique('iam', 'user_roles', 'api_name', 'api_name is unique');

SELECT has_column('iam', 'user_roles', 'label', 'has label');
SELECT col_not_null('iam', 'user_roles', 'label', 'label is NOT NULL');

SELECT has_column('iam', 'user_roles', 'description', 'has description');
SELECT col_has_default('iam', 'user_roles', 'description', 'description has default');

SELECT has_column('iam', 'user_roles', 'parent_id', 'has parent_id');
SELECT fk_ok('iam', 'user_roles', 'parent_id', 'iam', 'user_roles', 'id', 'FK parent_id -> user_roles.id');

-- Timestamps
SELECT col_not_null('iam', 'user_roles', 'created_at', 'created_at is NOT NULL');
SELECT col_not_null('iam', 'user_roles', 'updated_at', 'updated_at is NOT NULL');

SELECT finish();
ROLLBACK;
