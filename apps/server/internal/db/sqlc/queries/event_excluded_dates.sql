-- name: CreateExcludedDate :one
INSERT INTO event_excluded_dates (event_id, date)
VALUES ($1, $2)
ON CONFLICT (event_id, date) DO UPDATE SET event_id = event_excluded_dates.event_id
RETURNING *;

-- name: ListExcludedDates :many
SELECT * FROM event_excluded_dates WHERE event_id = $1 ORDER BY date;

-- name: DeleteExcludedDatesForEvent :exec
DELETE FROM event_excluded_dates WHERE event_id = $1;
