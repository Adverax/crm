BEGIN;
SELECT plan(10);

-- Table exists
SELECT has_table('metadata', 'profile_navigation', 'has metadata.profile_navigation table');

-- Columns
SELECT has_column('metadata', 'profile_navigation', 'id', 'has id column');
SELECT has_column('metadata', 'profile_navigation', 'profile_id', 'has profile_id column');
SELECT has_column('metadata', 'profile_navigation', 'config', 'has config column');
SELECT has_column('metadata', 'profile_navigation', 'created_at', 'has created_at column');
SELECT has_column('metadata', 'profile_navigation', 'updated_at', 'has updated_at column');

-- Types
SELECT col_type_is('metadata', 'profile_navigation', 'id', 'uuid', 'id is uuid');
SELECT col_type_is('metadata', 'profile_navigation', 'config', 'jsonb', 'config is jsonb');
SELECT col_type_is('metadata', 'profile_navigation', 'profile_id', 'uuid', 'profile_id is uuid');

-- Indexes
SELECT has_index('metadata', 'profile_navigation', 'idx_profile_navigation_profile_id', 'has profile_id index');

SELECT * FROM finish();
ROLLBACK;
