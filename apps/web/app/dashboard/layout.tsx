import { redirect } from "next/navigation"

import { me } from "@/lib/api/auth"
import { ApiError } from "@/lib/api/client"
import { listOrgSubscriptions } from "@/lib/api/plans"

import { DashboardShell } from "./dashboard-shell"

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  let user
  let hasExpiredPlan = false
  try {
    user = await me()
    const subs = await listOrgSubscriptions().catch(() => null)
    if (subs) {
      hasExpiredPlan = subs.subscriptions.some(
        (s) => s.status === "past_due" || s.status === "expired"
      )
    }
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      redirect("/login")
    }
    throw err
  }

  if (!user.has_organization) {
    redirect("/onboarding")
  }

  return (
    <DashboardShell user={user} hasExpiredPlan={hasExpiredPlan}>
      {children}
    </DashboardShell>
  )
}
