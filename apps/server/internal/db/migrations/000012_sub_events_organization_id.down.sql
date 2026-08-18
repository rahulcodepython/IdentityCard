DROP INDEX IF EXISTS sub_events_organization_id_idx;
ALTER TABLE sub_events DROP CONSTRAINT IF EXISTS sub_events_organization_id_fkey;
ALTER TABLE sub_events DROP COLUMN IF EXISTS organization_id;
