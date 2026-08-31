-- Join-date semantics (see events.Service.Publish): a person added while
-- their event is still a draft has no joined_at until the event is
-- published, at which point everyone still unjoined gets joined_at =
-- published_at. A person added after the event is already published gets
-- joined_at = now() at insert time (set by the people module in a later
-- phase). NULL is what makes "still waiting on publish" representable.
ALTER TABLE people ALTER COLUMN joined_at DROP NOT NULL;
ALTER TABLE people ALTER COLUMN joined_at DROP DEFAULT;
