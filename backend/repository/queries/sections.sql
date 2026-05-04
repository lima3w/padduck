-- name: CreateSection :one
INSERT INTO sections (name, description, created_by)
VALUES ($1, $2, $3)
RETURNING id, name, description, created_by, created_at, updated_at;

-- name: GetSectionByID :one
SELECT id, name, description, created_by, created_at, updated_at
FROM sections
WHERE id = $1;

-- name: ListSectionsByCreator :many
SELECT id, name, description, created_by, created_at, updated_at
FROM sections
WHERE created_by = $1
ORDER BY created_at DESC;

-- name: ListAllSections :many
SELECT id, name, description, created_by, created_at, updated_at
FROM sections
ORDER BY created_at DESC;

-- name: UpdateSection :one
UPDATE sections
SET name = $2, description = $3
WHERE id = $1
RETURNING id, name, description, created_by, created_at, updated_at;

-- name: DeleteSection :exec
DELETE FROM sections
WHERE id = $1;
