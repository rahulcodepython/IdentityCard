import Link from "next/link"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardAction,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { listPlans } from "@/lib/api/plans"
import type { Plan } from "@/lib/validation/plans"

function formatPrice(plan: Plan) {
  if (plan.price.billing === "custom") return "Contact us"

  const amount = new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency: plan.price.currency,
    maximumFractionDigits: 0,
  }).format(plan.price.amount / 100)

  const suffix = plan.price.billing === "yearly" ? "/year" : ""
  return `${amount}${suffix}`
}

function PlanCard({ plan }: { plan: Plan }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{plan.name}</CardTitle>
        <CardDescription>{formatPrice(plan)}</CardDescription>
        <CardAction>
          <Button render={<Link href={`/register?plan=${plan.code}`} />}>
            Select
          </Button>
        </CardAction>
      </CardHeader>
    </Card>
  )
}

export default async function PlansPage() {
  const plans = await listPlans()
  const flashPlans = plans.filter((plan) => plan.kind === "flash")
  const standardPlans = plans.filter((plan) => plan.kind === "standard")

  return (
    <div className="mx-auto flex min-h-svh max-w-4xl flex-col gap-10 p-6">
      <div className="flex flex-col gap-1">
        <h1 className="font-heading text-2xl font-medium">Plans</h1>
        <p className="text-sm text-muted-foreground">
          Pick a plan to create your organization.{" "}
          <Link href="/login" className="underline underline-offset-4">
            Already have an account? Sign in
          </Link>
        </p>
      </div>

      <section className="flex flex-col gap-3">
        <h2 className="text-sm font-medium text-muted-foreground">
          Flash — a single one-day event
        </h2>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          {flashPlans.map((plan) => (
            <PlanCard key={plan.id} plan={plan} />
          ))}
        </div>
      </section>

      <section className="flex flex-col gap-3">
        <h2 className="text-sm font-medium text-muted-foreground">
          Standard — established events, multi-day and recurring
        </h2>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          {standardPlans.map((plan) => (
            <PlanCard key={plan.id} plan={plan} />
          ))}
        </div>
      </section>

      <CardFooter className="justify-center rounded-xl border bg-transparent p-4 text-xs text-muted-foreground">
        Prices shown are placeholders. Choosing a plan does not charge you yet —
        payment is added in a later phase.
      </CardFooter>
    </div>
  )
}
