-- Add 'territory' to group_type CHECK constraint (ADR-0015)
-- Inline CHECK constraints get auto-named as {table}_{column}_check
ALTER TABLE iam.groups DROP CONSTRAINT IF EXISTS groups_group_type_check;
ALTER TABLE iam.groups DROP CONSTRAINT IF EXISTS iam_groups_group_type_check;
ALTER TABLE iam.groups ADD CONSTRAINT groups_group_type_check
    CHECK (group_type IN ('personal', 'role', 'role_and_subordinates', 'public', 'territory'));
