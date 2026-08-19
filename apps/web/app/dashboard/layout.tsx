"use client"

import { useEffect } from "react"
import { useRouter } from "next/navigation"
import { RiLoader4Line } from "@remixicon/react"

import { useMeQuery } from "@/query-hooks/auth.api"
import { useListSubscriptionsQuery } from "@/query-hooks/plans.api"
import { useSessionStore } from "@/store/session.store"

import { DashboardShell } from "./dashboard-shell"

// Client-driven auth gate (no more server-fetched cookie forwarding — see
// store/session.store.ts + react-query/client.ts). hydrate() bridges the
// httpOnly ic_access cookie into the store on first mount; once
// authenticated, useMeQuery/useListSubscriptionsQuery call Go directly with
// the Bearer token the axios interceptor attaches.
export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const router = useRouter()
  const status = useSessionStore((s) => s.status)
  const hydrate = useSessionStore((s) => s.hydrate)

  useEffect(() => {
    if (status === "idle") {
      void hydrate()
    }
  }, [status, hydrate])

  useEffect(() => {
    if (status === "unauthenticated") {
      router.replace("/login")
    }
  }, [status, router])

  const authenticated = status === "authenticated"
  const meQuery = useMeQuery(authenticated)
  const subsQuery = useListSubscriptionsQuery(authenticated)

  useEffect(() => {
    // A 401 here means the axios interceptor already tried (and failed) a
    // silent refresh — the session is genuinely gone.
    if (meQuery.isError) {
      router.replace("/login")
    }
  }, [meQuery.isError, router])

  useEffect(() => {
    if (meQuery.data && !meQuery.data.has_organization) {
      router.replace("/onboarding")
    }
  }, [meQuery.data, router])

  const ready = authenticated && !!meQuery.data && meQuery.data.has_organization

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
    <DashboardShell user={meQuery.data} hasExpiredPlan={hasExpiredPlan}>
      {children}
    </DashboardShell>
  )
}
