-- Copyright 2026 Adverax. All rights reserved.
-- Licensed under the Adverax Commercial License.
-- See ee/LICENSE for details.
-- Unauthorized use, copying, or distribution is prohibited.

-- Territory stored functions (ADR-0015: hybrid Go + PL/pgSQL approach)

-- 1. Rebuild territory hierarchy closure table using recursive CTE
CREATE OR REPLACE FUNCTION ee.rebuild_territory_hierarchy(p_model_id UUID)
RETURNS void AS $$
BEGIN
    DELETE FROM security.effective_territory_hierarchy
    WHERE ancestor_territory_id IN (
        SELECT id FROM ee.territories WHERE model_id = p_model_id
    );

    INSERT INTO security.effective_territory_hierarchy
        (ancestor_territory_id, descendant_territory_id, depth)
    WITH RECURSIVE closure AS (
        -- Self entries: each territory is its own ancestor at depth 0
        SELECT id AS ancestor, id AS descendant, 0 AS depth
        FROM ee.territories
        WHERE model_id = p_model_id
        UNION ALL
        -- Walk up: for each territory, add its parent as ancestor
        SELECT t.parent_id AS ancestor, c.descendant, c.depth + 1
        FROM closure c
        JOIN ee.territories t ON t.id = c.ancestor
        WHERE t.parent_id IS NOT NULL
          AND t.model_id = p_model_id
    )
    SELECT ancestor, descendant, depth FROM closure;
END;
$$ LANGUAGE plpgsql;

-- 2. Generate share entries for a single record assigned to a territory
-- Uses closure table + object_defaults to create entries for ancestor chain
CREATE OR REPLACE FUNCTION ee.generate_record_share_entries(
    p_record_id    UUID,
    p_object_id    UUID,
    p_territory_id UUID,
    p_share_table  TEXT
) RETURNS void AS $$
DECLARE
    rec RECORD;
BEGIN
    FOR rec IN
        SELECT g.id AS group_id, tod.access_level
        FROM security.effective_territory_hierarchy eth
        JOIN ee.territory_object_defaults tod
            ON tod.territory_id = eth.ancestor_territory_id
            AND tod.object_id = p_object_id
        JOIN iam.groups g
            ON g.related_territory_id = eth.ancestor_territory_id
            AND g.group_type = 'territory'
        WHERE eth.descendant_territory_id = p_territory_id
    LOOP
        EXECUTE format(
            'INSERT INTO %I (record_id, group_id, access_level, reason)
             VALUES ($1, $2, $3, $4)
             ON CONFLICT (record_id, group_id, reason) DO UPDATE SET access_level = $3',
            p_share_table
        ) USING p_record_id, rec.group_id, rec.access_level, 'territory';
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- 3. Full model activation: archive old, create groups, rebuild caches, generate shares
CREATE OR REPLACE FUNCTION ee.activate_territory_model(p_new_model_id UUID)
RETURNS void AS $$
DECLARE
    v_old_model_id UUID;
    v_territory RECORD;
    v_assignment RECORD;
    v_group_id UUID;
    v_share_table TEXT;
BEGIN
    -- 1. Archive old active model (if exists)
    SELECT id INTO v_old_model_id
    FROM ee.territory_models WHERE status = 'active';

    IF v_old_model_id IS NOT NULL THEN
        UPDATE ee.territory_models
        SET status = 'archived', archived_at = now(), updated_at = now()
        WHERE id = v_old_model_id;

        -- CASCADE: delete territory groups -> group_members -> effective_group_members
        DELETE FROM iam.groups
        WHERE related_territory_id IN (
            SELECT id FROM ee.territories WHERE model_id = v_old_model_id
        );

        -- Clean effective caches for old model
        DELETE FROM security.effective_territory_hierarchy
        WHERE ancestor_territory_id IN (
            SELECT id FROM ee.territories WHERE model_id = v_old_model_id
        );

        DELETE FROM security.effective_user_territory
        WHERE territory_id IN (
            SELECT id FROM ee.territories WHERE model_id = v_old_model_id
        );

        -- Clean territory share entries from all share tables
        FOR v_share_table IN
            SELECT table_name || '__share'
            FROM metadata.object_definitions
            WHERE visibility = 'private'
        LOOP
            EXECUTE format(
                'DELETE FROM %I WHERE reason = $1', v_share_table
            ) USING 'territory';
        END LOOP;
    END IF;

    -- 2. Activate new model
    UPDATE ee.territory_models
    SET status = 'active', activated_at = now(), updated_at = now()
    WHERE id = p_new_model_id;

    -- 3. Create territory groups + populate members (one group per territory)
    FOR v_territory IN
        SELECT id, api_name, label
        FROM ee.territories WHERE model_id = p_new_model_id
    LOOP
        INSERT INTO iam.groups (api_name, label, group_type, related_territory_id)
        VALUES (
            'territory_' || v_territory.api_name,
            v_territory.label,
            'territory',
            v_territory.id
        )
        RETURNING id INTO v_group_id;

        INSERT INTO iam.group_members (group_id, member_user_id)
        SELECT v_group_id, uta.user_id
        FROM ee.user_territory_assignments uta
        WHERE uta.territory_id = v_territory.id;
    END LOOP;

    -- 4. Rebuild effective_territory_hierarchy via recursive CTE
    PERFORM ee.rebuild_territory_hierarchy(p_new_model_id);

    -- 5. Rebuild effective_user_territory (flat list from assignments)
    INSERT INTO security.effective_user_territory (user_id, territory_id)
    SELECT uta.user_id, uta.territory_id
    FROM ee.user_territory_assignments uta
    JOIN ee.territories t ON t.id = uta.territory_id
    WHERE t.model_id = p_new_model_id;

    -- 6. Rebuild effective_group_members for new territory groups
    INSERT INTO security.effective_group_members (group_id, user_id)
    SELECT gm.group_id, gm.member_user_id
    FROM iam.group_members gm
    JOIN iam.groups g ON g.id = gm.group_id
    WHERE g.group_type = 'territory'
      AND g.related_territory_id IN (
          SELECT id FROM ee.territories WHERE model_id = p_new_model_id
      );

    -- 7. Generate share entries for all record assignments
    FOR v_assignment IN
        SELECT rta.record_id, rta.object_id, rta.territory_id, od.table_name
        FROM ee.record_territory_assignments rta
        JOIN ee.territories t ON t.id = rta.territory_id
        JOIN metadata.object_definitions od ON od.id = rta.object_id
        WHERE t.model_id = p_new_model_id
    LOOP
        PERFORM ee.generate_record_share_entries(
            v_assignment.record_id,
            v_assignment.object_id,
            v_assignment.territory_id,
            v_assignment.table_name || '__share'
        );
    END LOOP;
END;
$$ LANGUAGE plpgsql;
