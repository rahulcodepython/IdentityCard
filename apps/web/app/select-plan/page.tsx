import { headers } from "next/headers"
import { redirect } from "next/navigation"

import { auth } from "@/lib/auth"
import { plansResponseSchema } from "@/schema/plans.types"

import { SelectPlanForm } from "./select-plan-form"

const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8080"

// Reached only by a signed-in user with no organization — never
// purchased a plan yet (see middleware.ts + app/dashboard/layout.tsx's
// redirect). Org creation itself happens as a side effect of a
// successful purchase here (see lib/actions/checkout.ts), not from a
// standalone form.
export default async function SelectPlanPage() {
    const session = await auth.api.getSession({ headers: await headers() })
    if (!session) redirect("/login")
    if (session.session.activeOrganizationId) redirect("/dashboard")

    const res = await fetch(`${API_BASE_URL}/plans`, { cache: "no-store" })
    const body = await res.json()
    const plans = plansResponseSchema.parse(body.data)

    return (
        <div className="flex min-h-svh flex-col items-center gap-8 p-6 py-16">
            <div className="flex flex-col items-center gap-2 text-center">
                <h1 className="text-2xl font-bold">Choose a plan</h1>
                <p className="max-w-md text-balance text-sm text-muted-foreground">
                    Pick a plan to create your organization — you can invite teammates and start adding events
                    right after.
                </p>
            </div>

            <SelectPlanForm plans={plans} defaultName={session.user.name} />
        </div>
    )
}
