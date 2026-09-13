import { NextRequest, NextResponse } from "next/server"
import { symmetricEncrypt, generateRandomString } from "better-auth/crypto"
import { createOTP } from "@better-auth/utils/otp"
import { z } from "zod"

import { pool } from "@/lib/auth"
import { BadRequestResponse, InternalServerErrorResponse, SuccessResponse } from "@/lib/response"

const ISSUER = "IdentityCard"

const requestBodySchema = z.object({
    email: z.string().email(),
    name: z.string().min(1),
    organizationName: z.string().min(1),
})

export async function POST(req: NextRequest) {
    try {
        const body = await req.json()
        const parsedBody = requestBodySchema.safeParse(body)

        if (!parsedBody.success) {
            return BadRequestResponse("Invalid input. Please provide a valid email, name, and organization name.")
        }

        const secret = generateRandomString(32)
        const otp = createOTP(secret, { digits: 6, period: 30 })
        const totpURI = otp.url(ISSUER, parsedBody.data.email)

        const encryptedSecret = await symmetricEncrypt({
            key: process.env.BETTER_AUTH_SECRET!,
            data: secret,
        })

        const identifier = `totp-reg-${parsedBody.data.email}`
        const expiresAt = new Date(Date.now() + 15 * 60 * 1000)
        const payload = JSON.stringify({
            email: parsedBody.data.email,
            name: parsedBody.data.name,
            organizationName: parsedBody.data.organizationName,
            encryptedSecret,
        })

        const query = `
            WITH existing_user AS (
                SELECT "twoFactorEnabled"
                FROM "user"
                WHERE lower(email) = lower($1)
                LIMIT 1
            ),
            inserted_verification AS (
                INSERT INTO "verification" ("id", "identifier", "value", "expiresAt", "createdAt", "updatedAt")
                SELECT gen_random_uuid(), $2, $3, $4, now(), now()
                WHERE NOT EXISTS (
                    SELECT 1 FROM existing_user WHERE "twoFactorEnabled" = true
                )
                RETURNING id
            )
            SELECT
                CASE 
                    WHEN (SELECT "twoFactorEnabled" FROM existing_user) = true THEN '2FA_ALREADY_ENABLED'
                    WHEN (SELECT count(*) FROM inserted_verification) > 0 THEN 'SUCCESS'
                    ELSE 'INSERT_FAILED'
                END AS status;
        `

        const result = await pool.query<{ status: "2FA_ALREADY_ENABLED" | "SUCCESS" | "INSERT_FAILED" }>(
            query,
            [parsedBody.data.email, identifier, payload, expiresAt]
        )

        const status = result.rows[0]?.status

        switch (status) {
            case "2FA_ALREADY_ENABLED":
                return BadRequestResponse("Two-factor authentication is already enabled for this user.")
            case "INSERT_FAILED":
                return InternalServerErrorResponse("Unable to initialize authenticator registration. Please try again.")
            case "SUCCESS":
                return SuccessResponse({ totpURI, secret })
            default:
                return InternalServerErrorResponse("Unexpected error occurred. Please try again.")
        }
    } catch (err) {
        console.error("Failed to prepare TOTP registration:", err)
        return InternalServerErrorResponse("Unable to initialize authenticator registration. Please try again.")
    }
}