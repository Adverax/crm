BEGIN;
SELECT plan(3);

-- Verify group_type CHECK constraint allows 'territory' (ADR-0015)
SELECT has_check('iam', 'groups', 'iam.groups has CHECK constraint');

-- Insert a territory group and verify it succeeds
SELECT lives_ok(
    $$INSERT INTO iam.groups (api_name, label, group_type)
      VALUES ('test_territory_group', 'Test Territory', 'territory')$$,
    'can insert group with group_type = territory'
);

-- Verify invalid group_type is still rejected
SELECT throws_ok(
    $$INSERT INTO iam.groups (api_name, label, group_type)
      VALUES ('test_invalid_group', 'Invalid', 'invalid_type')$$,
    '23514',
    NULL,
    'invalid group_type is rejected by CHECK constraint'
);

SELECT finish();
ROLLBACK;
