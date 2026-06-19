-- +migrate Down
DROP INDEX IF EXISTS scheduled_reports_org;
DROP INDEX IF EXISTS api_tokens_org;
DROP INDEX IF EXISTS webhooks_org;
DROP INDEX IF EXISTS audit_logs_org_created;

ALTER TABLE scheduled_reports  DROP COLUMN IF EXISTS organization_id;
ALTER TABLE audit_logs         DROP COLUMN IF EXISTS organization_id;
ALTER TABLE webhook_deliveries DROP COLUMN IF EXISTS organization_id;
ALTER TABLE webhook_endpoints  DROP COLUMN IF EXISTS organization_id;
ALTER TABLE api_tokens         DROP COLUMN IF EXISTS organization_id;
