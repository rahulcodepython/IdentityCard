import type { ReactNode } from "react"

import {
  Card,
  CardAction,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import type { Plan } from "@/schema/plans.types"

// Shared by the public /plans catalog, the landing page's pricing
// preview, and the in-dashboard billing page — one place defines how a
// plan's price renders.
export function formatPlanPrice(plan: Plan) {
  const format = (paise: number) =>
    new Intl.NumberFormat("en-IN", {
      style: "currency",
      currency: plan.currency,
      maximumFractionDigits: 0,
    }).format(paise / 100)

  if (plan.kind === "custom") {
    return plan.per_event_amount ? `${format(plan.per_event_amount)}/event` : "Contact us"
  }
  if (plan.amount == null) return "Contact us"

  const suffix =
    plan.billing_cycle === "daily"
      ? "/day"
      : plan.billing_cycle === "monthly"
        ? "/month"
        : plan.billing_cycle === "yearly"
          ? "/year"
          : ""
  return `${format(plan.amount)}${suffix}`
}

export const BILLING_CYCLE_LABEL: Record<Plan["billing_cycle"], string> = {
  daily: "Billed daily",
  monthly: "Billed monthly",
  yearly: "Billed yearly",
  one_time: "One-time",
}

export function PlanCard({
  plan,
  action,
  current = false,
}: {
  plan: Plan
  action?: ReactNode
  current?: boolean
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          {plan.name}
          {
          current && <span className="rounded-lg bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
              Active
            </span>
          }
        </CardTitle>
        <CardDescription>
          {formatPlanPrice(plan)} · {BILLING_CYCLE_LABEL[plan.billing_cycle]}
        </CardDescription>
        {action && <CardAction>{action}</CardAction>}
      </CardHeader>
    </Card>
  )
}
