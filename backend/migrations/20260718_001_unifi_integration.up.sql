-- +migrate Up

-- Links between Padduck objects and their UniFi counterparts (#38). A VLAN
-- link carries the UniFi network name; a device link carries live status and
-- firmware. Kept separate from vlans/devices so UniFi stays an optional layer.
CREATE TABLE unifi_links (
    id BIGSERIAL PRIMARY KEY,
    vlan_id BIGINT REFERENCES vlans(id) ON DELETE CASCADE,
    device_id BIGINT REFERENCES devices(id) ON DELETE CASCADE,
    unifi_id TEXT NOT NULL,
    unifi_name TEXT NOT NULL DEFAULT '',
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_synced_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unifi_links_one_target CHECK ((vlan_id IS NOT NULL) <> (device_id IS NOT NULL))
);
CREATE UNIQUE INDEX uq_unifi_links_vlan ON unifi_links(vlan_id) WHERE vlan_id IS NOT NULL;
CREATE UNIQUE INDEX uq_unifi_links_device ON unifi_links(device_id) WHERE device_id IS NOT NULL;
CREATE UNIQUE INDEX uq_unifi_links_vlan_uid ON unifi_links(unifi_id) WHERE vlan_id IS NOT NULL;
CREATE UNIQUE INDEX uq_unifi_links_device_uid ON unifi_links(unifi_id) WHERE device_id IS NOT NULL;

-- Differences between an existing IP record and what UniFi reports. The sync
-- never overwrites a populated field; it records the difference here for an
-- admin to accept (apply UniFi's value) or dismiss (keep Padduck's).
CREATE TABLE unifi_conflicts (
    id BIGSERIAL PRIMARY KEY,
    ip_id BIGINT NOT NULL REFERENCES ip_addresses(id) ON DELETE CASCADE,
    field TEXT NOT NULL,
    padduck_value TEXT NOT NULL DEFAULT '',
    unifi_value TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'open',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMPTZ,
    CONSTRAINT unifi_conflicts_field_check CHECK (field IN ('hostname', 'mac_address')),
    CONSTRAINT unifi_conflicts_status_check CHECK (status IN ('open', 'accepted', 'dismissed'))
);
CREATE UNIQUE INDEX uq_unifi_conflicts_open ON unifi_conflicts(ip_id, field) WHERE status = 'open';
CREATE INDEX idx_unifi_conflicts_status ON unifi_conflicts(status);
