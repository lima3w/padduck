-- +migrate Down
DELETE FROM configs WHERE key = 'audit_log_retention_days';
