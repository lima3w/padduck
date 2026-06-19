-- +migrate Down
ALTER TABLE api_tokens DROP COLUMN IF EXISTS impersonated_org_id;
ALTER TABLE users DROP COLUMN IF EXISTS is_platform_admin;
