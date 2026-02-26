CREATE TABLE metadata.object_views (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id  UUID         REFERENCES iam.profiles(id) ON DELETE CASCADE,
    api_name    VARCHAR(100) NOT NULL UNIQUE,
    label       VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    config      JSONB        NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT chk_ov_api_name_format CHECK (api_name ~ '^[a-z][a-z0-9_]*$')
);

CREATE INDEX idx_object_views_profile_id ON metadata.object_views(profile_id);
