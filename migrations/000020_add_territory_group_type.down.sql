-- Revert group_type CHECK to original values (without 'territory')
ALTER TABLE iam.groups DROP CONSTRAINT IF EXISTS groups_group_type_check;
ALTER TABLE iam.groups ADD CONSTRAINT groups_group_type_check
    CHECK (group_type IN ('personal', 'role', 'role_and_subordinates', 'public'));
