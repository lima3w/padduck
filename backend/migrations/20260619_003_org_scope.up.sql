-- +migrate Up
ALTER TABLE api_tokens         ADD COLUMN organization_id BIGINT REFERENCES organizations(id);
ALTER TABLE webhook_endpoints  ADD COLUMN organization_id BIGINT REFERENCES organizations(id);
ALTER TABLE webhook_deliveries ADD COLUMN organization_id BIGINT REFERENCES organizations(id);
ALTER TABLE audit_logs         ADD COLUMN organization_id BIGINT REFERENCES organizations(id);
ALTER TABLE scheduled_reports  ADD COLUMN organization_id BIGINT REFERENCES organizations(id);

-- Seed existing rows with the default org
UPDATE api_tokens         SET organization_id = 1 WHERE organization_id IS NULL;
UPDATE webhook_endpoints  SET organization_id = 1 WHERE organization_id IS NULL;
UPDATE webhook_deliveries SET organization_id = 1 WHERE organization_id IS NULL;
UPDATE audit_logs         SET organization_id = 1 WHERE organization_id IS NULL;
UPDATE scheduled_reports  SET organization_id = 1 WHERE organization_id IS NULL;

CREATE INDEX audit_logs_org_created   ON audit_logs (organization_id, created_at DESC);
CREATE INDEX webhooks_org             ON webhook_endpoints (organization_id);
CREATE INDEX api_tokens_org           ON api_tokens (organization_id);
CREATE INDEX scheduled_reports_org    ON scheduled_reports (organization_id);
