-- +migrate Down
ALTER TABLE users DROP COLUMN IF EXISTS organization_id;
DROP TABLE IF EXISTS organizations;
