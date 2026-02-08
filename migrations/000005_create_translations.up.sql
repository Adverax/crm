CREATE TABLE IF NOT EXISTS metadata.translations (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_type   VARCHAR(100) NOT NULL,
    resource_id     UUID         NOT NULL,
    field_name      VARCHAR(100) NOT NULL,
    locale          VARCHAR(10)  NOT NULL,
    value           TEXT         NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),

    UNIQUE (resource_type, resource_id, field_name, locale)
);

CREATE INDEX IF NOT EXISTS idx_translations_resource
    ON metadata.translations (resource_type, resource_id);
