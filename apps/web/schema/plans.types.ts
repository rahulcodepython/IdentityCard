import { z } from "zod";

export const pricingConfigSchema = z.object({
    credit_unit_price: z.number(),
    currency: z.string(),
    annual_renewal_amount: z.number(),
});
export type PricingConfig = z.infer<typeof pricingConfigSchema>;

export const billingTransactionSchema = z.object({
    id: z.string().uuid(),
    organization_id: z.string().uuid(),
    type: z.enum(["credit_purchase", "credit_consumed", "annual_renewal"]),
    credits_delta: z.number(),
    amount: z.number(),
    currency: z.string(),
    event_id: z.string().uuid().nullable().optional(),
    description: z.string().nullable().optional(),
    created_at: z.string(),
});
export type BillingTransaction = z.infer<typeof billingTransactionSchema>;

export const billingOverviewResponseSchema = z.object({
    credit_balance: z.number(),
    annual_fee_status: z.enum(["free", "active", "past_due"]),
    current_period_start: z.string().nullable().optional(),
    current_period_end: z.string().nullable().optional(),
    event_count: z.number(),
    pricing: pricingConfigSchema,
    transactions: z.array(billingTransactionSchema),
});
export type BillingOverview = z.infer<typeof billingOverviewResponseSchema>;

export const purchaseCreditsRequestSchema = z.object({
    quantity: z.number().int().min(1, "Quantity must be at least 1").max(1000),
});
export type PurchaseCreditsInput = z.infer<typeof purchaseCreditsRequestSchema>;
