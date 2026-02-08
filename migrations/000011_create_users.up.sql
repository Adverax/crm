CREATE TABLE IF NOT EXISTS iam.users (
    id         UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    username   VARCHAR(100)  NOT NULL UNIQUE,
    email      VARCHAR(255)  NOT NULL UNIQUE,
    first_name VARCHAR(100)  NOT NULL DEFAULT '',
    last_name  VARCHAR(100)  NOT NULL DEFAULT '',
    profile_id UUID          NOT NULL
                             REFERENCES iam.profiles(id) ON DELETE RESTRICT,
    role_id    UUID          REFERENCES iam.user_roles(id) ON DELETE SET NULL,
    is_active  BOOLEAN       NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_users_profile_id
    ON iam.users (profile_id);

CREATE INDEX IF NOT EXISTS idx_users_role_id
    ON iam.users (role_id)
    WHERE role_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_users_is_active
    ON iam.users (is_active)
    WHERE is_active = true;
