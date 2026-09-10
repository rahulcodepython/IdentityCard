import { NextRequest, NextResponse } from "next/server"
import { cookies } from "next/headers"
import { symmetricDecrypt, makeSignature } from "better-auth/crypto"
import { createOTP } from "@better-auth/utils/otp"

import { pool, auth } from "@/lib/auth"
import { ROLE_ADMIN } from "@/lib/constants"

export async function POST(req: NextRequest) {
    try {
        const body = await req.json().catch(() => ({}))
        const cleanEmail = (body.email as string | undefined)?.trim().toLowerCase() ?? ""
        const cleanCode = (body.code as string | undefined)?.trim() ?? ""

        if (!cleanEmail || !cleanCode) {
            return NextResponse.json(
                { success: false, error: "Email and code are required." },
                { status: 400 }
            )
        }

        const verRes = await pool.query<{ id: string; value: string }>(
            `SELECT id, value FROM "verification"
             WHERE identifier = $1 AND "expiresAt" > now()
             ORDER BY "createdAt" DESC LIMIT 1`,
            [`totp-reg-${cleanEmail}`]
        )

        if (verRes.rows.length === 0) {
            return NextResponse.json(
                { success: false, error: "Registration session expired or not found. Please start over." },
                { status: 400 }
            )
        }

        const { name, organizationName, encryptedSecret } = JSON.parse(
            verRes.rows[0].value
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
            return NextResponse.json(
                { success: false, error: "Invalid authenticator code. Please check your app and try again." },
                { status: 400 }
            )
        }

        const userRes = await pool.query<{ id: string }>(
            `SELECT id FROM "user" WHERE lower(email) = lower($1) LIMIT 1`,
            [cleanEmail]
        )
        let userId: string

        if (userRes.rows.length === 0) {
            const insertUserRes = await pool.query<{ id: string }>(
                `INSERT INTO "user" ("name", "email", "emailVerified", "twoFactorEnabled", "createdAt", "updatedAt")
                 VALUES ($1, $2, true, true, now(), now())
                 RETURNING id`,
                [name || "User", cleanEmail]
            )
            userId = insertUserRes.rows[0].id
        } else {
            userId = userRes.rows[0].id
            await pool.query(
                `UPDATE "user" SET "twoFactorEnabled" = true, "emailVerified" = true, "updatedAt" = now() WHERE id = $1`,
                [userId]
            )
        }

        const existingTf = await pool.query<{ id: string }>(
            `SELECT id FROM "twoFactor" WHERE "userId" = $1 LIMIT 1`,
            [userId]
        )
        if (existingTf.rows.length > 0) {
            await pool.query(
                `UPDATE "twoFactor" SET "secret" = $1, "verified" = true WHERE id = $2`,
                [encryptedSecret, existingTf.rows[0].id]
            )
        } else {
            await pool.query(
                `INSERT INTO "twoFactor" ("secret", "backupCodes", "userId", "verified") VALUES ($1, '[]', $2, true)`,
                [encryptedSecret, userId]
            )
        }

        const existingMember = await pool.query<{ organizationId: string }>(
            `SELECT "organizationId" FROM "member" WHERE "userId" = $1 LIMIT 1`,
            [userId]
        )
        let activeOrgId: string
        const orgDisplayName = organizationName || `${name || "My"}'s Organization`

        if (existingMember.rows.length > 0) {
            activeOrgId = existingMember.rows[0].organizationId
            if (organizationName) {
                await pool.query(
                    `UPDATE "organization" SET "name" = $1 WHERE "id" = $2`,
                    [organizationName, activeOrgId]
                )
            }
        } else {
            const baseSlug = (organizationName || name || "org")
                .toLowerCase()
                .replace(/[^a-z0-9]/g, "-")
                .replace(/-+/g, "-")
                .replace(/^-|-$/g, "") || "org"
            const slug = `${baseSlug}-${userId.slice(0, 8)}`
            const orgRes = await pool.query<{ id: string }>(
                `INSERT INTO "organization" ("name", "slug") VALUES ($1, $2) RETURNING "id"`,
                [orgDisplayName, slug]
            )
            activeOrgId = orgRes.rows[0].id
            await pool.query(
                `INSERT INTO "member" ("organizationId", "userId", "role") VALUES ($1, $2, $3)`,
                [activeOrgId, userId, ROLE_ADMIN]
            )
        }

        const authContext = await auth.$context
        const session = await authContext.internalAdapter.createSession(userId)
        await pool.query(
            `UPDATE "session" SET "activeOrganizationId" = $1 WHERE "token" = $2`,
            [activeOrgId, session.token]
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

        return NextResponse.json({ success: true })
    } catch (err) {
        console.error("Failed to complete TOTP registration:", err)
        return NextResponse.json(
            { success: false, error: "Something went wrong while completing registration." },
            { status: 500 }
        )
    }
}
