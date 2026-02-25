BEGIN;
SELECT plan(26);

-- Table: metadata.procedures
SELECT has_table('metadata', 'procedures', 'has metadata.procedures table');
SELECT has_column('metadata', 'procedures', 'id', 'has id column');
SELECT has_column('metadata', 'procedures', 'code', 'has code column');
SELECT has_column('metadata', 'procedures', 'name', 'has name column');
SELECT has_column('metadata', 'procedures', 'description', 'has description column');
SELECT has_column('metadata', 'procedures', 'draft_version_id', 'has draft_version_id column');
SELECT has_column('metadata', 'procedures', 'published_version_id', 'has published_version_id column');
SELECT has_column('metadata', 'procedures', 'created_at', 'has created_at column');
SELECT has_column('metadata', 'procedures', 'updated_at', 'has updated_at column');

SELECT col_type_is('metadata', 'procedures', 'id', 'uuid', 'id is uuid');
SELECT col_type_is('metadata', 'procedures', 'code', 'character varying(100)', 'code is varchar(100)');
SELECT col_type_is('metadata', 'procedures', 'name', 'character varying(255)', 'name is varchar(255)');

SELECT has_check('metadata', 'procedures', 'has check constraints');

-- Table: metadata.procedure_versions
SELECT has_table('metadata', 'procedure_versions', 'has metadata.procedure_versions table');
SELECT has_column('metadata', 'procedure_versions', 'id', 'has id column');
SELECT has_column('metadata', 'procedure_versions', 'procedure_id', 'has procedure_id column');
SELECT has_column('metadata', 'procedure_versions', 'version', 'has version column');
SELECT has_column('metadata', 'procedure_versions', 'definition', 'has definition column');
SELECT has_column('metadata', 'procedure_versions', 'status', 'has status column');
SELECT has_column('metadata', 'procedure_versions', 'change_summary', 'has change_summary column');
SELECT has_column('metadata', 'procedure_versions', 'created_by', 'has created_by column');
SELECT has_column('metadata', 'procedure_versions', 'created_at', 'has created_at column');
SELECT has_column('metadata', 'procedure_versions', 'published_at', 'has published_at column');

SELECT col_type_is('metadata', 'procedure_versions', 'definition', 'jsonb', 'definition is jsonb');

-- Indexes
SELECT has_index('metadata', 'procedure_versions', 'idx_procedure_versions_procedure_id', 'has procedure_id index');
SELECT has_index('metadata', 'procedure_versions', 'idx_procedure_versions_status', 'has status index');

SELECT * FROM finish();
ROLLBACK;
