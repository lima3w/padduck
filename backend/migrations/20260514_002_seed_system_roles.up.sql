-- +migrate Up

-- Insert system roles
INSERT INTO roles (name, description, is_system) VALUES
    ('admin',    'Full system access including user management and audit logs', TRUE),
    ('operator', 'Full IPAM access with limited user visibility',               TRUE),
    ('viewer',   'Read-only access to IPAM resources and user list',            TRUE)
ON CONFLICT DO NOTHING;

-- Admin: all 24 permissions
WITH r AS (SELECT id FROM roles WHERE name = 'admin')
INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.permission
FROM r, (VALUES
    ('ipam:section:list'),
    ('ipam:section:read'),
    ('ipam:section:write'),
    ('ipam:section:delete'),
    ('ipam:subnet:list'),
    ('ipam:subnet:read'),
    ('ipam:subnet:write'),
    ('ipam:subnet:delete'),
    ('ipam:ip_address:list'),
    ('ipam:ip_address:read'),
    ('ipam:ip_address:assign'),
    ('ipam:ip_address:release'),
    ('ipam:vrf:list'),
    ('ipam:vrf:read'),
    ('ipam:vrf:write'),
    ('ipam:vrf:delete'),
    ('ipam:vlan:list'),
    ('ipam:vlan:read'),
    ('ipam:vlan:write'),
    ('ipam:vlan:delete'),
    ('auth:user:list'),
    ('auth:user:read'),
    ('auth:user:write'),
    ('auth:audit:read')
) AS p(permission)
ON CONFLICT DO NOTHING;

-- Operator: all ipam: permissions + auth:user:list + auth:user:read (no auth:user:write, no auth:audit:read)
WITH r AS (SELECT id FROM roles WHERE name = 'operator')
INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.permission
FROM r, (VALUES
    ('ipam:section:list'),
    ('ipam:section:read'),
    ('ipam:section:write'),
    ('ipam:section:delete'),
    ('ipam:subnet:list'),
    ('ipam:subnet:read'),
    ('ipam:subnet:write'),
    ('ipam:subnet:delete'),
    ('ipam:ip_address:list'),
    ('ipam:ip_address:read'),
    ('ipam:ip_address:assign'),
    ('ipam:ip_address:release'),
    ('ipam:vrf:list'),
    ('ipam:vrf:read'),
    ('ipam:vrf:write'),
    ('ipam:vrf:delete'),
    ('ipam:vlan:list'),
    ('ipam:vlan:read'),
    ('ipam:vlan:write'),
    ('ipam:vlan:delete'),
    ('auth:user:list'),
    ('auth:user:read')
) AS p(permission)
ON CONFLICT DO NOTHING;

-- Viewer: list+read only
WITH r AS (SELECT id FROM roles WHERE name = 'viewer')
INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.permission
FROM r, (VALUES
    ('ipam:section:list'),
    ('ipam:section:read'),
    ('ipam:subnet:list'),
    ('ipam:subnet:read'),
    ('ipam:ip_address:list'),
    ('ipam:ip_address:read'),
    ('ipam:vrf:list'),
    ('ipam:vrf:read'),
    ('ipam:vlan:list'),
    ('ipam:vlan:read'),
    ('auth:user:list'),
    ('auth:user:read')
) AS p(permission)
ON CONFLICT DO NOTHING;
