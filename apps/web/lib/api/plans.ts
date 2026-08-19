import "server-only"

import { apiFetch } from "@/lib/api/client"
import {
  orgSubscriptionsResponseSchema,
  plansResponseSchema,
  type SubscribeInput,
  subscribeSchema,
  subscriptionResponseSchema,
} from "@/lib/validation/plans"

export async function listPlans() {
  const data = await apiFetch("/plans")
  return plansResponseSchema.parse(data)
}

export async function listOrgSubscriptions() {
  const data = await apiFetch("/plans/subscriptions")
  return orgSubscriptionsResponseSchema.parse(data)
}

export async function subscribeToPlan(input: SubscribeInput) {
  const body = subscribeSchema.parse(input)
  const data = await apiFetch("/plans/subscriptions", {
    method: "POST",
    body: JSON.stringify(body),
  })
  return subscriptionResponseSchema.parse(data)
}

export async function renewSubscription(subscriptionId: string) {
  const data = await apiFetch(`/plans/subscriptions/${subscriptionId}/renew`, {
    method: "POST",
  })
  return subscriptionResponseSchema.parse(data)
}
