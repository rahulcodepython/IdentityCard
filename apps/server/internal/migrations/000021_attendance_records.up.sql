-- One row per person per calendar date they're scanned on. The first scan
-- of the day records entry, the second records exit — see
-- attendance.Service.Scan. entry/exit_status is the scanner's read of
-- "early/on_time/late" against that day's permitted schedule (the event's,
-- or the narrower one from any sub-event the person belongs to).
CREATE TABLE attendance_records (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id  UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    event_id         UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    person_id        UUID NOT NULL REFERENCES people (id) ON DELETE CASCADE,
    date             DATE NOT NULL,
    entry_at         TIMESTAMPTZ,
    entry_status     TEXT CHECK (entry_status IN ('early', 'on_time', 'late')),
    entry_device_id  UUID REFERENCES devices (id) ON DELETE SET NULL,
    exit_at          TIMESTAMPTZ,
    exit_status      TEXT CHECK (exit_status IN ('early', 'on_time', 'late')),
    exit_device_id   UUID REFERENCES devices (id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (person_id, date)
);

CREATE INDEX attendance_records_event_id_idx ON attendance_records (event_id);
