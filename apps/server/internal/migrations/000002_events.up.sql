CREATE TABLE events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CHECK (end_date >= start_date)
);
