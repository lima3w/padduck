-- name: CreateSubnet :one
INSERT INTO subnets (section_id, network_address, prefix_length, description)
VALUES ($1, $2, $3, $4)
RETURNING id, section_id, network_address, prefix_length, description, created_at, updated_at;

-- name: GetSubnetByID :one
SELECT id, section_id, network_address, prefix_length, description, created_at, updated_at
FROM subnets
WHERE id = $1;

-- name: ListSubnetsBySection :many
SELECT id, section_id, network_address, prefix_length, description, created_at, updated_at
FROM subnets
WHERE section_id = $1
ORDER BY network_address;

-- name: DeleteSubnet :exec
DELETE FROM subnets
WHERE id = $1;
