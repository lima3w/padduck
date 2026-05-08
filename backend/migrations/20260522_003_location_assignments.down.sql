-- +migrate Down

ALTER TABLE devices DROP COLUMN IF EXISTS location_id;
ALTER TABLE subnets DROP COLUMN IF EXISTS location_id;
