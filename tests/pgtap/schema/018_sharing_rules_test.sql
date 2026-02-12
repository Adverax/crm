BEGIN;
SELECT plan(20);

SELECT has_table('security', 'sharing_rules', 'table security.sharing_rules exists');

SELECT has_column('security', 'sharing_rules', 'id', 'has id');
SELECT col_type_is('security', 'sharing_rules', 'id', 'uuid', 'id is uuid');
SELECT col_is_pk('security', 'sharing_rules', 'id', 'id is PK');

SELECT has_column('security', 'sharing_rules', 'object_id', 'has object_id');
SELECT col_not_null('security', 'sharing_rules', 'object_id', 'object_id is NOT NULL');
SELECT fk_ok('security', 'sharing_rules', 'object_id', 'metadata', 'object_definitions', 'id', 'FK object_id -> object_definitions.id');

SELECT has_column('security', 'sharing_rules', 'rule_type', 'has rule_type');
SELECT col_not_null('security', 'sharing_rules', 'rule_type', 'rule_type is NOT NULL');

SELECT has_column('security', 'sharing_rules', 'source_group_id', 'has source_group_id');
SELECT col_not_null('security', 'sharing_rules', 'source_group_id', 'source_group_id is NOT NULL');
SELECT fk_ok('security', 'sharing_rules', 'source_group_id', 'iam', 'groups', 'id', 'FK source_group_id -> groups.id');

SELECT has_column('security', 'sharing_rules', 'target_group_id', 'has target_group_id');
SELECT col_not_null('security', 'sharing_rules', 'target_group_id', 'target_group_id is NOT NULL');
SELECT fk_ok('security', 'sharing_rules', 'target_group_id', 'iam', 'groups', 'id', 'FK target_group_id -> groups.id');

SELECT has_column('security', 'sharing_rules', 'access_level', 'has access_level');
SELECT col_not_null('security', 'sharing_rules', 'access_level', 'access_level is NOT NULL');

SELECT has_column('security', 'sharing_rules', 'criteria_field', 'has criteria_field');
SELECT has_column('security', 'sharing_rules', 'criteria_op', 'has criteria_op');
SELECT has_column('security', 'sharing_rules', 'criteria_value', 'has criteria_value');

SELECT finish();
ROLLBACK;
