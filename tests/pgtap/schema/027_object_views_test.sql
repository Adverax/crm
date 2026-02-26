BEGIN;
SELECT plan(12);

-- Table exists
SELECT has_table('metadata', 'object_views', 'has metadata.object_views table');

-- Columns
SELECT has_column('metadata', 'object_views', 'id', 'has id column');
SELECT has_column('metadata', 'object_views', 'profile_id', 'has profile_id column');
SELECT has_column('metadata', 'object_views', 'api_name', 'has api_name column');
SELECT has_column('metadata', 'object_views', 'label', 'has label column');
SELECT has_column('metadata', 'object_views', 'description', 'has description column');
SELECT has_column('metadata', 'object_views', 'config', 'has config column');
SELECT has_column('metadata', 'object_views', 'created_at', 'has created_at column');
SELECT has_column('metadata', 'object_views', 'updated_at', 'has updated_at column');

-- Types
SELECT col_type_is('metadata', 'object_views', 'id', 'uuid', 'id is uuid');
SELECT col_type_is('metadata', 'object_views', 'config', 'jsonb', 'config is jsonb');

-- Constraints
SELECT has_check('metadata', 'object_views', 'has check constraints');

SELECT * FROM finish();
ROLLBACK;
