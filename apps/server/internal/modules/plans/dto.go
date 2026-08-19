package plans

import "github.com/google/uuid"

type PlanResponse struct {
	ID             uuid.UUID `json:"id"`
	Code           string    `json:"code"`
	Kind           string    `json:"kind"`
	BillingCycle   string    `json:"billing_cycle"`
	Name           string    `json:"name"`
	Amount         *int64    `json:"amount,omitempty"`           // paise; nil for kind=custom (see PerEventAmount)
	PerEventAmount *int64    `json:"per_event_amount,omitempty"` // only set for kind=custom
	Currency       string    `json:"currency"`
	EventQuota     *int      `json:"event_quota,omitempty"` // nil = unlimited, or "caller supplies it" for custom
}

type SubscriptionResponse struct {
	ID               uuid.UUID `json:"id"`
	PlanCode         string    `json:"plan_code"`
	Kind             string    `json:"kind"`
	BillingCycle     string    `json:"billing_cycle"`
	Status           string    `json:"status"`
	EventQuota       *int      `json:"event_quota,omitempty"`
	StartedAt        string    `json:"started_at"`
	CurrentPeriodEnd *string   `json:"current_period_end,omitempty"`
	GraceDeadline    *string   `json:"grace_deadline,omitempty"`
}

// OrgSubscriptionsResponse is the org's whole billing picture: every
// subscription it has ever purchased — top-ups stack rather than
// replacing each other, see Service.Subscribe — plus a computed summary
// of total usable event capacity, mirroring the pooling logic in
// events.Service.checkEventLimit so the frontend doesn't have to
// re-derive it.
type OrgSubscriptionsResponse struct {
	Subscriptions []SubscriptionResponse `json:"subscriptions"`
	TotalQuota    *int                   `json:"total_quota,omitempty"` // nil when Unlimited
	UsedQuota     int                    `json:"used_quota"`
	Unlimited     bool                   `json:"unlimited"`
}

// SubscribeRequest purchases a new subscription. It is always additive —
// see Service.Subscribe — never replaces an existing one, so an org
// buying more Custom capacity on top of its current plan just calls this
// endpoint again. EventQuantity is required only when PlanCode names a
// kind='custom' plan (validated in Service.Subscribe once the plan's kind
// is known — a struct tag can't cleanly express that cross-field rule for
// both custom codes, custom_monthly and custom_yearly).
type SubscribeRequest struct {
	PlanCode      string `json:"plan_code" validate:"required"`
	EventQuantity *int   `json:"event_quantity,omitempty" validate:"omitempty,min=1"`
}
