-- Sub-events get the same fixed_range/selective authoring modes as
-- events (not recurring or flash — a sub-event's whole purpose is
-- splitting people across a subset of the parent's already-materialized
-- days, so it doesn't need its own open-ended recurrence rule).
TRUNCATE sub_events CASCADE;

ALTER TABLE sub_events ADD COLUMN schedule_mode TEXT NOT NULL
    CHECK (schedule_mode IN ('fixed_range', 'selective'));
