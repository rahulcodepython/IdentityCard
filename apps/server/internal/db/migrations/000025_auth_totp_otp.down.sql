ALTER TABLE organizations ADD COLUMN onboarding_completed_at TIMESTAMPTZ;

ALTER TABLE users
    DROP COLUMN email_verified_at,
    DROP COLUMN otp_expires_at,
    DROP COLUMN otp_code,
    DROP COLUMN totp_secret,
    ADD COLUMN password_hash TEXT;
