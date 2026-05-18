-- +migrate Up

-- Device permissions were added after the original system roles. Grant read to
-- roles with IPAM access, write/delete to subnet managers, and admin to admins.
INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'devices:read'
FROM role_permissions
WHERE permission LIKE 'ipam:%'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'devices:read'
        AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'devices:write'
FROM role_permissions
WHERE permission = 'ipam:subnet:write'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'devices:write'
        AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'devices:delete'
FROM role_permissions
WHERE permission = 'ipam:subnet:delete'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'devices:delete'
        AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'devices:admin'
FROM role_permissions
WHERE permission = 'auth:user:write'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'devices:admin'
        AND rp2.resource_type IS NULL
  );

-- Admin operation permissions are used by report/admin pages added after the
-- original role seed. Grant them to roles that already manage users.
INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'auth:admin:read'
FROM role_permissions
WHERE permission = 'auth:user:write'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'auth:admin:read'
        AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'auth:admin:write'
FROM role_permissions
WHERE permission = 'auth:user:write'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'auth:admin:write'
        AND rp2.resource_type IS NULL
  );

-- Request workflow permissions: operators can submit requests; admins can
-- submit and review/manage requests.
INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:subnet_request:submit'
FROM role_permissions
WHERE permission = 'ipam:subnet:write'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'ipam:subnet_request:submit'
        AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:subnet_request:review'
FROM role_permissions
WHERE permission = 'auth:user:write'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'ipam:subnet_request:review'
        AND rp2.resource_type IS NULL
  );

-- Customers: all IPAM roles can list/read; subnet managers can mutate.
INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:customer:list'
FROM role_permissions
WHERE permission LIKE 'ipam:%'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'ipam:customer:list'
        AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:customer:read'
FROM role_permissions
WHERE permission LIKE 'ipam:%'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'ipam:customer:read'
        AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:customer:write'
FROM role_permissions
WHERE permission = 'ipam:subnet:write'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'ipam:customer:write'
        AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:customer:delete'
FROM role_permissions
WHERE permission = 'ipam:subnet:delete'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'ipam:customer:delete'
        AND rp2.resource_type IS NULL
  );

-- Autonomous systems: all IPAM roles can list/read; subnet managers can mutate.
INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:autonomous_system:list'
FROM role_permissions
WHERE permission LIKE 'ipam:%'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'ipam:autonomous_system:list'
        AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:autonomous_system:read'
FROM role_permissions
WHERE permission LIKE 'ipam:%'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'ipam:autonomous_system:read'
        AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:autonomous_system:write'
FROM role_permissions
WHERE permission = 'ipam:subnet:write'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'ipam:autonomous_system:write'
        AND rp2.resource_type IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT DISTINCT role_id, 'ipam:autonomous_system:delete'
FROM role_permissions
WHERE permission = 'ipam:subnet:delete'
  AND resource_type IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp2
      WHERE rp2.role_id = role_permissions.role_id
        AND rp2.permission = 'ipam:autonomous_system:delete'
        AND rp2.resource_type IS NULL
  );
