-- A flash event is defined as a single-day event.
ALTER TABLE events ADD CONSTRAINT events_flash_single_day_check
    CHECK (kind <> 'flash' OR start_date = end_date);
