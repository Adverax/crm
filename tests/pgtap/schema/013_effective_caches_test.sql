BEGIN;
SELECT plan(28);

-- effective_ols
SELECT has_table('security', 'effective_ols', 'table security.effective_ols exists');

SELECT has_column('security', 'effective_ols', 'user_id');
SELECT col_not_null('security', 'effective_ols', 'user_id');
SELECT fk_ok('security', 'effective_ols', 'user_id', 'iam', 'users', 'id');

SELECT has_column('security', 'effective_ols', 'object_id');
SELECT col_not_null('security', 'effective_ols', 'object_id');
SELECT fk_ok('security', 'effective_ols', 'object_id', 'metadata', 'object_definitions', 'id');

SELECT has_column('security', 'effective_ols', 'permissions');
SELECT col_not_null('security', 'effective_ols', 'permissions');

-- effective_fls
SELECT has_table('security', 'effective_fls', 'table security.effective_fls exists');

SELECT has_column('security', 'effective_fls', 'user_id');
SELECT col_not_null('security', 'effective_fls', 'user_id');
SELECT fk_ok('security', 'effective_fls', 'user_id', 'iam', 'users', 'id');

SELECT has_column('security', 'effective_fls', 'field_id');
SELECT col_not_null('security', 'effective_fls', 'field_id');
SELECT fk_ok('security', 'effective_fls', 'field_id', 'metadata', 'field_definitions', 'id');

SELECT has_column('security', 'effective_fls', 'permissions');
SELECT col_not_null('security', 'effective_fls', 'permissions');

-- effective_field_lists
SELECT has_table('security', 'effective_field_lists', 'table security.effective_field_lists exists');

SELECT has_column('security', 'effective_field_lists', 'user_id');
SELECT col_not_null('security', 'effective_field_lists', 'user_id');

SELECT has_column('security', 'effective_field_lists', 'object_id');
SELECT col_not_null('security', 'effective_field_lists', 'object_id');

SELECT has_column('security', 'effective_field_lists', 'mask');
SELECT col_not_null('security', 'effective_field_lists', 'mask');

SELECT has_column('security', 'effective_field_lists', 'field_names');

-- security_outbox
SELECT has_table('security', 'security_outbox', 'table security.security_outbox exists');

SELECT has_column('security', 'security_outbox', 'event_type');
SELECT col_not_null('security', 'security_outbox', 'event_type');

SELECT finish();
ROLLBACK;
