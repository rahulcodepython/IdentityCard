-- 000040_clean_billing_system.up.sql
-- 1. Drop old tables, triggers, and functions
DROP TABLE IF EXISTS billing_lifecycle_notifications CASCADE;
DROP TABLE IF EXISTS credits CASCADE;
DROP TABLE IF EXISTS billing CASCADE;
DROP TABLE IF EXISTS transactions CASCADE;
DROP TABLE IF EXISTS plans CASCADE;
DROP FUNCTION IF EXISTS set_billing_prune_date() CASCADE;

-- 2. Simplify events table (drop old event_type constraints and column)
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_flash_single_day_check;
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_event_type_check;
ALTER TABLE events DROP COLUMN IF EXISTS event_type;

-- 3. Create organization_billing
CREATE TABLE organization_billing (
    organization_id      UUID PRIMARY KEY REFERENCES "organization" (id) ON DELETE CASCADE,
    credit_balance       INT NOT NULL DEFAULT 1 CHECK (credit_balance >= 0),
    annual_fee_status    VARCHAR(32) NOT NULL DEFAULT 'free' CHECK (annual_fee_status IN ('free', 'active', 'past_due')),
    current_period_start DATE,
    current_period_end   DATE,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT false,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Initialize organization_billing for any existing organizations with 1 credit
INSERT INTO organization_billing (organization_id, credit_balance, annual_fee_status)
SELECT id, 1, 'free' FROM "organization"
ON CONFLICT (organization_id) DO NOTHING;

-- Trigger to automatically initialize billing with 1 free credit for any new organization
CREATE OR REPLACE FUNCTION initialize_organization_billing()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO organization_billing (organization_id, credit_balance, annual_fee_status)
    VALUES (NEW.id, 1, 'free')
    ON CONFLICT (organization_id) DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_init_org_billing ON "organization";
CREATE TRIGGER trg_init_org_billing
AFTER INSERT ON "organization"
FOR EACH ROW
EXECUTE FUNCTION initialize_organization_billing();

-- 4. Create billing_transactions
CREATE TABLE billing_transactions (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id      UUID NOT NULL REFERENCES "organization" (id) ON DELETE CASCADE,
    type                 VARCHAR(32) NOT NULL CHECK (type IN ('credit_purchase', 'credit_consumed', 'annual_renewal')),
    credits_delta        INT NOT NULL DEFAULT 0,
    amount               BIGINT NOT NULL DEFAULT 0,
    currency             VARCHAR(8) NOT NULL DEFAULT 'INR',
    event_id             UUID REFERENCES events (id) ON DELETE SET NULL,
    description          TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_billing_transactions_org ON billing_transactions (organization_id, created_at DESC);
