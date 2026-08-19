import { PricingClient } from "@/components/pricing-client"
import { listPlans } from "@/lib/api/plans"

export default async function Pricing() {
  try {
    await listPlans()
  } catch (err) {
    // Fallback gracefully if backend API is offline during SSR
  }

  return <PricingClient />
}
