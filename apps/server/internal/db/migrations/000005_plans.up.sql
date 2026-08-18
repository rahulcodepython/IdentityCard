CREATE TABLE plans (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code         TEXT NOT NULL UNIQUE,
    kind         TEXT NOT NULL CHECK (kind IN ('flash', 'standard')),
    tier         TEXT NOT NULL, -- e.g. very_small, small, medium, large, yearly, recurring, custom
    name         TEXT NOT NULL,
    price_config JSONB NOT NULL DEFAULT '{}'::JSONB, -- billing-engine detail, defined in a later phase
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
