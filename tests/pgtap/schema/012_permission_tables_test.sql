BEGIN;
SELECT plan(31);

-- Схема существует
SELECT has_schema('security');

-- object_permissions
SELECT has_table('security', 'object_permissions', 'table security.object_permissions exists');

SELECT has_column('security', 'object_permissions', 'id');
SELECT col_type_is('security', 'object_permissions', 'id', 'uuid');
SELECT col_is_pk('security', 'object_permissions', 'id');

SELECT has_column('security', 'object_permissions', 'permission_set_id');
SELECT col_not_null('security', 'object_permissions', 'permission_set_id');
SELECT fk_ok('security', 'object_permissions', 'permission_set_id', 'iam', 'permission_sets', 'id');

SELECT has_column('security', 'object_permissions', 'object_id');
SELECT col_not_null('security', 'object_permissions', 'object_id');
SELECT fk_ok('security', 'object_permissions', 'object_id', 'metadata', 'object_definitions', 'id');

SELECT has_column('security', 'object_permissions', 'permissions');
SELECT col_not_null('security', 'object_permissions', 'permissions');
SELECT col_has_default('security', 'object_permissions', 'permissions');

SELECT col_not_null('security', 'object_permissions', 'created_at');
SELECT col_not_null('security', 'object_permissions', 'updated_at');

-- field_permissions
SELECT has_table('security', 'field_permissions', 'table security.field_permissions exists');

SELECT has_column('security', 'field_permissions', 'id');
SELECT col_is_pk('security', 'field_permissions', 'id');

SELECT has_column('security', 'field_permissions', 'permission_set_id');
SELECT col_not_null('security', 'field_permissions', 'permission_set_id');
SELECT fk_ok('security', 'field_permissions', 'permission_set_id', 'iam', 'permission_sets', 'id');

SELECT has_column('security', 'field_permissions', 'field_id');
SELECT col_not_null('security', 'field_permissions', 'field_id');
SELECT fk_ok('security', 'field_permissions', 'field_id', 'metadata', 'field_definitions', 'id');

SELECT has_column('security', 'field_permissions', 'permissions');
SELECT col_not_null('security', 'field_permissions', 'permissions');

-- permission_set_to_users
SELECT has_table('iam', 'permission_set_to_users', 'table iam.permission_set_to_users exists');

SELECT has_column('iam', 'permission_set_to_users', 'permission_set_id');
SELECT fk_ok('iam', 'permission_set_to_users', 'permission_set_id', 'iam', 'permission_sets', 'id');

SELECT has_column('iam', 'permission_set_to_users', 'user_id');
SELECT fk_ok('iam', 'permission_set_to_users', 'user_id', 'iam', 'users', 'id');

SELECT finish();
ROLLBACK;
