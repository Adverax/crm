BEGIN;
SELECT plan(10);

-- Table exists
SELECT has_table('metadata', 'profile_dashboards', 'has metadata.profile_dashboards table');

-- Columns
SELECT has_column('metadata', 'profile_dashboards', 'id', 'has id column');
SELECT has_column('metadata', 'profile_dashboards', 'profile_id', 'has profile_id column');
SELECT has_column('metadata', 'profile_dashboards', 'config', 'has config column');
SELECT has_column('metadata', 'profile_dashboards', 'created_at', 'has created_at column');
SELECT has_column('metadata', 'profile_dashboards', 'updated_at', 'has updated_at column');

-- Types
SELECT col_type_is('metadata', 'profile_dashboards', 'id', 'uuid', 'id is uuid');
SELECT col_type_is('metadata', 'profile_dashboards', 'config', 'jsonb', 'config is jsonb');
SELECT col_type_is('metadata', 'profile_dashboards', 'profile_id', 'uuid', 'profile_id is uuid');

-- Indexes
SELECT has_index('metadata', 'profile_dashboards', 'idx_profile_dashboards_profile_id', 'has profile_id index');

SELECT * FROM finish();
ROLLBACK;
