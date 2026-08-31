-- schedule_mode replaces kind: an event is now authored one of four ways
-- (flash/fixed_range/selective/recurring — see internal/modules/events),
-- not just flash-vs-established. end_date becomes nullable because a
-- recurring event can be open-ended ("runs indefinitely until the org
-- lets the subscription lapse" — confirmed with the user). Prototype
-- phase (see CLAUDE.md) — clear any dev-database events rather than
-- migrate their kind value forward.
TRUNCATE events CASCADE;

ALTER TABLE events DROP CONSTRAINT events_flash_single_day_check;
ALTER TABLE events DROP CONSTRAINT events_kind_check;
ALTER TABLE events DROP CONSTRAINT events_check;
ALTER TABLE events DROP COLUMN kind;

ALTER TABLE events ALTER COLUMN end_date DROP NOT NULL;

ALTER TABLE events ADD COLUMN schedule_mode TEXT NOT NULL
    CHECK (schedule_mode IN ('flash', 'fixed_range', 'selective', 'recurring'));

ALTER TABLE events ADD CONSTRAINT events_end_date_check
    CHECK (end_date IS NULL OR end_date >= start_date);
ALTER TABLE events ADD CONSTRAINT events_flash_single_day_check
    CHECK (schedule_mode <> 'flash' OR start_date = end_date);
