-- Copyright 2026 Adverax. All rights reserved.
-- Licensed under the Adverax Commercial License.
-- See ee/LICENSE for details.
-- Unauthorized use, copying, or distribution is prohibited.

BEGIN;
SELECT plan(22);

-- ee.record_territory_assignments table
SELECT has_table('ee', 'record_territory_assignments', 'table ee.record_territory_assignments exists');

-- id
SELECT has_column('ee', 'record_territory_assignments', 'id', 'has id');
SELECT col_type_is('ee', 'record_territory_assignments', 'id', 'uuid', 'id is uuid');
SELECT col_has_default('ee', 'record_territory_assignments', 'id', 'id has default');
SELECT col_is_pk('ee', 'record_territory_assignments', 'id', 'id is PK');

-- record_id
SELECT has_column('ee', 'record_territory_assignments', 'record_id', 'has record_id');
SELECT col_type_is('ee', 'record_territory_assignments', 'record_id', 'uuid', 'record_id is uuid');
SELECT col_not_null('ee', 'record_territory_assignments', 'record_id', 'record_id is NOT NULL');

-- object_id
SELECT has_column('ee', 'record_territory_assignments', 'object_id', 'has object_id');
SELECT col_not_null('ee', 'record_territory_assignments', 'object_id', 'object_id is NOT NULL');
SELECT fk_ok('ee', 'record_territory_assignments', 'object_id', 'metadata', 'object_definitions', 'id',
    'object_id FK to metadata.object_definitions');

-- territory_id
SELECT has_column('ee', 'record_territory_assignments', 'territory_id', 'has territory_id');
SELECT col_not_null('ee', 'record_territory_assignments', 'territory_id', 'territory_id is NOT NULL');
SELECT fk_ok('ee', 'record_territory_assignments', 'territory_id', 'ee', 'territories', 'id',
    'territory_id FK to ee.territories');

-- reason
SELECT has_column('ee', 'record_territory_assignments', 'reason', 'has reason');
SELECT col_not_null('ee', 'record_territory_assignments', 'reason', 'reason is NOT NULL');
SELECT col_has_default('ee', 'record_territory_assignments', 'reason', 'reason has default');
SELECT has_check('ee', 'record_territory_assignments', 'reason has CHECK');

-- created_at
SELECT has_column('ee', 'record_territory_assignments', 'created_at', 'has created_at');
SELECT col_not_null('ee', 'record_territory_assignments', 'created_at', 'created_at is NOT NULL');

-- Indexes
SELECT has_index('ee', 'record_territory_assignments', 'idx_record_territory_record', 'index on record_id, object_id');
SELECT has_index('ee', 'record_territory_assignments', 'idx_record_territory_territory', 'index on territory_id');

SELECT finish();
ROLLBACK;
