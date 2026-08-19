-- name: CreateEvent :one
INSERT INTO events (organization_id, subscription_id, name, schedule_mode, start_date, end_date, venue)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetEvent :one
SELECT * FROM events WHERE id = $1 AND organization_id = $2;

-- name: ListEvents :many
SELECT * FROM events WHERE organization_id = $1 ORDER BY start_date DESC, created_at DESC;

-- name: CountEventsBySubscription :one
-- Used by events.Service.checkEventLimit to enforce a subscription's
-- event_quota (see plans.Service.ActiveSubscriptionsForOrgOrderedByAge) —
-- a subscription with no quota (NULL, i.e. kind='unlimited') is never
-- checked against this at all.
SELECT count(*) FROM events WHERE subscription_id = $1;

-- name: DeleteEventsForSubscription :exec
-- Hard-deletes every event a subscription funded once its grace period
-- (or, for Flash, its fixed post-event retention window) has expired —
-- see internal/jobs. Cascades through event_days, sub_events, people,
-- attendance_records, and event_forms via their existing ON DELETE
-- CASCADE foreign keys on events.id.
DELETE FROM events WHERE subscription_id = $1;

-- name: GetEventEndDateForSubscription :one
-- A Flash subscription funds exactly one event, so this is unambiguous.
SELECT end_date FROM events WHERE subscription_id = $1;

-- name: UpdateEventNameAndDates :one
UPDATE events SET name = $3, start_date = $4, end_date = $5, venue = $6, updated_at = now()
WHERE id = $1 AND organization_id = $2 AND status = 'draft'
RETURNING *;

-- name: ListRecurringEventsForMaterialization :many
-- Every published, still-running recurring event — internal/jobs extends
-- each one's materialized event_days window daily (see
-- events.Service.ExtendRecurringHorizon).
SELECT e.* FROM events e
JOIN event_recurrence r ON r.event_id = e.id
WHERE e.schedule_mode = 'recurring'
  AND e.status = 'published'
  AND (r.ends_on IS NULL OR r.ends_on > $1::date)
ORDER BY e.id;

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
