-- +migrate Down
DROP TABLE IF EXISTS automation_policies;
DROP INDEX IF EXISTS idx_webhook_endpoints_tag_filters;
DROP INDEX IF EXISTS idx_webhook_endpoints_object_types;
DROP INDEX IF EXISTS idx_webhook_endpoints_events;
ALTER TABLE webhook_endpoints
    DROP COLUMN IF EXISTS filter_conditions,
    DROP COLUMN IF EXISTS tag_filters,
    DROP COLUMN IF EXISTS object_types;
