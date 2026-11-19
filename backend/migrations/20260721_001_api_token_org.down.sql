-- +migrate Down
ALTER TABLE api_tokens ALTER COLUMN organization_id DROP NOT NULL;
ALTER TABLE api_tokens ALTER COLUMN organization_id DROP DEFAULT;
