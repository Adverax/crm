-- Copyright 2026 Adverax. All rights reserved.
-- Licensed under the Adverax Commercial License.
-- See ee/LICENSE for details.
-- Unauthorized use, copying, or distribution is prohibited.

BEGIN;
SELECT plan(7);

-- Verify function exists
SELECT has_function('ee', 'generate_record_share_entries', ARRAY['uuid', 'uuid', 'uuid', 'text'],
    'function ee.generate_record_share_entries exists');
SELECT function_returns('ee', 'generate_record_share_entries', ARRAY['uuid', 'uuid', 'uuid', 'text'], 'void');

-- Setup: model + territories (EMEA -> France -> Paris)
INSERT INTO ee.territory_models (id, api_name, label, status)
VALUES ('a0000000-0000-4000-a000-000000000001', 'test_model', 'Test Model', 'active');

INSERT INTO ee.territories (id, model_id, parent_id, api_name, label) VALUES
    ('b0000000-0000-4000-a000-000000000001', 'a0000000-0000-4000-a000-000000000001', NULL, 'emea', 'EMEA'),
    ('b0000000-0000-4000-a000-000000000002', 'a0000000-0000-4000-a000-000000000001', 'b0000000-0000-4000-a000-000000000001', 'france', 'France'),
    ('b0000000-0000-4000-a000-000000000003', 'a0000000-0000-4000-a000-000000000001', 'b0000000-0000-4000-a000-000000000002', 'paris', 'Paris');

-- Build closure table
SELECT ee.rebuild_territory_hierarchy('a0000000-0000-4000-a000-000000000001');

-- Create territory groups
INSERT INTO iam.groups (id, api_name, label, group_type, related_territory_id) VALUES
    ('c0000000-0000-4000-a000-000000000001', 'territory_emea', 'EMEA', 'territory', 'b0000000-0000-4000-a000-000000000001'),
    ('c0000000-0000-4000-a000-000000000002', 'territory_france', 'France', 'territory', 'b0000000-0000-4000-a000-000000000002'),
    ('c0000000-0000-4000-a000-000000000003', 'territory_paris', 'Paris', 'territory', 'b0000000-0000-4000-a000-000000000003');

-- Setup object: use the existing seed object (Account)
-- We need an object_id from metadata.object_definitions
INSERT INTO metadata.object_definitions (id, api_name, label, plural_label, table_name, object_type, visibility)
VALUES ('d0000000-0000-4000-a000-000000000001', 'test_account', 'Test Account', 'Test Accounts', 'obj_test_account', 'standard', 'private');

-- Object defaults: EMEA=read, France=read_write, Paris=none
INSERT INTO ee.territory_object_defaults (territory_id, object_id, access_level) VALUES
    ('b0000000-0000-4000-a000-000000000001', 'd0000000-0000-4000-a000-000000000001', 'read'),
    ('b0000000-0000-4000-a000-000000000002', 'd0000000-0000-4000-a000-000000000001', 'read_write');

-- Create a fake share table for testing
CREATE TABLE obj_test_account (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

CREATE TABLE obj_test_account__share (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    record_id    UUID        NOT NULL REFERENCES obj_test_account(id) ON DELETE CASCADE,
    group_id     UUID        NOT NULL,
    access_level VARCHAR(20) NOT NULL CHECK (access_level IN ('read', 'read_write')),
    reason       VARCHAR(30) NOT NULL CHECK (reason IN ('manual', 'sharing_rule', 'territory')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (record_id, group_id, reason)
);

-- Create a test record
INSERT INTO obj_test_account (id) VALUES ('e0000000-0000-4000-a000-000000000001');

-- Execute: assign record to Paris territory
SELECT lives_ok(
    $$SELECT ee.generate_record_share_entries(
        'e0000000-0000-4000-a000-000000000001',
        'd0000000-0000-4000-a000-000000000001',
        'b0000000-0000-4000-a000-000000000003',
        'obj_test_account__share'
    )$$,
    'generate_record_share_entries executes without error'
);

-- Verify: Paris has no object_default -> no share entry for Paris group
SELECT is(
    (SELECT count(*)::int FROM obj_test_account__share
     WHERE group_id = 'c0000000-0000-4000-a000-000000000003'),
    0, 'no share entry for Paris (no object_default)'
);

-- Verify: France has object_default read_write -> share entry exists
SELECT is(
    (SELECT access_level FROM obj_test_account__share
     WHERE group_id = 'c0000000-0000-4000-a000-000000000002'),
    'read_write', 'France share entry has read_write access'
);

-- Verify: EMEA has object_default read -> share entry exists
SELECT is(
    (SELECT access_level FROM obj_test_account__share
     WHERE group_id = 'c0000000-0000-4000-a000-000000000001'),
    'read', 'EMEA share entry has read access'
);

-- Verify total: exactly 2 share entries (France + EMEA)
SELECT is(
    (SELECT count(*)::int FROM obj_test_account__share WHERE reason = 'territory'),
    2, 'exactly 2 territory share entries created'
);

SELECT finish();
ROLLBACK;
