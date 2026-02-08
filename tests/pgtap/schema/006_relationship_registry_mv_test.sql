BEGIN;

SELECT plan(5);

-- 1. Materialized view exists
SELECT has_materialized_view(
    'metadata', 'relationship_registry',
    'Materialized view metadata.relationship_registry should exist'
);

-- 2-8. Required columns
SELECT has_column('metadata', 'relationship_registry', 'field_id',
    'relationship_registry should have field_id column');
SELECT has_column('metadata', 'relationship_registry', 'child_object_id',
    'relationship_registry should have child_object_id column');
SELECT has_column('metadata', 'relationship_registry', 'parent_object_id',
    'relationship_registry should have parent_object_id column');
SELECT has_column('metadata', 'relationship_registry', 'reference_subtype',
    'relationship_registry should have reference_subtype column');

SELECT finish();

ROLLBACK;
