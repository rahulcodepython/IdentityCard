package entities

import "github.com/google/uuid"

type PlanResponse struct {
	ID               uuid.UUID `json:"id"`
	Code             string    `json:"code"`
	Kind             string    `json:"kind"`
	BillingCycle     string    `json:"billing_cycle"`
	Name             string    `json:"name"`
	Amount           *int64    `json:"amount,omitempty"`           // paise; nil for kind=custom (see PerEventAmount)
	PerEventAmount   *int64    `json:"per_event_amount,omitempty"` // only set for kind=custom
	Currency         string    `json:"currency"`
	EventQuota       *int      `json:"event_quota,omitempty"` // nil = unlimited, or "caller supplies it" for custom
	NominalIncrement int64     `json:"nominal_increment"`     // per-cycle price escalation, paise
}

// BillingResponse is one billing lineage's current period row — see
// PlansService.ListBilling, which always returns the latest row per
// lineage (a lineage's historical paid rows aren't surfaced).
type BillingResponse struct {
	ID            uuid.UUID `json:"id"`
	LineageRootID uuid.UUID `json:"lineage_root_id"`
	PlanCode      string    `json:"plan_code"`
	Kind          string    `json:"kind"`
	BillingCycle  string    `json:"billing_cycle"`
	BillingNumber int       `json:"billing_number"`
	PeriodStart   string    `json:"period_start"`
	PeriodEnd     string    `json:"period_end"`
	Status        string    `json:"status"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	PaidAt        *string   `json:"paid_at,omitempty"`
}

type CreditResponse struct {
	ID              uuid.UUID  `json:"id"`
	Type            string     `json:"type"`
	EventID         *uuid.UUID `json:"event_id,omitempty"`
	IsRestricted    bool       `json:"is_restricted"`
	RestrictedSince *string    `json:"restricted_since,omitempty"`
	CreatedAt       string     `json:"created_at"`
}

type TransactionResponse struct {
	ID        uuid.UUID `json:"id"`
	PlanCode  string    `json:"plan_code"`
	Amount    int64     `json:"amount"`
	Currency  string    `json:"currency"`
	CreatedAt string    `json:"created_at"`
}

// OrgBillingResponse is the org's whole billing picture: every billing
// lineage it has ever started (each a separate purchase — top-ups stack
// rather than replacing each other, see PlansService.Purchase), plus its
// full credit ledger and a computed available-by-type count so the
// frontend doesn't have to re-derive it.
type OrgBillingResponse struct {
	Billings         []BillingResponse `json:"billings"`
	Credits          []CreditResponse  `json:"credits"`
	AvailableByType  map[string]int    `json:"available_by_type"`
}

// PurchaseRequest buys a new billing lineage — always additive (see
// PlansService.Purchase), never replaces an existing one, so an org
// buying more Custom capacity on top of its current plan just calls this
// endpoint again. EventQuantity is required only when PlanCode names a
// kind='custom' plan (validated in PlansService.Purchase once the plan's
// kind is known — a struct tag can't cleanly express that cross-field
// rule for both custom codes, custom_monthly and custom_yearly).
type PurchaseRequest struct {
	PlanCode      string `json:"plan_code" validate:"required"`
	EventQuantity *int   `json:"event_quantity,omitempty" validate:"omitempty,min=1"`
}

// UpgradeRequest switches an existing billing lineage's plan for future
// renewals — see PlansService.Upgrade.
type UpgradeRequest struct {
	PlanCode string `json:"plan_code" validate:"required"`
}
