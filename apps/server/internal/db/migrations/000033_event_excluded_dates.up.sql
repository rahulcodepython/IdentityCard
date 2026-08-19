-- Dates that would otherwise be active (per a fixed_range span or a
-- recurring weekday pattern) but are explicitly skipped — holidays in the
-- middle of an office's term, etc. Not meaningful for flash/selective,
-- where the admin only ever lists the dates they actually want.
CREATE TABLE event_excluded_dates (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id   UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    date       DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (event_id, date)
);

CREATE INDEX event_excluded_dates_event_id_idx ON event_excluded_dates (event_id);
