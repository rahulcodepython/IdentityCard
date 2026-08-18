-- Placeholder catalog + prices so the plans page and registration flow have
-- something real to select from. Amounts are illustrative (INR, smallest
-- unit i.e. paise) — swap via a follow-up migration once real pricing is
-- decided; the "custom" tier intentionally carries no amount (sales-assisted).
INSERT INTO plans (code, kind, tier, name, price_config) VALUES
    ('flash', 'flash', 'flash', 'Flash Event',
        '{"amount": 199900, "currency": "INR", "billing": "per_event"}'),
    ('standard_very_small', 'standard', 'very_small', 'Very Small (up to 5 days)',
        '{"amount": 299900, "currency": "INR", "billing": "one_time"}'),
    ('standard_small', 'standard', 'small', 'Small (up to 7 days)',
        '{"amount": 499900, "currency": "INR", "billing": "one_time"}'),
    ('standard_medium', 'standard', 'medium', 'Medium (1-2 months)',
        '{"amount": 999900, "currency": "INR", "billing": "one_time"}'),
    ('standard_large', 'standard', 'large', 'Large (6 months)',
        '{"amount": 2499900, "currency": "INR", "billing": "one_time"}'),
    ('standard_yearly', 'standard', 'yearly', 'Yearly',
        '{"amount": 4999900, "currency": "INR", "billing": "yearly"}'),
    ('standard_recurring', 'standard', 'recurring', 'Recurring (up to 20 events/yr)',
        '{"amount": 7999900, "currency": "INR", "billing": "yearly", "event_limit": 20}'),
    ('standard_custom', 'standard', 'custom', 'Custom / Unlimited',
        '{"currency": "INR", "billing": "custom"}');
