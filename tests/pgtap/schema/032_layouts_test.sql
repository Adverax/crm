BEGIN;
SELECT plan(17);

-- Table exists
SELECT has_table('metadata', 'layouts', 'has metadata.layouts table');

-- Columns
SELECT has_column('metadata', 'layouts', 'id', 'has id column');
SELECT has_column('metadata', 'layouts', 'object_view_id', 'has object_view_id column');
SELECT has_column('metadata', 'layouts', 'form_factor', 'has form_factor column');
SELECT has_column('metadata', 'layouts', 'mode', 'has mode column');
SELECT has_column('metadata', 'layouts', 'config', 'has config column');
SELECT has_column('metadata', 'layouts', 'created_at', 'has created_at column');
SELECT has_column('metadata', 'layouts', 'updated_at', 'has updated_at column');

-- Column types
SELECT col_type_is('metadata', 'layouts', 'id', 'uuid', 'id is uuid');
SELECT col_type_is('metadata', 'layouts', 'object_view_id', 'uuid', 'object_view_id is uuid');
SELECT col_type_is('metadata', 'layouts', 'form_factor', 'character varying(20)', 'form_factor is varchar(20)');
SELECT col_type_is('metadata', 'layouts', 'mode', 'character varying(20)', 'mode is varchar(20)');
SELECT col_type_is('metadata', 'layouts', 'config', 'jsonb', 'config is jsonb');

-- Primary key
SELECT col_is_pk('metadata', 'layouts', 'id', 'id is primary key');

-- FK to object_views
SELECT fk_ok('metadata', 'layouts', 'object_view_id', 'metadata', 'object_views', 'id', 'FK to object_views');

-- Unique constraint
SELECT has_index('metadata', 'layouts', 'layouts_ov_ff_mode_unique', 'has unique index on (object_view_id, form_factor, mode)');

-- Index
SELECT has_index('metadata', 'layouts', 'idx_layouts_object_view_id', 'has index on object_view_id');

SELECT finish();
ROLLBACK;
