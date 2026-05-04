-- name: CreateUser :one
INSERT INTO users (username, email)
VALUES ($1, $2)
RETURNING id, username, email, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, username, email, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, username, email, created_at, updated_at
FROM users
WHERE username = $1;

-- name: ListUsers :many
SELECT id, username, email, created_at, updated_at
FROM users
ORDER BY created_at DESC;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
