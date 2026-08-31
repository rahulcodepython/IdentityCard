-- A recurring event's authored rule (schedule_mode='recurring' only) —
-- separate from event_days, which stays the materialized truth every
-- other module reads. ends_on NULL means open-ended: the org keeps
-- running it for as long as its subscription stays paid.
CREATE TABLE event_recurrence (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id   UUID NOT NULL UNIQUE REFERENCES events (id) ON DELETE CASCADE,
    starts_on  DATE NOT NULL,
    ends_on    DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CHECK (ends_on IS NULL OR ends_on >= starts_on)
);

-- Which weekdays are active and what time window applies on each —
-- independently, since e.g. Friday can close earlier than Mon-Thu. Absent
-- weekday = not active that week.
CREATE TABLE event_recurrence_weekdays (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id   UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    weekday    SMALLINT NOT NULL CHECK (weekday BETWEEN 0 AND 6), -- 0=Monday..6=Sunday
    entry_time TIME NOT NULL,
    exit_time  TIME NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (event_id, weekday),
    CHECK (exit_time > entry_time)
);

CREATE INDEX event_recurrence_weekdays_event_id_idx ON event_recurrence_weekdays (event_id);
