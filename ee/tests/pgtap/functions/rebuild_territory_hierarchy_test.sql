-- Copyright 2026 Adverax. All rights reserved.
-- Licensed under the Adverax Commercial License.
-- See ee/LICENSE for details.
-- Unauthorized use, copying, or distribution is prohibited.

BEGIN;
SELECT plan(8);

-- Verify function exists
SELECT has_function('ee', 'rebuild_territory_hierarchy', ARRAY['uuid'], 'function ee.rebuild_territory_hierarchy exists');
SELECT function_returns('ee', 'rebuild_territory_hierarchy', ARRAY['uuid'], 'void');

-- Setup: create model and territories (3-level tree)
-- EMEA -> France -> Paris
INSERT INTO ee.territory_models (id, api_name, label, status)
VALUES ('a0000000-0000-4000-a000-000000000001', 'test_model', 'Test Model', 'planning');

INSERT INTO ee.territories (id, model_id, parent_id, api_name, label)
VALUES
    ('b0000000-0000-4000-a000-000000000001', 'a0000000-0000-4000-a000-000000000001', NULL, 'emea', 'EMEA'),
    ('b0000000-0000-4000-a000-000000000002', 'a0000000-0000-4000-a000-000000000001', 'b0000000-0000-4000-a000-000000000001', 'france', 'France'),
    ('b0000000-0000-4000-a000-000000000003', 'a0000000-0000-4000-a000-000000000001', 'b0000000-0000-4000-a000-000000000002', 'paris', 'Paris');

-- Execute function
SELECT lives_ok(
    $$SELECT ee.rebuild_territory_hierarchy('a0000000-0000-4000-a000-000000000001')$$,
    'rebuild_territory_hierarchy executes without error'
);

-- Verify closure table entries
-- Self entries (depth=0): 3
SELECT is(
    (SELECT count(*)::int FROM security.effective_territory_hierarchy WHERE depth = 0),
    3, 'three self entries at depth 0'
);

-- EMEA is ancestor of France (depth=1)
SELECT is(
    (SELECT count(*)::int FROM security.effective_territory_hierarchy
     WHERE ancestor_territory_id = 'b0000000-0000-4000-a000-000000000001'
       AND descendant_territory_id = 'b0000000-0000-4000-a000-000000000002'
       AND depth = 1),
    1, 'EMEA is ancestor of France at depth 1'
);

-- EMEA is ancestor of Paris (depth=2)
SELECT is(
    (SELECT count(*)::int FROM security.effective_territory_hierarchy
     WHERE ancestor_territory_id = 'b0000000-0000-4000-a000-000000000001'
       AND descendant_territory_id = 'b0000000-0000-4000-a000-000000000003'
       AND depth = 2),
    1, 'EMEA is ancestor of Paris at depth 2'
);

-- France is ancestor of Paris (depth=1)
SELECT is(
    (SELECT count(*)::int FROM security.effective_territory_hierarchy
     WHERE ancestor_territory_id = 'b0000000-0000-4000-a000-000000000002'
       AND descendant_territory_id = 'b0000000-0000-4000-a000-000000000003'
       AND depth = 1),
    1, 'France is ancestor of Paris at depth 1'
);

-- Total entries: 3 self + 2 (EMEA->France, France->Paris) + 1 (EMEA->Paris) = 6
SELECT is(
    (SELECT count(*)::int FROM security.effective_territory_hierarchy),
    6, 'total 6 closure entries for 3-level tree'
);

SELECT finish();
ROLLBACK;
