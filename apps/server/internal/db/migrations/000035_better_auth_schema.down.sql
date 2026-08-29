ALTER TABLE "session" DROP CONSTRAINT IF EXISTS session_active_organization_id_fkey;

DROP TABLE IF EXISTS "twoFactor";
DROP TABLE IF EXISTS "invitation";
DROP TABLE IF EXISTS "member";
DROP TABLE IF EXISTS "organization";
DROP TABLE IF EXISTS "jwks";
DROP TABLE IF EXISTS "verification";
DROP TABLE IF EXISTS "account";
DROP TABLE IF EXISTS "session";
DROP TABLE IF EXISTS "user";
