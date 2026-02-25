CREATE TABLE metadata.credentials (
    id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    code                 VARCHAR(100) UNIQUE NOT NULL,
    name                 VARCHAR(255) NOT NULL,
    description          TEXT         NOT NULL DEFAULT '',
    type                 VARCHAR(20)  NOT NULL,
    base_url             VARCHAR(500) NOT NULL,
    auth_data_encrypted  BYTEA        NOT NULL,
    auth_data_nonce      BYTEA        NOT NULL,
    is_active            BOOLEAN      NOT NULL DEFAULT true,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_credential_code_format CHECK (code ~ '^[a-z][a-z0-9_]*$'),
    CONSTRAINT chk_credential_type CHECK (type IN ('api_key', 'basic', 'oauth2_client')),
    CONSTRAINT chk_credential_base_url_https CHECK (base_url LIKE 'https://%')
);

CREATE TABLE metadata.credential_tokens (
    credential_id          UUID        PRIMARY KEY REFERENCES metadata.credentials(id) ON DELETE CASCADE,
    access_token_encrypted BYTEA       NOT NULL,
    access_token_nonce     BYTEA       NOT NULL,
    token_type             VARCHAR(50) NOT NULL DEFAULT 'Bearer',
    expires_at             TIMESTAMPTZ NOT NULL,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE metadata.credential_usage_log (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    credential_id   UUID         NOT NULL REFERENCES metadata.credentials(id) ON DELETE CASCADE,
    procedure_code  VARCHAR(100),
    request_url     VARCHAR(500) NOT NULL,
    response_status INT,
    success         BOOLEAN      NOT NULL DEFAULT true,
    error_message   TEXT,
    duration_ms     INT,
    user_id         UUID,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_credential_usage_credential_id ON metadata.credential_usage_log(credential_id);
CREATE INDEX idx_credential_usage_created_at ON metadata.credential_usage_log(created_at);
