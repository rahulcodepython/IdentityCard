-- Replaces subscriptions/subscription_periods with a credits-based
-- purchase model: a purchase mints one or more single-use credits (each
-- funds exactly one event), and a separate billing table tracks
-- recurring payment obligations independently of how credits were
-- originally granted. Prototype phase (CLAUDE.md): straight
-- drop/recreate.
TRUNCATE TABLE events CASCADE;

ALTER TABLE events DROP CONSTRAINT events_subscription_id_fkey;
ALTER TABLE events DROP COLUMN subscription_id;

DROP TABLE subscription_periods;
DROP TABLE subscriptions;

ALTER TABLE plans ADD COLUMN nominal_increment BIGINT NOT NULL DEFAULT 0;

-- Flash billing now genuinely recurs (once a day — see the credits
-- model's flash-renewal rule), so it's no longer a true "one_time"
-- product; billing_cycle gets a 'daily' value for it instead.
ALTER TABLE plans DROP CONSTRAINT plans_check;
ALTER TABLE plans DROP CONSTRAINT plans_billing_cycle_check;
UPDATE plans SET billing_cycle = 'daily' WHERE kind = 'flash';
ALTER TABLE plans ADD CONSTRAINT plans_billing_cycle_check
    CHECK (billing_cycle IN ('daily', 'monthly', 'yearly'));
ALTER TABLE plans ADD CONSTRAINT plans_flash_daily_check
    CHECK (kind <> 'flash' OR billing_cycle = 'daily');

-- One row per completed payment — both an initial purchase and every
-- later renewal payment.
CREATE TABLE transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES "organization" (id) ON DELETE CASCADE,
    plan_id         UUID NOT NULL REFERENCES plans (id),
    amount          BIGINT NOT NULL,
    currency        TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX transactions_organization_id_idx ON transactions (organization_id);

-- One row per billing period, chained into a lineage via
-- lineage_root_id: on the first row of a lineage this equals its own id;
-- every renewal row copies the same value forward, so "current status of
-- this lineage" is always the latest row where lineage_root_id matches.
CREATE TABLE billing (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES "organization" (id) ON DELETE CASCADE,
    plan_id         UUID NOT NULL REFERENCES plans (id),
    lineage_root_id UUID NOT NULL,
    billing_number  INT NOT NULL CHECK (billing_number >= 1),
    period_start    DATE NOT NULL,
    period_end      DATE NOT NULL CHECK (period_end > period_start),
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'pending', 'paid', 'cancel')),
    amount          BIGINT NOT NULL,
    paid_at         TIMESTAMPTZ,
    transaction_id  UUID REFERENCES transactions (id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE billing ADD CONSTRAINT billing_lineage_root_id_fkey
    FOREIGN KEY (lineage_root_id) REFERENCES billing (id);

CREATE INDEX billing_organization_id_idx ON billing (organization_id);
CREATE INDEX billing_lineage_root_id_idx ON billing (lineage_root_id);

-- One row per single-use event-creation credit. billing_id points at a
-- lineage_root_id value (the stable lineage anchor), not a specific
-- billing period row.
CREATE TABLE credits (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id   UUID NOT NULL REFERENCES "organization" (id) ON DELETE CASCADE,
    billing_id        UUID NOT NULL REFERENCES billing (id),
    type              TEXT NOT NULL CHECK (type IN ('flash', 'events', 'all')),
    event_id          UUID UNIQUE REFERENCES events (id) ON DELETE SET NULL,
    is_restricted     BOOLEAN NOT NULL DEFAULT false,
    restricted_since  TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX credits_organization_id_idx ON credits (organization_id);
CREATE INDEX credits_billing_id_idx ON credits (billing_id);
-- The FIFO, unconsumed, non-restricted lookup checkEventLimit's
-- replacement runs on every event creation.
CREATE INDEX credits_available_idx ON credits (organization_id, type, created_at)
    WHERE event_id IS NULL AND is_restricted = false;
