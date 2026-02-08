-- Well-known UUIDs for seed data
-- system_administrator_base PS:  00000000-0000-4000-a000-000000000001
-- system_administrator profile:  00000000-0000-4000-a000-000000000002
-- admin user:                    00000000-0000-4000-a000-000000000003

INSERT INTO iam.permission_sets (id, api_name, label, description, ps_type)
VALUES (
    '00000000-0000-4000-a000-000000000001',
    'system_administrator_base',
    'System Administrator Base',
    'Base permission set for system administrator profile',
    'grant'
);

INSERT INTO iam.profiles (id, api_name, label, description, base_permission_set_id)
VALUES (
    '00000000-0000-4000-a000-000000000002',
    'system_administrator',
    'System Administrator',
    'Full access to all objects and fields',
    '00000000-0000-4000-a000-000000000001'
);

INSERT INTO iam.users (id, username, email, first_name, last_name, profile_id, is_active)
VALUES (
    '00000000-0000-4000-a000-000000000003',
    'admin',
    'admin@crm.local',
    'System',
    'Administrator',
    '00000000-0000-4000-a000-000000000002',
    true
);
