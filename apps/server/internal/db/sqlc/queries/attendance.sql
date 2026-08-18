-- name: GetAttendanceRecord :one
SELECT * FROM attendance_records WHERE person_id = $1 AND date = $2;

-- name: CreateAttendanceEntry :one
INSERT INTO attendance_records (organization_id, event_id, person_id, date, entry_at, entry_status, entry_device_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: RecordAttendanceExit :one
UPDATE attendance_records
SET exit_at = $3, exit_status = $4, exit_device_id = $5, updated_at = now()
WHERE person_id = $1 AND date = $2
RETURNING *;

-- name: ListAttendanceForEvent :many
SELECT
    attendance_records.*,
    people.name AS person_name
FROM attendance_records
JOIN people ON people.id = attendance_records.person_id
WHERE attendance_records.event_id = $1 AND attendance_records.organization_id = $2
ORDER BY attendance_records.date, people.name;
