-- 1. Create event_forms table (independent snapshot, no foreign keys to templates)
CREATE TABLE IF NOT EXISTS event_forms (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id        UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    name            TEXT NOT NULL DEFAULT 'Event Registration Form',
    fields          JSONB NOT NULL DEFAULT '[]'::jsonb,
    max_applicants  INT NOT NULL DEFAULT -1,
    expires_at      TIMESTAMPTZ NOT NULL,
    status          TEXT NOT NULL DEFAULT 'waiting' CHECK (status IN ('waiting', 'live')),
    is_locked       BOOLEAN NOT NULL DEFAULT false,
    locked_at       TIMESTAMPTZ NULL DEFAULT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_event_forms_event UNIQUE (event_id),
    CONSTRAINT chk_event_forms_max_applicants CHECK (max_applicants = -1 OR max_applicants >= 1)
);

CREATE INDEX IF NOT EXISTS idx_event_forms_event_id ON event_forms (event_id);
CREATE INDEX IF NOT EXISTS idx_event_forms_is_locked ON event_forms (is_locked);
CREATE INDEX IF NOT EXISTS idx_event_forms_status ON event_forms (status);

-- 2. Create applicants table
CREATE TABLE IF NOT EXISTS applicants (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    email       TEXT NOT NULL,
    data        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_applicants_email ON applicants (email);
CREATE INDEX IF NOT EXISTS idx_applicants_created_at ON applicants (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_applicants_data_gin ON applicants USING GIN (data);

-- 3. Create event_applicants table
CREATE TABLE IF NOT EXISTS event_applicants (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id        UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    user_id         TEXT NOT NULL REFERENCES applicants (id) ON DELETE CASCADE,
    email           TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_event_applicants_event_email UNIQUE (event_id, email)
);

CREATE INDEX IF NOT EXISTS idx_event_applicants_event_id ON event_applicants (event_id);
CREATE INDEX IF NOT EXISTS idx_event_applicants_user_id ON event_applicants (user_id);
CREATE INDEX IF NOT EXISTS idx_event_applicants_created_at ON event_applicants (created_at DESC);
