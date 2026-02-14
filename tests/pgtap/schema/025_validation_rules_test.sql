BEGIN;
SELECT plan(16);

-- Table exists
SELECT has_table('metadata', 'validation_rules', 'has metadata.validation_rules table');

-- Columns
SELECT has_column('metadata', 'validation_rules', 'id', 'has id column');
SELECT has_column('metadata', 'validation_rules', 'object_id', 'has object_id column');
SELECT has_column('metadata', 'validation_rules', 'api_name', 'has api_name column');
SELECT has_column('metadata', 'validation_rules', 'label', 'has label column');
SELECT has_column('metadata', 'validation_rules', 'expression', 'has expression column');
SELECT has_column('metadata', 'validation_rules', 'error_message', 'has error_message column');
SELECT has_column('metadata', 'validation_rules', 'error_code', 'has error_code column');
SELECT has_column('metadata', 'validation_rules', 'severity', 'has severity column');
SELECT has_column('metadata', 'validation_rules', 'when_expression', 'has when_expression column');
SELECT has_column('metadata', 'validation_rules', 'applies_to', 'has applies_to column');
SELECT has_column('metadata', 'validation_rules', 'sort_order', 'has sort_order column');
SELECT has_column('metadata', 'validation_rules', 'is_active', 'has is_active column');
SELECT has_column('metadata', 'validation_rules', 'created_at', 'has created_at column');
SELECT has_column('metadata', 'validation_rules', 'updated_at', 'has updated_at column');

-- Index
SELECT has_index('metadata', 'validation_rules', 'idx_validation_rules_object_id', 'has object_id index');

SELECT * FROM finish();
ROLLBACK;
