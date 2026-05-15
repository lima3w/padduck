-- +migrate Down

ALTER TABLE vlans DROP COLUMN IF EXISTS domain_id;
DROP TABLE IF EXISTS vlan_domains;
