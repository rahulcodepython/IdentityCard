-- Replaces 000011_seed_plans (that data is gone — 000026 dropped the old
-- plans table outright). Amounts are illustrative placeholders (INR,
-- smallest unit i.e. paise), same spirit as 000011's original disclaimer —
-- swap via a follow-up migration once real pricing is decided; low stakes
-- for now since checkout is stubbed (no real payment gateway yet).
INSERT INTO plans (code, kind, billing_cycle, name, amount, per_event_amount, event_quota) VALUES
    ('flash',              'flash',     'one_time', 'Flash',              199900,  NULL,  1),
    ('base_monthly',       'base',      'monthly',  'Base (Monthly)',     299900,  NULL,  1),
    ('base_yearly',        'base',      'yearly',   'Base (Yearly)',      2999900, NULL,  1),
    ('custom_monthly',     'custom',    'monthly',  'Custom (Monthly)',   NULL,    49900, NULL),
    ('custom_yearly',      'custom',    'yearly',   'Custom (Yearly)',    NULL,    449900, NULL),
    ('unlimited_monthly',  'unlimited', 'monthly',  'Unlimited (Monthly)', 999900, NULL,  NULL),
    ('unlimited_yearly',   'unlimited', 'yearly',   'Unlimited (Yearly)', 9999900, NULL,  NULL);
