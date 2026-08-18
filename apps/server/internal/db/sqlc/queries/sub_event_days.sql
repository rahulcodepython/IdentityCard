-- name: CreateSubEventDay :one
INSERT INTO sub_event_days (sub_event_id, date, entry_time, exit_time)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListSubEventDays :many
SELECT * FROM sub_event_days WHERE sub_event_id = $1 ORDER BY date;

-- name: DeleteSubEventDaysForSubEvent :exec
DELETE FROM sub_event_days WHERE sub_event_id = $1;
