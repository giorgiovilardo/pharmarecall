-- name: GetUserByEmail :one
SELECT id, email, password_hash, name, role, pharmacy_id, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash, name, role, pharmacy_id, created_at, updated_at
FROM users
WHERE id = $1;
