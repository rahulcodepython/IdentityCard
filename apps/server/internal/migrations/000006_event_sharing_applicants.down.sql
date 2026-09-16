DROP TABLE IF EXISTS event_applicants;
DROP TABLE IF EXISTS applicants;
DROP TABLE IF EXISTS event_forms;
ALTER TABLE forms DROP COLUMN IF EXISTS published_at;
ALTER TABLE forms DROP COLUMN IF EXISTS is_published;
