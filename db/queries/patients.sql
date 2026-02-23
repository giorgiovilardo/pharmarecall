-- name: CreatePatient :one
INSERT INTO patients (pharmacy_id, first_name, last_name, phone, email, delivery_address, fulfillment, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, pharmacy_id, first_name, last_name, phone, email, delivery_address, fulfillment, notes, consensus, consensus_date, created_at, updated_at;

-- name: ListPatientsByPharmacy :many
SELECT id, first_name, last_name, phone, email, consensus
FROM patients
WHERE pharmacy_id = sqlc.arg(pharmacy_id)::BIGINT
ORDER BY last_name, first_name;

-- name: GetPatientByID :one
SELECT id, pharmacy_id, first_name, last_name, phone, email, delivery_address, fulfillment, notes, consensus, consensus_date, created_at, updated_at
FROM patients
WHERE id = $1;

-- name: UpdatePatient :exec
UPDATE patients
SET first_name = $2, last_name = $3, phone = $4, email = $5, delivery_address = $6, fulfillment = $7, notes = $8, updated_at = now()
WHERE id = $1;

-- name: SetPatientConsensus :exec
UPDATE patients
SET consensus = true, consensus_date = now(), updated_at = now()
WHERE id = $1;
