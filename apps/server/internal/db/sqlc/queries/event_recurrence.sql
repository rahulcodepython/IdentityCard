-- name: UpsertEventRecurrence :one
-- Update (draft-only, wholesale replace — see events.Service.Update)
-- reuses this same write path as Create, since event_id is unique.
INSERT INTO event_recurrence (event_id, starts_on, ends_on)
VALUES ($1, $2, $3)
ON CONFLICT (event_id) DO UPDATE SET starts_on = EXCLUDED.starts_on, ends_on = EXCLUDED.ends_on, updated_at = now()
RETURNING *;

-- name: GetEventRecurrence :one
SELECT * FROM event_recurrence WHERE event_id = $1;

-- name: UpdateEventRecurrenceEndsOn :one
-- The only field of a recurring event's rule editable post-publish (see
-- events.Service.Publish) — an org can decide to stop an open-ended
-- event, or push a fixed end date further out.
UPDATE event_recurrence SET ends_on = $2, updated_at = now()
WHERE event_id = $1
RETURNING *;

-- name: CreateRecurrenceWeekday :one
INSERT INTO event_recurrence_weekdays (event_id, weekday, entry_time, exit_time)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListRecurrenceWeekdays :many
SELECT * FROM event_recurrence_weekdays WHERE event_id = $1 ORDER BY weekday;

-- name: DeleteRecurrenceWeekdaysForEvent :exec
DELETE FROM event_recurrence_weekdays WHERE event_id = $1;
