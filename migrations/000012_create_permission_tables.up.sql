CREATE SCHEMA IF NOT EXISTS security;

CREATE TABLE IF NOT EXISTS security.object_permissions (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    permission_set_id UUID        NOT NULL
                                  REFERENCES iam.permission_sets(id) ON DELETE CASCADE,
    object_id         UUID        NOT NULL
                                  REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    permissions       INT         NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (permission_set_id, object_id)
);

CREATE INDEX IF NOT EXISTS idx_object_permissions_permission_set_id
    ON security.object_permissions (permission_set_id);

CREATE INDEX IF NOT EXISTS idx_object_permissions_object_id
    ON security.object_permissions (object_id);

CREATE TABLE IF NOT EXISTS security.field_permissions (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    permission_set_id UUID        NOT NULL
                                  REFERENCES iam.permission_sets(id) ON DELETE CASCADE,
    field_id          UUID        NOT NULL
                                  REFERENCES metadata.field_definitions(id) ON DELETE CASCADE,
    permissions       INT         NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (permission_set_id, field_id)
);

CREATE INDEX IF NOT EXISTS idx_field_permissions_permission_set_id
    ON security.field_permissions (permission_set_id);

CREATE INDEX IF NOT EXISTS idx_field_permissions_field_id
    ON security.field_permissions (field_id);

CREATE TABLE IF NOT EXISTS iam.permission_set_to_users (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    permission_set_id UUID        NOT NULL
                                  REFERENCES iam.permission_sets(id) ON DELETE CASCADE,
    user_id           UUID        NOT NULL
                                  REFERENCES iam.users(id) ON DELETE CASCADE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (permission_set_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_permission_set_to_users_permission_set_id
    ON iam.permission_set_to_users (permission_set_id);

CREATE INDEX IF NOT EXISTS idx_permission_set_to_users_user_id
    ON iam.permission_set_to_users (user_id);
