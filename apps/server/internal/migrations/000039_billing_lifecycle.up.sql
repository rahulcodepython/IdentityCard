-- 000039_billing_lifecycle.up.sql
-- Adds prune_date to billing, idempotent notification logging, and transactional outbox

-- 1. Add prune_date column with default 30-day grace window
ALTER TABLE billing ADD COLUMN prune_date DATE;

-- Populate existing records with period_end + 30 days
UPDATE billing SET prune_date = (period_end + 30);

-- Enforce NOT NULL constraint
ALTER TABLE billing ALTER COLUMN prune_date SET NOT NULL;

-- Automatically default prune_date to period_end + 30 days if omitted on INSERT or UPDATE of period_end
CREATE OR REPLACE FUNCTION set_billing_prune_date()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.prune_date IS NULL THEN
        NEW.prune_date := NEW.period_end + 30;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_billing_prune_date
BEFORE INSERT OR UPDATE OF period_end ON billing
FOR EACH ROW
EXECUTE FUNCTION set_billing_prune_date();

-- Composite index for fast gateway authorization checks
CREATE INDEX idx_billing_gateway_auth 
ON billing (organization_id, period_start DESC);

-- Index for the daily pruning engine
CREATE INDEX idx_billing_prune_sweep 
ON billing (prune_date, status) 
WHERE status IN ('pending', 'cancel');

-- 2. Idempotent lifecycle notification log table
CREATE TABLE billing_lifecycle_notifications (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id   UUID NOT NULL REFERENCES "organization" (id) ON DELETE CASCADE,
    lineage_root_id   UUID NOT NULL REFERENCES billing (id) ON DELETE CASCADE,
    notification_type VARCHAR(32) NOT NULL,
    sent_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (lineage_root_id, notification_type)
);

CREATE INDEX idx_lifecycle_notifications_lookup 
ON billing_lifecycle_notifications (lineage_root_id, notification_type);

-- 3. Transactional outbox table for resilient asynchronous jobs
CREATE TABLE job_outbox (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    queue         VARCHAR(64) NOT NULL DEFAULT 'default',
    job_type      VARCHAR(128) NOT NULL,
    payload       JSONB NOT NULL,
    status        VARCHAR(32) NOT NULL DEFAULT 'pending',
    attempts      INT NOT NULL DEFAULT 0,
    max_attempts  INT NOT NULL DEFAULT 5,
    last_error    TEXT,
    run_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    locked_at     TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_job_outbox_dequeue 
ON job_outbox (status, run_at) 
WHERE status IN ('pending', 'failed');
