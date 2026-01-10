-- IoT Platform - User Service Queries
-- SQL queries with sqlc annotations for type-safe code generation

-- name: CreateUser :one
INSERT INTO users (
    id,
    email,
    password_hash,
    name,
    role,
    created_at,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetPasswordHash :one
SELECT password_hash FROM users
WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET
    name = sqlc.arg(name),
    role = sqlc.arg(role),
    is_active = sqlc.arg(is_active)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateLastLogin :exec
UPDATE users
SET last_login = $2
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListUsersByRole :many
SELECT * FROM users
WHERE role = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;
