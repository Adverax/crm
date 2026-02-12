BEGIN;
SELECT plan(26);

-- effective_role_hierarchy
SELECT has_table('security', 'effective_role_hierarchy', 'table security.effective_role_hierarchy exists');

SELECT has_column('security', 'effective_role_hierarchy', 'ancestor_role_id', 'has ancestor_role_id');
SELECT col_not_null('security', 'effective_role_hierarchy', 'ancestor_role_id', 'ancestor_role_id is NOT NULL');
SELECT fk_ok('security', 'effective_role_hierarchy', 'ancestor_role_id', 'iam', 'user_roles', 'id', 'FK ancestor_role_id -> user_roles.id');

SELECT has_column('security', 'effective_role_hierarchy', 'descendant_role_id', 'has descendant_role_id');
SELECT col_not_null('security', 'effective_role_hierarchy', 'descendant_role_id', 'descendant_role_id is NOT NULL');
SELECT fk_ok('security', 'effective_role_hierarchy', 'descendant_role_id', 'iam', 'user_roles', 'id', 'FK descendant_role_id -> user_roles.id');

SELECT has_column('security', 'effective_role_hierarchy', 'depth', 'has depth');
SELECT col_not_null('security', 'effective_role_hierarchy', 'depth', 'depth is NOT NULL');

-- effective_visible_owner
SELECT has_table('security', 'effective_visible_owner', 'table security.effective_visible_owner exists');

SELECT has_column('security', 'effective_visible_owner', 'user_id', 'has user_id');
SELECT col_not_null('security', 'effective_visible_owner', 'user_id', 'user_id is NOT NULL');
SELECT fk_ok('security', 'effective_visible_owner', 'user_id', 'iam', 'users', 'id', 'FK user_id -> users.id');

SELECT has_column('security', 'effective_visible_owner', 'visible_owner_id', 'has visible_owner_id');
SELECT col_not_null('security', 'effective_visible_owner', 'visible_owner_id', 'visible_owner_id is NOT NULL');
SELECT fk_ok('security', 'effective_visible_owner', 'visible_owner_id', 'iam', 'users', 'id', 'FK visible_owner_id -> users.id');

-- effective_group_members
SELECT has_table('security', 'effective_group_members', 'table security.effective_group_members exists');

SELECT has_column('security', 'effective_group_members', 'group_id', 'has group_id');
SELECT col_not_null('security', 'effective_group_members', 'group_id', 'group_id is NOT NULL');
SELECT fk_ok('security', 'effective_group_members', 'group_id', 'iam', 'groups', 'id', 'FK group_id -> groups.id');

SELECT has_column('security', 'effective_group_members', 'user_id', 'has user_id');
SELECT col_not_null('security', 'effective_group_members', 'user_id', 'user_id is NOT NULL');
SELECT fk_ok('security', 'effective_group_members', 'user_id', 'iam', 'users', 'id', 'FK user_id -> users.id');

-- effective_object_hierarchy
SELECT has_table('security', 'effective_object_hierarchy', 'table security.effective_object_hierarchy exists');

SELECT has_column('security', 'effective_object_hierarchy', 'ancestor_object_id', 'has ancestor_object_id');
SELECT col_not_null('security', 'effective_object_hierarchy', 'ancestor_object_id', 'ancestor_object_id is NOT NULL');

SELECT finish();
ROLLBACK;
