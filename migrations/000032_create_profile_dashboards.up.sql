-- Profile Dashboards: home page widgets per profile (ADR-0032)
CREATE TABLE metadata.profile_dashboards (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id      UUID NOT NULL REFERENCES iam.profiles(id) ON DELETE CASCADE,
    config          JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (profile_id)
);

CREATE INDEX idx_profile_dashboards_profile_id ON metadata.profile_dashboards(profile_id);
