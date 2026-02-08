BEGIN;
SELECT plan(16);

-- Таблица существует
SELECT has_table('metadata', 'polymorphic_targets', 'table metadata.polymorphic_targets exists');

-- Колонки
SELECT has_column('metadata', 'polymorphic_targets', 'id');
SELECT col_type_is('metadata', 'polymorphic_targets', 'id', 'uuid');
SELECT col_is_pk('metadata', 'polymorphic_targets', 'id');

SELECT has_column('metadata', 'polymorphic_targets', 'field_id');
SELECT col_not_null('metadata', 'polymorphic_targets', 'field_id');

SELECT has_column('metadata', 'polymorphic_targets', 'object_id');
SELECT col_not_null('metadata', 'polymorphic_targets', 'object_id');

SELECT has_column('metadata', 'polymorphic_targets', 'created_at');
SELECT col_not_null('metadata', 'polymorphic_targets', 'created_at');

-- FK
SELECT fk_ok('metadata', 'polymorphic_targets', 'field_id', 'metadata', 'field_definitions', 'id');
SELECT fk_ok('metadata', 'polymorphic_targets', 'object_id', 'metadata', 'object_definitions', 'id');

-- Unique constraint
SELECT has_unique('metadata', 'polymorphic_targets', 'polymorphic_targets has UNIQUE constraint');

-- Индексы
SELECT has_index('metadata', 'polymorphic_targets', 'idx_polymorphic_targets_field_id', 'index on field_id exists');
SELECT has_index('metadata', 'polymorphic_targets', 'idx_polymorphic_targets_object_id', 'index on object_id exists');

SELECT finish();
ROLLBACK;
