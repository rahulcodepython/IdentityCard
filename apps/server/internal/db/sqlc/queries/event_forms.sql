-- name: CreateEventForm :one
INSERT INTO event_forms (organization_id, event_id, sub_event_id, token, capacity)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListEventForms :many
SELECT * FROM event_forms WHERE event_id = $1 AND organization_id = $2 ORDER BY created_at DESC;

-- name: GetEventForm :one
SELECT * FROM event_forms WHERE id = $1 AND event_id = $2 AND organization_id = $3;

-- name: UpdateEventForm :one
UPDATE event_forms SET capacity = $4, is_active = $5, updated_at = now()
WHERE id = $1 AND event_id = $2 AND organization_id = $3
RETURNING *;

-- name: DeleteEventForm :execrows
DELETE FROM event_forms WHERE id = $1 AND event_id = $2 AND organization_id = $3;

-- name: GetPublicFormByToken :one
-- Everything the public submit page needs in one query: the form itself
-- plus enough parent context (org/event id for the person upsert, names
-- for display) that no authenticated read is required to serve it.
SELECT
    event_forms.id,
    event_forms.token,
    event_forms.capacity,
    event_forms.submissions_count,
    event_forms.is_active,
    event_forms.event_id,
    event_forms.sub_event_id,
    event_forms.organization_id,
    events.name AS event_name,
    sub_events.name AS sub_event_name
FROM event_forms
JOIN events ON events.id = event_forms.event_id
LEFT JOIN sub_events ON sub_events.id = event_forms.sub_event_id
WHERE event_forms.token = $1;

-- name: IncrementEventFormSubmissions :one
-- Atomic capacity check-and-increment; zero rows back means the form is
-- full (or inactive) — the caller must not create the person in that case.
UPDATE event_forms SET submissions_count = submissions_count + 1, updated_at = now()
WHERE id = $1 AND is_active = true AND (capacity IS NULL OR submissions_count < capacity)
RETURNING *;

-- name: DecrementEventFormSubmissions :exec
-- Undoes the speculative charge above when the submission turns out to be
-- an existing registrant resubmitting, not a new one — see forms.Service.Submit.
UPDATE event_forms SET submissions_count = GREATEST(submissions_count - 1, 0), updated_at = now()
WHERE id = $1;
