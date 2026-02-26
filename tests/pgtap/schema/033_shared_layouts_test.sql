BEGIN;
SELECT plan(14);

-- Table exists
SELECT has_table('metadata', 'shared_layouts', 'has metadata.shared_layouts table');

-- Columns
SELECT has_column('metadata', 'shared_layouts', 'id', 'has id column');
SELECT has_column('metadata', 'shared_layouts', 'api_name', 'has api_name column');
SELECT has_column('metadata', 'shared_layouts', 'type', 'has type column');
SELECT has_column('metadata', 'shared_layouts', 'label', 'has label column');
SELECT has_column('metadata', 'shared_layouts', 'config', 'has config column');
SELECT has_column('metadata', 'shared_layouts', 'created_at', 'has created_at column');
SELECT has_column('metadata', 'shared_layouts', 'updated_at', 'has updated_at column');

-- Column types
SELECT col_type_is('metadata', 'shared_layouts', 'id', 'uuid', 'id is uuid');
SELECT col_type_is('metadata', 'shared_layouts', 'api_name', 'character varying(63)', 'api_name is varchar(63)');
SELECT col_type_is('metadata', 'shared_layouts', 'type', 'character varying(20)', 'type is varchar(20)');
SELECT col_type_is('metadata', 'shared_layouts', 'config', 'jsonb', 'config is jsonb');

-- Primary key
SELECT col_is_pk('metadata', 'shared_layouts', 'id', 'id is primary key');

-- Unique constraint on api_name
SELECT has_index('metadata', 'shared_layouts', 'shared_layouts_api_name_unique', 'has unique index on api_name');

-- Check constraints
SELECT has_check('metadata', 'shared_layouts', 'has check constraints');

SELECT finish();
ROLLBACK;
