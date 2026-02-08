-- name: CreatePicklistValue :one
INSERT INTO metadata.picklist_values (
    picklist_definition_id, value, label, sort_order, is_default, is_active
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, picklist_definition_id, value, label, sort_order, is_default, is_active,
    created_at, updated_at;

-- name: GetPicklistValueByID :one
SELECT id, picklist_definition_id, value, label, sort_order, is_default, is_active,
    created_at, updated_at
FROM metadata.picklist_values
WHERE id = $1;

-- name: ListPicklistValuesByDefinitionID :many
SELECT id, picklist_definition_id, value, label, sort_order, is_default, is_active,
    created_at, updated_at
FROM metadata.picklist_values
WHERE picklist_definition_id = $1
ORDER BY sort_order, value;

-- name: UpdatePicklistValue :one
UPDATE metadata.picklist_values SET
    label = $2,
    sort_order = $3,
    is_default = $4,
    is_active = $5,
    updated_at = now()
WHERE id = $1
RETURNING id, picklist_definition_id, value, label, sort_order, is_default, is_active,
    created_at, updated_at;

-- name: DeletePicklistValue :exec
DELETE FROM metadata.picklist_values WHERE id = $1;

-- name: DeletePicklistValuesByDefinitionID :exec
DELETE FROM metadata.picklist_values WHERE picklist_definition_id = $1;
