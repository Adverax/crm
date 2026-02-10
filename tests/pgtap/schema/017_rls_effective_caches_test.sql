BEGIN;
SELECT plan(25);

-- effective_role_hierarchy
SELECT has_table('security', 'effective_role_hierarchy', 'table security.effective_role_hierarchy exists');

SELECT has_column('security', 'effective_role_hierarchy', 'ancestor_role_id');
SELECT col_not_null('security', 'effective_role_hierarchy', 'ancestor_role_id');
SELECT fk_ok('security', 'effective_role_hierarchy', 'ancestor_role_id', 'iam', 'user_roles', 'id');

SELECT has_column('security', 'effective_role_hierarchy', 'descendant_role_id');
SELECT col_not_null('security', 'effective_role_hierarchy', 'descendant_role_id');
SELECT fk_ok('security', 'effective_role_hierarchy', 'descendant_role_id', 'iam', 'user_roles', 'id');

SELECT has_column('security', 'effective_role_hierarchy', 'depth');
SELECT col_not_null('security', 'effective_role_hierarchy', 'depth');

-- effective_visible_owner
SELECT has_table('security', 'effective_visible_owner', 'table security.effective_visible_owner exists');

SELECT has_column('security', 'effective_visible_owner', 'user_id');
SELECT col_not_null('security', 'effective_visible_owner', 'user_id');
SELECT fk_ok('security', 'effective_visible_owner', 'user_id', 'iam', 'users', 'id');

SELECT has_column('security', 'effective_visible_owner', 'visible_owner_id');
SELECT col_not_null('security', 'effective_visible_owner', 'visible_owner_id');
SELECT fk_ok('security', 'effective_visible_owner', 'visible_owner_id', 'iam', 'users', 'id');

-- effective_group_members
SELECT has_table('security', 'effective_group_members', 'table security.effective_group_members exists');

SELECT has_column('security', 'effective_group_members', 'group_id');
SELECT col_not_null('security', 'effective_group_members', 'group_id');
SELECT fk_ok('security', 'effective_group_members', 'group_id', 'iam', 'groups', 'id');

SELECT has_column('security', 'effective_group_members', 'user_id');
SELECT col_not_null('security', 'effective_group_members', 'user_id');
SELECT fk_ok('security', 'effective_group_members', 'user_id', 'iam', 'users', 'id');

-- effective_object_hierarchy
SELECT has_table('security', 'effective_object_hierarchy', 'table security.effective_object_hierarchy exists');

SELECT has_column('security', 'effective_object_hierarchy', 'ancestor_object_id');
SELECT col_not_null('security', 'effective_object_hierarchy', 'ancestor_object_id');

SELECT finish();
ROLLBACK;
