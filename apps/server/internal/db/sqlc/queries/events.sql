-- name: CreateEvent :one
INSERT INTO events (organization_id, event_type, start_date, end_date)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetEvent :one
SELECT * FROM events WHERE id = $1 AND organization_id = $2;

-- name: ListEvents :many
SELECT * FROM events WHERE organization_id = $1 ORDER BY start_date DESC, created_at DESC;

-- name: UpdateEventDates :one
UPDATE events SET start_date = $3, end_date = $4, updated_at = now()
WHERE id = $1 AND organization_id = $2 AND status = 'draft'
RETURNING *;

-- name: PublishEvent :one
UPDATE events SET status = 'published', published_at = now(), updated_at = now()
WHERE id = $1 AND organization_id = $2 AND status = 'draft'
RETURNING *;

-- name: DeleteDraftEvent :execrows
DELETE FROM events WHERE id = $1 AND organization_id = $2 AND status = 'draft';

-- name: DeleteEvent :exec
-- Unconditional hard delete — used by the credits-restriction/flash
-- cleanup cron (internal/jobs), a system sweep, not a user-facing action
-- gated on draft status.
DELETE FROM events WHERE id = $1;

-- name: ListFlashEventsOlderThan :many
-- Flash events whose (single) event_days.date is older than the cutoff —
-- the cleanup cron's first deletion criterion.
SELECT e.* FROM events e
JOIN event_days d ON d.event_id = e.id
WHERE e.event_type = 'flash' AND d.date < $1::date;

-- name: MarkUnjoinedPeopleJoinedAt :exec
-- Publish-time join-date backfill: anyone added to this event while it was
-- still a draft has no joined_at yet (see people.joined_at, made nullable
-- specifically for this) — they all join at the moment of publish.
UPDATE people SET joined_at = $2, updated_at = now()
WHERE event_id = $1 AND joined_at IS NULL;
