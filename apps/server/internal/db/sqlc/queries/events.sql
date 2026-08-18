-- name: CreateEvent :one
INSERT INTO events (organization_id, name, kind, start_date, end_date, venue)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetEvent :one
SELECT * FROM events WHERE id = $1 AND organization_id = $2;

-- name: ListEvents :many
SELECT * FROM events WHERE organization_id = $1 ORDER BY start_date DESC, created_at DESC;

-- name: UpdateEventNameAndDates :one
UPDATE events SET name = $3, start_date = $4, end_date = $5, venue = $6, updated_at = now()
WHERE id = $1 AND organization_id = $2 AND status = 'draft'
RETURNING *;

-- name: PublishEvent :one
UPDATE events SET status = 'published', published_at = now(), updated_at = now()
WHERE id = $1 AND organization_id = $2 AND status = 'draft'
RETURNING *;

-- name: DeleteDraftEvent :execrows
DELETE FROM events WHERE id = $1 AND organization_id = $2 AND status = 'draft';

-- name: MarkUnjoinedPeopleJoinedAt :exec
-- Publish-time join-date backfill: anyone added to this event while it was
-- still a draft has no joined_at yet (see people.joined_at, made nullable
-- specifically for this) — they all join at the moment of publish.
UPDATE people SET joined_at = $2, updated_at = now()
WHERE event_id = $1 AND joined_at IS NULL;
