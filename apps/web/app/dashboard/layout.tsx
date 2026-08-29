"use client"

import { useEffect } from "react"
import { useRouter } from "next/navigation"
import { RiLoader4Line } from "@remixicon/react"

import { useListSubscriptionsQuery } from "@/query-hooks/plans.api"
import { useSessionStore } from "@/store/session.store"

import { DashboardShell } from "./dashboard-shell"

// Auth gate driven entirely by the zustand session store, populated once
// by components/session-provider.tsx (mounted in the root layout) — no
// fetch happens here. A signed-in user with no active organization
// (never purchased a plan — see lib/auth.ts's organizationLimit/purchase
// gating) can't reach anything under /dashboard; every route here
// assumes an org to scope its data by.
export default function DashboardLayout({
    children,
}: {
    children: React.ReactNode
}) {
    const router = useRouter()
    const status = useSessionStore((s) => s.status)
    const user = useSessionStore((s) => s.user)
    const role = useSessionStore((s) => s.role)
    const activeOrganizationId = useSessionStore((s) => s.activeOrganizationId)

    useEffect(() => {
        if (status === "unauthenticated") {
            router.replace("/login")
        }
    }, [status, router])

    const authenticated = status === "authenticated"
    const subsQuery = useListSubscriptionsQuery(authenticated && !!activeOrganizationId)

    const ready = authenticated && !!user

    if (!ready) {
        return (
            <div className="flex min-h-svh items-center justify-center">
                <RiLoader4Line className="size-6 animate-spin text-muted-foreground" />
            </div>
        )
    }

    const hasExpiredPlan =
        subsQuery.data?.subscriptions.some(
            (s) => s.status === "past_due" || s.status === "expired"
        ) ?? false

    return (
        <DashboardShell user={user} role={role} hasExpiredPlan={hasExpiredPlan}>
            {children}
        </DashboardShell>
    )
}
