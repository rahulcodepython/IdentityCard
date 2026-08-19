-- History of paid billing cycles for a subscription — mirrors the
-- "materialize history via rows, don't mutate" convention already used
-- for event_days under events (000013). Only written for monthly/yearly
-- subscriptions; a one_time (Flash) purchase never gets a period row.
CREATE TABLE subscription_periods (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES subscriptions (id) ON DELETE CASCADE,
    period_start    DATE NOT NULL,
    period_end      DATE NOT NULL,
    paid_at         TIMESTAMPTZ NOT NULL DEFAULT now(), -- always "now" — stubbed, no real payment/pending state yet, same as subscriptions themselves
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (subscription_id, period_start)
);

CREATE INDEX subscription_periods_subscription_id_idx ON subscription_periods (subscription_id);
