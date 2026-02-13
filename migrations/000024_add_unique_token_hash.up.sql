-- Replace regular indexes with UNIQUE indexes on token_hash columns.
-- Ensures no duplicate token hashes and prevents TOCTOU race conditions.

DROP INDEX IF EXISTS iam.idx_refresh_tokens_token_hash;
CREATE UNIQUE INDEX idx_refresh_tokens_token_hash ON iam.refresh_tokens(token_hash);

DROP INDEX IF EXISTS iam.idx_password_reset_tokens_token_hash;
CREATE UNIQUE INDEX idx_password_reset_tokens_token_hash ON iam.password_reset_tokens(token_hash);
