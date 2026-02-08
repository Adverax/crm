CREATE TABLE IF NOT EXISTS security.effective_ols (
    user_id     UUID        NOT NULL
                            REFERENCES iam.users(id) ON DELETE CASCADE,
    object_id   UUID        NOT NULL
                            REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    permissions INT         NOT NULL DEFAULT 0,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, object_id)
);

CREATE INDEX IF NOT EXISTS idx_effective_ols_user_id
    ON security.effective_ols (user_id);

CREATE TABLE IF NOT EXISTS security.effective_fls (
    user_id     UUID        NOT NULL
                            REFERENCES iam.users(id) ON DELETE CASCADE,
    field_id    UUID        NOT NULL
                            REFERENCES metadata.field_definitions(id) ON DELETE CASCADE,
    permissions INT         NOT NULL DEFAULT 0,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, field_id)
);

CREATE INDEX IF NOT EXISTS idx_effective_fls_user_id
    ON security.effective_fls (user_id);

CREATE TABLE IF NOT EXISTS security.effective_field_lists (
    user_id   UUID        NOT NULL
                          REFERENCES iam.users(id) ON DELETE CASCADE,
    object_id UUID        NOT NULL
                          REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    mask      INT         NOT NULL,
    field_names TEXT[]     NOT NULL DEFAULT '{}',
    computed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, object_id, mask)
);

CREATE INDEX IF NOT EXISTS idx_effective_field_lists_user_id
    ON security.effective_field_lists (user_id);

CREATE TABLE IF NOT EXISTS security.security_outbox (
    id           BIGSERIAL   PRIMARY KEY,
    event_type   VARCHAR(50) NOT NULL,
    entity_type  VARCHAR(50) NOT NULL,
    entity_id    UUID        NOT NULL,
    payload      JSONB       NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_security_outbox_unprocessed
    ON security.security_outbox (created_at)
    WHERE processed_at IS NULL;

CREATE OR REPLACE FUNCTION security.notify_outbox() RETURNS trigger AS $$
BEGIN
    PERFORM pg_notify('security_outbox', NEW.id::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_outbox_notify
    AFTER INSERT ON security.security_outbox
    FOR EACH ROW EXECUTE FUNCTION security.notify_outbox();
