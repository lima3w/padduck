-- +migrate Up

CREATE TABLE IF NOT EXISTS vrfs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    route_distinguisher VARCHAR(255),
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS vlans (
    id BIGSERIAL PRIMARY KEY,
    vrf_id BIGINT REFERENCES vrfs(id) ON DELETE SET NULL,
    vlan_id INTEGER NOT NULL CHECK (vlan_id >= 1 AND vlan_id <= 4094),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(vlan_id)
);

CREATE INDEX idx_vlans_vrf_id ON vlans(vrf_id);
CREATE INDEX idx_vlans_vlan_id ON vlans(vlan_id);
