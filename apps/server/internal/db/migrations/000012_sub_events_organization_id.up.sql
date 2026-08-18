-- sub_events was missing organization_id, breaking the "every tenant-owned
-- table carries organization_id, queries always scope by it" convention.
-- Backfilled from the parent event since sub_events.event_id already
-- determines the org.
ALTER TABLE sub_events ADD COLUMN organization_id UUID;

UPDATE sub_events
SET organization_id = events.organization_id
FROM events
WHERE events.id = sub_events.event_id;

ALTER TABLE sub_events ALTER COLUMN organization_id SET NOT NULL;
ALTER TABLE sub_events
    ADD CONSTRAINT sub_events_organization_id_fkey
    FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE CASCADE;

CREATE INDEX sub_events_organization_id_idx ON sub_events (organization_id);
