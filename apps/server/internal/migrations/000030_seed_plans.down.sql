DELETE FROM plans WHERE code IN (
    'flash', 'base_monthly', 'base_yearly', 'custom_monthly', 'custom_yearly',
    'unlimited_monthly', 'unlimited_yearly'
);
