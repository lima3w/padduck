-- name: CreateIPAddress :one
INSERT INTO ip_addresses (subnet_id, address, hostname, status, assigned_to)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, subnet_id, address, hostname, status, assigned_to, created_at, updated_at;

-- name: GetIPAddressByID :one
SELECT id, subnet_id, address, hostname, status, assigned_to, created_at, updated_at
FROM ip_addresses
WHERE id = $1;

-- name: GetIPAddressByAddress :one
SELECT id, subnet_id, address, hostname, status, assigned_to, created_at, updated_at
FROM ip_addresses
WHERE address = $1;

-- name: ListIPAddressesBySubnet :many
SELECT id, subnet_id, address, hostname, status, assigned_to, created_at, updated_at
FROM ip_addresses
WHERE subnet_id = $1
ORDER BY address;

-- name: ListAvailableIPsBySubnet :many
SELECT id, subnet_id, address, hostname, status, assigned_to, created_at, updated_at
FROM ip_addresses
WHERE subnet_id = $1 AND status = 'available'
ORDER BY address;

-- name: UpdateIPAddressStatus :one
UPDATE ip_addresses
SET status = $2, assigned_to = $3
WHERE id = $1
RETURNING id, subnet_id, address, hostname, status, assigned_to, created_at, updated_at;

-- name: DeleteIPAddress :exec
DELETE FROM ip_addresses
WHERE id = $1;
