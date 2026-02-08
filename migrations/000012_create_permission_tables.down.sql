DROP INDEX IF EXISTS iam.idx_permission_set_to_users_user_id;
DROP INDEX IF EXISTS iam.idx_permission_set_to_users_permission_set_id;
DROP TABLE IF EXISTS iam.permission_set_to_users;

DROP INDEX IF EXISTS security.idx_field_permissions_field_id;
DROP INDEX IF EXISTS security.idx_field_permissions_permission_set_id;
DROP TABLE IF EXISTS security.field_permissions;

DROP INDEX IF EXISTS security.idx_object_permissions_object_id;
DROP INDEX IF EXISTS security.idx_object_permissions_permission_set_id;
DROP TABLE IF EXISTS security.object_permissions;

DROP SCHEMA IF EXISTS security;
