// Package plans owns the plan catalog and subscriptions — including the
// billing-lifecycle sweep queries used by internal/jobs. Subscriptions are
// stubbed (no real payment call yet, same as before); "purchasing" one is
// just inserting the row.
package plans

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

type Repository struct {
	q *dbgen.Queries
}

func NewRepository(q *dbgen.Queries) *Repository {
	return &Repository{q: q}
}

// WithTx returns a Repository whose queries run inside tx — used by
// Service methods that write a subscription and its first
// subscription_periods row atomically.
func (r *Repository) WithTx(tx pgx.Tx) *Repository {
	return &Repository{q: r.q.WithTx(tx)}
}

func (r *Repository) List(ctx context.Context) ([]dbgen.Plan, error) {
	return r.q.ListPlans(ctx)
}

func (r *Repository) GetPlanByCode(ctx context.Context, code string) (dbgen.Plan, error) {
	return r.q.GetPlanByCode(ctx, code)
}

func (r *Repository) GetPlanByID(ctx context.Context, id uuid.UUID) (dbgen.Plan, error) {
	return r.q.GetPlanByID(ctx, id)
}

func (r *Repository) CreateSubscription(ctx context.Context, orgID, planID uuid.UUID, kind, billingCycle string, eventQuota pgtype.Int4, currentPeriodEnd pgtype.Date) (dbgen.Subscription, error) {
	return r.q.CreateSubscription(ctx, dbgen.CreateSubscriptionParams{
		OrganizationID:   orgID,
		PlanID:           planID,
		Kind:             kind,
		BillingCycle:     billingCycle,
		EventQuota:       eventQuota,
		CurrentPeriodEnd: currentPeriodEnd,
	})
}

func (r *Repository) CreateSubscriptionPeriod(ctx context.Context, subscriptionID uuid.UUID, start, end pgtype.Date) error {
	_, err := r.q.CreateSubscriptionPeriod(ctx, dbgen.CreateSubscriptionPeriodParams{
		SubscriptionID: subscriptionID,
		PeriodStart:    start,
		PeriodEnd:      end,
	})
	return err
}

func (r *Repository) ListForOrganization(ctx context.Context, orgID uuid.UUID) ([]dbgen.Subscription, error) {
	return r.q.ListSubscriptionsForOrganization(ctx, orgID)
}

// ListActiveForOrganization is ordered oldest-first (started_at ASC) —
// see Service.ActiveSubscriptionsForOrgOrderedByAge, which relies on this
// order for FIFO capacity spending.
func (r *Repository) ListActiveForOrganization(ctx context.Context, orgID uuid.UUID) ([]dbgen.Subscription, error) {
	return r.q.ListActiveSubscriptionsForOrganization(ctx, orgID)
}

func (r *Repository) GetForOrganization(ctx context.Context, orgID, id uuid.UUID) (dbgen.Subscription, error) {
	return r.q.GetSubscriptionForOrganization(ctx, dbgen.GetSubscriptionForOrganizationParams{ID: id, OrganizationID: orgID})
}

func (r *Repository) Renew(ctx context.Context, id uuid.UUID, currentPeriodEnd pgtype.Date) (dbgen.Subscription, error) {
	return r.q.RenewSubscription(ctx, dbgen.RenewSubscriptionParams{ID: id, CurrentPeriodEnd: currentPeriodEnd})
}

// -- billing sweeps (used by internal/jobs) --

func (r *Repository) ListPastDueCandidates(ctx context.Context, today pgtype.Date) ([]dbgen.Subscription, error) {
	return r.q.ListPastDueCandidateSubscriptions(ctx, today)
}

func (r *Repository) MarkPastDue(ctx context.Context, id uuid.UUID, pastDueSince, graceDeadline pgtype.Date) error {
	return r.q.MarkSubscriptionPastDue(ctx, dbgen.MarkSubscriptionPastDueParams{
		ID:            id,
		PastDueSince:  pastDueSince,
		GraceDeadline: graceDeadline,
	})
}

func (r *Repository) ListGraceExpired(ctx context.Context, today pgtype.Date) ([]dbgen.Subscription, error) {
	return r.q.ListGraceExpiredSubscriptions(ctx, today)
}

func (r *Repository) MarkExpired(ctx context.Context, id uuid.UUID) error {
	return r.q.MarkSubscriptionExpired(ctx, id)
}

func (r *Repository) ListActiveFlash(ctx context.Context) ([]dbgen.Subscription, error) {
	return r.q.ListActiveFlashSubscriptions(ctx)
}
