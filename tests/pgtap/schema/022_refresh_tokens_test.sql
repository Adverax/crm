BEGIN;
SELECT plan(18);

-- iam.refresh_tokens table
SELECT has_table('iam', 'refresh_tokens', 'table iam.refresh_tokens exists');

SELECT has_column('iam', 'refresh_tokens', 'id', 'has id');
SELECT col_type_is('iam', 'refresh_tokens', 'id', 'uuid', 'id is uuid');
SELECT col_has_default('iam', 'refresh_tokens', 'id', 'id has default');
SELECT col_is_pk('iam', 'refresh_tokens', 'id', 'id is PK');

SELECT has_column('iam', 'refresh_tokens', 'user_id', 'has user_id');
SELECT col_not_null('iam', 'refresh_tokens', 'user_id', 'user_id is NOT NULL');
SELECT fk_ok('iam', 'refresh_tokens', 'user_id', 'iam', 'users', 'id', 'FK user_id -> users.id');

SELECT has_column('iam', 'refresh_tokens', 'token_hash', 'has token_hash');
SELECT col_type_is('iam', 'refresh_tokens', 'token_hash', 'character varying(64)', 'token_hash is varchar(64)');
SELECT col_not_null('iam', 'refresh_tokens', 'token_hash', 'token_hash is NOT NULL');

SELECT has_column('iam', 'refresh_tokens', 'expires_at', 'has expires_at');
SELECT col_not_null('iam', 'refresh_tokens', 'expires_at', 'expires_at is NOT NULL');

SELECT has_column('iam', 'refresh_tokens', 'created_at', 'has created_at');
SELECT col_not_null('iam', 'refresh_tokens', 'created_at', 'created_at is NOT NULL');

SELECT has_column('iam', 'refresh_tokens', 'updated_at', 'has updated_at');
SELECT col_not_null('iam', 'refresh_tokens', 'updated_at', 'updated_at is NOT NULL');

SELECT has_index('iam', 'refresh_tokens', 'idx_refresh_tokens_user_id', 'has user_id index');

SELECT finish();
ROLLBACK;
