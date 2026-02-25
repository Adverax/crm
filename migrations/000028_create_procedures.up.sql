CREATE TABLE metadata.procedures (
    id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    code                 VARCHAR(100) UNIQUE NOT NULL,
    name                 VARCHAR(255) NOT NULL,
    description          TEXT         NOT NULL DEFAULT '',
    draft_version_id     UUID,
    published_version_id UUID,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_procedure_code_format CHECK (code ~ '^[a-z][a-z0-9_]*$')
);

CREATE TABLE metadata.procedure_versions (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    procedure_id   UUID         NOT NULL REFERENCES metadata.procedures(id) ON DELETE CASCADE,
    version        INT          NOT NULL,
    definition     JSONB        NOT NULL,
    status         VARCHAR(20)  NOT NULL DEFAULT 'draft',
    change_summary TEXT         NOT NULL DEFAULT '',
    created_by     UUID,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    published_at   TIMESTAMPTZ,
    CONSTRAINT procedure_versions_unique UNIQUE (procedure_id, version),
    CONSTRAINT procedure_versions_status_check CHECK (status IN ('draft', 'published', 'superseded'))
);

ALTER TABLE metadata.procedures
    ADD CONSTRAINT fk_procedures_draft_version
        FOREIGN KEY (draft_version_id) REFERENCES metadata.procedure_versions(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_procedures_published_version
        FOREIGN KEY (published_version_id) REFERENCES metadata.procedure_versions(id) ON DELETE SET NULL;

CREATE INDEX idx_procedure_versions_procedure_id ON metadata.procedure_versions(procedure_id);
CREATE INDEX idx_procedure_versions_status ON metadata.procedure_versions(status);
