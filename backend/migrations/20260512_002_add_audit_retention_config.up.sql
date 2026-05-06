-- +migrate Up
INSERT INTO configs (key, value) VALUES ('audit_log_retention_days', '90')
ON CONFLICT (key) DO NOTHING;
