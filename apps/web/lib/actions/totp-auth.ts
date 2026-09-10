"use server"

import { cookies } from "next/headers"
import {
    symmetricDecrypt,
    symmetricEncrypt,
    generateRandomString,
    makeSignature,
} from "better-auth/crypto"
import { createOTP } from "@better-auth/utils/otp"

import { pool, auth } from "@/lib/auth"
import { ROLE_ADMIN } from "@/lib/roles"

const ISSUER = "IdentityCard"

export type PrepareTotpResult =
    | {
          success: true
          totpURI: string
          secret: string
      }
    | {
          success: false
          error: string
      }

export async function prepareTotpRegistration({
    email,
    name,
    organizationName,
}: {
    email: string
    name: string
    organizationName: string
}): Promise<PrepareTotpResult> {
    const cleanEmail = email.trim().toLowerCase()
    const cleanName = name.trim()
    const cleanOrg = organizationName.trim()

    if (!cleanEmail || !cleanName || !cleanOrg) {
        return {
            success: false,
            error: "Please provide your name, organization name, and email.",
        }
    }

    try {
        const userRes = await pool.query<{
            id: string
            twoFactorEnabled: boolean
        }>(
            `SELECT id, "twoFactorEnabled" FROM "user" WHERE lower(email) = lower($1) LIMIT 1`,
            [cleanEmail],
        )
        if (userRes.rows.length > 0 && userRes.rows[0].twoFactorEnabled) {
            return {
                success: false,
                error: "An account with this email already has two-factor authentication enabled. Please sign in instead.",
            }
        }

        const secret = generateRandomString(32)
        const otp = createOTP(secret, { digits: 6, period: 30 })
        const totpURI = otp.url(ISSUER, cleanEmail)

        const encryptedSecret = await symmetricEncrypt({
            key: process.env.BETTER_AUTH_SECRET!,
            data: secret,
        })

        const identifier = `totp-reg-${cleanEmail}`
        const expiresAt = new Date(Date.now() + 15 * 60 * 1000)
        const payload = JSON.stringify({
            email: cleanEmail,
            name: cleanName,
            organizationName: cleanOrg,
            encryptedSecret,
        })

        await pool.query(
            `INSERT INTO "verification" ("id", "identifier", "value", "expiresAt", "createdAt", "updatedAt")
             VALUES (gen_random_uuid(), $1, $2, $3, now(), now())`,
            [identifier, payload, expiresAt],
        )

        return {
            success: true,
            totpURI,
            secret,
        }
    } catch (err) {
        console.error("Failed to prepare TOTP registration:", err)
        return {
            success: false,
            error: "Unable to initialize authenticator registration. Please try again.",
        }
    }
}

export type CompleteTotpResult =
    | {
          success: true
      }
    | {
          success: false
          error: string
      }

export async function completeTotpRegistration({
    email,
    code,
}: {
    email: string
    code: string
}): Promise<CompleteTotpResult> {
    const cleanEmail = email.trim().toLowerCase()
    const cleanCode = code.trim()

    if (!cleanEmail || !cleanCode) {
        return { success: false, error: "Email and code are required." }
    }

    try {
        const verRes = await pool.query<{ id: string; value: string }>(
            `SELECT id, value FROM "verification"
             WHERE identifier = $1 AND "expiresAt" > now()
             ORDER BY "createdAt" DESC LIMIT 1`,
            [`totp-reg-${cleanEmail}`],
        )

        if (verRes.rows.length === 0) {
            return {
                success: false,
                error: "Registration session expired or not found. Please start over.",
            }
        }

        const { name, organizationName, encryptedSecret } = JSON.parse(
            verRes.rows[0].value,
        ) as {
            name: string
            organizationName: string
            encryptedSecret: string
        }

        const secret = await symmetricDecrypt({
            key: process.env.BETTER_AUTH_SECRET!,
            data: encryptedSecret,
        })

        const isValid = await createOTP(secret, {
            digits: 6,
            period: 30,
        }).verify(cleanCode)
        if (!isValid) {
            return {
                success: false,
                error: "Invalid authenticator code. Please check your app and try again.",
            }
        }

        const userRes = await pool.query<{ id: string }>(
            `SELECT id FROM "user" WHERE lower(email) = lower($1) LIMIT 1`,
            [cleanEmail],
        )
        let userId: string

        if (userRes.rows.length === 0) {
            const insertUserRes = await pool.query<{ id: string }>(
                `INSERT INTO "user" ("name", "email", "emailVerified", "twoFactorEnabled", "createdAt", "updatedAt")
                 VALUES ($1, $2, true, true, now(), now())
                 RETURNING id`,
                [name || "User", cleanEmail],
            )
            userId = insertUserRes.rows[0].id
        } else {
            userId = userRes.rows[0].id
            await pool.query(
                `UPDATE "user" SET "twoFactorEnabled" = true, "emailVerified" = true, "updatedAt" = now() WHERE id = $1`,
                [userId],
            )
        }

        const existingTf = await pool.query<{ id: string }>(
            `SELECT id FROM "twoFactor" WHERE "userId" = $1 LIMIT 1`,
            [userId],
        )
        if (existingTf.rows.length > 0) {
            await pool.query(
                `UPDATE "twoFactor" SET "secret" = $1, "verified" = true WHERE id = $2`,
                [encryptedSecret, existingTf.rows[0].id],
            )
        } else {
            await pool.query(
                `INSERT INTO "twoFactor" ("secret", "backupCodes", "userId", "verified") VALUES ($1, '[]', $2, true)`,
                [encryptedSecret, userId],
            )
        }

        const existingMember = await pool.query<{ organizationId: string }>(
            `SELECT "organizationId" FROM "member" WHERE "userId" = $1 LIMIT 1`,
            [userId],
        )
        let activeOrgId: string
        const orgDisplayName =
            organizationName || `${name || "My"}'s Organization`

        if (existingMember.rows.length > 0) {
            activeOrgId = existingMember.rows[0].organizationId
            if (organizationName) {
                await pool.query(
                    `UPDATE "organization" SET "name" = $1 WHERE "id" = $2`,
                    [organizationName, activeOrgId],
                )
            }
        } else {
            const baseSlug =
                (organizationName || name || "org")
                    .toLowerCase()
                    .replace(/[^a-z0-9]/g, "-")
                    .replace(/-+/g, "-")
                    .replace(/^-|-$/g, "") || "org"
            const slug = `${baseSlug}-${userId.slice(0, 8)}`
            const orgRes = await pool.query<{ id: string }>(
                `INSERT INTO "organization" ("name", "slug") VALUES ($1, $2) RETURNING "id"`,
                [orgDisplayName, slug],
            )
            activeOrgId = orgRes.rows[0].id
            await pool.query(
                `INSERT INTO "member" ("organizationId", "userId", "role") VALUES ($1, $2, $3)`,
                [activeOrgId, userId, ROLE_ADMIN],
            )
        }

        const authContext = await auth.$context
        const session = await authContext.internalAdapter.createSession(userId)
        await pool.query(
            `UPDATE "session" SET "activeOrganizationId" = $1 WHERE "token" = $2`,
            [activeOrgId, session.token],
        )

        const signedToken = `${session.token}.${await makeSignature(session.token, process.env.BETTER_AUTH_SECRET!)}`
        const cookieStore = await cookies()
        cookieStore.set("better-auth.session_token", signedToken, {
            httpOnly: true,
            sameSite: "lax",
            path: "/",
            secure: process.env.NODE_ENV === "production",
            maxAge: 60 * 60 * 24 * 7,
        })

        await pool.query(`DELETE FROM "verification" WHERE identifier = $1`, [
            `totp-reg-${cleanEmail}`,
        ])

        return { success: true }
    } catch (err) {
        console.error("Failed to complete TOTP registration:", err)
        return {
            success: false,
            error: "Something went wrong while completing registration.",
        }
    }
}

export type SignInWithTotpResult =
    | {
          success: true
      }
    | {
          success: false
          error: string
      }

export async function signInWithTotp({
    email,
    code,
}: {
    email: string
    code: string
}): Promise<SignInWithTotpResult> {
    const cleanEmail = email.trim().toLowerCase()
    const cleanCode = code.trim()

    if (!cleanEmail || !cleanCode) {
        return { success: false, error: "Email and code are required." }
    }

    try {
        const userRes = await pool.query<{
            id: string
            email: string
            name: string
            twoFactorEnabled: boolean
            secret: string | null
            verified: boolean | null
        }>(
            `SELECT u.id, u.email, u.name, u."twoFactorEnabled", tf.secret, tf.verified
             FROM "user" u
             LEFT JOIN "twoFactor" tf ON tf."userId" = u.id
             WHERE lower(u.email) = lower($1)
             LIMIT 1`,
            [cleanEmail],
        )

        if (userRes.rows.length === 0) {
            return {
                success: false,
                error: "No account found with this email.",
            }
        }

        const user = userRes.rows[0]
        if (!user.twoFactorEnabled || !user.secret || !user.verified) {
            return {
                success: false,
                error: "Authenticator (TOTP) is not set up on this account. Please sign in with Email OTP.",
            }
        }

        const secret = await symmetricDecrypt({
            key: process.env.BETTER_AUTH_SECRET!,
            data: user.secret,
        })

        const isValid = await createOTP(secret, {
            digits: 6,
            period: 30,
        }).verify(cleanCode)
        if (!isValid) {
            return {
                success: false,
                error: "Invalid authenticator code. Please check your app and try again.",
            }
        }

        const memberRes = await pool.query<{ organizationId: string }>(
            `SELECT "organizationId" FROM "member" WHERE "userId" = $1 LIMIT 1`,
            [user.id],
        )
        const activeOrgId = memberRes.rows[0]?.organizationId ?? null

        const authContext = await auth.$context
        const session = await authContext.internalAdapter.createSession(user.id)
        if (activeOrgId) {
            await pool.query(
                `UPDATE "session" SET "activeOrganizationId" = $1 WHERE "token" = $2`,
                [activeOrgId, session.token],
            )
        }

        const signedToken = `${session.token}.${await makeSignature(session.token, process.env.BETTER_AUTH_SECRET!)}`
        const cookieStore = await cookies()
        cookieStore.set("better-auth.session_token", signedToken, {
            httpOnly: true,
            sameSite: "lax",
            path: "/",
            secure: process.env.NODE_ENV === "production",
            maxAge: 60 * 60 * 24 * 7,
        })

        return { success: true }
    } catch (err) {
        console.error("Failed to sign in with TOTP:", err)
        return { success: false, error: "Something went wrong during sign in." }
    }
}
