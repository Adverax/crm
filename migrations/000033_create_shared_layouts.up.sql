CREATE TABLE metadata.shared_layouts (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    api_name   VARCHAR(63)  NOT NULL,
    type       VARCHAR(20)  NOT NULL,
    label      VARCHAR(255) NOT NULL DEFAULT '',
    config     JSONB        NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT shared_layouts_api_name_unique UNIQUE (api_name),
    CONSTRAINT shared_layouts_type_check CHECK (type IN ('field', 'section', 'list')),
    CONSTRAINT chk_shared_layout_api_name_format CHECK (api_name ~ '^[a-z][a-z0-9_]*$')
);
