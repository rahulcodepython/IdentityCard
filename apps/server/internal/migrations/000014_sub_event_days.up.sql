-- A sub-event's own day schedule, when it applies to only a subset of the
-- parent event's days (e.g. a "day 2 workshop" sub-event within a 5-day
-- established event). Empty for a sub-event = it runs on every day of the
-- parent event, using the parent's entry/exit times.
--
-- Each date+time pair is application-validated (in subevents.Service) to
-- be one of the parent event's own event_days rows, rather than enforced
-- via a composite foreign key — keeps the schema simpler for data that's
-- always admin-entered through this API, never bulk-loaded externally.
CREATE TABLE sub_event_days (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sub_event_id UUID NOT NULL REFERENCES sub_events (id) ON DELETE CASCADE,
    date         DATE NOT NULL,
    entry_time   TIME NOT NULL,
    exit_time    TIME NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (sub_event_id, date),
    CHECK (exit_time > entry_time)
);

CREATE INDEX sub_event_days_sub_event_id_idx ON sub_event_days (sub_event_id);
