CREATE TABLE metadata.object_views (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    object_id   UUID         NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    profile_id  UUID         REFERENCES iam.profiles(id) ON DELETE CASCADE,
    api_name    VARCHAR(100) NOT NULL UNIQUE,
    label       VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    is_default  BOOLEAN      NOT NULL DEFAULT false,
    config      JSONB        NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT uq_object_views_object_profile UNIQUE (object_id, profile_id),
    CONSTRAINT chk_ov_api_name_format CHECK (api_name ~ '^[a-z][a-z0-9_]*$')
);

CREATE INDEX idx_object_views_object_id ON metadata.object_views(object_id);
CREATE INDEX idx_object_views_profile_id ON metadata.object_views(profile_id);
