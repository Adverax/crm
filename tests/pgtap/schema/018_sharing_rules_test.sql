BEGIN;
SELECT plan(20);

SELECT has_table('security', 'sharing_rules', 'table security.sharing_rules exists');

SELECT has_column('security', 'sharing_rules', 'id');
SELECT col_type_is('security', 'sharing_rules', 'id', 'uuid');
SELECT col_is_pk('security', 'sharing_rules', 'id');

SELECT has_column('security', 'sharing_rules', 'object_id');
SELECT col_not_null('security', 'sharing_rules', 'object_id');
SELECT fk_ok('security', 'sharing_rules', 'object_id', 'metadata', 'object_definitions', 'id');

SELECT has_column('security', 'sharing_rules', 'rule_type');
SELECT col_not_null('security', 'sharing_rules', 'rule_type');

SELECT has_column('security', 'sharing_rules', 'source_group_id');
SELECT col_not_null('security', 'sharing_rules', 'source_group_id');
SELECT fk_ok('security', 'sharing_rules', 'source_group_id', 'iam', 'groups', 'id');

SELECT has_column('security', 'sharing_rules', 'target_group_id');
SELECT col_not_null('security', 'sharing_rules', 'target_group_id');
SELECT fk_ok('security', 'sharing_rules', 'target_group_id', 'iam', 'groups', 'id');

SELECT has_column('security', 'sharing_rules', 'access_level');
SELECT col_not_null('security', 'sharing_rules', 'access_level');

SELECT has_column('security', 'sharing_rules', 'criteria_field');
SELECT has_column('security', 'sharing_rules', 'criteria_op');
SELECT has_column('security', 'sharing_rules', 'criteria_value');

SELECT finish();
ROLLBACK;
