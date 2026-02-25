BEGIN;
SELECT plan(19);

-- ===== metadata.automation_rules =====
SELECT has_table('metadata', 'automation_rules', 'has automation_rules table');

SELECT has_column('metadata', 'automation_rules', 'id', 'has id column');
SELECT col_type_is('metadata', 'automation_rules', 'id', 'uuid', 'id is uuid');
SELECT col_is_pk('metadata', 'automation_rules', 'id', 'id is PK');

SELECT has_column('metadata', 'automation_rules', 'object_id', 'has object_id column');
SELECT col_not_null('metadata', 'automation_rules', 'object_id', 'object_id is NOT NULL');

SELECT has_column('metadata', 'automation_rules', 'name', 'has name column');
SELECT col_not_null('metadata', 'automation_rules', 'name', 'name is NOT NULL');

SELECT has_column('metadata', 'automation_rules', 'event_type', 'has event_type column');
SELECT col_not_null('metadata', 'automation_rules', 'event_type', 'event_type is NOT NULL');

SELECT has_column('metadata', 'automation_rules', 'condition', 'has condition column');

SELECT has_column('metadata', 'automation_rules', 'procedure_code', 'has procedure_code column');
SELECT col_not_null('metadata', 'automation_rules', 'procedure_code', 'procedure_code is NOT NULL');

SELECT has_column('metadata', 'automation_rules', 'execution_mode', 'has execution_mode column');
SELECT has_column('metadata', 'automation_rules', 'sort_order', 'has sort_order column');
SELECT has_column('metadata', 'automation_rules', 'is_active', 'has is_active column');

SELECT has_column('metadata', 'automation_rules', 'created_at', 'has created_at column');
SELECT has_column('metadata', 'automation_rules', 'updated_at', 'has updated_at column');

SELECT fk_ok('metadata', 'automation_rules', 'object_id', 'metadata', 'object_definitions', 'id', 'FK to object_definitions');

SELECT * FROM finish();
ROLLBACK;
