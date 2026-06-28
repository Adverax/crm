-- Migration 000034 DOWN: Reverse gate graph → flat read-based config.
-- Only single-gate portals can be losslessly reversed. Multi-gate portals lose action gates.

UPDATE metadata.portals
SET config = jsonb_build_object(
    'args', COALESCE(config->'args', '[]'::jsonb),
    'read', jsonb_build_object(
        'fields', COALESCE(
            config->'gates'->>(config->>'entry_gate')->'layout'->'fields',
            '[]'::jsonb
        ),
        'actions', '[]'::jsonb,
        'queries', (
            SELECT COALESCE(jsonb_agg(
                jsonb_build_object(
                    'name', step->>'name',
                    'soql', step->>'soql'
                ) ||
                CASE WHEN step->>'when' IS NOT NULL AND step->>'when' != ''
                     THEN jsonb_build_object('when', step->>'when')
                     ELSE '{}'::jsonb
                END
            ), '[]'::jsonb)
            FROM jsonb_array_elements(
                COALESCE(
                    config->'gates'->>(config->>'entry_gate')->'body',
                    '[]'::jsonb
                )
            ) AS step
            WHERE step->>'type' = 'soql'
        )
    )
),
updated_at = NOW()
WHERE config ? 'entry_gate';
