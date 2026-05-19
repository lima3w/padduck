-- +migrate Up
ALTER TABLE webhook_endpoints
    ADD COLUMN IF NOT EXISTS object_types TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS tag_filters TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS filter_conditions JSONB NOT NULL DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_webhook_endpoints_events ON webhook_endpoints USING GIN (events);
CREATE INDEX IF NOT EXISTS idx_webhook_endpoints_object_types ON webhook_endpoints USING GIN (object_types);
CREATE INDEX IF NOT EXISTS idx_webhook_endpoints_tag_filters ON webhook_endpoints USING GIN (tag_filters);

CREATE TABLE IF NOT EXISTS automation_policies (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    workflow TEXT NOT NULL,
    action TEXT NOT NULL,
    effect TEXT NOT NULL DEFAULT 'allow' CHECK (effect IN ('allow', 'deny', 'manual_review')),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    conditions JSONB NOT NULL DEFAULT '{}'::jsonb,
    message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_automation_policies_lookup
    ON automation_policies(workflow, action, enabled);
