-- name: CreateEventDay :one
INSERT INTO event_days (event_id, date, entry_time, exit_time)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListEventDays :many
SELECT * FROM event_days WHERE event_id = $1 ORDER BY date;

-- name: DeleteEventDaysForEvent :exec
DELETE FROM event_days WHERE event_id = $1;
