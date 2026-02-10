-- Well-known UUID for admin personal group
-- admin_personal_group: 00000000-0000-4000-a000-000000000010

INSERT INTO iam.groups (id, api_name, label, group_type, related_user_id)
VALUES (
    '00000000-0000-4000-a000-000000000010',
    'personal_admin',
    'admin (Personal)',
    'personal',
    '00000000-0000-4000-a000-000000000003'
);

INSERT INTO iam.group_members (group_id, member_user_id)
VALUES (
    '00000000-0000-4000-a000-000000000010',
    '00000000-0000-4000-a000-000000000003'
);
