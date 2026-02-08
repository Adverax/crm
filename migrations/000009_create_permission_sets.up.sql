CREATE TABLE IF NOT EXISTS iam.permission_sets (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    api_name    VARCHAR(100)  NOT NULL UNIQUE,
    label       VARCHAR(255)  NOT NULL,
    description TEXT          NOT NULL DEFAULT '',
    ps_type     VARCHAR(10)   NOT NULL DEFAULT 'grant'
                              CHECK (ps_type IN ('grant', 'deny')),
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT now()
);
