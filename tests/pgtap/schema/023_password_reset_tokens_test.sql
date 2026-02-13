BEGIN;
SELECT plan(15);

-- iam.password_reset_tokens table
SELECT has_table('iam', 'password_reset_tokens', 'table iam.password_reset_tokens exists');

SELECT has_column('iam', 'password_reset_tokens', 'id', 'has id');
SELECT col_type_is('iam', 'password_reset_tokens', 'id', 'uuid', 'id is uuid');
SELECT col_has_default('iam', 'password_reset_tokens', 'id', 'id has default');
SELECT col_is_pk('iam', 'password_reset_tokens', 'id', 'id is PK');

SELECT has_column('iam', 'password_reset_tokens', 'user_id', 'has user_id');
SELECT col_not_null('iam', 'password_reset_tokens', 'user_id', 'user_id is NOT NULL');
SELECT fk_ok('iam', 'password_reset_tokens', 'user_id', 'iam', 'users', 'id', 'FK user_id -> users.id');

SELECT has_column('iam', 'password_reset_tokens', 'token_hash', 'has token_hash');
SELECT col_type_is('iam', 'password_reset_tokens', 'token_hash', 'character varying(64)', 'token_hash is varchar(64)');
SELECT col_not_null('iam', 'password_reset_tokens', 'token_hash', 'token_hash is NOT NULL');

SELECT has_column('iam', 'password_reset_tokens', 'expires_at', 'has expires_at');
SELECT col_not_null('iam', 'password_reset_tokens', 'expires_at', 'expires_at is NOT NULL');

SELECT has_column('iam', 'password_reset_tokens', 'used_at', 'has used_at');

SELECT has_column('iam', 'password_reset_tokens', 'created_at', 'has created_at');

SELECT finish();
ROLLBACK;
