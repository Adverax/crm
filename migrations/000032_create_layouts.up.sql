CREATE TABLE metadata.layouts (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    portal_id       UUID         NOT NULL REFERENCES metadata.portals(id) ON DELETE CASCADE,
    form_factor     VARCHAR(20)  NOT NULL DEFAULT 'desktop',
    mode            VARCHAR(20)  NOT NULL DEFAULT 'read',
    config          JSONB        NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT layouts_form_factor_check CHECK (form_factor IN ('desktop', 'tablet', 'mobile')),
    CONSTRAINT layouts_mode_check CHECK (mode IN ('read', 'view')),
    CONSTRAINT layouts_portal_ff_mode_unique UNIQUE (portal_id, form_factor, mode)
);

CREATE INDEX idx_layouts_portal_id ON metadata.layouts(portal_id);
