import { NextRequest } from "next/server"
import { cookies } from "next/headers"
import { symmetricDecrypt, makeSignature, generateRandomString } from "better-auth/crypto"
import { createOTP } from "@better-auth/utils/otp"
import { z } from "zod"

import { pool } from "@/lib/auth"
import { ROLE_OWNER } from "@/lib/constants"
import { BadRequestResponse, InternalServerErrorResponse, SuccessResponse } from "@/lib/response"

const requestSchema = z.object({
    email: z.string().email(),
    code: z.string().length(6),
})

const registrationPayloadSchema = z.object({
    name: z.string().optional(),
    organizationName: z.string().optional(),
    encryptedSecret: z.string(),
})

interface VerificationRow {
    value: string
}

interface RegistrationCteResult {
    user_id: string
    organization_id: string
}

export async function POST(req: NextRequest) {
    try {
        const body = await req.json()
        const parsedBody = requestSchema.safeParse(body)

        if (!parsedBody.success) {
            return BadRequestResponse("Invalid email or code format.")
        }

        const identifier = `totp-reg-${parsedBody.data.email}`

        // Single Read Round-Trip: Fetch and parse verification payload
        const verRes = await pool.query<VerificationRow>(
            `SELECT value FROM "verification"
             WHERE identifier = $1 AND "expiresAt" > now()
             ORDER BY "createdAt" DESC LIMIT 1`,
            [identifier]
        )

        if (verRes.rows.length === 0) {
            return BadRequestResponse("Registration session expired or not found. Please start over.")
        }

        let parsedJson: unknown
        try {
            parsedJson = JSON.parse(verRes.rows[0].value)
        } catch {
            return BadRequestResponse("Invalid registration data. Please start over.")
        }

        const parsedPayload = registrationPayloadSchema.safeParse(parsedJson)
        if (!parsedPayload.success) {
            return BadRequestResponse("Invalid registration data. Please start over.")
        }

        const secret = await symmetricDecrypt({
            key: process.env.BETTER_AUTH_SECRET!,
            data: parsedPayload.data.encryptedSecret,
        })

        const isValid = await createOTP(secret, {
            digits: 6,
            period: 30,
        }).verify(parsedBody.data.code)

        if (!isValid) {
            return BadRequestResponse("Invalid authenticator code. Please check your app and try again.")
        }

        // Generate session attributes and slug prefix in application layer
        const sessionToken = generateRandomString(32)
        const sessionExpiresAt = new Date(Date.now() + 7 * 24 * 60 * 60 * 1000)
        const rawOrgName = parsedPayload.data.organizationName
        const rawUserName = parsedPayload.data.name || "User"
        const orgDisplayName = rawOrgName || `${rawUserName}'s Organization`
        const baseSlug = (rawOrgName || rawUserName || "org")
            .toLowerCase()
            .replace(/[^a-z0-9]/g, "-")
            .replace(/-+/g, "-")
            .replace(/^-|-$/g, "") || "org"

        // Single Write Round-Trip: All mutations orchestrated via CTE
        const registrationQuery = `
            WITH upsert_user AS (
                INSERT INTO "user" ("id", "name", "email", "emailVerified", "twoFactorEnabled", "createdAt", "updatedAt")
                VALUES (gen_random_uuid(), $1, lower($2), true, true, now(), now())
                ON CONFLICT ("email") DO UPDATE
                SET "twoFactorEnabled" = true,
                    "emailVerified" = true,
                    "updatedAt" = now()
                RETURNING id
            ),
            upsert_two_factor AS (
                INSERT INTO "twoFactor" ("id", "secret", "backupCodes", "userId", "verified")
                SELECT gen_random_uuid(), $3, '[]', id, true
                FROM upsert_user
                ON CONFLICT ("userId") DO UPDATE
                SET "secret" = EXCLUDED."secret",
                    "verified" = true
                RETURNING id
            ),
            existing_member AS (
                SELECT m."organizationId", u.id AS user_id
                FROM upsert_user u
                LEFT JOIN "member" m ON m."userId" = u.id
                LIMIT 1
            ),
            update_existing_org AS (
                UPDATE "organization"
                SET "name" = $4
                WHERE id = (SELECT "organizationId" FROM existing_member WHERE "organizationId" IS NOT NULL)
                  AND $5::text IS NOT NULL
                RETURNING id
            ),
            insert_new_org AS (
                INSERT INTO "organization" ("id", "name", "slug")
                SELECT gen_random_uuid(), $4, $6 || '-' || substr(user_id::text, 1, 8)
                FROM existing_member
                WHERE "organizationId" IS NULL
                RETURNING id
            ),
            resolved_org AS (
                SELECT id AS org_id FROM update_existing_org
                UNION ALL
                SELECT id AS org_id FROM insert_new_org
                UNION ALL
                SELECT "organizationId" AS org_id FROM existing_member WHERE "organizationId" IS NOT NULL AND $5::text IS NULL
                LIMIT 1
            ),
            insert_new_member AS (
                INSERT INTO "member" ("id", "organizationId", "userId", "role", "createdAt")
                SELECT gen_random_uuid(), ro.org_id, em.user_id, $7, now()
                FROM existing_member em
                CROSS JOIN resolved_org ro
                WHERE em."organizationId" IS NULL
                RETURNING id
            ),
            insert_session AS (
                INSERT INTO "session" ("id", "userId", "token", "expiresAt", "activeOrganizationId", "createdAt", "updatedAt")
                SELECT gen_random_uuid(), em.user_id, $8, $9, ro.org_id, now(), now()
                FROM existing_member em
                CROSS JOIN resolved_org ro
                RETURNING id
            ),
            delete_verification AS (
                DELETE FROM "verification"
                WHERE identifier = $10
                RETURNING id
            )
            SELECT 
                (SELECT user_id FROM existing_member) AS user_id,
                (SELECT org_id FROM resolved_org) AS organization_id;
        `

        const result = await pool.query<RegistrationCteResult>(registrationQuery, [
            rawUserName,                           // $1
            parsedBody.data.email,                 // $2
            parsedPayload.data.encryptedSecret,    // $3
            orgDisplayName,                        // $4
            rawOrgName ?? null,                    // $5
            baseSlug,                              // $6
            ROLE_OWNER,                            // $7
            sessionToken,                          // $8
            sessionExpiresAt,                      // $9
            identifier,                            // $10
        ])

        if (result.rows.length === 0) {
            return InternalServerErrorResponse("Failed to process account setup.")
        }

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
        console.error("Failed to complete TOTP registration:", err)
        return InternalServerErrorResponse("Something went wrong while completing registration.")
    }
}