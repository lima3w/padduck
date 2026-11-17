-- +migrate Up

-- Optional VLAN association for DHCP servers and leases (#36). A server may be
-- scoped to one VLAN; a lease's VLAN is derived from its subnet (or, failing
-- that, its server). Both are nullable: not every DHCP server is VLAN-scoped.
ALTER TABLE dhcp_servers
    ADD COLUMN vlan_id BIGINT REFERENCES vlans(id) ON DELETE SET NULL,
    ADD COLUMN auto_sync BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN last_synced_at TIMESTAMPTZ,
    ADD COLUMN last_sync_error TEXT NOT NULL DEFAULT '',
    ADD COLUMN last_sync_count INTEGER;

ALTER TABLE dhcp_leases
    ADD COLUMN vlan_id BIGINT REFERENCES vlans(id) ON DELETE SET NULL;

CREATE INDEX idx_dhcp_servers_vlan_id ON dhcp_servers(vlan_id);
CREATE INDEX idx_dhcp_leases_vlan_id ON dhcp_leases(vlan_id);
CREATE INDEX idx_dhcp_leases_subnet_id ON dhcp_leases(subnet_id);
CREATE INDEX idx_dhcp_leases_ip_id ON dhcp_leases(ip_id);
