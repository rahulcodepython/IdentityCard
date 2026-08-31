-- A scanner-bot device. Never logs in as an org member — it's paired via
-- a one-time OTP shown to a human admin, then authenticates every scan
-- with an opaque key (see devices.Service.Authenticate) stored only as a
-- hash here, exactly like an API key.
CREATE TABLE devices (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    status         TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'verified', 'revoked')),
    otp_code       TEXT,
    otp_expires_at TIMESTAMPTZ,
    key_hash       TEXT,
    verified_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX devices_organization_id_idx ON devices (organization_id);
-- Pairing looks a device up by OTP alone (it doesn't know its org yet).
CREATE INDEX devices_otp_code_idx ON devices (otp_code) WHERE status = 'pending';
CREATE UNIQUE INDEX devices_key_hash_idx ON devices (key_hash) WHERE status = 'verified';
