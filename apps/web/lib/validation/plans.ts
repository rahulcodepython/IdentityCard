import { z } from "zod"

// Mirrors apps/server/internal/modules/plans/dto.go.

export const priceConfigSchema = z.object({
  amount: z.number(), // smallest currency unit (e.g. paise); 0 for "contact sales" (custom tier)
  currency: z.string(),
  billing: z.enum(["one_time", "per_event", "yearly", "custom"]),
})

export const planSchema = z.object({
  id: z.string().uuid(),
  code: z.string(),
  kind: z.enum(["flash", "standard"]),
  tier: z.string(),
  name: z.string(),
  price: priceConfigSchema,
})
export type Plan = z.infer<typeof planSchema>

export const plansResponseSchema = z.array(planSchema)
