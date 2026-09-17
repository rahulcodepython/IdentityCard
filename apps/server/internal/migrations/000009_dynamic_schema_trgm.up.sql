-- 1. Enable pg_trgm extension for fast arbitrary substring indexing
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 2. Fast JSONB containment and path operations index for user-defined dynamic schema fields
CREATE INDEX IF NOT EXISTS idx_applicants_data_path_ops ON applicants USING GIN (data jsonb_path_ops);

-- 3. Trigram GIN index on text cast of data for fast substring search across unknown custom keys
CREATE INDEX IF NOT EXISTS idx_applicants_data_trgm ON applicants USING GIN ((data::text) gin_trgm_ops);

-- 4. Trigram indexes on name and email for fast partial matching
CREATE INDEX IF NOT EXISTS idx_applicants_name_trgm ON applicants USING GIN (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_applicants_email_trgm ON applicants USING GIN (email gin_trgm_ops);
