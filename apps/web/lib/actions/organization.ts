"use server"

import { headers } from "next/headers"

import { auth, pool } from "@/lib/auth"
import { ROLE_SUPER_ADMIN } from "@/lib/roles"

function slugify(name: string) {
    const base = name
        .toLowerCase()
        .trim()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/^-+|-+$/g, "")
    return base || "org"
}

export async function setRegistrationOrganization(organizationName: string) {
    const cleanName = organizationName.trim()
    if (!cleanName) return

    const reqHeaders = await headers()
    const session = await auth.api.getSession({ headers: reqHeaders })
    if (!session) throw new Error("Not signed in.")

    const existingMember = await pool.query<{ organizationId: string }>(
        `SELECT "organizationId" FROM "member" WHERE "userId" = $1 LIMIT 1`,
        [session.user.id]
    )

    if (existingMember.rows.length > 0) {
        const orgId = existingMember.rows[0].organizationId
        await pool.query(
            `UPDATE "organization" SET "name" = $1 WHERE "id" = $2`,
            [cleanName, orgId]
        )
        await auth.api.setActiveOrganization({
            headers: reqHeaders,
            body: { organizationId: orgId },
        })
    } else {
        const slug = `${slugify(cleanName)}-${session.user.id.slice(0, 8)}`
        const orgRes = await pool.query<{ id: string }>(
            `INSERT INTO "organization" ("name", "slug") VALUES ($1, $2) RETURNING "id"`,
            [cleanName, slug]
        )
        const orgId = orgRes.rows[0].id
        await pool.query(
            `INSERT INTO "member" ("organizationId", "userId", "role") VALUES ($1, $2, $3)`,
            [orgId, session.user.id, ROLE_SUPER_ADMIN]
        )
        await auth.api.setActiveOrganization({
            headers: reqHeaders,
            body: { organizationId: orgId },
        })
    }
}

