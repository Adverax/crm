BEGIN;
SELECT plan(16);

-- Таблица существует
SELECT has_table('iam', 'permission_sets', 'table iam.permission_sets exists');

-- Колонки и типы
SELECT has_column('iam', 'permission_sets', 'id', 'has id');
SELECT col_type_is('iam', 'permission_sets', 'id', 'uuid', 'id is uuid');
SELECT col_has_default('iam', 'permission_sets', 'id', 'id has default');
SELECT col_is_pk('iam', 'permission_sets', 'id', 'id is PK');

SELECT has_column('iam', 'permission_sets', 'api_name', 'has api_name');
SELECT col_type_is('iam', 'permission_sets', 'api_name', 'character varying(100)', 'api_name is character varying(100)');
SELECT col_not_null('iam', 'permission_sets', 'api_name', 'api_name is NOT NULL');
SELECT col_is_unique('iam', 'permission_sets', 'api_name', 'api_name is unique');

SELECT has_column('iam', 'permission_sets', 'label', 'has label');
SELECT col_not_null('iam', 'permission_sets', 'label', 'label is NOT NULL');

SELECT has_column('iam', 'permission_sets', 'description', 'has description');
SELECT col_has_default('iam', 'permission_sets', 'description', 'description has default');

SELECT has_column('iam', 'permission_sets', 'ps_type', 'has ps_type');
SELECT col_not_null('iam', 'permission_sets', 'ps_type', 'ps_type is NOT NULL');
SELECT has_check('iam', 'permission_sets', 'permission_sets has CHECK constraint');

SELECT finish();
ROLLBACK;
