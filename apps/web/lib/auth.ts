import "server-only"

import { betterAuth } from "better-auth"
import { admin, jwt, organization, emailOTP, twoFactor } from "better-auth/plugins"
import { nextCookies } from "better-auth/next-js"
import { Pool } from "pg"
import { ROLE_OWNER } from "@/lib/constants"
import { EmailOTPTemplate } from "@/lib/email"

export const pool = new Pool({ connectionString: process.env.DATABASE_URL })

const SECRET = process.env.BETTER_AUTH_SECRET!!

export const auth = betterAuth({
    baseURL: process.env.BETTER_AUTH_URL,
    secret: SECRET,
    database: pool,
    advanced: {
        database: {
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
                        await pool.query(
                            `WITH new_org AS (
                                INSERT INTO "organization" ("name", "slug") 
                                VALUES ($1, $2) 
                                RETURNING "id"
                            )
                            INSERT INTO "member" ("organizationId", "userId", "role")
                            SELECT "id", $3, $4 
                            FROM new_org`,
                            [orgName, slug, user.id, ROLE_OWNER]
                        )
                    } catch (err) {
                        console.error("Failed to auto-create default organization for user:", err)
                    }
                },
            },
        },
    },
    plugins: [
        admin({
            defaultRole: "user",
            adminRole: "admin",
        }),
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
                        try {
                            const result = await pool.query<{ role: string }>(
                                `SELECT role FROM "member" WHERE "organizationId" = $1 AND "userId" = $2 LIMIT 1`,
                                [organizationId, user.id]
                            )
                            role = result.rows[0]?.role ?? null
                        } catch {
                            role = null
                        }
                    }
                    return {
                        id: user.id,
                        email: user.email,
                        name: user.name,
                        organizationId: organizationId ?? null,
                        role,
                        userRole: (user as { role?: string }).role ?? "user",
                    }
                },
            },
        }),
        organization({
            creatorRole: ROLE_OWNER,
            async sendInvitationEmail({ id, email, organization, inviter }) {
                const url = `${process.env.BETTER_AUTH_URL}/accept-invitation?id=${id}`
                console.info(`[auth:invitation] To: ${email} | Org: ${organization.name} | Inviter: ${inviter.user.name} | URL: ${url}`)
            },
        }),
        emailOTP({
            otpLength: 6,
            expiresIn: 300,
            async sendVerificationOTP({ email, otp, type }) {
                const copy = EmailOTPTemplate[type] ?? EmailOTPTemplate["sign-in"]
                console.info(`[auth:otp] To: ${email} | Type: ${type} | Code: ${otp} | Subject: ${copy.subject}`)
            },
        }),
        twoFactor({
            issuer: "IdentityCard",
            allowPasswordless: true,
        }),
        nextCookies(),
    ],
})
