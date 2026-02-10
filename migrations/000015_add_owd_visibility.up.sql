ALTER TABLE metadata.object_definitions
    ADD COLUMN visibility VARCHAR(30) NOT NULL DEFAULT 'private'
    CHECK (visibility IN ('private', 'public_read', 'public_read_write', 'controlled_by_parent'));
