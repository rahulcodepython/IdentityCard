-- Every event now records which subscription funds it — this is what
-- makes the per-subscription grace/deletion sweep in internal/jobs
-- correct (a lapsed subscription only ever deletes the events it
-- personally funded). Prototype phase (see CLAUDE.md) — no production
-- event data exists yet, so rather than leave a NOT NULL column with
-- nothing valid to backfill, any events currently sitting in a dev
-- database are cleared here instead of migrated forward.
TRUNCATE events CASCADE;

ALTER TABLE events ADD COLUMN subscription_id UUID NOT NULL REFERENCES subscriptions (id);

CREATE INDEX events_subscription_id_idx ON events (subscription_id);
