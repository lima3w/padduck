-- +migrate Up

CREATE TABLE firewall_zones (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    color TEXT NOT NULL DEFAULT '#2563eb',
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT firewall_zones_status_check CHECK (status IN ('active', 'planned', 'disabled', 'retired')),
    CONSTRAINT firewall_zones_color_check CHECK (color ~ '^#[0-9A-Fa-f]{6}$')
);

CREATE TABLE firewall_zone_mappings (
    id BIGSERIAL PRIMARY KEY,
    zone_id BIGINT NOT NULL REFERENCES firewall_zones(id) ON DELETE CASCADE,
    object_type TEXT NOT NULL,
    object_id BIGINT,
    cidr CIDR,
    direction TEXT NOT NULL DEFAULT 'both',
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT firewall_zone_mappings_target_check CHECK (object_id IS NOT NULL OR cidr IS NOT NULL),
    CONSTRAINT firewall_zone_mappings_object_check CHECK (object_type IN ('section', 'subnet', 'ip_address', 'device', 'rack', 'location', 'vlan', 'vrf', 'nat_rule', 'dhcp_server', 'dhcp_lease', 'physical_circuit', 'logical_circuit', 'cidr')),
    CONSTRAINT firewall_zone_mappings_direction_check CHECK (direction IN ('inbound', 'outbound', 'both')),
    CONSTRAINT firewall_zone_mappings_status_check CHECK (status IN ('active', 'planned', 'disabled', 'retired'))
);

CREATE INDEX idx_firewall_zone_mappings_zone_id ON firewall_zone_mappings(zone_id);
CREATE INDEX idx_firewall_zone_mappings_object ON firewall_zone_mappings(object_type, object_id);
CREATE INDEX idx_firewall_zone_mappings_cidr ON firewall_zone_mappings(cidr);

INSERT INTO configs (key, value) VALUES ('feature_firewall_enabled', 'true') ON CONFLICT (key) DO NOTHING;

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, permission_to_add
FROM role_permissions
CROSS JOIN (VALUES ('ipam:firewall:list'), ('ipam:firewall:read')) AS p(permission_to_add)
WHERE role_permissions.permission LIKE 'ipam:%'
  AND role_permissions.resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = p.permission_to_add
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, permission_to_add
FROM role_permissions
CROSS JOIN (VALUES ('ipam:firewall:write'), ('ipam:firewall:delete')) AS p(permission_to_add)
WHERE role_permissions.permission = 'ipam:subnet:write'
  AND role_permissions.resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = p.permission_to_add
  );
