-- +migrate Down
ALTER TABLE subnets DROP COLUMN IF EXISTS is_container;
ALTER TABLE subnets DROP COLUMN IF EXISTS parent_subnet_id;
