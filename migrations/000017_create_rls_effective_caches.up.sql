-- Role hierarchy closure table (ADR-0012)
CREATE TABLE security.effective_role_hierarchy (
    ancestor_role_id   UUID NOT NULL REFERENCES iam.user_roles(id) ON DELETE CASCADE,
    descendant_role_id UUID NOT NULL REFERENCES iam.user_roles(id) ON DELETE CASCADE,
    depth              INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (ancestor_role_id, descendant_role_id)
);

CREATE INDEX idx_effective_role_hierarchy_descendant ON security.effective_role_hierarchy (descendant_role_id);

-- Visible owner cache: which record owners a user can see via role hierarchy
CREATE TABLE security.effective_visible_owner (
    user_id         UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    visible_owner_id UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, visible_owner_id)
);

CREATE INDEX idx_effective_visible_owner_visible ON security.effective_visible_owner (visible_owner_id);

-- Flattened group membership (resolves nested groups)
CREATE TABLE security.effective_group_members (
    group_id UUID NOT NULL REFERENCES iam.groups(id) ON DELETE CASCADE,
    user_id  UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX idx_effective_group_members_user ON security.effective_group_members (user_id);

-- Object hierarchy closure for controlled_by_parent
CREATE TABLE security.effective_object_hierarchy (
    ancestor_object_id   UUID NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    descendant_object_id UUID NOT NULL REFERENCES metadata.object_definitions(id) ON DELETE CASCADE,
    depth                INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (ancestor_object_id, descendant_object_id)
);

CREATE INDEX idx_effective_object_hierarchy_descendant ON security.effective_object_hierarchy (descendant_object_id);
