import "server-only"

import { apiFetch } from "@/lib/api/client"
import { plansResponseSchema } from "@/lib/validation/plans"

export async function listPlans() {
  const data = await apiFetch("/plans")
  return plansResponseSchema.parse(data)
}
