import { NextRequest } from "next/server"
import { z } from "zod"

import { auth, pool } from "@/lib/auth"
import { ROLE_OWNER } from "@/lib/constants"
import {
    BadRequestResponse,
    InternalServerErrorResponse,
    SuccessResponse,
    UnauthorizedResponse,
} from "@/lib/response"

const setupOrganizationSchema = z.object({
    organizationName: z.string().trim().min(1, "Organization name is required."),
    userEmail: z.string().trim().email().optional(),
})

interface OrganizationSetupCteResult {
    org_id: string
}

function slugify(name: string): string {
    const base = name
        .toLowerCase()
        .trim()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/^-+|-+$/g, "")
    return base || "org"
}

export async function POST(req: NextRequest) {
    try {
        const body: unknown = await req.json().catch(() => ({}))
        const parsed = setupOrganizationSchema.safeParse(body)

        if (!parsed.success) {
            return BadRequestResponse(parsed.error.issues[0]?.message ?? "Invalid request body.")
        }

        const { organizationName, userEmail } = parsed.data

        let userId: string | null = null

        const session = await auth.api.getSession({ headers: req.headers })
        if (session?.user?.id) {
            userId = session.user.id
        }

        if (!userId && !userEmail) {
            return UnauthorizedResponse("Not signed in.")
        }

        const baseSlug = slugify(organizationName)

        // Single SQL round-trip CTE:
        // 1. Resolves target user ID (from session or email)
        // 2. Checks existing membership
        // 3. Conditionally updates existing org OR inserts new org and member in one atomic execution
        const setupQuery = `
            WITH target_user AS (
                SELECT id 
                FROM "user"
                WHERE ($1::text IS NOT NULL AND id = $1)
                   OR ($1::text IS NULL AND lower(email) = lower($2))
                LIMIT 1
            ),
            existing_member AS (
                SELECT m."organizationId", tu.id AS user_id
                FROM target_user tu
                LEFT JOIN "member" m ON m."userId" = tu.id
                LIMIT 1
            ),
            update_existing_org AS (
                UPDATE "organization"
                SET "name" = $3
                WHERE id = (SELECT "organizationId" FROM existing_member WHERE "organizationId" IS NOT NULL)
                RETURNING id
            ),
            insert_new_org AS (
                INSERT INTO "organization" ("id", "name", "slug")
                SELECT 
                    gen_random_uuid(), 
                    $3, 
                    $4 || '-' || substr(user_id::text, 1, 8)
                FROM existing_member
                WHERE "organizationId" IS NULL
                RETURNING id
            ),
            resolved_org AS (
                SELECT id AS org_id FROM update_existing_org
                UNION ALL
                SELECT id AS org_id FROM insert_new_org
                LIMIT 1
            ),
            insert_new_member AS (
                INSERT INTO "member" ("id", "organizationId", "userId", "role", "createdAt")
                SELECT 
                    gen_random_uuid(), 
                    ro.org_id, 
                    em.user_id, 
                    $5, 
                    now()
                FROM existing_member em
                CROSS JOIN resolved_org ro
                WHERE em."organizationId" IS NULL
                RETURNING id
            )
            SELECT org_id FROM resolved_org;
        `

        const result = await pool.query<OrganizationSetupCteResult>(setupQuery, [
            userId,             // $1
            userEmail ?? null,  // $2
            organizationName,   // $3
            baseSlug,           // $4
            ROLE_OWNER,         // $5
        ])

        if (result.rows.length === 0) {
            return UnauthorizedResponse("Not signed in.")
        }

        await auth.api.setActiveOrganization({
            headers: req.headers,
            body: { organizationId: result.rows[0].org_id },
        })

        return SuccessResponse({ success: true, organizationId: result.rows[0].org_id })
    } catch (err) {
        return InternalServerErrorResponse("Failed to setup organization.")
    }
}