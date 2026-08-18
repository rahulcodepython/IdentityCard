UPDATE people SET joined_at = now() WHERE joined_at IS NULL;
ALTER TABLE people ALTER COLUMN joined_at SET DEFAULT now();
ALTER TABLE people ALTER COLUMN joined_at SET NOT NULL;
