-- name: CreateOrder :one
INSERT INTO orders (prescription_id, cycle_start_date, estimated_depletion_date, status)
VALUES ($1, $2, $3, $4)
RETURNING id, prescription_id, cycle_start_date, estimated_depletion_date, status, created_at, updated_at;

-- name: GetActiveOrderByPrescription :one
SELECT id, prescription_id, cycle_start_date, estimated_depletion_date, status, created_at, updated_at
FROM orders
WHERE prescription_id = sqlc.arg(prescription_id)::BIGINT
  AND status IN ('pending', 'prepared')
  AND cycle_start_date = sqlc.arg(cycle_start_date)::DATE
LIMIT 1;

-- name: GetOrderByID :one
SELECT id, prescription_id, cycle_start_date, estimated_depletion_date, status, created_at, updated_at
FROM orders
WHERE id = $1;

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = $2, updated_at = now()
WHERE id = $1;

-- name: ListDashboardOrders :many
SELECT
    o.id AS order_id,
    o.prescription_id,
    o.cycle_start_date,
    o.estimated_depletion_date,
    o.status AS order_status,
    p.medication_name,
    p.units_per_box,
    p.daily_consumption,
    p.box_start_date,
    pat.id AS patient_id,
    pat.first_name,
    pat.last_name,
    pat.fulfillment,
    pat.delivery_address,
    pat.phone,
    pat.email
FROM orders o
JOIN prescriptions p ON o.prescription_id = p.id
JOIN patients pat ON p.patient_id = pat.id
WHERE pat.pharmacy_id = sqlc.arg(pharmacy_id)::BIGINT
ORDER BY o.estimated_depletion_date ASC;

-- name: FulfillActiveOrderByPrescription :exec
UPDATE orders
SET status = 'fulfilled', updated_at = now()
WHERE prescription_id = $1
  AND status IN ('pending', 'prepared');

-- name: ListPrescriptionsInLookahead :many
SELECT
    p.id AS prescription_id,
    p.units_per_box,
    p.daily_consumption,
    p.box_start_date,
    pat.id AS patient_id
FROM prescriptions p
JOIN patients pat ON p.patient_id = pat.id
WHERE pat.pharmacy_id = sqlc.arg(pharmacy_id)::BIGINT
  AND pat.consensus = true
ORDER BY p.id;
