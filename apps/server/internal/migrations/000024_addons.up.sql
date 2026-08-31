-- Add-ons are purchased on top of an existing subscription (unlike
-- plans.kind='flash', which is a standalone plan chosen at registration
-- time) — e.g. a org on a standard plan buying a one-off Flash Event
-- add-on without switching plans.
CREATE TABLE addons (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code         TEXT NOT NULL UNIQUE,
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    price_config JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One row per purchase (stubbed — no real payment call yet, same as
-- subscriptions.CreateSubscription). An org can buy the same add-on more
-- than once (e.g. two separate flash events), so there's no unique
-- constraint on (organization_id, addon_id).
CREATE TABLE subscription_addons (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    addon_id        UUID NOT NULL REFERENCES addons (id),
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active')),
    purchased_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX subscription_addons_organization_id_idx ON subscription_addons (organization_id);

INSERT INTO addons (code, name, description, price_config) VALUES
    ('flash_addon', 'Flash Event Add-on',
     'Run one additional flash (single-day) event on top of your current plan.',
     '{"amount": 199900, "currency": "INR", "billing": "per_event"}');
