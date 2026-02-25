BEGIN;
SELECT plan(30);

-- ===== metadata.credentials =====
SELECT has_table('metadata', 'credentials', 'has credentials table');

SELECT has_column('metadata', 'credentials', 'id', 'has id column');
SELECT col_type_is('metadata', 'credentials', 'id', 'uuid', 'id is uuid');
SELECT col_is_pk('metadata', 'credentials', 'id', 'id is PK');

SELECT has_column('metadata', 'credentials', 'code', 'has code column');
SELECT col_is_unique('metadata', 'credentials', 'code', 'code is unique');

SELECT has_column('metadata', 'credentials', 'name', 'has name column');
SELECT has_column('metadata', 'credentials', 'description', 'has description column');
SELECT has_column('metadata', 'credentials', 'type', 'has type column');
SELECT has_column('metadata', 'credentials', 'base_url', 'has base_url column');
SELECT has_column('metadata', 'credentials', 'auth_data_encrypted', 'has auth_data_encrypted column');
SELECT col_type_is('metadata', 'credentials', 'auth_data_encrypted', 'bytea', 'auth_data_encrypted is bytea');
SELECT has_column('metadata', 'credentials', 'auth_data_nonce', 'has auth_data_nonce column');
SELECT col_type_is('metadata', 'credentials', 'auth_data_nonce', 'bytea', 'auth_data_nonce is bytea');
SELECT has_column('metadata', 'credentials', 'is_active', 'has is_active column');
SELECT has_column('metadata', 'credentials', 'created_at', 'has created_at column');
SELECT has_column('metadata', 'credentials', 'updated_at', 'has updated_at column');

SELECT has_check('metadata', 'credentials', 'has check constraints on credentials');

-- ===== metadata.credential_tokens =====
SELECT has_table('metadata', 'credential_tokens', 'has credential_tokens table');

SELECT has_column('metadata', 'credential_tokens', 'credential_id', 'has credential_id column');
SELECT col_is_pk('metadata', 'credential_tokens', 'credential_id', 'credential_id is PK');
SELECT col_type_is('metadata', 'credential_tokens', 'credential_id', 'uuid', 'credential_id is uuid');

SELECT has_column('metadata', 'credential_tokens', 'access_token_encrypted', 'has access_token_encrypted column');
SELECT has_column('metadata', 'credential_tokens', 'token_type', 'has token_type column');
SELECT has_column('metadata', 'credential_tokens', 'expires_at', 'has expires_at column');

SELECT fk_ok('metadata', 'credential_tokens', 'credential_id', 'metadata', 'credentials', 'id', 'credential_tokens FK to credentials');

-- ===== metadata.credential_usage_log =====
SELECT has_table('metadata', 'credential_usage_log', 'has credential_usage_log table');

SELECT has_column('metadata', 'credential_usage_log', 'id', 'usage_log has id column');
SELECT has_column('metadata', 'credential_usage_log', 'credential_id', 'usage_log has credential_id column');

SELECT fk_ok('metadata', 'credential_usage_log', 'credential_id', 'metadata', 'credentials', 'id', 'usage_log FK to credentials');

SELECT * FROM finish();
ROLLBACK;
