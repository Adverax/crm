-- Materialized view for relationship registry (ADR-0006).
-- Derived from field_definitions + object_definitions + polymorphic_targets.
-- Refreshed on metadata changes (field create/update/delete).

CREATE MATERIALIZED VIEW IF NOT EXISTS metadata.relationship_registry AS
SELECT
    fd.id                      AS field_id,
    fd.api_name                AS field_api_name,
    fd.config->>'relationship_name' AS relationship_name,
    fd.object_id               AS child_object_id,
    child_obj.api_name         AS child_object_api_name,
    fd.referenced_object_id    AS parent_object_id,
    parent_obj.api_name        AS parent_object_api_name,
    fd.field_subtype            AS reference_subtype,
    COALESCE(fd.config->>'on_delete',
        CASE fd.field_subtype
            WHEN 'association' THEN 'set_null'
            WHEN 'composition' THEN 'cascade'
            ELSE ''
        END
    )                          AS on_delete
FROM metadata.field_definitions fd
JOIN metadata.object_definitions child_obj  ON child_obj.id  = fd.object_id
LEFT JOIN metadata.object_definitions parent_obj ON parent_obj.id = fd.referenced_object_id
WHERE fd.field_type = 'reference'
  AND fd.field_subtype IN ('association', 'composition')

UNION ALL

SELECT
    fd.id                      AS field_id,
    fd.api_name                AS field_api_name,
    fd.config->>'relationship_name' AS relationship_name,
    fd.object_id               AS child_object_id,
    child_obj.api_name         AS child_object_api_name,
    pt.object_id               AS parent_object_id,
    parent_obj.api_name        AS parent_object_api_name,
    fd.field_subtype            AS reference_subtype,
    ''                         AS on_delete
FROM metadata.field_definitions fd
JOIN metadata.object_definitions child_obj   ON child_obj.id  = fd.object_id
JOIN metadata.polymorphic_targets pt         ON pt.field_id   = fd.id
JOIN metadata.object_definitions parent_obj  ON parent_obj.id = pt.object_id
WHERE fd.field_type = 'reference'
  AND fd.field_subtype = 'polymorphic'
WITH NO DATA;

-- Unique index on MV for REFRESH MATERIALIZED VIEW CONCURRENTLY
CREATE UNIQUE INDEX IF NOT EXISTS idx_relationship_registry_field_parent
    ON metadata.relationship_registry (field_id, parent_object_id);

-- Initial population
REFRESH MATERIALIZED VIEW metadata.relationship_registry;
