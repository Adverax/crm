-- Migration 000034: Transform portal config from flat read-based to gate graph model (ADR-0037).
-- Converts: {args, read: {fields, actions, queries}} → {args, entry_gate: "main", gates: {"main": {...}}}

UPDATE metadata.portals
SET config = jsonb_build_object(
    'args', COALESCE(config->'args', '[]'::jsonb),
    'entry_gate', 'main',
    'gates', jsonb_build_object(
        'main', jsonb_build_object(
            'label', label,
            'body', (
                SELECT COALESCE(jsonb_agg(
                    jsonb_build_object(
                        'name', q->>'name',
                        'type', 'soql',
                        'soql', q->>'soql'
                    ) ||
                    CASE WHEN q->>'when' IS NOT NULL AND q->>'when' != ''
                         THEN jsonb_build_object('when', q->>'when')
                         ELSE '{}'::jsonb
                    END
                ), '[]'::jsonb)
                FROM jsonb_array_elements(COALESCE(config->'read'->'queries', '[]'::jsonb)) AS q
            ),
            'layout', jsonb_build_object(
                'fields', COALESCE(config->'read'->'fields', '[]'::jsonb)
            ),
            'outcomes', '[]'::jsonb
        )
    )
),
updated_at = NOW()
WHERE config ? 'read';
