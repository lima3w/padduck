-- +migrate Down

ALTER TABLE subnets DROP COLUMN IF EXISTS auto_reserve_last;
ALTER TABLE subnets DROP COLUMN IF EXISTS auto_reserve_first;
ALTER TABLE subnets DROP COLUMN IF EXISTS gateway;
