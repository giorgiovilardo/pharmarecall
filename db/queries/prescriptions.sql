-- name: CreatePrescription :one
INSERT INTO prescriptions (patient_id, medication_name, units_per_box, daily_consumption, box_start_date)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, patient_id, medication_name, units_per_box, daily_consumption, box_start_date, created_at, updated_at;

-- name: ListPrescriptionsByPatient :many
SELECT id, patient_id, medication_name, units_per_box, daily_consumption, box_start_date, created_at, updated_at
FROM prescriptions
WHERE patient_id = $1
ORDER BY medication_name;

-- name: GetPrescriptionByID :one
SELECT id, patient_id, medication_name, units_per_box, daily_consumption, box_start_date, created_at, updated_at
FROM prescriptions
WHERE id = $1;

-- name: UpdatePrescription :exec
UPDATE prescriptions
SET medication_name = $2, units_per_box = $3, daily_consumption = $4, box_start_date = $5, updated_at = now()
WHERE id = $1;

-- name: InsertRefillHistory :exec
INSERT INTO refill_history (prescription_id, box_start_date, box_end_date)
VALUES ($1, $2, $3);
