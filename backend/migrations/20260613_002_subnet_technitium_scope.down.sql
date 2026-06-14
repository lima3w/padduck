-- +migrate Down
ALTER TABLE subnets DROP COLUMN IF EXISTS technitium_scope_name;
