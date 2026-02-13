-- Revert UNIQUE indexes to regular indexes.

DROP INDEX IF EXISTS iam.idx_refresh_tokens_token_hash;
CREATE INDEX idx_refresh_tokens_token_hash ON iam.refresh_tokens(token_hash);

DROP INDEX IF EXISTS iam.idx_password_reset_tokens_token_hash;
CREATE INDEX idx_password_reset_tokens_token_hash ON iam.password_reset_tokens(token_hash);
