-- +migrate Up

-- Repair location write/delete grants for installations where the original
-- backfill used the wrong device permission prefix.
INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:location:write'
FROM role_permissions
WHERE permission IN (
    'ipam:section:write',
    'ipam:subnet:write',
    'ipam:vrf:write',
    'ipam:vlan:write',
    'devices:write',
    'devices:admin'
)
  AND resource_type IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp2
    WHERE rp2.role_id = role_permissions.role_id
      AND rp2.permission = 'ipam:location:write'
      AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:location:delete'
FROM role_permissions
WHERE permission IN (
    'ipam:section:delete',
    'ipam:subnet:delete',
    'ipam:vrf:delete',
    'ipam:vlan:delete',
    'devices:delete',
    'devices:admin'
)
  AND resource_type IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp2
    WHERE rp2.role_id = role_permissions.role_id
      AND rp2.permission = 'ipam:location:delete'
      AND rp2.resource_type IS NULL
  );
