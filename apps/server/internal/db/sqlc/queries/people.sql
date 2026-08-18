-- name: UpsertPerson :one
-- Shared by manual entry, CSV import, and public-form submission (see
-- people.Service.upsert): one person per (event_id, email, mobile). A
-- resubmission updates their details but never touches joined_at, and the
-- caller can tell an insert from an update via the returned `inserted` flag
-- (Postgres's xmax=0 trick) to decide whether to charge a form's capacity.
INSERT INTO people (organization_id, event_id, email, mobile, name, image_url, age, gender, joined_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (event_id, email, mobile) DO UPDATE SET
    name = EXCLUDED.name,
    image_url = EXCLUDED.image_url,
    age = EXCLUDED.age,
    gender = EXCLUDED.gender,
    updated_at = now()
RETURNING *, (xmax = 0) AS inserted;

-- name: GetPerson :one
SELECT * FROM people WHERE id = $1 AND event_id = $2 AND organization_id = $3;

-- name: ListPeopleForEvent :many
SELECT * FROM people
WHERE event_id = $1
  AND organization_id = $2
  AND (
    sqlc.narg('sub_event_id')::uuid IS NULL
    OR EXISTS (
      SELECT 1 FROM people_sub_events pse
      WHERE pse.person_id = people.id AND pse.sub_event_id = sqlc.narg('sub_event_id')
    )
  )
  AND (
    sqlc.narg('search')::text IS NULL
    OR name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR email ILIKE '%' || sqlc.narg('search')::text || '%'
    OR mobile ILIKE '%' || sqlc.narg('search')::text || '%'
  )
ORDER BY created_at DESC;

-- name: ListPeopleByIDs :many
SELECT * FROM people
WHERE event_id = $1 AND organization_id = $2 AND id = ANY(sqlc.arg('ids')::uuid[]);

-- name: UpdatePerson :one
UPDATE people SET name = $4, image_url = $5, age = $6, gender = $7, updated_at = now()
WHERE id = $1 AND event_id = $2 AND organization_id = $3
RETURNING *;

-- name: DeletePerson :execrows
DELETE FROM people WHERE id = $1 AND event_id = $2 AND organization_id = $3;

-- name: ListPersonSubEventIDs :many
SELECT sub_event_id FROM people_sub_events WHERE person_id = $1;

-- name: AddPersonToSubEvent :exec
INSERT INTO people_sub_events (person_id, sub_event_id) VALUES ($1, $2)
ON CONFLICT (person_id, sub_event_id) DO NOTHING;

-- name: DeletePersonSubEvents :exec
DELETE FROM people_sub_events WHERE person_id = $1;

-- name: MarkCardSent :exec
UPDATE people SET card_sent_at = now() WHERE id = $1;
