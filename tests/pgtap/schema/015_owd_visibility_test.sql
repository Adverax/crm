BEGIN;
SELECT plan(5);

-- Column exists
SELECT has_column('metadata', 'object_definitions', 'visibility',
    'column metadata.object_definitions.visibility exists');

-- Type
SELECT col_type_is('metadata', 'object_definitions', 'visibility', 'character varying(30)',
    'visibility is varchar(30)');

-- NOT NULL
SELECT col_not_null('metadata', 'object_definitions', 'visibility',
    'visibility is NOT NULL');

-- Default value
SELECT col_has_default('metadata', 'object_definitions', 'visibility',
    'visibility has a default value');

SELECT col_default_is('metadata', 'object_definitions', 'visibility', 'private',
    'visibility defaults to private');

SELECT finish();
ROLLBACK;
