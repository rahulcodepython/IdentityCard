CREATE TABLE people (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    event_id        UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    email           CITEXT NOT NULL,
    mobile          TEXT NOT NULL,
    name            TEXT NOT NULL,
    image_url       TEXT,
    age             SMALLINT,
    gender          TEXT,
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (event_id, email, mobile)
);

CREATE INDEX people_organization_id_idx ON people (organization_id);
CREATE INDEX people_event_id_idx ON people (event_id);
