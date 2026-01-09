-- IoT Platform - Device Manager Queries
-- SQL queries with sqlc annotations for type-safe code generation

-- name: CreateDevice :one
INSERT INTO devices (
    id,
    name,
    type,
    status,
    created_at,
    last_seen,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetDevice :one
SELECT * FROM devices
WHERE id = $1;

-- name: ListDevices :many
SELECT * FROM devices
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountDevices :one
SELECT COUNT(*) FROM devices;

-- name: UpdateDevice :one
UPDATE devices
SET
    name = sqlc.arg(name),
    status = sqlc.arg(status),
    metadata = sqlc.arg(metadata),
    last_seen = sqlc.arg(last_seen)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteDevice :exec
DELETE FROM devices
WHERE id = $1;

-- name: ListDevicesByType :many
SELECT * FROM devices
WHERE type = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListDevicesByStatus :many
SELECT * FROM devices
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountDevicesByStatus :one
SELECT COUNT(*) FROM devices
WHERE status = $1;
