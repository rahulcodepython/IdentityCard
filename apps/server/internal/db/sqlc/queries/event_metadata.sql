-- name: CreateEventMetadata :one
INSERT INTO event_metadata (event_id, name, venue, organizer_name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetEventMetadata :one
SELECT * FROM event_metadata WHERE event_id = $1;

-- name: UpdateEventMetadata :one
UPDATE event_metadata SET name = $2, venue = $3, organizer_name = $4, updated_at = now()
WHERE event_id = $1
RETURNING *;

-- name: UpdateEventImage :one
UPDATE event_metadata SET image_object_key = $2, updated_at = now()
WHERE event_id = $1
RETURNING *;

-- name: UpdateEventOrganizerSignature :one
UPDATE event_metadata SET organizer_signature_object_key = $2, updated_at = now()
WHERE event_id = $1
RETURNING *;
