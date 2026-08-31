import { z } from "zod";

export const planKindSchema = z.enum(["flash", "base", "custom", "unlimited"]);
export type PlanKind = z.infer<typeof planKindSchema>;

export const billingCycleSchema = z.enum(["daily", "monthly", "yearly", "one_time"]);
export type BillingCycle = z.infer<typeof billingCycleSchema>;

export const subscriptionStatusSchema = z.enum(["active", "pending", "cancelled"]);
export type SubscriptionStatus = z.infer<typeof subscriptionStatusSchema>;

export const planSchema = z.object({
    id: z.string().uuid(),
    code: z.string(),
    kind: planKindSchema,
    billing_cycle: billingCycleSchema,
    name: z.string(),
    amount: z.number().nullable().optional(),
    per_event_amount: z.number().nullable().optional(),
    currency: z.string(),
    event_quota: z.number().nullable().optional(),
    nominal_increment: z.number(),
});
export type Plan = z.infer<typeof planSchema>;

export const plansResponseSchema = z.array(planSchema);

export const billingResponseSchema = z.object({
    id: z.string().uuid(),
    lineage_root_id: z.string().uuid(),
    plan_code: z.string(),
    kind: planKindSchema,
    billing_cycle: billingCycleSchema,
    billing_number: z.number(),
    period_start: z.string(),
    period_end: z.string(),
    status: subscriptionStatusSchema,
    amount: z.number(),
    currency: z.string(),
    paid_at: z.string().nullable().optional(),
});
export type Billing = z.infer<typeof billingResponseSchema>;

export const creditResponseSchema = z.object({
    id: z.string().uuid(),
    type: z.string(),
    event_id: z.string().uuid().nullable().optional(),
    is_restricted: z.boolean(),
    restricted_since: z.string().nullable().optional(),
    created_at: z.string(),
});
export type Credit = z.infer<typeof creditResponseSchema>;

export const orgBillingResponseSchema = z.object({
    billings: z.array(billingResponseSchema),
    credits: z.array(creditResponseSchema),
    available_by_type: z.record(z.string(), z.number()),
});
export type OrgBilling = z.infer<typeof orgBillingResponseSchema>;

// Alias OrgSubscriptions to OrgBilling for backward compatibility with UI components
export type OrgSubscriptions = OrgBilling;
export const orgSubscriptionsResponseSchema = orgBillingResponseSchema;

export const purchaseRequestSchema = z.object({
    plan_code: z.string().min(1, "Choose a plan"),
    event_quantity: z.number().min(1).optional(),
});
export type PurchaseInput = z.infer<typeof purchaseRequestSchema>;

export const upgradeRequestSchema = z.object({
    plan_code: z.string().min(1, "Choose a plan"),
});
export type UpgradeInput = z.infer<typeof upgradeRequestSchema>;
