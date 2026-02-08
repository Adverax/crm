BEGIN;
SELECT plan(16);

-- Таблица существует
SELECT has_table('iam', 'permission_sets', 'table iam.permission_sets exists');

-- Колонки и типы
SELECT has_column('iam', 'permission_sets', 'id');
SELECT col_type_is('iam', 'permission_sets', 'id', 'uuid');
SELECT col_has_default('iam', 'permission_sets', 'id');
SELECT col_is_pk('iam', 'permission_sets', 'id');

SELECT has_column('iam', 'permission_sets', 'api_name');
SELECT col_type_is('iam', 'permission_sets', 'api_name', 'character varying(100)');
SELECT col_not_null('iam', 'permission_sets', 'api_name');
SELECT col_is_unique('iam', 'permission_sets', 'api_name');

SELECT has_column('iam', 'permission_sets', 'label');
SELECT col_not_null('iam', 'permission_sets', 'label');

SELECT has_column('iam', 'permission_sets', 'description');
SELECT col_has_default('iam', 'permission_sets', 'description');

SELECT has_column('iam', 'permission_sets', 'ps_type');
SELECT col_not_null('iam', 'permission_sets', 'ps_type');
SELECT has_check('iam', 'permission_sets', 'permission_sets has CHECK constraint');

SELECT finish();
ROLLBACK;
