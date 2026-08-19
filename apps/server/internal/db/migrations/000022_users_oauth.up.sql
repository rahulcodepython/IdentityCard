-- Google-only accounts have no password. password_hash stays required for
-- password accounts at the application layer (auth.Service.Register still
-- always sets it); it's only nullable here so an OAuth-created row can
-- omit it instead of storing a dummy hash.
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
ALTER TABLE users ADD COLUMN google_id CITEXT UNIQUE;
ALTER TABLE users ADD COLUMN avatar_url TEXT;
