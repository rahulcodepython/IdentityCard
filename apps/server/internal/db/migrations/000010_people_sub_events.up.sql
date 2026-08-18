CREATE TABLE people_sub_events (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    person_id    UUID NOT NULL REFERENCES people (id) ON DELETE CASCADE,
    sub_event_id UUID NOT NULL REFERENCES sub_events (id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (person_id, sub_event_id)
);

CREATE INDEX people_sub_events_sub_event_id_idx ON people_sub_events (sub_event_id);
