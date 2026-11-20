-- +migrate Down
DROP INDEX IF EXISTS resource_intents_org_status_idx;
ALTER TABLE resource_intents ALTER COLUMN organization_id DROP NOT NULL;
ALTER TABLE resource_intents ALTER COLUMN organization_id DROP DEFAULT;
