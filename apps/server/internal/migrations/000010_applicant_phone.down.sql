DROP INDEX IF EXISTS idx_applicants_phone_trgm;
DROP INDEX IF EXISTS idx_applicants_phone;
ALTER TABLE applicants DROP COLUMN IF EXISTS phone;
