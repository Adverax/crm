-- name: ListRelationships :many
SELECT
    field_id,
    field_api_name,
    relationship_name,
    child_object_id,
    child_object_api_name,
    parent_object_id,
    parent_object_api_name,
    reference_subtype,
    on_delete
FROM metadata.relationship_registry;
