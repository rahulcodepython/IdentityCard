-- The per-day schedule for an event. A flash event has exactly one row; an
-- established event has one row per day it's actually active — which may
-- be a non-contiguous subset of [start_date, end_date] ("selective dates"
-- in the spec). events.start_date/end_date stay as a denormalized summary
-- (min/max of these rows), recomputed by the service whenever days change.
CREATE TABLE event_days (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id   UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    date       DATE NOT NULL,
    entry_time TIME NOT NULL,
    exit_time  TIME NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (event_id, date),
    CHECK (exit_time > entry_time) -- same-day window only; overnight events are a later phase
);

CREATE INDEX event_days_event_id_idx ON event_days (event_id);
