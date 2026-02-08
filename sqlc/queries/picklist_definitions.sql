-- name: CreatePicklistDefinition :one
INSERT INTO metadata.picklist_definitions (
    api_name, label, description
) VALUES ($1, $2, $3)
RETURNING id, api_name, label, description, created_at, updated_at;

-- name: GetPicklistDefinitionByID :one
SELECT id, api_name, label, description, created_at, updated_at
FROM metadata.picklist_definitions
WHERE id = $1;

-- name: GetPicklistDefinitionByAPIName :one
SELECT id, api_name, label, description, created_at, updated_at
FROM metadata.picklist_definitions
WHERE api_name = $1;

-- name: ListPicklistDefinitions :many
SELECT id, api_name, label, description, created_at, updated_at
FROM metadata.picklist_definitions
ORDER BY api_name;

-- name: UpdatePicklistDefinition :one
UPDATE metadata.picklist_definitions SET
    label = $2,
    description = $3,
    updated_at = now()
WHERE id = $1
RETURNING id, api_name, label, description, created_at, updated_at;

-- name: DeletePicklistDefinition :exec
DELETE FROM metadata.picklist_definitions WHERE id = $1;
