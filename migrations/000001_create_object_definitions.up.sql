CREATE SCHEMA IF NOT EXISTS metadata;

CREATE TABLE IF NOT EXISTS metadata.object_definitions (
    -- Идентификация
    id                       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    api_name                 VARCHAR(100)  NOT NULL UNIQUE
                             CHECK (api_name ~ '^[a-z][a-z0-9_]*$'),
    label                    VARCHAR(255)  NOT NULL,
    plural_label             VARCHAR(255)  NOT NULL,
    description              TEXT          NOT NULL DEFAULT '',

    -- Физическое хранение (ADR-0007)
    table_name               VARCHAR(63)   NOT NULL,

    -- Классификация
    object_type              VARCHAR(20)   NOT NULL CHECK (object_type IN ('standard', 'custom')),

    -- Поведенческие флаги (уровень схемы)
    is_platform_managed      BOOLEAN       NOT NULL DEFAULT false,
    is_visible_in_setup      BOOLEAN       NOT NULL DEFAULT true,
    is_custom_fields_allowed BOOLEAN       NOT NULL DEFAULT true,
    is_deleteable_object     BOOLEAN       NOT NULL DEFAULT true,

    -- Возможности записей
    is_createable            BOOLEAN       NOT NULL DEFAULT true,
    is_updateable            BOOLEAN       NOT NULL DEFAULT true,
    is_deleteable            BOOLEAN       NOT NULL DEFAULT true,
    is_queryable             BOOLEAN       NOT NULL DEFAULT true,
    is_searchable            BOOLEAN       NOT NULL DEFAULT true,

    -- Фичи (подключаемые подсистемы)
    has_activities            BOOLEAN       NOT NULL DEFAULT false,
    has_notes                 BOOLEAN       NOT NULL DEFAULT false,
    has_history_tracking      BOOLEAN       NOT NULL DEFAULT false,
    has_sharing_rules         BOOLEAN       NOT NULL DEFAULT false,

    -- Системные timestamps
    created_at               TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_object_definitions_table_name
    ON metadata.object_definitions (table_name);
