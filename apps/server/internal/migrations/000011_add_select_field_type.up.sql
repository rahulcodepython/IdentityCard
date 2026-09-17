-- Insert 'select' into field_types reference table
INSERT INTO field_types (type, label, category, default_config) VALUES
    ('select', 'Dropdown Select', 'choice', '{"has_options": true, "placeholder": "Choose an option"}'::jsonb)
ON CONFLICT (type) DO NOTHING;
