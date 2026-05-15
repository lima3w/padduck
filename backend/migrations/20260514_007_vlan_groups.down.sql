-- +migrate Down

ALTER TABLE vlans DROP COLUMN IF EXISTS group_id;
DROP TABLE IF EXISTS vlan_groups;
