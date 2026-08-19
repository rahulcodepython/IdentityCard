TRUNCATE events CASCADE;

ALTER TABLE events DROP CONSTRAINT events_flash_single_day_check;
ALTER TABLE events DROP CONSTRAINT events_end_date_check;

ALTER TABLE events DROP COLUMN schedule_mode;

ALTER TABLE events ALTER COLUMN end_date SET NOT NULL;

ALTER TABLE events ADD COLUMN kind TEXT NOT NULL CHECK (kind IN ('established', 'flash'));
ALTER TABLE events ADD CONSTRAINT events_check CHECK (end_date >= start_date);
ALTER TABLE events ADD CONSTRAINT events_flash_single_day_check
    CHECK (kind <> 'flash' OR start_date = end_date);
