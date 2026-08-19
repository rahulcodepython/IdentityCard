-- One row per purchase. Unlike the old model, an org can hold many
-- concurrently 'active' subscriptions — buying a Custom top-up while
-- already on Base/Custom/Unlimited never cancels anything, it just adds
-- another row (see plans.Service.Subscribe). Total usable event capacity
-- is the sum of event_quota across an org's 'active' subscriptions (any
-- active 'unlimited' one overrides the sum entirely — see
-- plans.Service.EventCreationContext-equivalent logic in
-- events.Service.checkEventLimit). Each row tracks its own billing cycle
-- and grace/deletion timeline independently, so a lapsed subscription only
-- ever costs the org the events it personally funded.
CREATE TABLE subscriptions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    plan_id             UUID NOT NULL REFERENCES plans (id),
    kind                TEXT NOT NULL CHECK (kind IN ('flash', 'base', 'custom', 'unlimited')), -- denormalized from plans.kind for cheap filtering in sweep queries
    billing_cycle       TEXT NOT NULL CHECK (billing_cycle IN ('monthly', 'yearly', 'one_time')),
    status              TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'past_due', 'expired')),
    event_quota         INT,           -- copied from plans.event_quota at purchase time (NULL = unlimited)
    started_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    current_period_end  DATE,          -- NULL for one_time (Flash) subscriptions, which never renew
    past_due_since      DATE,          -- set once current_period_end passes unpaid; cleared on renew
    grace_deadline      DATE,          -- past_due_since + 1 month (monthly) / +6 months (yearly); NULL for Flash, whose 1-month post-event retention deadline is computed on the fly from its event's end_date (see internal/jobs) rather than stored here
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX subscriptions_organization_id_idx ON subscriptions (organization_id);
CREATE INDEX subscriptions_status_idx ON subscriptions (status);
