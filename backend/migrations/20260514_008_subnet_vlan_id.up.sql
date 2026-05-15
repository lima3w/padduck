-- +migrate Up

ALTER TABLE subnets ADD COLUMN vlan_id BIGINT REFERENCES vlans(id) ON DELETE SET NULL;

CREATE INDEX idx_subnets_vlan_id ON subnets(vlan_id);
