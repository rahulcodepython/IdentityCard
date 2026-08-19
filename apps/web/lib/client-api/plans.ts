import { z } from "zod"

import { apiRequest } from "@/react-query/client"
import {
  orgSubscriptionsResponseSchema,
  type OrgSubscriptions,
  plansResponseSchema,
  type SubscribeInput,
  subscribeSchema,
  subscriptionResponseSchema,
} from "@/schema/plans.types"

export async function listPlans() {
  return apiRequest({ url: "/plans", method: "GET" }, plansResponseSchema)
}

export async function listOrgSubscriptions() {
  return apiRequest(
    { url: "/plans/subscriptions", method: "GET" },
    orgSubscriptionsResponseSchema
  )
}

export async function subscribeToPlan(input: SubscribeInput) {
  const body = subscribeSchema.parse(input)
  return apiRequest(
    { url: "/plans/subscriptions", method: "POST", data: body },
    subscriptionResponseSchema
  )
}

export async function renewSubscription(subscriptionId: string) {
  return apiRequest(
    {
      url: `/plans/subscriptions/${subscriptionId}/renew`,
      method: "POST",
    },
    subscriptionResponseSchema
  )
}