CREATE TABLE metadata.automation_rules (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    object_id       UUID         NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    name            TEXT         NOT NULL,
    description     TEXT         NOT NULL DEFAULT '',
    event_type      TEXT         NOT NULL,
    condition       TEXT,
    procedure_code  TEXT         NOT NULL,
    execution_mode  TEXT         NOT NULL DEFAULT 'per_record',
    sort_order      INT          NOT NULL DEFAULT 0,
    is_active       BOOLEAN      NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (object_id, name),
    CONSTRAINT chk_automation_rule_event_type CHECK (
        event_type IN (
            'before_insert', 'after_insert',
            'before_update', 'after_update',
            'before_delete', 'after_delete'
        )
    ),
    CONSTRAINT chk_automation_rule_execution_mode CHECK (
        execution_mode IN ('per_record', 'per_batch')
    )
);

CREATE INDEX idx_automation_rules_object_id ON metadata.automation_rules(object_id);
CREATE INDEX idx_automation_rules_active ON metadata.automation_rules(object_id, event_type) WHERE is_active = true;
