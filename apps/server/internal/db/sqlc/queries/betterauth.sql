-- Queries against better-auth's own tables (see
-- internal/db/migrations/000035_better_auth_schema.up.sql). Go doesn't
-- own this data — better-auth's Next.js side does — but it shares the
-- same Postgres database, so org settings (name/logo, folded onto
-- better-auth's own "organization" table via additionalFields) and the
-- one-off cmd/seed account creation both go through plain SQL here
-- rather than a round trip to the Next.js app.

-- name: GetOrganizationByID :one
SELECT * FROM "organization" WHERE id = $1;

-- name: UpdateOrganizationName :one
UPDATE "organization" SET name = $2 WHERE id = $1 RETURNING *;

-- name: UpdateOrganizationLogo :one
UPDATE "organization" SET "logoObjectKey" = $2 WHERE id = $1 RETURNING *;

-- name: DeleteOrganizationLogo :one
UPDATE "organization" SET "logoObjectKey" = NULL WHERE id = $1 RETURNING *;

-- name: DeleteOrganization :exec
DELETE FROM "organization" WHERE id = $1;

-- cmd/seed only, below — a real signup never goes through Go.

-- name: GetUserByEmail :one
SELECT * FROM "user" WHERE email = $1;

-- name: CreateUser :one
-- Seeded already-verified — the seeded account signs in for real via
-- better-auth's email-OTP flow, it just needs a matching user row to
-- exist first.
INSERT INTO "user" (name, email, "emailVerified")
VALUES ($1, $2, true)
RETURNING *;

-- name: CreateOrganization :one
INSERT INTO "organization" (name, slug)
VALUES ($1, $2)
RETURNING *;

-- name: CreateMember :one
INSERT INTO "member" ("organizationId", "userId", role)
VALUES ($1, $2, $3)
RETURNING *;
