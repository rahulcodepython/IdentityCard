"use server";

import { headers } from "next/headers";

import { auth } from "@/lib/auth";
import { API_V1_PREFIX } from "@/lib/constants";
import { paykit } from "@/lib/paykit";
import { billingResponseSchema } from "@/schema/plans.types";

const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8080";

function slugify(name: string) {
    const base = name
        .toLowerCase()
        .trim()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/^-+|-+$/g, "");
    return base || "org";
}

export async function purchasePlan(input: {
    planCode: string;
    eventQuantity?: number;
    organizationName?: string;
}) {
    const reqHeaders = await headers();
    const session = await auth.api.getSession({ headers: reqHeaders });
    if (!session) throw new Error("Not signed in.");

    if (!session.session.activeOrganizationId) {
        if (!input.organizationName?.trim()) {
            throw new Error("Organization name is required.");
        }

        let slug = slugify(input.organizationName);
        try {
            await auth.api.checkOrganizationSlug({ headers: reqHeaders, body: { slug } });
        } catch {
            slug = `${slug}-${Math.random().toString(36).slice(2, 6)}`;
        }

        await auth.api.createOrganization({
            headers: reqHeaders,
            body: { name: input.organizationName.trim(), slug },
        });
    }

    await paykit.checkouts.create({
        session_type: "one_time",
        customer: { email: session.user.email },
        item_id: input.planCode,
        quantity: input.eventQuantity ?? 1,
        success_url: "/dashboard/billing",
        cancel_url: "/dashboard/billing",
        metadata: { planCode: input.planCode },
    });

    const { token } = await auth.api.getToken({ headers: reqHeaders });

    const res = await fetch(`${API_BASE_URL}${API_V1_PREFIX}/plans/purchase`, {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify({ plan_code: input.planCode, event_quantity: input.eventQuantity }),
    });
    const body = await res.json().catch(() => null);
    if (!res.ok) {
        throw new Error(body?.error?.message ?? body?.message ?? "Subscription failed.");
    }
    return billingResponseSchema.parse(body.data);
}
