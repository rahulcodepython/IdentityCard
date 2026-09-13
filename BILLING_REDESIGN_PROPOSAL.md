# Clean & Simple Billing Architecture: Event Credits + Zero-Base Conditional Annual Renewal

## 1. Core Principles & Business Model

This architecture eliminates all complicated plan tiers and replaces them with an intuitive, fair model:

1. **Zero Events = 100% Free**:
    * An organization with **0 events** is never billed, never prompted for a subscription, and never asked for credit card details.
    * Users can create an organization, explore the dashboard, invite team members, configure organization settings, and pair scanner devices with zero paywalls.

2. **1 Event Credit = 1 New Event**:
    * Every time an organization wants to create an event, it costs **1 Event Credit**.
    * Credits can be purchased anytime (e.g., 1 credit for ₹1,499, or packs with bundle discounts).
    * When creating an event, 1 credit is deducted atomically.

3. **Conditional Flat Annual Renewal**:
    * When the annual billing anniversary arrives:
        * The system checks: **Does this organization retain any events (`COUNT(events) > 0`), regardless of whether they are active, expired, or archived?**
        * **If YES**: A **Flat Annual Platform & Retention Fee** (e.g., ₹2,999/year) is charged. This keeps all event history, attendee rosters, past ID cards, analytics, and platform features operational.
        * **If NO (`COUNT(events) == 0`)**: The renewal is **₹0 (No charge)**. If an organization deletes all their events or never created one, they are never charged.

4. **Flat Values (No Mid-Cycle Proration Headaches)**:
    * The annual renewal is a **fixed flat amount**, not a fluctuating per-event formula.
    * Adding 1 event or 10 events throughout the year does not alter the recurring annual fee; the variable cost is cleanly paid upfront via **Event Creation Credits**.

---

## 2. Why This Eliminates the Annual Loophole

### The Loophole You Identified:
> *"A user creates 1 event, buys a 1-year subscription, and then creates multiple events mid-year for a tiny initial cost while keeping them hosted for the full year."*

### How This Design Closes It Completely:
1. **Creation is Never Free**:
    * Having an active annual subscription does **NOT** grant free event creation.
    * Every single event requires spending **1 Event Credit**. If an organization runs 10 events in a year, they must purchase 10 credits.
2. **Predictable Flat Renewal**:
    * The annual fee is a flat platform preservation fee for holding event data on the platform.
    * There is no mid-cycle billing dispute, no complex proration calculation, and no unbilled event creation.

---

## 3. Customer Journey Walkthrough

```
[ Step 1: Sign Up & Create Org ]
  • Event count: 0
  • Billing status: Free (₹0)
  • No credit card prompt. Full platform access.
               │
               ▼
[ Step 2: Create First Event ]
  • Requires 1 Event Credit.
  • User purchases 1 credit for ₹1,499.
  • Event is created.
  • Next billing anniversary is set to 1 year from today (Period End = Today + 1 Year).
               │
               ▼
[ Step 3: Adding More Events During Year 1 ]
  • Want to create Event #2? ──> Buy 1 Credit (₹1,499) ──> Created.
  • Want to create Event #3? ──> Buy 1 Credit (₹1,499) ──> Created.
  • Annual renewal date remains fixed at the 1-year anniversary.
               │
               ▼
[ Step 4: Year 1 Renewal Anniversary Arrives ]
  • System checks: Does organization have any events (active/expired/archived)?
        ├── YES (events exist):
        │     • Charge flat annual renewal fee (₹2,999).
        │     • Next period extended by 1 year. All event history preserved.
        │
        └── NO (all events deleted or 0 events):
              • Billed ₹0. Status remains Free.
```

---

## 4. Simplified Database Schema

Only two tables are needed to support this entire system:

```sql
-- 1. Organization Billing & Subscription State
CREATE TABLE organization_billing (
    organization_id      UUID PRIMARY KEY REFERENCES "organization" (id) ON DELETE CASCADE,
    credit_balance       INT NOT NULL DEFAULT 0 CHECK (credit_balance >= 0),
    annual_fee_status    VARCHAR(32) NOT NULL DEFAULT 'free', -- 'free', 'active', 'past_due'
    current_period_start DATE,
    current_period_end   DATE,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT false,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 2. Audit Ledger for Payments & Credit Transactions
CREATE TABLE billing_transactions (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id      UUID NOT NULL REFERENCES "organization" (id) ON DELETE CASCADE,
    type                 VARCHAR(32) NOT NULL, -- 'credit_purchase', 'credit_consumed', 'annual_renewal'
    credits_delta        INT NOT NULL DEFAULT 0, -- +1, +5, or -1
    amount               BIGINT NOT NULL DEFAULT 0, -- Price in paise/cents (0 for consumption)
    currency             VARCHAR(8) NOT NULL DEFAULT 'INR',
    event_id             UUID REFERENCES events (id) ON DELETE SET NULL,
    gateway_payment_id   VARCHAR(255),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

---

## 5. Event Creation Workflow ("Allow Create Event")

When a user calls `POST /api/v1/organization/:orgId/events`:

```go
func (s *EventService) Create(ctx context.Context, orgID uuid.UUID, req CreateEventRequest) (EventResponse, error) {
    // 1. Verify User Role (Admin or Member)
    // Handled by router middleware: RequireRole(RoleAdmin, RoleMember)

    var eventID uuid.UUID

    // 2. Atomic Database Transaction
    err := postgres.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
        // Deduct 1 credit atomically
        const deductCreditQuery = `
            UPDATE organization_billing
            SET credit_balance = credit_balance - 1,
                -- If this is their very first event, initialize annual period
                annual_fee_status = CASE WHEN annual_fee_status = 'free' THEN 'active' ELSE annual_fee_status END,
                current_period_start = COALESCE(current_period_start, CURRENT_DATE),
                current_period_end = COALESCE(current_period_end, CURRENT_DATE + INTERVAL '1 year')
            WHERE organization_id = $1 AND credit_balance >= 1
            RETURNING credit_balance;
        `
        var remainingCredits int
        err := tx.QueryRow(ctx, deductCreditQuery, orgID).Scan(&remainingCredits)
        if err != nil {
            // No credits available
            return utils.NewError(http.StatusPaymentRequired, "1 Event Credit is required to create an event. Please purchase credits to continue.", err)
        }

        // Create the event record
        event, err := s.repo.CreateEventTx(ctx, tx, orgID, req)
        if err != nil {
            return err
        }
        eventID = event.ID

        // Record credit consumption in audit ledger
        const logAuditQuery = `
            INSERT INTO billing_transactions (organization_id, type, credits_delta, event_id)
            VALUES ($1, 'credit_consumed', -1, $2);
        `
        _, err = tx.Exec(ctx, logAuditQuery, orgID, eventID)
        return err
    })

    if err != nil {
        return EventResponse{}, err
    }

    return s.Get(ctx, orgID, eventID)
}
```

### When Credit Balance is 0:
* The endpoint returns `402 Payment Required`.
* The frontend renders an inline purchase modal:
  * **Option 1**: Purchase 1 Credit (₹1,499)
  * **Option 2**: Purchase 5 Credits (₹5,999 — Save 20%)
* Once payment succeeds, the credit balance increments and the event is created immediately.

---

## 6. How the Annual Renewal Works

The background scheduler (or Stripe/Razorpay webhook) checks subscriptions on their `current_period_end` date:

```
[ Current Date reaches current_period_end ]
                    │
                    ▼
     SELECT COUNT(*) FROM events 
     WHERE organization_id = :orgId;
                    │
       ┌────────────┴────────────┐
       ▼                         ▼
 [ Count > 0 ]              [ Count == 0 ]
Events exist in account     No events in account
       │                         │
       ▼                         ▼
Charge flat annual fee      Charge ₹0 (Free)
(e.g., ₹2,999)              Status resets to 'free'.
Extend period by 1 year.    No credit card charged.
```

### What Happens if Annual Renewal Fails?
1. **Grace Period (30 Days)**:
    * Status updates to `'past_due'`.
    * Organizers can still log in and view past event records, download past attendee lists, and view ID cards.
    * New event creation is paused until the flat annual renewal is settled.
2. **After Grace Period (No Payment)**:
    * Event public landing pages and QR check-in scanning are paused.
    * Account data remains intact until owner renews or chooses to export and delete.
