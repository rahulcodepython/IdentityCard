-- Full replacement of the plan/subscription/addon model — see CLAUDE.md:
-- this is still prototype phase, so old rows are dropped rather than
-- migrated forward. Addons are folded into the new model outright: a
-- Custom-kind subscription purchase now IS the "buy more event capacity"
-- action addons used to provide (see 000027_subscriptions), so there is
-- no addons replacement table.
DROP TABLE IF EXISTS subscription_addons;
DROP TABLE IF EXISTS addons;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS plans;

CREATE TABLE plans (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code              TEXT NOT NULL UNIQUE,
    kind              TEXT NOT NULL CHECK (kind IN ('flash', 'base', 'custom', 'unlimited')),
    billing_cycle     TEXT NOT NULL CHECK (billing_cycle IN ('monthly', 'yearly', 'one_time')),
    name              TEXT NOT NULL,
    amount            BIGINT,          -- paise; NULL only for kind='custom' (amount = per_event_amount * quantity, computed at purchase time)
    per_event_amount  BIGINT,          -- only set for kind='custom'
    currency          TEXT NOT NULL DEFAULT 'INR',
    event_quota       INT,             -- 1 for flash/base; NULL for custom (caller supplies quantity at purchase) and unlimited (no cap)
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CHECK (kind <> 'flash' OR billing_cycle = 'one_time'),
    CHECK (kind = 'custom' OR amount IS NOT NULL)
);
