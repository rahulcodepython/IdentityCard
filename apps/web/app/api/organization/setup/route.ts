import { NextRequest, NextResponse } from "next/server"

import { auth, pool } from "@/lib/auth"
import { ROLE_ADMIN } from "@/lib/constants"

function slugify(name: string) {
    const base = name
        .toLowerCase()
        .trim()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/^-+|-+$/g, "")
    return base || "org"
}

export async function POST(req: NextRequest) {
    try {
        const body = await req.json().catch(() => ({}))
        const organizationName = (body.organizationName as string | undefined)?.trim() ?? ""
        const userEmail = (body.userEmail as string | undefined)?.trim()?.toLowerCase()

        if (!organizationName) {
            return NextResponse.json(
                { error: "Organization name is required." },
                { status: 400 }
            )
        }

        let userId: string | null = null

        try {
            const session = await auth.api.getSession({ headers: req.headers })
            if (session?.user?.id) {
                userId = session.user.id
            }
        } catch {
            // Fall back to userEmail lookup
        }

        if (!userId && userEmail) {
            const userRes = await pool.query<{ id: string }>(
                `SELECT id FROM "user" WHERE lower(email) = lower($1) LIMIT 1`,
                [userEmail]
            )
            if (userRes.rows.length > 0) {
                userId = userRes.rows[0].id
            }
        }

        if (!userId) {
            return NextResponse.json(
                { error: "Not signed in." },
                { status: 401 }
            )
        }

        const existingMember = await pool.query<{ organizationId: string }>(
            `SELECT "organizationId" FROM "member" WHERE "userId" = $1 LIMIT 1`,
            [userId]
        )

        let orgId: string
        if (existingMember.rows.length > 0) {
            orgId = existingMember.rows[0].organizationId
            await pool.query(
                `UPDATE "organization" SET "name" = $1 WHERE "id" = $2`,
                [organizationName, orgId]
            )
        } else {
            const slug = `${slugify(organizationName)}-${userId.slice(0, 8)}`
            const orgRes = await pool.query<{ id: string }>(
                `INSERT INTO "organization" ("name", "slug") VALUES ($1, $2) RETURNING "id"`,
                [organizationName, slug]
            )
            orgId = orgRes.rows[0].id
            await pool.query(
                `INSERT INTO "member" ("organizationId", "userId", "role") VALUES ($1, $2, $3)`,
                [orgId, userId, ROLE_ADMIN]
            )
        }

        try {
            await auth.api.setActiveOrganization({
                headers: req.headers,
                body: { organizationId: orgId },
            })
        } catch {
            // Non-critical, session reloads on dashboard entry
        }

        return NextResponse.json({ success: true, organizationId: orgId })
    } catch (err) {
        console.error("Failed to setup organization:", err)
        return NextResponse.json(
            { error: "Failed to setup organization." },
            { status: 500 }
        )
    }
}
