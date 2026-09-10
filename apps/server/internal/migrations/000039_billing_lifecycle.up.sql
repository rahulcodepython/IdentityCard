-- 000039_billing_lifecycle.up.sql
-- Adds prune_date to billing, idempotent notification logging, and transactional outbox

-- 1. Add prune_date column with default 30-day grace window
ALTER TABLE billing 
ADD COLUMN prune_date DATE NOT NULL DEFAULT (period_end + INTERVAL '30 days');

UPDATE billing SET prune_date = period_end + INTERVAL '30 days';

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
    organization_id   UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
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
