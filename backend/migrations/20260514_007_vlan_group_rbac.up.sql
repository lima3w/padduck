-- +migrate Up

-- Backfill ipam:vlan_group:list and ipam:vlan_group:read into every role
-- that already has any VLAN permission.
INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:vlan_group:list'
FROM role_permissions
WHERE permission LIKE 'ipam:vlan%'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'ipam:vlan_group:list'
        AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:vlan_group:read'
FROM role_permissions
WHERE permission LIKE 'ipam:vlan%'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'ipam:vlan_group:read'
        AND rp2.resource_type IS NULL
  );

-- Backfill write/delete to roles that already have VLAN write/delete.
INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:vlan_group:write'
FROM role_permissions
WHERE permission = 'ipam:vlan:write'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'ipam:vlan_group:write'
        AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:vlan_group:delete'
FROM role_permissions
WHERE permission = 'ipam:vlan:delete'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'ipam:vlan_group:delete'
        AND rp2.resource_type IS NULL
  );
