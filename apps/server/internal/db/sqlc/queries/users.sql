-- name: CreateUser :one
-- New accounts start unverified (email_verified_at is null) with a TOTP
-- secret generated up front — the QR is only ever shown once, on the
-- signup verification screen (see auth.Service.Register).
INSERT INTO users (email, name, totp_secret)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateOAuthUser :one
-- Google-only accounts skip verification entirely — Google already
-- verified the email — so email_verified_at is set immediately and no
-- totp_secret is generated.
INSERT INTO users (email, name, google_id, avatar_url, email_verified_at)
VALUES ($1, $2, $3, $4, now())
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByGoogleID :one
SELECT * FROM users WHERE google_id = $1;

-- name: LinkGoogleID :one
-- Links an existing account to a Google identity the first time they log
-- in with Google using the same (Google-verified) email.
UPDATE users SET google_id = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetUserOTP :one
UPDATE users SET otp_code = $2, otp_expires_at = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: VerifyAndConsumeOTP :one
-- Single-use: a matching row is only ever found once, since otp_code is
-- cleared in the same statement. email_verified_at is set here too — the
-- first successful verification (OTP or TOTP) is what completes
-- registration; there's no separate "purpose" to track.
UPDATE users
SET otp_code = NULL, otp_expires_at = NULL,
    email_verified_at = COALESCE(email_verified_at, now()),
    updated_at = now()
WHERE email = $1 AND otp_code = $2 AND otp_expires_at > now()
RETURNING *;

-- name: MarkEmailVerified :one
-- Idempotent: a second call just re-returns the same timestamp via
-- COALESCE rather than bumping it.
UPDATE users SET email_verified_at = COALESCE(email_verified_at, now()), updated_at = now()
WHERE id = $1
RETURNING *;
