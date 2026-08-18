-- name: CreateDevice :one
INSERT INTO devices (organization_id, name, otp_code, otp_expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListDevices :many
SELECT * FROM devices WHERE organization_id = $1 ORDER BY created_at DESC;

-- name: GetDevice :one
SELECT * FROM devices WHERE id = $1 AND organization_id = $2;

-- name: GetPendingDeviceByOTP :one
-- The pairing device doesn't know its org yet — OTP alone is the lookup key.
SELECT * FROM devices
WHERE otp_code = $1 AND status = 'pending' AND otp_expires_at > now();

-- name: MarkDeviceVerified :one
UPDATE devices
SET status = 'verified', key_hash = $2, verified_at = now(),
    otp_code = NULL, otp_expires_at = NULL, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetDeviceByKeyHash :one
SELECT * FROM devices WHERE key_hash = $1 AND status = 'verified';

-- name: RevokeDevice :execrows
UPDATE devices SET status = 'revoked', key_hash = NULL, updated_at = now()
WHERE id = $1 AND organization_id = $2 AND status != 'revoked';
