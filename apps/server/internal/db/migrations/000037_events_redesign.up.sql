-- Collapses schedule_mode (flash/fixed_range/selective/recurring) plus
-- event_recurrence/event_recurrence_weekdays/event_excluded_dates/
-- sub_event_days into three explicit, immutable event types
-- (flash/standard/grouped) — see PROJECT_MEMORY.md for the full
-- rationale. Prototype phase (CLAUDE.md): straight drop/recreate, no
-- data to preserve.
TRUNCATE TABLE events, sub_events CASCADE;

DROP TABLE event_recurrence_weekdays;
DROP TABLE event_recurrence;
DROP TABLE event_excluded_dates;
DROP TABLE sub_event_days;

-- events: schedule_mode -> event_type. end_date is NOT NULL again — every
-- remaining type has a defined end (no more open-ended recurring).
-- venue moves to event_metadata.
ALTER TABLE events DROP CONSTRAINT events_flash_single_day_check;
ALTER TABLE events DROP CONSTRAINT events_schedule_mode_check;
ALTER TABLE events DROP CONSTRAINT events_end_date_check;
ALTER TABLE events DROP COLUMN schedule_mode;
ALTER TABLE events DROP COLUMN venue;
ALTER TABLE events ALTER COLUMN end_date SET NOT NULL;

ALTER TABLE events ADD COLUMN event_type TEXT NOT NULL
    CHECK (event_type IN ('flash', 'standard', 'grouped'));
ALTER TABLE events ADD CONSTRAINT events_end_date_check
    CHECK (end_date >= start_date);
ALTER TABLE events ADD CONSTRAINT events_flash_single_day_check
    CHECK (event_type <> 'flash' OR start_date = end_date);

-- sub_events: schedule_mode + sub_event_days -> a single date/entry/exit
-- directly on the row. A sub-event is always exactly one day now.
ALTER TABLE sub_events DROP CONSTRAINT sub_events_schedule_mode_check;
ALTER TABLE sub_events DROP COLUMN schedule_mode;

ALTER TABLE sub_events ADD COLUMN date       DATE NOT NULL;
ALTER TABLE sub_events ADD COLUMN entry_time TIME NOT NULL;
ALTER TABLE sub_events ADD COLUMN exit_time  TIME NOT NULL;
ALTER TABLE sub_events ADD CONSTRAINT sub_events_time_check
    CHECK (exit_time > entry_time);

-- Everything display-only about an event — events itself stays purely
-- structural. Extending this later is just an ADD COLUMN here.
CREATE TABLE event_metadata (
    id                             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id                       UUID NOT NULL UNIQUE REFERENCES events (id) ON DELETE CASCADE,
    name                           TEXT NOT NULL,
    venue                          TEXT,
    image_object_key               TEXT,
    organizer_name                 TEXT,
    organizer_signature_object_key TEXT,
    created_at                     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                     TIMESTAMPTZ NOT NULL DEFAULT now()
);
