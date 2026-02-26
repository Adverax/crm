CREATE TABLE IF NOT EXISTS metadata.field_definitions (
    -- Идентификация
    id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    object_id            UUID         NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    api_name             VARCHAR(100) NOT NULL
                         CHECK (api_name ~ '^[a-z][a-z0-9_]*$'),
    label                VARCHAR(255) NOT NULL,
    description          TEXT         NOT NULL DEFAULT '',
    help_text            TEXT         NOT NULL DEFAULT '',

    -- Типизация
    field_type           VARCHAR(20)  NOT NULL,
    field_subtype        VARCHAR(20),

    -- Reference-связь (прямая колонка для FK constraint)
    referenced_object_id UUID         REFERENCES metadata.object_definitions(id),

    -- Структурные constraints
    is_required          BOOLEAN      NOT NULL DEFAULT false,
    is_unique            BOOLEAN      NOT NULL DEFAULT false,

    -- Type-specific параметры
    config               JSONB        NOT NULL DEFAULT '{}',

    -- Классификация
    is_system_field      BOOLEAN      NOT NULL DEFAULT false,
    is_custom            BOOLEAN      NOT NULL DEFAULT false,
    is_platform_managed  BOOLEAN      NOT NULL DEFAULT false,
    sort_order           INTEGER      NOT NULL DEFAULT 0,

    -- Timestamps
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),

    UNIQUE (object_id, api_name)
);

CREATE INDEX IF NOT EXISTS idx_field_definitions_object_id
    ON metadata.field_definitions (object_id);

CREATE INDEX IF NOT EXISTS idx_field_definitions_referenced_object_id
    ON metadata.field_definitions (referenced_object_id)
    WHERE referenced_object_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_field_definitions_field_type
    ON metadata.field_definitions (field_type);
