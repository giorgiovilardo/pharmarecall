-- name: GetUserByEmail :one
SELECT id, email, password_hash, name, role, pharmacy_id, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash, name, role, pharmacy_id, created_at, updated_at
FROM users
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, name, role, pharmacy_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, email, password_hash, name, role, pharmacy_id, created_at, updated_at;

-- name: ListUsersByPharmacy :many
SELECT id, email, password_hash, name, role, pharmacy_id, created_at, updated_at
FROM users
WHERE pharmacy_id = sqlc.arg(pharmacy_id)::BIGINT
ORDER BY name;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = now()
WHERE id = $1;
