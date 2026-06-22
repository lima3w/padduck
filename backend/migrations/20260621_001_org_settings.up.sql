-- +migrate Up
CREATE TABLE organization_settings (
  organization_id      BIGINT       PRIMARY KEY REFERENCES organizations(id) ON DELETE CASCADE,
  max_subnets          INT,
  max_ip_addresses     INT,
  max_users            INT,
  max_webhooks         INT,
  max_api_tokens       INT,
  audit_retention_days INT,
  smtp_host            TEXT,
  smtp_port            INT,
  smtp_from            TEXT,
  created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
