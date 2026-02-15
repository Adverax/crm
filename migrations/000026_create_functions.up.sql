CREATE TABLE metadata.functions (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT         NOT NULL DEFAULT '',
    params      JSONB        NOT NULL DEFAULT '[]',
    return_type VARCHAR(20)  NOT NULL DEFAULT 'any'
                CHECK (return_type IN ('string', 'number', 'boolean', 'list', 'map', 'any')),
    body        TEXT         NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_name_format CHECK (name ~ '^[a-z][a-z0-9_]*$'),
    CONSTRAINT chk_body_size CHECK (length(body) <= 4096),
    CONSTRAINT chk_params_size CHECK (jsonb_array_length(params) <= 10)
);

CREATE INDEX idx_functions_name ON metadata.functions(name);
