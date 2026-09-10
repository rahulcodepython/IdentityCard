import { NextRequest, NextResponse } from "next/server"
import { symmetricEncrypt, generateRandomString } from "better-auth/crypto"
import { createOTP } from "@better-auth/utils/otp"

import { pool } from "@/lib/auth"

const ISSUER = "IdentityCard"

export async function POST(req: NextRequest) {
    try {
        const body = await req.json().catch(() => ({}))
        const cleanEmail = (body.email as string | undefined)?.trim().toLowerCase() ?? ""
        const cleanName = (body.name as string | undefined)?.trim() ?? ""
        const cleanOrg = (body.organizationName as string | undefined)?.trim() ?? ""

        if (!cleanEmail || !cleanName || !cleanOrg) {
            return NextResponse.json(
                { success: false, error: "Please provide your name, organization name, and email." },
                { status: 400 }
            )
        }

        const userRes = await pool.query<{ id: string; twoFactorEnabled: boolean }>(
            `SELECT id, "twoFactorEnabled" FROM "user" WHERE lower(email) = lower($1) LIMIT 1`,
            [cleanEmail]
        )
        if (userRes.rows.length > 0 && userRes.rows[0].twoFactorEnabled) {
            return NextResponse.json(
                { success: false, error: "An account with this email already has two-factor authentication enabled. Please sign in instead." },
                { status: 400 }
            )
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
            [identifier, payload, expiresAt]
        )

        return NextResponse.json({
            success: true,
            totpURI,
            secret,
        })
    } catch (err) {
        console.error("Failed to prepare TOTP registration:", err)
        return NextResponse.json(
            { success: false, error: "Unable to initialize authenticator registration. Please try again." },
            { status: 500 }
        )
    }
}
