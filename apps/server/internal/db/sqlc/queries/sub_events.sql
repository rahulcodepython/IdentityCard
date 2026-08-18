-- name: CreateSubEvent :one
INSERT INTO sub_events (organization_id, event_id, name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSubEvent :one
SELECT * FROM sub_events WHERE id = $1 AND event_id = $2 AND organization_id = $3;

-- name: ListSubEvents :many
SELECT * FROM sub_events WHERE event_id = $1 AND organization_id = $2 ORDER BY created_at;

-- name: UpdateSubEventName :one
UPDATE sub_events SET name = $4, updated_at = now()
WHERE id = $1 AND event_id = $2 AND organization_id = $3
RETURNING *;

-- name: DeleteSubEvent :execrows
DELETE FROM sub_events WHERE id = $1 AND event_id = $2 AND organization_id = $3;

-- name: ValidateSubEventIDs :many
-- Returns the subset of the given ids that are real sub-events of this
-- event/org — the caller diffs the count to catch any bogus id (used when
-- assigning a person to sub-events).
SELECT id FROM sub_events
WHERE id = ANY(sqlc.arg('ids')::uuid[]) AND event_id = $1 AND organization_id = $2;
