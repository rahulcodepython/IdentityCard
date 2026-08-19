-- name: CreateEventDay :one
INSERT INTO event_days (event_id, date, entry_time, exit_time)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListEventDays :many
SELECT * FROM event_days WHERE event_id = $1 ORDER BY date;

-- name: DeleteEventDaysForEvent :exec
DELETE FROM event_days WHERE event_id = $1;

-- name: DeleteEventDay :exec
-- Removes a single materialized day — used when an exclusion date added
-- after publish (see events.Service.AddExcludedDate) falls on a date
-- that's already been materialized for a recurring event.
DELETE FROM event_days WHERE event_id = $1 AND date = $2;

-- name: LatestEventDay :one
-- The furthest-out materialized date for one event — the horizon
-- extension job (events.Service.ExtendRecurringHorizon) resumes
-- generating from the day after this, so it never re-checks or
-- re-inserts already-materialized days. Returns no rows if the event has
-- no days yet (a freshly created recurring event).
SELECT * FROM event_days WHERE event_id = $1 ORDER BY date DESC LIMIT 1;
