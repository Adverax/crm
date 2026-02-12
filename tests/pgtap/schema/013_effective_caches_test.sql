BEGIN;
SELECT plan(29);

-- effective_ols
SELECT has_table('security', 'effective_ols', 'table security.effective_ols exists');

SELECT has_column('security', 'effective_ols', 'user_id', 'has user_id');
SELECT col_not_null('security', 'effective_ols', 'user_id', 'user_id is NOT NULL');
SELECT fk_ok('security', 'effective_ols', 'user_id', 'iam', 'users', 'id', 'FK user_id -> users.id');

SELECT has_column('security', 'effective_ols', 'object_id', 'has object_id');
SELECT col_not_null('security', 'effective_ols', 'object_id', 'object_id is NOT NULL');
SELECT fk_ok('security', 'effective_ols', 'object_id', 'metadata', 'object_definitions', 'id', 'FK object_id -> object_definitions.id');

SELECT has_column('security', 'effective_ols', 'permissions', 'has permissions');
SELECT col_not_null('security', 'effective_ols', 'permissions', 'permissions is NOT NULL');

-- effective_fls
SELECT has_table('security', 'effective_fls', 'table security.effective_fls exists');

SELECT has_column('security', 'effective_fls', 'user_id', 'has user_id');
SELECT col_not_null('security', 'effective_fls', 'user_id', 'user_id is NOT NULL');
SELECT fk_ok('security', 'effective_fls', 'user_id', 'iam', 'users', 'id', 'FK user_id -> users.id');

SELECT has_column('security', 'effective_fls', 'field_id', 'has field_id');
SELECT col_not_null('security', 'effective_fls', 'field_id', 'field_id is NOT NULL');
SELECT fk_ok('security', 'effective_fls', 'field_id', 'metadata', 'field_definitions', 'id', 'FK field_id -> field_definitions.id');

SELECT has_column('security', 'effective_fls', 'permissions', 'has permissions');
SELECT col_not_null('security', 'effective_fls', 'permissions', 'permissions is NOT NULL');

-- effective_field_lists
SELECT has_table('security', 'effective_field_lists', 'table security.effective_field_lists exists');

SELECT has_column('security', 'effective_field_lists', 'user_id', 'has user_id');
SELECT col_not_null('security', 'effective_field_lists', 'user_id', 'user_id is NOT NULL');

SELECT has_column('security', 'effective_field_lists', 'object_id', 'has object_id');
SELECT col_not_null('security', 'effective_field_lists', 'object_id', 'object_id is NOT NULL');

SELECT has_column('security', 'effective_field_lists', 'mask', 'has mask');
SELECT col_not_null('security', 'effective_field_lists', 'mask', 'mask is NOT NULL');

SELECT has_column('security', 'effective_field_lists', 'field_names', 'has field_names');

-- security_outbox
SELECT has_table('security', 'security_outbox', 'table security.security_outbox exists');

SELECT has_column('security', 'security_outbox', 'event_type', 'has event_type');
SELECT col_not_null('security', 'security_outbox', 'event_type', 'event_type is NOT NULL');

SELECT finish();
ROLLBACK;
