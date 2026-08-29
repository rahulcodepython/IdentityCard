import "server-only"

import type {
    Checkout,
    CreateCheckoutSchema,
    UpdateCheckoutSchema,
    Customer,
    CreateCustomerParams,
    UpdateCustomerParams,
    Subscription,
    CreateSubscriptionSchema,
    UpdateSubscriptionSchema,
    Payment,
    CreatePaymentSchema,
    UpdatePaymentSchema,
    CapturePaymentSchema,
    Refund,
    CreateRefundSchema,
    PayKitProvider,
    WebhookHandlerConfig,
    WebhookEventPayload,
} from "@paykit-sdk/core"

// A hand-written PayKit Provider for testing-phase checkout — every plan
// purchase "succeeds" the instant it's created, no real payment
// collection happens. This is the seam decision #4 in the auth/billing
// rewrite plan calls for: everything that touches PayKit (checkouts.ts,
// the select-plan flow) goes through the unified PayKit API, so swapping
// this out for `@paykit-sdk/razorpay` later is a one-file change, not a
// rewrite of the checkout flow itself.
//
// Go's `plans`/`subscriptions` tables remain the actual system of record
// for what an org purchased and how much event quota it has (see
// PlansService.Subscribe) — this provider only orchestrates the
// checkout step; nothing it stores here needs to survive a server
// restart. Payment/refund/webhook methods are never called by that flow
// today (there's no real money moving), so they're minimal stubs, not
// full implementations — they'll get built out for real once Razorpay
// (or another live provider) replaces this one.
const checkouts = new Map<string, Checkout>()
const customers = new Map<string, Customer>()
const subscriptions = new Map<string, Subscription>()

function newId(prefix: string) {
    return `${prefix}_${crypto.randomUUID()}`
}

export class ManualProvider implements PayKitProvider {
    readonly providerName = "manual"
    readonly isSandbox = true
    readonly providerVersion = "1.0.0"
    readonly _native = null

    async createCheckout(params: CreateCheckoutSchema): Promise<Checkout> {
        const checkout: Checkout = {
            id: newId("checkout"),
            customer: params.customer,
            payment_url: params.session_type === "one_time" ? params.success_url : "/dashboard/billing",
            metadata: params.metadata,
            session_type: params.session_type,
            products: [{ id: params.item_id, quantity: params.quantity }],
            currency: "inr",
            amount: 0,
            subscription: params.subscription ?? null,
        }
        checkouts.set(checkout.id, checkout)
        return checkout
    }

    async retrieveCheckout(id: string): Promise<Checkout | null> {
        return checkouts.get(id) ?? null
    }

    async updateCheckout(id: string, params: UpdateCheckoutSchema): Promise<Checkout> {
        const existing = checkouts.get(id)
        if (!existing) throw new Error(`manual provider: checkout ${id} not found`)
        const updated = { ...existing, ...params } as Checkout
        checkouts.set(id, updated)
        return updated
    }

    async deleteCheckout(id: string): Promise<null> {
        checkouts.delete(id)
        return null
    }

    async createCustomer(params: CreateCustomerParams): Promise<Customer> {
        const customer: Customer = {
            id: newId("cus"),
            email: params.email,
            name: params.name ?? params.email,
            phone: params.phone ?? null,
            metadata: params.metadata,
            created_at: new Date(),
            updated_at: null,
        }
        customers.set(customer.id, customer)
        return customer
    }

    async updateCustomer(id: string, params: UpdateCustomerParams): Promise<Customer> {
        const existing = customers.get(id)
        if (!existing) throw new Error(`manual provider: customer ${id} not found`)
        const updated = { ...existing, ...params, updated_at: new Date() }
        customers.set(id, updated)
        return updated
    }

    async retrieveCustomer(id: string): Promise<Customer | null> {
        return customers.get(id) ?? null
    }

    async deleteCustomer(id: string): Promise<null> {
        customers.delete(id)
        return null
    }

    async createSubscription(params: CreateSubscriptionSchema): Promise<Subscription> {
        const now = new Date()
        const subscription: Subscription = {
            id: newId("sub"),
            customer: params.customer,
            amount: params.amount,
            currency: params.currency,
            status: "active",
            current_period_start: now,
            current_period_end: now,
            item_id: params.item_id,
            billing_interval: params.billing_interval,
            metadata: params.metadata,
            custom_fields: null,
            requires_action: false,
            payment_url: null,
        }
        subscriptions.set(subscription.id, subscription)
        return subscription
    }

    async updateSubscription(id: string, params: UpdateSubscriptionSchema): Promise<Subscription> {
        const existing = subscriptions.get(id)
        if (!existing) throw new Error(`manual provider: subscription ${id} not found`)
        const updated = { ...existing, ...params } as Subscription
        subscriptions.set(id, updated)
        return updated
    }

    async cancelSubscription(id: string): Promise<Subscription> {
        const existing = subscriptions.get(id)
        if (!existing) throw new Error(`manual provider: subscription ${id} not found`)
        const updated: Subscription = { ...existing, status: "canceled" }
        subscriptions.set(id, updated)
        return updated
    }

    async deleteSubscription(id: string): Promise<null> {
        subscriptions.delete(id)
        return null
    }

    async retrieveSubscription(id: string): Promise<Subscription | null> {
        return subscriptions.get(id) ?? null
    }

    // Not exercised by the select-plan/billing flow — no real payment
    // collection happens on the manual provider, so there's nothing to
    // model here yet.
    async createPayment(_params: CreatePaymentSchema): Promise<Payment> {
        throw new Error("manual provider: payments aren't implemented — no live gateway configured yet")
    }
    async updatePayment(_id: string, _params: UpdatePaymentSchema): Promise<Payment> {
        throw new Error("manual provider: payments aren't implemented — no live gateway configured yet")
    }
    async retrievePayment(_id: string): Promise<Payment | null> {
        return null
    }
    async deletePayment(_id: string): Promise<null> {
        return null
    }
    async capturePayment(_id: string, _params: CapturePaymentSchema): Promise<Payment> {
        throw new Error("manual provider: payments aren't implemented — no live gateway configured yet")
    }
    async cancelPayment(_id: string): Promise<Payment> {
        throw new Error("manual provider: payments aren't implemented — no live gateway configured yet")
    }
    async createRefund(_params: CreateRefundSchema): Promise<Refund> {
        throw new Error("manual provider: refunds aren't implemented — no live gateway configured yet")
    }
    async handleWebhook(
        _payload: WebhookHandlerConfig,
        _webhookSecret: string | null
    ): Promise<Array<WebhookEventPayload>> {
        return []
    }
}
