"use server"

import { headers } from "next/headers"

import { auth, pool } from "@/lib/auth"
import { ROLE_ADMIN } from "@/lib/roles"

function slugify(name: string) {
    const base = name
        .toLowerCase()
        .trim()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/^-+|-+$/g, "")
    return base || "org"
}

export async function setRegistrationOrganization(organizationName: string, userEmail?: string) {
    const cleanName = organizationName.trim()
    if (!cleanName) return

    const reqHeaders = await headers()
    let userId: string | null = null

    try {
        const session = await auth.api.getSession({ headers: reqHeaders })
        if (session?.user?.id) {
            userId = session.user.id
        }
    } catch {
        // Continue to userEmail fallback
    }

    if (!userId && userEmail) {
        const userRes = await pool.query<{ id: string }>(
            `SELECT id FROM "user" WHERE lower(email) = lower($1) LIMIT 1`,
            [userEmail.trim().toLowerCase()]
        )
        if (userRes.rows.length > 0) {
            userId = userRes.rows[0].id
        }
    }

    if (!userId) throw new Error("Not signed in.")

    const existingMember = await pool.query<{ organizationId: string }>(
        `SELECT "organizationId" FROM "member" WHERE "userId" = $1 LIMIT 1`,
        [userId]
    )

    let orgId: string
    if (existingMember.rows.length > 0) {
        orgId = existingMember.rows[0].organizationId
        await pool.query(
            `UPDATE "organization" SET "name" = $1 WHERE "id" = $2`,
            [cleanName, orgId]
        )
    } else {
        const slug = `${slugify(cleanName)}-${userId.slice(0, 8)}`
        const orgRes = await pool.query<{ id: string }>(
            `INSERT INTO "organization" ("name", "slug") VALUES ($1, $2) RETURNING "id"`,
            [cleanName, slug]
        )
        orgId = orgRes.rows[0].id
        await pool.query(
            `INSERT INTO "member" ("organizationId", "userId", "role") VALUES ($1, $2, $3)`,
            [orgId, userId, ROLE_ADMIN]
        )
    }

    await pool.query(
        `UPDATE "session" SET "activeOrganizationId" = $1 WHERE "userId" = $2`,
        [orgId, userId]
    )

    try {
        await auth.api.setActiveOrganization({
            headers: reqHeaders,
            body: { organizationId: orgId },
        })
    } catch {
        // Handled via DB session update above
    }
}
