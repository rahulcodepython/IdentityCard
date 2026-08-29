-- better-auth's schema — generated via `npx auth generate` against
-- apps/web/lib/auth.ts (jwt + organization + emailOTP + twoFactor
-- plugins), then hand-edited:
--   * advanced.database.generateId: "uuid" already makes every id column
--     `uuid default gen_random_uuid()` instead of better-auth's normal
--     text id, matching the rest of this schema.
--   * session.activeOrganizationId is widened from the generator's
--     default `text` to `uuid` (with an FK) for the same reason.
--   * organization.logoObjectKey/onboardingCompletedAt are this app's
--     additionalFields (see lib/auth.ts) — org settings live on this one
--     table instead of a separate Go-owned satellite table.
--
-- Table/column names are camelCase, quoted — that's the org plugin's own
-- convention (Kysely "camel" casing), left as-is rather than fighting it
-- with modelName/fields overrides. Everything else in this schema is
-- snake_case; this is the one deliberate exception, confined to the
-- tables better-auth owns.
CREATE TABLE "user" (
    "id"               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "name"             TEXT NOT NULL,
    "email"            TEXT NOT NULL UNIQUE,
    "emailVerified"    BOOLEAN NOT NULL,
    "image"            TEXT,
    "createdAt"        TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updatedAt"        TIMESTAMPTZ NOT NULL DEFAULT now(),
    "twoFactorEnabled" BOOLEAN
);

CREATE TABLE "session" (
    "id"                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "expiresAt"            TIMESTAMPTZ NOT NULL,
    "token"                TEXT NOT NULL UNIQUE,
    "createdAt"            TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updatedAt"            TIMESTAMPTZ NOT NULL,
    "ipAddress"            TEXT,
    "userAgent"            TEXT,
    "userId"               UUID NOT NULL REFERENCES "user" ("id") ON DELETE CASCADE,
    "activeOrganizationId" UUID
);

CREATE TABLE "account" (
    "id"                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "issuer"                TEXT NOT NULL,
    "accountId"             TEXT NOT NULL,
    "providerId"            TEXT NOT NULL,
    "userId"                UUID NOT NULL REFERENCES "user" ("id") ON DELETE CASCADE,
    "accessToken"           TEXT,
    "refreshToken"          TEXT,
    "idToken"               TEXT,
    "accessTokenExpiresAt"  TIMESTAMPTZ,
    "refreshTokenExpiresAt" TIMESTAMPTZ,
    "scope"                 TEXT,
    "password"              TEXT,
    "createdAt"             TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updatedAt"             TIMESTAMPTZ NOT NULL
);

CREATE TABLE "verification" (
    "id"         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "identifier" TEXT NOT NULL,
    "value"      TEXT NOT NULL,
    "expiresAt"  TIMESTAMPTZ NOT NULL,
    "createdAt"  TIMESTAMPTZ NOT NULL DEFAULT now(),
    "updatedAt"  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE "jwks" (
    "id"         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "publicKey"  TEXT NOT NULL,
    "privateKey" TEXT NOT NULL,
    "createdAt"  TIMESTAMPTZ NOT NULL,
    "expiresAt"  TIMESTAMPTZ,
    "alg"        TEXT,
    "crv"        TEXT
);

CREATE TABLE "organization" (
    "id"                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "name"                  TEXT NOT NULL,
    "slug"                  TEXT NOT NULL UNIQUE,
    "logo"                  TEXT,
    "createdAt"             TIMESTAMPTZ NOT NULL DEFAULT now(),
    "metadata"              TEXT,
    "logoObjectKey"         TEXT,
    "onboardingCompletedAt" TIMESTAMPTZ
);

CREATE TABLE "member" (
    "id"             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "organizationId" UUID NOT NULL REFERENCES "organization" ("id") ON DELETE CASCADE,
    "userId"         UUID NOT NULL REFERENCES "user" ("id") ON DELETE CASCADE,
    "role"           TEXT NOT NULL,
    "createdAt"      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE "invitation" (
    "id"             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "organizationId" UUID NOT NULL REFERENCES "organization" ("id") ON DELETE CASCADE,
    "email"          TEXT NOT NULL,
    "role"           TEXT,
    "status"         TEXT NOT NULL,
    "expiresAt"      TIMESTAMPTZ NOT NULL,
    "createdAt"      TIMESTAMPTZ NOT NULL DEFAULT now(),
    "inviterId"      UUID NOT NULL REFERENCES "user" ("id") ON DELETE CASCADE
);

CREATE TABLE "twoFactor" (
    "id"                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "secret"                  TEXT NOT NULL,
    "backupCodes"             TEXT NOT NULL,
    "userId"                  UUID NOT NULL REFERENCES "user" ("id") ON DELETE CASCADE,
    "verified"                BOOLEAN,
    "failedVerificationCount" INTEGER,
    "lockedUntil"             TIMESTAMPTZ
);

ALTER TABLE "session"
    ADD CONSTRAINT session_active_organization_id_fkey
    FOREIGN KEY ("activeOrganizationId") REFERENCES "organization" ("id") ON DELETE SET NULL;

CREATE INDEX session_user_id_idx ON "session" ("userId");
CREATE INDEX account_user_id_idx ON "account" ("userId");
CREATE INDEX verification_identifier_idx ON "verification" ("identifier");
CREATE INDEX member_organization_id_idx ON "member" ("organizationId");
CREATE INDEX member_user_id_idx ON "member" ("userId");
CREATE INDEX invitation_organization_id_idx ON "invitation" ("organizationId");
CREATE INDEX invitation_email_idx ON "invitation" ("email");
CREATE INDEX two_factor_secret_idx ON "twoFactor" ("secret");
CREATE INDEX two_factor_user_id_idx ON "twoFactor" ("userId");
CREATE UNIQUE INDEX account_issuer_account_id_uidx ON "account" ("issuer", "accountId");
