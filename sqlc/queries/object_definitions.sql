-- name: CreateObjectDefinition :one
INSERT INTO metadata.object_definitions (
    api_name, label, plural_label, description, table_name, object_type,
    is_platform_managed, is_visible_in_setup, is_custom_fields_allowed, is_deleteable_object,
    is_createable, is_updateable, is_deleteable, is_queryable, is_searchable,
    has_activities, has_notes, has_history_tracking, has_sharing_rules
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10,
    $11, $12, $13, $14, $15,
    $16, $17, $18, $19
) RETURNING id, api_name, label, plural_label, description, table_name, object_type,
    is_platform_managed, is_visible_in_setup, is_custom_fields_allowed, is_deleteable_object,
    is_createable, is_updateable, is_deleteable, is_queryable, is_searchable,
    has_activities, has_notes, has_history_tracking, has_sharing_rules,
    created_at, updated_at;

-- name: GetObjectDefinitionByID :one
SELECT id, api_name, label, plural_label, description, table_name, object_type,
    is_platform_managed, is_visible_in_setup, is_custom_fields_allowed, is_deleteable_object,
    is_createable, is_updateable, is_deleteable, is_queryable, is_searchable,
    has_activities, has_notes, has_history_tracking, has_sharing_rules,
    created_at, updated_at
FROM metadata.object_definitions
WHERE id = $1;

-- name: GetObjectDefinitionByAPIName :one
SELECT id, api_name, label, plural_label, description, table_name, object_type,
    is_platform_managed, is_visible_in_setup, is_custom_fields_allowed, is_deleteable_object,
    is_createable, is_updateable, is_deleteable, is_queryable, is_searchable,
    has_activities, has_notes, has_history_tracking, has_sharing_rules,
    created_at, updated_at
FROM metadata.object_definitions
WHERE api_name = $1;

-- name: ListObjectDefinitions :many
SELECT id, api_name, label, plural_label, description, table_name, object_type,
    is_platform_managed, is_visible_in_setup, is_custom_fields_allowed, is_deleteable_object,
    is_createable, is_updateable, is_deleteable, is_queryable, is_searchable,
    has_activities, has_notes, has_history_tracking, has_sharing_rules,
    created_at, updated_at
FROM metadata.object_definitions
ORDER BY api_name
LIMIT $1 OFFSET $2;

-- name: ListAllObjectDefinitions :many
SELECT id, api_name, label, plural_label, description, table_name, object_type,
    is_platform_managed, is_visible_in_setup, is_custom_fields_allowed, is_deleteable_object,
    is_createable, is_updateable, is_deleteable, is_queryable, is_searchable,
    has_activities, has_notes, has_history_tracking, has_sharing_rules,
    created_at, updated_at
FROM metadata.object_definitions
ORDER BY api_name;

-- name: UpdateObjectDefinition :one
UPDATE metadata.object_definitions SET
    label = $2,
    plural_label = $3,
    description = $4,
    is_visible_in_setup = $5,
    is_custom_fields_allowed = $6,
    is_deleteable_object = $7,
    is_createable = $8,
    is_updateable = $9,
    is_deleteable = $10,
    is_queryable = $11,
    is_searchable = $12,
    has_activities = $13,
    has_notes = $14,
    has_history_tracking = $15,
    has_sharing_rules = $16,
    updated_at = now()
WHERE id = $1
RETURNING id, api_name, label, plural_label, description, table_name, object_type,
    is_platform_managed, is_visible_in_setup, is_custom_fields_allowed, is_deleteable_object,
    is_createable, is_updateable, is_deleteable, is_queryable, is_searchable,
    has_activities, has_notes, has_history_tracking, has_sharing_rules,
    created_at, updated_at;

-- name: DeleteObjectDefinition :exec
DELETE FROM metadata.object_definitions WHERE id = $1;

-- name: CountObjectDefinitions :one
SELECT count(*) FROM metadata.object_definitions;
