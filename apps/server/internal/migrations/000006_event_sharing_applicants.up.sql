-- 1. Alter forms to add publication status
ALTER TABLE forms ADD COLUMN IF NOT EXISTS is_published BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE forms ADD COLUMN IF NOT EXISTS published_at TIMESTAMPTZ DEFAULT NULL;
CREATE INDEX IF NOT EXISTS idx_forms_is_published ON forms (is_published);

-- 2. Create event_forms table
CREATE TABLE IF NOT EXISTS event_forms (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id        UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    form_id         UUID NOT NULL REFERENCES forms (id) ON DELETE RESTRICT,
    max_applicants  INT NOT NULL DEFAULT -1,
    expires_at      TIMESTAMPTZ NOT NULL,
    status          TEXT NOT NULL DEFAULT 'waiting' CHECK (status IN ('waiting', 'live')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_event_forms_event UNIQUE (event_id),
    CONSTRAINT chk_event_forms_max_applicants CHECK (max_applicants = -1 OR max_applicants >= 1)
);

CREATE INDEX IF NOT EXISTS idx_event_forms_event_id ON event_forms (event_id);
CREATE INDEX IF NOT EXISTS idx_event_forms_form_id ON event_forms (form_id);
CREATE INDEX IF NOT EXISTS idx_event_forms_status ON event_forms (status);

-- 3. Create applicants table
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

-- 4. Create event_applicants table
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
