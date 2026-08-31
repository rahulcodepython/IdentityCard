-- A shareable public sign-up link for an event, optionally scoped to one
-- sub-event, optionally capped at the first N submissions. token is the
-- opaque public-facing identifier (never expose event_forms.id publicly).
CREATE TABLE event_forms (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id   UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    event_id          UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    sub_event_id      UUID REFERENCES sub_events (id) ON DELETE CASCADE,
    token             TEXT NOT NULL UNIQUE,
    capacity          INT,
    submissions_count INT NOT NULL DEFAULT 0,
    is_active         BOOLEAN NOT NULL DEFAULT true,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CHECK (capacity IS NULL OR capacity > 0)
);

CREATE INDEX event_forms_event_id_idx ON event_forms (event_id);
