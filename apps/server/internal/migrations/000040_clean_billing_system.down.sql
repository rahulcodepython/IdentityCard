-- 000040_clean_billing_system.down.sql
DROP TABLE IF EXISTS billing_transactions CASCADE;
DROP TRIGGER IF EXISTS trg_init_org_billing ON "organization";
DROP FUNCTION IF EXISTS initialize_organization_billing() CASCADE;
DROP TABLE IF EXISTS organization_billing CASCADE;
