-- +migrate Up
ALTER TABLE subnets ADD COLUMN alert_threshold_pct INTEGER CHECK (alert_threshold_pct BETWEEN 1 AND 100);
ALTER TABLE subnets ADD COLUMN alert_email_override VARCHAR(255);
CREATE TABLE alert_cooldowns (
    id BIGSERIAL PRIMARY KEY,
    subnet_id BIGINT NOT NULL REFERENCES subnets(id) ON DELETE CASCADE UNIQUE,
    alerted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    alerted_pct NUMERIC(5,2) NOT NULL
);
