-- Groups table (ADR-0013)
CREATE TABLE iam.groups (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    api_name        VARCHAR(100) NOT NULL UNIQUE,
    label           VARCHAR(255) NOT NULL,
    group_type      VARCHAR(30)  NOT NULL CHECK (group_type IN ('personal', 'role', 'role_and_subordinates', 'public')),
    related_role_id UUID        REFERENCES iam.user_roles(id) ON DELETE CASCADE,
    related_user_id UUID        REFERENCES iam.users(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_iam_groups_group_type ON iam.groups (group_type);
CREATE INDEX idx_iam_groups_related_role_id ON iam.groups (related_role_id) WHERE related_role_id IS NOT NULL;
CREATE INDEX idx_iam_groups_related_user_id ON iam.groups (related_user_id) WHERE related_user_id IS NOT NULL;

-- Group members table
CREATE TABLE iam.group_members (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id        UUID        NOT NULL REFERENCES iam.groups(id) ON DELETE CASCADE,
    member_user_id  UUID        REFERENCES iam.users(id) ON DELETE CASCADE,
    member_group_id UUID        REFERENCES iam.groups(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_group_members_xor CHECK (
        (member_user_id IS NOT NULL AND member_group_id IS NULL) OR
        (member_user_id IS NULL AND member_group_id IS NOT NULL)
    ),
    CONSTRAINT uq_group_members_user UNIQUE (group_id, member_user_id),
    CONSTRAINT uq_group_members_group UNIQUE (group_id, member_group_id)
);

CREATE INDEX idx_iam_group_members_group_id ON iam.group_members (group_id);
CREATE INDEX idx_iam_group_members_member_user_id ON iam.group_members (member_user_id) WHERE member_user_id IS NOT NULL;
CREATE INDEX idx_iam_group_members_member_group_id ON iam.group_members (member_group_id) WHERE member_group_id IS NOT NULL;
