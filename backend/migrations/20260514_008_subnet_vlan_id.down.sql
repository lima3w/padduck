-- +migrate Down

ALTER TABLE subnets DROP COLUMN IF EXISTS vlan_id;
