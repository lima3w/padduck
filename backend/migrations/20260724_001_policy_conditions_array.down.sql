-- +migrate Down
-- Best effort: the object form cannot hold operators, so only eq and glob
-- conditions survive; other operators are dropped.
ALTER TABLE automation_policies ALTER COLUMN conditions SET DEFAULT '{}'::jsonb;
UPDATE automation_policies
SET conditions = COALESCE((
    SELECT jsonb_object_agg(c->>'field', c->>'value')
    FROM jsonb_array_elements(automation_policies.conditions) c
    WHERE c->>'operator' IN ('eq', 'glob')
), '{}'::jsonb)
WHERE jsonb_typeof(conditions) = 'array';
