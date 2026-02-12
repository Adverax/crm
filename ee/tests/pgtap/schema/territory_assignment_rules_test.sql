-- Copyright 2026 Adverax. All rights reserved.
-- Licensed under the Adverax Commercial License.
-- See ee/LICENSE for details.
-- Unauthorized use, copying, or distribution is prohibited.

BEGIN;
SELECT plan(32);

-- ee.territory_assignment_rules table
SELECT has_table('ee', 'territory_assignment_rules', 'table ee.territory_assignment_rules exists');

-- id
SELECT has_column('ee', 'territory_assignment_rules', 'id', 'has id');
SELECT col_type_is('ee', 'territory_assignment_rules', 'id', 'uuid', 'id is uuid');
SELECT col_has_default('ee', 'territory_assignment_rules', 'id', 'id has default');
SELECT col_is_pk('ee', 'territory_assignment_rules', 'id', 'id is PK');

-- territory_id
SELECT has_column('ee', 'territory_assignment_rules', 'territory_id', 'has territory_id');
SELECT col_not_null('ee', 'territory_assignment_rules', 'territory_id', 'territory_id is NOT NULL');
SELECT fk_ok('ee', 'territory_assignment_rules', 'territory_id', 'ee', 'territories', 'id',
    'territory_id FK to ee.territories');

-- object_id
SELECT has_column('ee', 'territory_assignment_rules', 'object_id', 'has object_id');
SELECT col_not_null('ee', 'territory_assignment_rules', 'object_id', 'object_id is NOT NULL');
SELECT fk_ok('ee', 'territory_assignment_rules', 'object_id', 'metadata', 'object_definitions', 'id',
    'object_id FK to metadata.object_definitions');

-- is_active
SELECT has_column('ee', 'territory_assignment_rules', 'is_active', 'has is_active');
SELECT col_type_is('ee', 'territory_assignment_rules', 'is_active', 'boolean', 'is_active is boolean');
SELECT col_not_null('ee', 'territory_assignment_rules', 'is_active', 'is_active is NOT NULL');
SELECT col_has_default('ee', 'territory_assignment_rules', 'is_active', 'is_active has default');

-- rule_order
SELECT has_column('ee', 'territory_assignment_rules', 'rule_order', 'has rule_order');
SELECT col_type_is('ee', 'territory_assignment_rules', 'rule_order', 'integer', 'rule_order is integer');
SELECT col_not_null('ee', 'territory_assignment_rules', 'rule_order', 'rule_order is NOT NULL');
SELECT col_has_default('ee', 'territory_assignment_rules', 'rule_order', 'rule_order has default');

-- criteria_field
SELECT has_column('ee', 'territory_assignment_rules', 'criteria_field', 'has criteria_field');
SELECT col_not_null('ee', 'territory_assignment_rules', 'criteria_field', 'criteria_field is NOT NULL');

-- criteria_op
SELECT has_column('ee', 'territory_assignment_rules', 'criteria_op', 'has criteria_op');
SELECT col_not_null('ee', 'territory_assignment_rules', 'criteria_op', 'criteria_op is NOT NULL');
SELECT has_check('ee', 'territory_assignment_rules', 'criteria_op has CHECK');

-- criteria_value
SELECT has_column('ee', 'territory_assignment_rules', 'criteria_value', 'has criteria_value');
SELECT col_not_null('ee', 'territory_assignment_rules', 'criteria_value', 'criteria_value is NOT NULL');

-- created_at / updated_at
SELECT has_column('ee', 'territory_assignment_rules', 'created_at', 'has created_at');
SELECT col_not_null('ee', 'territory_assignment_rules', 'created_at', 'created_at is NOT NULL');

SELECT has_column('ee', 'territory_assignment_rules', 'updated_at', 'has updated_at');
SELECT col_not_null('ee', 'territory_assignment_rules', 'updated_at', 'updated_at is NOT NULL');

-- Indexes
SELECT has_index('ee', 'territory_assignment_rules', 'idx_territory_assignment_rules_territory', 'index on territory_id');
SELECT has_index('ee', 'territory_assignment_rules', 'idx_territory_assignment_rules_object', 'index on object_id');

SELECT finish();
ROLLBACK;
