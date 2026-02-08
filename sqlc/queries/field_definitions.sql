-- name: CreateFieldDefinition :one
INSERT INTO metadata.field_definitions (
    object_id, api_name, label, description, help_text,
    field_type, field_subtype, referenced_object_id,
    is_required, is_unique, config,
    is_system_field, is_custom, is_platform_managed, sort_order
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8,
    $9, $10, $11,
    $12, $13, $14, $15
) RETURNING id, object_id, api_name, label, description, help_text,
    field_type, field_subtype, referenced_object_id,
    is_required, is_unique, config,
    is_system_field, is_custom, is_platform_managed, sort_order,
    created_at, updated_at;

-- name: GetFieldDefinitionByID :one
SELECT id, object_id, api_name, label, description, help_text,
    field_type, field_subtype, referenced_object_id,
    is_required, is_unique, config,
    is_system_field, is_custom, is_platform_managed, sort_order,
    created_at, updated_at
FROM metadata.field_definitions
WHERE id = $1;

-- name: GetFieldDefinitionByObjectAndName :one
SELECT id, object_id, api_name, label, description, help_text,
    field_type, field_subtype, referenced_object_id,
    is_required, is_unique, config,
    is_system_field, is_custom, is_platform_managed, sort_order,
    created_at, updated_at
FROM metadata.field_definitions
WHERE object_id = $1 AND api_name = $2;

-- name: ListFieldDefinitionsByObjectID :many
SELECT id, object_id, api_name, label, description, help_text,
    field_type, field_subtype, referenced_object_id,
    is_required, is_unique, config,
    is_system_field, is_custom, is_platform_managed, sort_order,
    created_at, updated_at
FROM metadata.field_definitions
WHERE object_id = $1
ORDER BY sort_order, api_name;

-- name: ListAllFieldDefinitions :many
SELECT id, object_id, api_name, label, description, help_text,
    field_type, field_subtype, referenced_object_id,
    is_required, is_unique, config,
    is_system_field, is_custom, is_platform_managed, sort_order,
    created_at, updated_at
FROM metadata.field_definitions
ORDER BY object_id, sort_order, api_name;

-- name: ListReferenceFieldDefinitions :many
SELECT id, object_id, api_name, label, description, help_text,
    field_type, field_subtype, referenced_object_id,
    is_required, is_unique, config,
    is_system_field, is_custom, is_platform_managed, sort_order,
    created_at, updated_at
FROM metadata.field_definitions
WHERE field_type = 'reference'
ORDER BY object_id, api_name;

-- name: UpdateFieldDefinition :one
UPDATE metadata.field_definitions SET
    label = $2,
    description = $3,
    help_text = $4,
    is_required = $5,
    is_unique = $6,
    config = $7,
    sort_order = $8,
    updated_at = now()
WHERE id = $1
RETURNING id, object_id, api_name, label, description, help_text,
    field_type, field_subtype, referenced_object_id,
    is_required, is_unique, config,
    is_system_field, is_custom, is_platform_managed, sort_order,
    created_at, updated_at;

-- name: DeleteFieldDefinition :exec
DELETE FROM metadata.field_definitions WHERE id = $1;
