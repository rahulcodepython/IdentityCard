-- Printed on the ID card ("place" in the spec) — optional, since not every
-- event has a fixed physical venue worth stating.
ALTER TABLE events ADD COLUMN venue TEXT;
