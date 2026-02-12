-- Copyright 2026 Adverax. All rights reserved.
-- Licensed under the Adverax Commercial License.
-- See ee/LICENSE for details.
-- Unauthorized use, copying, or distribution is prohibited.

BEGIN;
SELECT plan(9);

-- Verify function exists
SELECT has_function('ee', 'activate_territory_model', ARRAY['uuid'],
    'function ee.activate_territory_model exists');
SELECT function_returns('ee', 'activate_territory_model', ARRAY['uuid'], 'void');

-- Setup: create a model in planning status with 2 territories
INSERT INTO ee.territory_models (id, api_name, label, status)
VALUES ('a0000000-0000-4000-a000-000000000001', 'q1_2026', 'Q1 2026', 'planning');

INSERT INTO ee.territories (id, model_id, parent_id, api_name, label) VALUES
    ('b0000000-0000-4000-a000-000000000001', 'a0000000-0000-4000-a000-000000000001', NULL, 'north', 'North'),
    ('b0000000-0000-4000-a000-000000000002', 'a0000000-0000-4000-a000-000000000001', 'b0000000-0000-4000-a000-000000000001', 'northeast', 'Northeast');

-- Assign admin user to North territory
INSERT INTO ee.user_territory_assignments (user_id, territory_id)
VALUES ('00000000-0000-4000-a000-000000000003', 'b0000000-0000-4000-a000-000000000001');

-- Activate model
SELECT lives_ok(
    $$SELECT ee.activate_territory_model('a0000000-0000-4000-a000-000000000001')$$,
    'activate_territory_model executes without error'
);

-- Verify model status changed to active
SELECT is(
    (SELECT status FROM ee.territory_models WHERE id = 'a0000000-0000-4000-a000-000000000001'),
    'active', 'model status is active'
);

-- Verify activated_at is set
SELECT isnt(
    (SELECT activated_at FROM ee.territory_models WHERE id = 'a0000000-0000-4000-a000-000000000001'),
    NULL, 'activated_at is set'
);

-- Verify territory groups were created (one per territory)
SELECT is(
    (SELECT count(*)::int FROM iam.groups WHERE group_type = 'territory'),
    2, 'two territory groups created'
);

-- Verify group members (admin user in North territory group)
SELECT is(
    (SELECT count(*)::int FROM iam.group_members gm
     JOIN iam.groups g ON g.id = gm.group_id
     WHERE g.group_type = 'territory'
       AND gm.member_user_id = '00000000-0000-4000-a000-000000000003'),
    1, 'admin user is member of North territory group'
);

-- Verify effective_territory_hierarchy was built (2 self + 1 parent-child = 3)
SELECT is(
    (SELECT count(*)::int FROM security.effective_territory_hierarchy),
    3, 'effective_territory_hierarchy has 3 entries'
);

-- Verify effective_user_territory was populated
SELECT is(
    (SELECT count(*)::int FROM security.effective_user_territory
     WHERE user_id = '00000000-0000-4000-a000-000000000003'),
    1, 'admin user has 1 territory in effective_user_territory'
);

SELECT finish();
ROLLBACK;
