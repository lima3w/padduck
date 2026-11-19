-- +migrate Up
-- api_tokens.organization_id was added nullable and never populated by new
-- tokens, so per-org token counts (quotas) and org-scoped listings missed them.
UPDATE api_tokens t SET organization_id = u.organization_id
  FROM users u WHERE t.user_id = u.id AND t.organization_id IS NULL;
UPDATE api_tokens SET organization_id = 1 WHERE organization_id IS NULL;
ALTER TABLE api_tokens ALTER COLUMN organization_id SET DEFAULT 1;
ALTER TABLE api_tokens ALTER COLUMN organization_id SET NOT NULL;
