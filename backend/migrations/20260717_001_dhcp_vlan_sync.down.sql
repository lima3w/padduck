-- +migrate Down

DROP INDEX IF EXISTS idx_dhcp_leases_ip_id;
DROP INDEX IF EXISTS idx_dhcp_leases_subnet_id;
DROP INDEX IF EXISTS idx_dhcp_leases_vlan_id;
DROP INDEX IF EXISTS idx_dhcp_servers_vlan_id;

ALTER TABLE dhcp_leases DROP COLUMN IF EXISTS vlan_id;

ALTER TABLE dhcp_servers
    DROP COLUMN IF EXISTS last_sync_count,
    DROP COLUMN IF EXISTS last_sync_error,
    DROP COLUMN IF EXISTS last_synced_at,
    DROP COLUMN IF EXISTS auto_sync,
    DROP COLUMN IF EXISTS vlan_id;
