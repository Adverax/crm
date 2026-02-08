CREATE TABLE IF NOT EXISTS iam.profiles (
    id                     UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    api_name               VARCHAR(100)  NOT NULL UNIQUE,
    label                  VARCHAR(255)  NOT NULL,
    description            TEXT          NOT NULL DEFAULT '',
    base_permission_set_id UUID          NOT NULL
                           REFERENCES iam.permission_sets(id) ON DELETE RESTRICT,
    created_at             TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_profiles_base_permission_set_id
    ON iam.profiles (base_permission_set_id);
