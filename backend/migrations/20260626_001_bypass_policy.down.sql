-- +migrate Down
ALTER TABLE api_tokens DROP COLUMN IF EXISTS bypass_policy;
