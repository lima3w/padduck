-- +migrate Up
-- Rewrite legacy object-form conditions ({"field": "value"}) into the array
-- form ([{"field","operator","value"}]) so every stored policy uses one shape.
-- Values ending in * become glob conditions; everything else is eq.
UPDATE automation_policies
SET conditions = COALESCE((
    SELECT jsonb_agg(jsonb_build_object(
        'field', e.key,
        'operator', CASE WHEN e.value LIKE '%*' THEN 'glob' ELSE 'eq' END,
        'value', e.value) ORDER BY e.key)
    FROM jsonb_each_text(automation_policies.conditions) e
), '[]'::jsonb)
WHERE jsonb_typeof(conditions) = 'object';

ALTER TABLE automation_policies ALTER COLUMN conditions SET DEFAULT '[]'::jsonb;
