-- name: CreatePharmacy :one
INSERT INTO pharmacies (name, address, phone, email)
VALUES ($1, $2, $3, $4)
RETURNING id, name, address, phone, email, created_at, updated_at;

-- name: ListPharmacies :many
SELECT
    p.id,
    p.name,
    p.address,
    p.phone,
    p.email,
    p.created_at,
    p.updated_at,
    COUNT(DISTINCT u.id)::BIGINT AS personnel_count
FROM pharmacies p
LEFT JOIN users u ON u.pharmacy_id = p.id
GROUP BY p.id
ORDER BY p.name;

-- name: GetPharmacyByID :one
SELECT id, name, address, phone, email, created_at, updated_at
FROM pharmacies
WHERE id = $1;

-- name: UpdatePharmacy :exec
UPDATE pharmacies
SET name = $2, address = $3, phone = $4, email = $5, updated_at = now()
WHERE id = $1;
