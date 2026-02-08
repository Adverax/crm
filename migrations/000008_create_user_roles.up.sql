CREATE SCHEMA IF NOT EXISTS iam;

CREATE TABLE IF NOT EXISTS iam.user_roles (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    api_name    VARCHAR(100)  NOT NULL UNIQUE,
    label       VARCHAR(255)  NOT NULL,
    description TEXT          NOT NULL DEFAULT '',
    parent_id   UUID          REFERENCES iam.user_roles(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_roles_parent_id
    ON iam.user_roles (parent_id)
    WHERE parent_id IS NOT NULL;
