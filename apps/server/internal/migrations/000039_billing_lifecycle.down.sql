-- 000039_billing_lifecycle.down.sql
DROP TABLE IF EXISTS job_outbox;
DROP TABLE IF EXISTS billing_lifecycle_notifications;
DROP INDEX IF EXISTS idx_billing_prune_sweep;
DROP INDEX IF EXISTS idx_billing_gateway_auth;
ALTER TABLE billing DROP COLUMN IF EXISTS prune_date;
