-- name: CreatePolymorphicTarget :one
INSERT INTO metadata.polymorphic_targets (
    field_id, object_id
) VALUES ($1, $2)
RETURNING id, field_id, object_id, created_at;

-- name: ListPolymorphicTargetsByFieldID :many
SELECT id, field_id, object_id, created_at
FROM metadata.polymorphic_targets
WHERE field_id = $1
ORDER BY created_at;

-- name: ListAllPolymorphicTargets :many
SELECT id, field_id, object_id, created_at
FROM metadata.polymorphic_targets
ORDER BY field_id, object_id;

-- name: DeletePolymorphicTarget :exec
DELETE FROM metadata.polymorphic_targets WHERE id = $1;

-- name: DeletePolymorphicTargetsByFieldID :exec
DELETE FROM metadata.polymorphic_targets WHERE field_id = $1;
