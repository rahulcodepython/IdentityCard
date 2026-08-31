ALTER TABLE events DROP CONSTRAINT IF EXISTS events_organization_id_fkey;
ALTER TABLE sub_events DROP CONSTRAINT IF EXISTS sub_events_organization_id_fkey;
ALTER TABLE people DROP CONSTRAINT IF EXISTS people_organization_id_fkey;
ALTER TABLE event_forms DROP CONSTRAINT IF EXISTS event_forms_organization_id_fkey;
ALTER TABLE devices DROP CONSTRAINT IF EXISTS devices_organization_id_fkey;
ALTER TABLE attendance_records DROP CONSTRAINT IF EXISTS attendance_records_organization_id_fkey;
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_organization_id_fkey;

TRUNCATE TABLE events, sub_events, people, event_forms, devices, attendance_records, subscriptions CASCADE;

-- Reconstructs the pre-cutover shape as of 000025_auth_totp_otp (the last
-- migration to touch these tables) — not the original 000002/000003
-- shape.
CREATE TABLE users (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email             CITEXT NOT NULL UNIQUE,
    name              TEXT NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    google_id         CITEXT UNIQUE,
    avatar_url        TEXT,
    totp_secret       TEXT,
    otp_code          TEXT,
    otp_expires_at    TIMESTAMPTZ,
    email_verified_at TIMESTAMPTZ
);

CREATE TABLE organizations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    slug            CITEXT NOT NULL UNIQUE,
    logo_object_key TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE organization_members (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    roles           TEXT[] NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (organization_id, user_id),
    CONSTRAINT organization_members_roles_valid CHECK (
        roles <@ ARRAY['super_admin', 'admin', 'scanner']::TEXT[]
        AND array_length(roles, 1) > 0
    )
);

CREATE INDEX organization_members_user_id_idx ON organization_members (user_id);

CREATE UNIQUE INDEX organization_members_one_super_admin_idx
    ON organization_members (organization_id)
    WHERE 'super_admin' = ANY (roles);

ALTER TABLE events
    ADD CONSTRAINT events_organization_id_fkey
    FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE;

ALTER TABLE sub_events
    ADD CONSTRAINT sub_events_organization_id_fkey
    FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE;

ALTER TABLE people
    ADD CONSTRAINT people_organization_id_fkey
    FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE;

ALTER TABLE event_forms
    ADD CONSTRAINT event_forms_organization_id_fkey
    FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE;

ALTER TABLE devices
    ADD CONSTRAINT devices_organization_id_fkey
    FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE;

ALTER TABLE attendance_records
    ADD CONSTRAINT attendance_records_organization_id_fkey
    FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE;

ALTER TABLE subscriptions
    ADD CONSTRAINT subscriptions_organization_id_fkey
    FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE;
