-- Auth (users/organizations/organization_members) moved to better-auth
-- (see 000035_better_auth_schema) — no traces of Go-owned auth stay
-- behind. Per CLAUDE.md (prototype phase, no production data), this is a
-- straight drop/repoint rather than a data-preserving migration: every
-- tenant-scoped table gets truncated (CASCADE pulls in their children —
-- event_days, sub_event_days, people_sub_events, subscription_periods,
-- etc.) since their existing organization_id values point at rows that
-- are about to stop existing.
TRUNCATE TABLE events, sub_events, people, event_forms, devices, attendance_records, subscriptions CASCADE;

DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations CASCADE;
DROP TABLE IF EXISTS users CASCADE;

ALTER TABLE events
    ADD CONSTRAINT events_organization_id_fkey
    FOREIGN KEY (organization_id) REFERENCES "organization" (id) ON DELETE CASCADE;

ALTER TABLE sub_events
    ADD CONSTRAINT sub_events_organization_id_fkey
    FOREIGN KEY (organization_id) REFERENCES "organization" (id) ON DELETE CASCADE;

ALTER TABLE people
    ADD CONSTRAINT people_organization_id_fkey
    FOREIGN KEY (organization_id) REFERENCES "organization" (id) ON DELETE CASCADE;

ALTER TABLE event_forms
    ADD CONSTRAINT event_forms_organization_id_fkey
    FOREIGN KEY (organization_id) REFERENCES "organization" (id) ON DELETE CASCADE;

ALTER TABLE devices
    ADD CONSTRAINT devices_organization_id_fkey
    FOREIGN KEY (organization_id) REFERENCES "organization" (id) ON DELETE CASCADE;

ALTER TABLE attendance_records
    ADD CONSTRAINT attendance_records_organization_id_fkey
    FOREIGN KEY (organization_id) REFERENCES "organization" (id) ON DELETE CASCADE;

ALTER TABLE subscriptions
    ADD CONSTRAINT subscriptions_organization_id_fkey
    FOREIGN KEY (organization_id) REFERENCES "organization" (id) ON DELETE CASCADE;
