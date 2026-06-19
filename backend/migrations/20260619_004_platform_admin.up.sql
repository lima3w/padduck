-- +migrate Up
ALTER TABLE users ADD COLUMN is_platform_admin BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE api_tokens ADD COLUMN impersonated_org_id BIGINT REFERENCES organizations(id);
