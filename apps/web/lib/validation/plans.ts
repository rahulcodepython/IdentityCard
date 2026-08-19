import { z } from "zod"

// Mirrors apps/server/internal/modules/plans/dto.go.

export const planKindSchema = z.enum(["flash", "base", "custom", "unlimited"])
export type PlanKind = z.infer<typeof planKindSchema>

export const billingCycleSchema = z.enum(["monthly", "yearly", "one_time"])
export type BillingCycle = z.infer<typeof billingCycleSchema>

export const subscriptionStatusSchema = z.enum(["active", "past_due", "expired"])
export type SubscriptionStatus = z.infer<typeof subscriptionStatusSchema>

export const planSchema = z.object({
  id: z.string().uuid(),
  code: z.string(),
  kind: planKindSchema,
  billing_cycle: billingCycleSchema,
  name: z.string(),
  amount: z.number().nullish(), // paise; null for kind=custom (see per_event_amount)
  per_event_amount: z.number().nullish(), // only set for kind=custom
  currency: z.string(),
  event_quota: z.number().nullish(), // null = unlimited, or "caller supplies it" for custom
})
export type Plan = z.infer<typeof planSchema>

export const plansResponseSchema = z.array(planSchema)

export const subscriptionResponseSchema = z.object({
  id: z.string().uuid(),
  plan_code: z.string(),
  kind: planKindSchema,
  billing_cycle: billingCycleSchema,
  status: subscriptionStatusSchema,
  event_quota: z.number().nullish(),
  started_at: z.string(),
  current_period_end: z.string().nullish(),
  grace_deadline: z.string().nullish(),
})
export type Subscription = z.infer<typeof subscriptionResponseSchema>

// The org's whole billing picture: every subscription it has ever
// purchased (top-ups stack rather than replacing each other) plus a
// computed summary of total usable event capacity.
export const orgSubscriptionsResponseSchema = z.object({
  subscriptions: z.array(subscriptionResponseSchema),
  total_quota: z.number().nullish(), // null when unlimited
  used_quota: z.number(),
  unlimited: z.boolean(),
})
export type OrgSubscriptions = z.infer<typeof orgSubscriptionsResponseSchema>

// Purchasing is always additive (see plans.Service.Subscribe) — never
// replaces an existing subscription, so buying more Custom capacity on
// top of the current plan just calls this again.
export const subscribeSchema = z.object({
  plan_code: z.string().min(1, "Choose a plan"),
  event_quantity: z.number().min(1).optional(),
})
export type SubscribeInput = z.infer<typeof subscribeSchema>
