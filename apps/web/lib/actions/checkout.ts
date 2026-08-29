"use server"

import { headers } from "next/headers"

import { auth } from "@/lib/auth"
import { paykit } from "@/lib/paykit"
import { subscriptionResponseSchema } from "@/schema/plans.types"

const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8080"

function slugify(name: string) {
    const base = name
        .toLowerCase()
        .trim()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/^-+|-+$/g, "")
    return base || "org"
}

// Runs the whole "buy a plan" flow server-side, in order:
//  1. If the signed-in user has no organization yet, create one now —
//     this is the ONLY place an organization ever gets created (see
//     decision #7 in the auth/billing rewrite: no standalone "create
//     organization" form anywhere). better-auth's org plugin auto-sets
//     it active and makes the creator super_admin (creatorRole, see
//     lib/auth.ts) — org creation itself already enforces one org per
//     user (organizationLimit: 1).
//  2. Run the checkout through PayKit (lib/paykit.ts) — today that's the
//     manual/instant-success provider; swapping in a live gateway later
//     only touches that file, not this one.
//  3. Mint a fresh bearer JWT (now carrying the new org id + role) and
//     call Go's POST /plans/subscriptions — Go remains the system of
//     record for what an org purchased and its event quota.
export async function purchasePlan(input: {
    planCode: string
    eventQuantity?: number
    organizationName?: string
}) {
    const reqHeaders = await headers()
    const session = await auth.api.getSession({ headers: reqHeaders })
    if (!session) throw new Error("Not signed in.")

    if (!session.session.activeOrganizationId) {
        if (!input.organizationName?.trim()) {
            throw new Error("Organization name is required.")
        }

        let slug = slugify(input.organizationName)
        try {
            await auth.api.checkOrganizationSlug({ headers: reqHeaders, body: { slug } })
        } catch {
            slug = `${slug}-${Math.random().toString(36).slice(2, 6)}`
        }

        await auth.api.createOrganization({
            headers: reqHeaders,
            body: { name: input.organizationName.trim(), slug },
        })
    }

    await paykit.checkouts.create({
        session_type: "one_time",
        customer: { email: session.user.email },
        item_id: input.planCode,
        quantity: input.eventQuantity ?? 1,
        success_url: "/dashboard/billing",
        cancel_url: "/select-plan",
        metadata: { planCode: input.planCode },
    })

    const { token } = await auth.api.getToken({ headers: reqHeaders })

    const res = await fetch(`${API_BASE_URL}/plans/subscriptions`, {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify({ plan_code: input.planCode, event_quantity: input.eventQuantity }),
    })
    const body = await res.json().catch(() => null)
    if (!res.ok) {
        throw new Error(body?.error?.message ?? "Subscription failed.")
    }
    return subscriptionResponseSchema.parse(body.data)
}
