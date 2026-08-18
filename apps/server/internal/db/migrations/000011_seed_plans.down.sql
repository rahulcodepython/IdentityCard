DELETE FROM plans WHERE code IN (
    'flash', 'standard_very_small', 'standard_small', 'standard_medium',
    'standard_large', 'standard_yearly', 'standard_recurring', 'standard_custom'
);
