-- 1. Add phone column to applicants table
ALTER TABLE applicants ADD COLUMN IF NOT EXISTS phone TEXT NOT NULL DEFAULT '';

-- 2. Backfill existing applicants from data jsonb (phone or mobile keys)
UPDATE applicants
SET phone = COALESCE(NULLIF(data->>'phone', ''), NULLIF(data->>'mobile', ''), '')
WHERE phone = '';

-- 3. Add B-tree index for exact matching and Trigram index for fast partial matching
CREATE INDEX IF NOT EXISTS idx_applicants_phone ON applicants (phone);
CREATE INDEX IF NOT EXISTS idx_applicants_phone_trgm ON applicants USING GIN (phone gin_trgm_ops);
