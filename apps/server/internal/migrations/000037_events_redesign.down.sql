DROP TABLE event_metadata;

ALTER TABLE sub_events DROP CONSTRAINT sub_events_time_check;
ALTER TABLE sub_events DROP COLUMN exit_time;
ALTER TABLE sub_events DROP COLUMN entry_time;
ALTER TABLE sub_events DROP COLUMN date;
ALTER TABLE sub_events ADD COLUMN schedule_mode TEXT NOT NULL
    CHECK (schedule_mode IN ('fixed_range', 'selective'));

ALTER TABLE events DROP CONSTRAINT events_flash_single_day_check;
ALTER TABLE events DROP CONSTRAINT events_end_date_check;
ALTER TABLE events DROP COLUMN event_type;
ALTER TABLE events ALTER COLUMN end_date DROP NOT NULL;

ALTER TABLE events ADD COLUMN venue TEXT;
ALTER TABLE events ADD COLUMN schedule_mode TEXT NOT NULL
    CHECK (schedule_mode IN ('flash', 'fixed_range', 'selective', 'recurring'));
ALTER TABLE events ADD CONSTRAINT events_end_date_check
    CHECK (end_date IS NULL OR end_date >= start_date);
ALTER TABLE events ADD CONSTRAINT events_flash_single_day_check
    CHECK (schedule_mode <> 'flash' OR start_date = end_date);

CREATE TABLE event_recurrence (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id   UUID NOT NULL UNIQUE REFERENCES events (id) ON DELETE CASCADE,
    starts_on  DATE NOT NULL,
    ends_on    DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CHECK (ends_on IS NULL OR ends_on >= starts_on)
);

CREATE TABLE event_recurrence_weekdays (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id   UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    weekday    SMALLINT NOT NULL CHECK (weekday BETWEEN 0 AND 6),
    entry_time TIME NOT NULL,
    exit_time  TIME NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (event_id, weekday),
    CHECK (exit_time > entry_time)
);

CREATE INDEX event_recurrence_weekdays_event_id_idx ON event_recurrence_weekdays (event_id);

CREATE TABLE event_excluded_dates (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id   UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    date       DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (event_id, date)
);

CREATE INDEX event_excluded_dates_event_id_idx ON event_excluded_dates (event_id);

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
