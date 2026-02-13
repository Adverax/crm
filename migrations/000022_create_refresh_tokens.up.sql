CREATE TABLE iam.refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user_id ON iam.refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON iam.refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON iam.refresh_tokens(expires_at);
