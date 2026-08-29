import "server-only"

import { betterAuth } from "better-auth"
import { jwt, organization, emailOTP, twoFactor } from "better-auth/plugins"
import { nextCookies } from "better-auth/next-js"
import { Pool } from "pg"

import { ac, roles } from "@/lib/auth-access-control"
import { ROLE_SUPER_ADMIN } from "@/lib/roles"
import { sendMail } from "@/lib/mailer"

// Shared with definePayload below (looks up a session's org role directly
// via SQL rather than round-tripping through auth.api) and available for
// any other Node-side code that needs the same Postgres connection
// better-auth itself uses.
export const pool = new Pool({ connectionString: process.env.DATABASE_URL })

const otpEmailCopy: Record<string, { subject: string; body: (otp: string) => string }> = {
    "sign-in": {
        subject: "Your IdentityCard sign-in code",
        body: (otp) => `Your sign-in code is ${otp}. It expires in 5 minutes.`,
    },
    "email-verification": {
        subject: "Verify your email",
        body: (otp) => `Your verification code is ${otp}. It expires in 5 minutes.`,
    },
    "forget-password": {
        subject: "Reset your password",
        body: (otp) => `Your password reset code is ${otp}. It expires in 5 minutes.`,
    },
    "change-email": {
        subject: "Confirm your new email",
        body: (otp) => `Your email change code is ${otp}. It expires in 5 minutes.`,
    },
}

export const auth = betterAuth({
    baseURL: process.env.BETTER_AUTH_URL,
    secret: process.env.BETTER_AUTH_SECRET,
    database: pool,
    advanced: {
        database: {
            // Keeps every better-auth table's id column a native Postgres
            // uuid (via gen_random_uuid()) instead of better-auth's default
            // string id — the whole Go/sqlc codebase is uuid.UUID-typed
            // throughout, so this avoids a disruptive type mismatch at
            // every FK boundary between the two schemas.
            generateId: "uuid",
        },
    },
    emailAndPassword: {
        enabled: false,
    },
    socialProviders: {
        google: {
            clientId: process.env.GOOGLE_CLIENT_ID || "",
            clientSecret: process.env.GOOGLE_CLIENT_SECRET || "",
        },
    },
    databaseHooks: {
        user: {
            create: {
                after: async (user) => {
                    const orgName = `${user.name || "My"}'s Organization`
                    const baseSlug = (user.name || "org")
                        .toLowerCase()
                        .replace(/[^a-z0-9]/g, "-")
                        .replace(/-+/g, "-")
                        .replace(/^-|-$/g, "") || "org"
                    const slug = `${baseSlug}-${user.id.slice(0, 8)}`

                    try {
                        const orgRes = await pool.query<{ id: string }>(
                            `INSERT INTO "organization" ("name", "slug") VALUES ($1, $2) RETURNING "id"`,
                            [orgName, slug]
                        )
                        const orgId = orgRes.rows[0].id
                        await pool.query(
                            `INSERT INTO "member" ("organizationId", "userId", "role") VALUES ($1, $2, $3)`,
                            [orgId, user.id, ROLE_SUPER_ADMIN]
                        )
                    } catch (err) {
                        console.error("Failed to auto-create default organization for user:", err)
                    }
                },
            },
        },
    },
    plugins: [
        jwt({
            jwks: {
                keyPairConfig: { alg: "EdDSA", crv: "Ed25519" },
            },
            jwt: {
                expirationTime: "15m",
                definePayload: async ({ user, session }) => {
                    const organizationId = (session as { activeOrganizationId?: string }).activeOrganizationId
                    let role: string | null = null
                    if (organizationId) {
                        const result = await pool.query<{ role: string }>(
                            `SELECT role FROM "member" WHERE "organizationId" = $1 AND "userId" = $2 LIMIT 1`,
                            [organizationId, user.id]
                        )
                        role = result.rows[0]?.role ?? null
                    }
                    return {
                        email: user.email,
                        name: user.name,
                        organizationId: organizationId ?? null,
                        role,
                    }
                },
            },
        }),
        organization({
            ac,
            roles,
            creatorRole: ROLE_SUPER_ADMIN,
            // One org per user, ever — a second org can only come from a
            // second user account. Actual creation only ever happens
            // server-side, as the result of a successful plan purchase
            // (see the PayKit checkout-success handler, Phase 6/7) — never
            // from a user-facing "create organization" form.
            organizationLimit: 1,
            schema: {
                organization: {
                    additionalFields: {
                        logoObjectKey: { type: "string", required: false },
                        onboardingCompletedAt: { type: "date", required: false },
                    },
                },
            },
            async sendInvitationEmail({ id, email, organization, inviter }) {
                const url = `${process.env.BETTER_AUTH_URL}/accept-invitation?id=${id}`
                await sendMail(
                    email,
                    `You've been invited to join ${organization.name} on IdentityCard`,
                    `${inviter.user.name} invited you to join ${organization.name}. Accept the invitation: ${url}`
                )
            },
        }),
        emailOTP({
            otpLength: 6,
            expiresIn: 300,
            async sendVerificationOTP({ email, otp, type }) {
                const copy = otpEmailCopy[type] ?? otpEmailCopy["sign-in"]
                await sendMail(email, copy.subject, copy.body(otp))
            },
        }),
        twoFactor({
            issuer: "IdentityCard",
            // The app is passwordless (no emailAndPassword credential
            // account ever exists), so enable/disable/verify must not
            // require a password — see shouldRequirePassword in
            // better-auth's own utils/password.ts.
            allowPasswordless: true,
        }),
        // Must be last — lets Server Actions/Route Handlers set the
        // session cookie via next/headers instead of a raw Response.
        nextCookies(),
    ],
})
