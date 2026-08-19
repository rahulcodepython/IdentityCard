-- Passwords are gone: accounts sign in with a TOTP code (an authenticator
-- app) or a one-time email code instead. totp_secret is generated at
-- signup; otp_code/otp_expires_at are single-use and cleared on
-- consumption (see internal/modules/auth). email_verified_at is null until
-- the first successful OTP/TOTP verification — a Google account gets it
-- set immediately at signup, since Google already verified the address.
ALTER TABLE users
    DROP COLUMN password_hash,
    ADD COLUMN totp_secret TEXT,
    ADD COLUMN otp_code TEXT,
    ADD COLUMN otp_expires_at TIMESTAMPTZ,
    ADD COLUMN email_verified_at TIMESTAMPTZ;

-- Onboarding is redefined as "does this user have an organization at all",
-- derived from organization_members — not a per-org flag anymore.
ALTER TABLE organizations DROP COLUMN onboarding_completed_at;
