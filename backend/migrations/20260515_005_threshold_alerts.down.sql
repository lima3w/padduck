-- +migrate Down
DROP TABLE IF EXISTS alert_cooldowns;
ALTER TABLE subnets DROP COLUMN IF EXISTS alert_email_override;
ALTER TABLE subnets DROP COLUMN IF EXISTS alert_threshold_pct;
