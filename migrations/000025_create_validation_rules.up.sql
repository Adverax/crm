CREATE TABLE IF NOT EXISTS metadata.validation_rules (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    object_id       UUID        NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    api_name        VARCHAR(100) NOT NULL
                    CHECK (api_name ~ '^[a-z][a-z0-9_]*$'),
    label           VARCHAR(255) NOT NULL,
    description     TEXT         NOT NULL DEFAULT '',
    expression      TEXT         NOT NULL,
    error_message   VARCHAR(500) NOT NULL,
    error_code      VARCHAR(100) NOT NULL DEFAULT 'validation_failed',
    severity        VARCHAR(20)  NOT NULL DEFAULT 'error'
                    CHECK (severity IN ('error', 'warning')),
    when_expression TEXT,
    applies_to      VARCHAR(30)  NOT NULL DEFAULT 'create,update',
    sort_order      INT          NOT NULL DEFAULT 0,
    is_active       BOOLEAN      NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE(object_id, api_name)
);

CREATE INDEX idx_validation_rules_object_id ON metadata.validation_rules(object_id);
