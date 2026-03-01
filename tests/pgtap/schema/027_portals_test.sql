BEGIN;
SELECT plan(12);

-- Table exists
SELECT has_table('metadata', 'portals', 'has metadata.portals table');

-- Columns
SELECT has_column('metadata', 'portals', 'id', 'has id column');
SELECT has_column('metadata', 'portals', 'profile_id', 'has profile_id column');
SELECT has_column('metadata', 'portals', 'api_name', 'has api_name column');
SELECT has_column('metadata', 'portals', 'label', 'has label column');
SELECT has_column('metadata', 'portals', 'description', 'has description column');
SELECT has_column('metadata', 'portals', 'config', 'has config column');
SELECT has_column('metadata', 'portals', 'created_at', 'has created_at column');
SELECT has_column('metadata', 'portals', 'updated_at', 'has updated_at column');

-- Types
SELECT col_type_is('metadata', 'portals', 'id', 'uuid', 'id is uuid');
SELECT col_type_is('metadata', 'portals', 'config', 'jsonb', 'config is jsonb');

-- Constraints
SELECT has_check('metadata', 'portals', 'has check constraints');

SELECT * FROM finish();
ROLLBACK;
