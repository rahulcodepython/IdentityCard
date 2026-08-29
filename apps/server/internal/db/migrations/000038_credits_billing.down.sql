DROP TABLE credits;
DROP TABLE billing;
DROP TABLE transactions;

ALTER TABLE plans DROP CONSTRAINT plans_flash_daily_check;
ALTER TABLE plans DROP CONSTRAINT plans_billing_cycle_check;
UPDATE plans SET billing_cycle = 'one_time' WHERE kind = 'flash';
ALTER TABLE plans ADD CONSTRAINT plans_billing_cycle_check
    CHECK (billing_cycle IN ('monthly', 'yearly', 'one_time'));
ALTER TABLE plans ADD CONSTRAINT plans_check
    CHECK (kind <> 'flash' OR billing_cycle = 'one_time');

ALTER TABLE plans DROP COLUMN nominal_increment;

CREATE TABLE subscriptions (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id      UUID NOT NULL REFERENCES "organization" (id) ON DELETE CASCADE,
    plan_id              UUID NOT NULL REFERENCES plans (id),
    kind                 TEXT NOT NULL CHECK (kind IN ('flash', 'base', 'custom', 'unlimited')),
    billing_cycle        TEXT NOT NULL CHECK (billing_cycle IN ('monthly', 'yearly', 'one_time')),
    status               TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'past_due', 'expired')),
    event_quota          INT,
    started_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    current_period_end   DATE,
    past_due_since       DATE,
    grace_deadline       DATE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX subscriptions_organization_id_idx ON subscriptions (organization_id);
CREATE INDEX subscriptions_status_idx ON subscriptions (status);

CREATE TABLE subscription_periods (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES subscriptions (id) ON DELETE CASCADE,
    period_start    DATE NOT NULL,
    period_end      DATE NOT NULL,
    paid_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (subscription_id, period_start)
);

ALTER TABLE events ADD COLUMN subscription_id UUID REFERENCES subscriptions (id);
CREATE INDEX events_subscription_id_idx ON events (subscription_id);
