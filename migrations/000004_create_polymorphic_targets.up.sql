CREATE TABLE IF NOT EXISTS metadata.polymorphic_targets (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    field_id   UUID        NOT NULL REFERENCES metadata.field_definitions(id) ON DELETE CASCADE,
    object_id  UUID        NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (field_id, object_id)
);

CREATE INDEX IF NOT EXISTS idx_polymorphic_targets_field_id
    ON metadata.polymorphic_targets (field_id);

CREATE INDEX IF NOT EXISTS idx_polymorphic_targets_object_id
    ON metadata.polymorphic_targets (object_id);
