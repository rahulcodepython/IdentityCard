import { NextRequest } from "next/server"
import { cookies } from "next/headers"
import { symmetricDecrypt, makeSignature, generateRandomString } from "better-auth/crypto"
import { createOTP } from "@better-auth/utils/otp"
import { z } from "zod"

import { pool } from "@/lib/auth"
import {
    BadRequestResponse,
    NotFoundResponse,
    InternalServerErrorResponse,
    SuccessResponse,
} from "@/lib/response"

const requestSchema = z.object({
    email: z.string().email(),
    code: z.string().min(6).max(6),
})

interface TotpAuthQueryRow {
    status: "USER_NOT_FOUND" | "TOTP_NOT_CONFIGURED" | "READY"
    user_id: string | null
    secret: string | null
    active_organization_id: string | null
}

export async function POST(req: NextRequest) {
    try {
        const body: unknown = await req.json().catch(() => ({}))
        const parsedBody = requestSchema.safeParse(body)

        if (!parsedBody.success) {
            return BadRequestResponse("Email and 6-digit code are required.")
        }

        const cleanEmail = parsedBody.data.email.trim().toLowerCase()
        const cleanCode = parsedBody.data.code.trim()

        // Single SQL round-trip CTE: checks existence, 2FA readiness, and resolves member organization
        const query = `
            WITH matched_user AS (
                SELECT 
                    u.id, 
                    u."twoFactorEnabled", 
                    tf.secret, 
                    tf.verified,
                    m."organizationId" AS active_org_id
                FROM "user" u
                LEFT JOIN "twoFactor" tf ON tf."userId" = u.id
                LEFT JOIN "member" m ON m."userId" = u.id
                WHERE lower(u.email) = lower($1)
                LIMIT 1
            )
            SELECT
                CASE
                    WHEN NOT EXISTS (SELECT 1 FROM matched_user) THEN 'USER_NOT_FOUND'
                    WHEN (SELECT "twoFactorEnabled" FROM matched_user) != true 
                         OR (SELECT secret FROM matched_user) IS NULL 
                         OR (SELECT verified FROM matched_user) != true THEN 'TOTP_NOT_CONFIGURED'
                    ELSE 'READY'
                END AS status,
                (SELECT id FROM matched_user) AS user_id,
                (SELECT secret FROM matched_user) AS secret,
                (SELECT active_org_id FROM matched_user) AS active_organization_id;
        `

        const res = await pool.query<TotpAuthQueryRow>(query, [cleanEmail])
        const row = res.rows[0]

        if (!row || row.status === "USER_NOT_FOUND") {
            return NotFoundResponse("No account found with this email.")
        }

        if (row.status === "TOTP_NOT_CONFIGURED" || !row.secret || !row.user_id) {
            return BadRequestResponse("Authenticator (TOTP) is not set up on this account. Please sign in with Email OTP.")
        }

        const decryptedSecret = await symmetricDecrypt({
            key: process.env.BETTER_AUTH_SECRET!,
            data: row.secret,
        })

        const isValid = await createOTP(decryptedSecret, {
            digits: 6,
            period: 30,
        }).verify(cleanCode)

        if (!isValid) {
            return BadRequestResponse("Invalid authenticator code. Please check your app and try again.")
        }

        const sessionToken = generateRandomString(32)
        const sessionExpiresAt = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000)

        // Single query insert for the resolved session
        await pool.query(
            `INSERT INTO "session" ("id", "userId", "token", "expiresAt", "activeOrganizationId", "createdAt", "updatedAt")
             VALUES (gen_random_uuid(), $1, $2, $3, $4, now(), now())`,
            [row.user_id, sessionToken, sessionExpiresAt, row.active_organization_id]
        )

        const signedToken = `${sessionToken}.${await makeSignature(sessionToken, process.env.BETTER_AUTH_SECRET!)}`
        const cookieStore = await cookies()
        cookieStore.set("better-auth.session_token", signedToken, {
            httpOnly: true,
            sameSite: "lax",
            path: "/",
            secure: process.env.NODE_ENV === "production",
            maxAge: 60 * 60 * 24 * 7,
        })

        return SuccessResponse({ success: true })
    } catch (err) {
        console.error("Failed to sign in with TOTP:", err)
        return InternalServerErrorResponse("Something went wrong during sign in.")
    }
}