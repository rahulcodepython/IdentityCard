-- IF EXISTS: 000026_plans.up.sql also drops these tables on its way up
-- (addons/subscription_addons were folded into the new plan model), so a
-- full `migrate down -all` reaches this migration with them already gone.
DROP TABLE IF EXISTS subscription_addons;
DROP TABLE IF EXISTS addons;
