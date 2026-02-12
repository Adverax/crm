BEGIN;
SELECT plan(31);

-- iam.groups table
SELECT has_table('iam', 'groups', 'table iam.groups exists');

SELECT has_column('iam', 'groups', 'id', 'has id');
SELECT col_type_is('iam', 'groups', 'id', 'uuid', 'id is uuid');
SELECT col_has_default('iam', 'groups', 'id', 'id has default');
SELECT col_is_pk('iam', 'groups', 'id', 'id is PK');

SELECT has_column('iam', 'groups', 'api_name', 'has api_name');
SELECT col_type_is('iam', 'groups', 'api_name', 'character varying(100)', 'api_name is character varying(100)');
SELECT col_not_null('iam', 'groups', 'api_name', 'api_name is NOT NULL');
SELECT col_is_unique('iam', 'groups', 'api_name', 'api_name is unique');

SELECT has_column('iam', 'groups', 'label', 'has label');
SELECT col_not_null('iam', 'groups', 'label', 'label is NOT NULL');

SELECT has_column('iam', 'groups', 'group_type', 'has group_type');
SELECT col_not_null('iam', 'groups', 'group_type', 'group_type is NOT NULL');

SELECT has_column('iam', 'groups', 'related_role_id', 'has related_role_id');
SELECT fk_ok('iam', 'groups', 'related_role_id', 'iam', 'user_roles', 'id', 'FK related_role_id -> user_roles.id');

SELECT has_column('iam', 'groups', 'related_user_id', 'has related_user_id');
SELECT fk_ok('iam', 'groups', 'related_user_id', 'iam', 'users', 'id', 'FK related_user_id -> users.id');

SELECT has_column('iam', 'groups', 'created_at', 'has created_at');
SELECT col_not_null('iam', 'groups', 'created_at', 'created_at is NOT NULL');

SELECT has_column('iam', 'groups', 'updated_at', 'has updated_at');
SELECT col_not_null('iam', 'groups', 'updated_at', 'updated_at is NOT NULL');

-- iam.group_members table
SELECT has_table('iam', 'group_members', 'table iam.group_members exists');

SELECT has_column('iam', 'group_members', 'id', 'has id');
SELECT col_is_pk('iam', 'group_members', 'id', 'id is PK');

SELECT has_column('iam', 'group_members', 'group_id', 'has group_id');
SELECT col_not_null('iam', 'group_members', 'group_id', 'group_id is NOT NULL');
SELECT fk_ok('iam', 'group_members', 'group_id', 'iam', 'groups', 'id', 'FK group_id -> groups.id');

SELECT has_column('iam', 'group_members', 'member_user_id', 'has member_user_id');
SELECT fk_ok('iam', 'group_members', 'member_user_id', 'iam', 'users', 'id', 'FK member_user_id -> users.id');

SELECT has_column('iam', 'group_members', 'member_group_id', 'has member_group_id');
SELECT fk_ok('iam', 'group_members', 'member_group_id', 'iam', 'groups', 'id', 'FK member_group_id -> groups.id');

SELECT finish();
ROLLBACK;
