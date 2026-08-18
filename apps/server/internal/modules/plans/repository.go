// Package plans owns the plan catalog and the (currently stubbed —
// Razorpay integration is a later phase) subscriptions that tie an
// organization to one.
package plans

import (
	"context"

	"github.com/google/uuid"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

type Repository struct {
	q *dbgen.Queries
}

func NewRepository(q *dbgen.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) List(ctx context.Context) ([]dbgen.Plan, error) {
	return r.q.ListPlans(ctx)
}

func (r *Repository) GetByCode(ctx context.Context, code string) (dbgen.Plan, error) {
	return r.q.GetPlanByCode(ctx, code)
}

// CreateSubscription always creates an 'active' subscription — there is no
// real payment step yet (see plans.Service.SelectPlan / auth Register),
// so "selecting" a plan and being subscribed to it are the same action.
func (r *Repository) CreateSubscription(ctx context.Context, orgID, planID uuid.UUID) (dbgen.Subscription, error) {
	return r.q.CreateSubscription(ctx, dbgen.CreateSubscriptionParams{
		OrganizationID: orgID,
		PlanID:         planID,
	})
}
