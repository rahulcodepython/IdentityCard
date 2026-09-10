import { NextRequest, NextResponse } from "next/server"
import { cookies } from "next/headers"
import { symmetricDecrypt, makeSignature } from "better-auth/crypto"
import { createOTP } from "@better-auth/utils/otp"

import { pool, auth } from "@/lib/auth"

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
            [cleanEmail]
        )

        if (userRes.rows.length === 0) {
            return NextResponse.json(
                { success: false, error: "No account found with this email." },
                { status: 404 }
            )
        }

        const user = userRes.rows[0]
        if (!user.twoFactorEnabled || !user.secret || !user.verified) {
            return NextResponse.json(
                { success: false, error: "Authenticator (TOTP) is not set up on this account. Please sign in with Email OTP." },
                { status: 400 }
            )
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
            return NextResponse.json(
                { success: false, error: "Invalid authenticator code. Please check your app and try again." },
                { status: 400 }
            )
        }

        const memberRes = await pool.query<{ organizationId: string }>(
            `SELECT "organizationId" FROM "member" WHERE "userId" = $1 LIMIT 1`,
            [user.id]
        )
        const activeOrgId = memberRes.rows[0]?.organizationId ?? null

        const authContext = await auth.$context
        const session = await authContext.internalAdapter.createSession(user.id)
        if (activeOrgId) {
            await pool.query(
                `UPDATE "session" SET "activeOrganizationId" = $1 WHERE "token" = $2`,
                [activeOrgId, session.token]
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

        return NextResponse.json({ success: true })
    } catch (err) {
        console.error("Failed to sign in with TOTP:", err)
        return NextResponse.json(
            { success: false, error: "Something went wrong during sign in." },
            { status: 500 }
        )
    }
}
