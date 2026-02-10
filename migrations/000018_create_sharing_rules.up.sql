CREATE TABLE security.sharing_rules (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    object_id       UUID        NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    rule_type       VARCHAR(30) NOT NULL CHECK (rule_type IN ('owner_based', 'criteria_based')),
    source_group_id UUID        NOT NULL REFERENCES iam.groups(id) ON DELETE CASCADE,
    target_group_id UUID        NOT NULL REFERENCES iam.groups(id) ON DELETE CASCADE,
    access_level    VARCHAR(20) NOT NULL CHECK (access_level IN ('read', 'read_write')),
    criteria_field  VARCHAR(255),
    criteria_op     VARCHAR(20),
    criteria_value  TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sharing_rules_object_id ON security.sharing_rules (object_id);
CREATE INDEX idx_sharing_rules_source_group ON security.sharing_rules (source_group_id);
CREATE INDEX idx_sharing_rules_target_group ON security.sharing_rules (target_group_id);
