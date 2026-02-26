CREATE TABLE IF NOT EXISTS metadata.picklist_definitions (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    api_name    VARCHAR(100) NOT NULL UNIQUE
                CHECK (api_name ~ '^[a-z][a-z0-9_]*$'),
    label       VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS metadata.picklist_values (
    id                     UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    picklist_definition_id UUID         NOT NULL REFERENCES metadata.picklist_definitions(id) ON DELETE CASCADE,
    value                  VARCHAR(255) NOT NULL,
    label                  VARCHAR(255) NOT NULL,
    sort_order             INTEGER      NOT NULL DEFAULT 0,
    is_default             BOOLEAN      NOT NULL DEFAULT false,
    is_active              BOOLEAN      NOT NULL DEFAULT true,
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),

    UNIQUE (picklist_definition_id, value)
);

CREATE INDEX IF NOT EXISTS idx_picklist_values_picklist_definition_id
    ON metadata.picklist_values (picklist_definition_id);
