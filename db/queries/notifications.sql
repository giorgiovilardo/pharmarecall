-- name: CreateNotification :exec
INSERT INTO notifications (pharmacy_id, prescription_id, transition_type)
VALUES ($1, $2, $3)
ON CONFLICT (prescription_id, transition_type) DO NOTHING;

-- name: ListNotificationsByPharmacy :many
SELECT
    n.id,
    n.pharmacy_id,
    n.prescription_id,
    n.transition_type,
    n.read,
    n.created_at,
    p.medication_name,
    p.units_per_box,
    p.daily_consumption,
    p.box_start_date,
    pat.id AS patient_id,
    pat.first_name,
    pat.last_name
FROM notifications n
JOIN prescriptions p ON n.prescription_id = p.id
JOIN patients pat ON p.patient_id = pat.id
WHERE n.pharmacy_id = sqlc.arg(pharmacy_id)::BIGINT
ORDER BY n.created_at DESC;

-- name: MarkNotificationRead :exec
UPDATE notifications
SET read = true
WHERE id = $1 AND pharmacy_id = sqlc.arg(pharmacy_id)::BIGINT;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications
SET read = true
WHERE pharmacy_id = sqlc.arg(pharmacy_id)::BIGINT AND read = false;

-- name: CountUnreadNotifications :one
SELECT count(*) FROM notifications
WHERE pharmacy_id = sqlc.arg(pharmacy_id)::BIGINT AND read = false;
