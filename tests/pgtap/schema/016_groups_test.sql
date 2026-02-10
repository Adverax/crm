BEGIN;
SELECT plan(30);

-- iam.groups table
SELECT has_table('iam', 'groups', 'table iam.groups exists');

SELECT has_column('iam', 'groups', 'id');
SELECT col_type_is('iam', 'groups', 'id', 'uuid');
SELECT col_has_default('iam', 'groups', 'id');
SELECT col_is_pk('iam', 'groups', 'id');

SELECT has_column('iam', 'groups', 'api_name');
SELECT col_type_is('iam', 'groups', 'api_name', 'character varying(100)');
SELECT col_not_null('iam', 'groups', 'api_name');
SELECT col_is_unique('iam', 'groups', 'api_name');

SELECT has_column('iam', 'groups', 'label');
SELECT col_not_null('iam', 'groups', 'label');

SELECT has_column('iam', 'groups', 'group_type');
SELECT col_not_null('iam', 'groups', 'group_type');

SELECT has_column('iam', 'groups', 'related_role_id');
SELECT fk_ok('iam', 'groups', 'related_role_id', 'iam', 'user_roles', 'id');

SELECT has_column('iam', 'groups', 'related_user_id');
SELECT fk_ok('iam', 'groups', 'related_user_id', 'iam', 'users', 'id');

SELECT has_column('iam', 'groups', 'created_at');
SELECT col_not_null('iam', 'groups', 'created_at');

SELECT has_column('iam', 'groups', 'updated_at');
SELECT col_not_null('iam', 'groups', 'updated_at');

-- iam.group_members table
SELECT has_table('iam', 'group_members', 'table iam.group_members exists');

SELECT has_column('iam', 'group_members', 'id');
SELECT col_is_pk('iam', 'group_members', 'id');

SELECT has_column('iam', 'group_members', 'group_id');
SELECT col_not_null('iam', 'group_members', 'group_id');
SELECT fk_ok('iam', 'group_members', 'group_id', 'iam', 'groups', 'id');

SELECT has_column('iam', 'group_members', 'member_user_id');
SELECT fk_ok('iam', 'group_members', 'member_user_id', 'iam', 'users', 'id');

SELECT has_column('iam', 'group_members', 'member_group_id');
SELECT fk_ok('iam', 'group_members', 'member_group_id', 'iam', 'groups', 'id');

SELECT finish();
ROLLBACK;
