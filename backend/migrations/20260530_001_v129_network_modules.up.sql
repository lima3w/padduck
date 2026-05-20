-- +migrate Up

ALTER TABLE locations
  ADD COLUMN IF NOT EXISTS city TEXT,
  ADD COLUMN IF NOT EXISTS region TEXT,
  ADD COLUMN IF NOT EXISTS country TEXT,
  ADD COLUMN IF NOT EXISTS facility_code TEXT,
  ADD COLUMN IF NOT EXISTS time_zone TEXT,
  ADD COLUMN IF NOT EXISTS contact_name TEXT,
  ADD COLUMN IF NOT EXISTS contact_email TEXT,
  ADD COLUMN IF NOT EXISTS contact_phone TEXT,
  ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active';

CREATE TABLE nat_rules (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL DEFAULT 'static',
    internal_cidr TEXT NOT NULL,
    external_cidr TEXT NOT NULL,
    protocol TEXT NOT NULL DEFAULT 'any',
    internal_port INTEGER,
    external_port INTEGER,
    device_id BIGINT REFERENCES devices(id) ON DELETE SET NULL,
    customer_id BIGINT REFERENCES customers(id) ON DELETE SET NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT nat_rules_protocol_check CHECK (protocol IN ('any', 'tcp', 'udp', 'icmp')),
    CONSTRAINT nat_rules_type_check CHECK (type IN ('static', 'dynamic', 'pat')),
    CONSTRAINT nat_rules_status_check CHECK (status IN ('active', 'disabled', 'planned', 'retired')),
    CONSTRAINT nat_rules_internal_port_check CHECK (internal_port IS NULL OR internal_port BETWEEN 1 AND 65535),
    CONSTRAINT nat_rules_external_port_check CHECK (external_port IS NULL OR external_port BETWEEN 1 AND 65535)
);

CREATE TABLE dhcp_servers (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    address TEXT NOT NULL,
    vendor TEXT NOT NULL DEFAULT '',
    version TEXT NOT NULL DEFAULT '',
    location_id BIGINT REFERENCES locations(id) ON DELETE SET NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT dhcp_servers_status_check CHECK (status IN ('active', 'disabled', 'planned', 'retired'))
);

CREATE TABLE dhcp_leases (
    id BIGSERIAL PRIMARY KEY,
    server_id BIGINT NOT NULL REFERENCES dhcp_servers(id) ON DELETE CASCADE,
    ip_address INET NOT NULL,
    mac_address TEXT NOT NULL,
    hostname TEXT NOT NULL DEFAULT '',
    subnet_id BIGINT REFERENCES subnets(id) ON DELETE SET NULL,
    ip_id BIGINT REFERENCES ip_addresses(id) ON DELETE SET NULL,
    customer_id BIGINT REFERENCES customers(id) ON DELETE SET NULL,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    state TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT dhcp_leases_state_check CHECK (state IN ('active', 'expired', 'reserved', 'declined', 'released')),
    CONSTRAINT dhcp_leases_unique_server_ip UNIQUE (server_id, ip_address)
);

CREATE TABLE circuit_providers (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    account_no TEXT NOT NULL DEFAULT '',
    support_email TEXT NOT NULL DEFAULT '',
    support_phone TEXT NOT NULL DEFAULT '',
    portal_url TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE physical_circuits (
    id BIGSERIAL PRIMARY KEY,
    provider_id BIGINT NOT NULL REFERENCES circuit_providers(id) ON DELETE RESTRICT,
    circuit_id TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL DEFAULT 'ethernet',
    status TEXT NOT NULL DEFAULT 'active',
    bandwidth_mbps INTEGER,
    location_a_id BIGINT REFERENCES locations(id) ON DELETE SET NULL,
    location_b_id BIGINT REFERENCES locations(id) ON DELETE SET NULL,
    customer_id BIGINT REFERENCES customers(id) ON DELETE SET NULL,
    install_date DATE,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT physical_circuits_unique_provider_circuit UNIQUE (provider_id, circuit_id),
    CONSTRAINT physical_circuits_status_check CHECK (status IN ('active', 'planned', 'down', 'retired')),
    CONSTRAINT physical_circuits_bandwidth_check CHECK (bandwidth_mbps IS NULL OR bandwidth_mbps > 0)
);

CREATE TABLE logical_circuits (
    id BIGSERIAL PRIMARY KEY,
    physical_circuit_id BIGINT REFERENCES physical_circuits(id) ON DELETE SET NULL,
    name TEXT NOT NULL UNIQUE,
    service_id TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL DEFAULT 'l2vpn',
    status TEXT NOT NULL DEFAULT 'active',
    vlan_id BIGINT REFERENCES vlans(id) ON DELETE SET NULL,
    vrf_id BIGINT REFERENCES vrfs(id) ON DELETE SET NULL,
    customer_id BIGINT REFERENCES customers(id) ON DELETE SET NULL,
    bandwidth_mbps INTEGER,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT logical_circuits_status_check CHECK (status IN ('active', 'planned', 'down', 'retired')),
    CONSTRAINT logical_circuits_bandwidth_check CHECK (bandwidth_mbps IS NULL OR bandwidth_mbps > 0)
);

CREATE TABLE customer_associations (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    object_type TEXT NOT NULL,
    object_id BIGINT NOT NULL,
    relationship TEXT NOT NULL DEFAULT 'owner',
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT customer_associations_object_check CHECK (object_type IN ('section', 'subnet', 'ip_address', 'device', 'rack', 'location', 'vlan', 'vrf', 'nat_rule', 'dhcp_server', 'dhcp_lease', 'physical_circuit', 'logical_circuit')),
    CONSTRAINT customer_associations_relationship_check CHECK (relationship IN ('owner', 'consumer', 'billing', 'technical', 'stakeholder')),
    CONSTRAINT customer_associations_unique UNIQUE (customer_id, object_type, object_id, relationship)
);

CREATE INDEX idx_nat_rules_customer_id ON nat_rules(customer_id);
CREATE INDEX idx_nat_rules_device_id ON nat_rules(device_id);
CREATE INDEX idx_dhcp_servers_location_id ON dhcp_servers(location_id);
CREATE INDEX idx_dhcp_leases_server_id ON dhcp_leases(server_id);
CREATE INDEX idx_dhcp_leases_customer_id ON dhcp_leases(customer_id);
CREATE INDEX idx_physical_circuits_provider_id ON physical_circuits(provider_id);
CREATE INDEX idx_physical_circuits_customer_id ON physical_circuits(customer_id);
CREATE INDEX idx_logical_circuits_customer_id ON logical_circuits(customer_id);
CREATE INDEX idx_customer_associations_customer_id ON customer_associations(customer_id);
CREATE INDEX idx_customer_associations_object ON customer_associations(object_type, object_id);

INSERT INTO configs (key, value) VALUES
  ('feature_nat_enabled', 'true'),
  ('feature_dhcp_enabled', 'true'),
  ('feature_circuits_enabled', 'true')
ON CONFLICT (key) DO NOTHING;

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, permission_to_add
FROM role_permissions
CROSS JOIN (VALUES
  ('ipam:nat:list'), ('ipam:nat:read'), ('ipam:dhcp:list'), ('ipam:dhcp:read'), ('ipam:circuit:list'), ('ipam:circuit:read')
) AS p(permission_to_add)
WHERE role_permissions.permission LIKE 'ipam:%'
  AND role_permissions.resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = p.permission_to_add
        AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, permission_to_add
FROM role_permissions
CROSS JOIN (VALUES
  ('ipam:nat:write'), ('ipam:nat:delete'), ('ipam:dhcp:write'), ('ipam:dhcp:delete'), ('ipam:circuit:write'), ('ipam:circuit:delete')
) AS p(permission_to_add)
WHERE role_permissions.permission = 'ipam:subnet:write'
  AND role_permissions.resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = p.permission_to_add
        AND rp2.resource_type IS NULL
  );
