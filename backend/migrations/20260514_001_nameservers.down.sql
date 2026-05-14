-- +migrate Down

ALTER TABLE subnets DROP COLUMN IF EXISTS nameserver_id;
DROP TABLE IF EXISTS nameservers;
