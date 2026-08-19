// Plans repository/service own the plan catalog and subscriptions —
// including the billing-lifecycle sweep queries used by internal/jobs.
// Subscriptions are stubbed (no real payment call yet); "purchasing" one
// is just inserting the row.
package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

type PlansRepository struct {
	q *dbgen.Queries
}

func NewPlansRepository(q *dbgen.Queries) *PlansRepository {
	return &PlansRepository{q: q}
}

// WithTx returns a PlansRepository whose queries run inside tx — used by
// PlansService methods that write a subscription and its first
// subscription_periods row atomically.
func (r *PlansRepository) WithTx(tx pgx.Tx) *PlansRepository {
	return &PlansRepository{q: r.q.WithTx(tx)}
}

func (r *PlansRepository) List(ctx context.Context) ([]dbgen.Plan, error) {
	return r.q.ListPlans(ctx)
}

func (r *PlansRepository) GetPlanByCode(ctx context.Context, code string) (dbgen.Plan, error) {
	return r.q.GetPlanByCode(ctx, code)
}

func (r *PlansRepository) GetPlanByID(ctx context.Context, id uuid.UUID) (dbgen.Plan, error) {
	return r.q.GetPlanByID(ctx, id)
}

func (r *PlansRepository) CreateSubscription(ctx context.Context, orgID, planID uuid.UUID, kind, billingCycle string, eventQuota pgtype.Int4, currentPeriodEnd pgtype.Date) (dbgen.Subscription, error) {
	return r.q.CreateSubscription(ctx, dbgen.CreateSubscriptionParams{
		OrganizationID:   orgID,
		PlanID:           planID,
		Kind:             kind,
		BillingCycle:     billingCycle,
		EventQuota:       eventQuota,
		CurrentPeriodEnd: currentPeriodEnd,
	})
}

func (r *PlansRepository) CreateSubscriptionPeriod(ctx context.Context, subscriptionID uuid.UUID, start, end pgtype.Date) error {
	_, err := r.q.CreateSubscriptionPeriod(ctx, dbgen.CreateSubscriptionPeriodParams{
		SubscriptionID: subscriptionID,
		PeriodStart:    start,
		PeriodEnd:      end,
	})
	return err
}

func (r *PlansRepository) ListForOrganization(ctx context.Context, orgID uuid.UUID) ([]dbgen.Subscription, error) {
	return r.q.ListSubscriptionsForOrganization(ctx, orgID)
}

// ListActiveForOrganization is ordered oldest-first (started_at ASC) —
// see PlansService.ActiveSubscriptionsForOrgOrderedByAge, which relies on
// this order for FIFO capacity spending.
func (r *PlansRepository) ListActiveForOrganization(ctx context.Context, orgID uuid.UUID) ([]dbgen.Subscription, error) {
	return r.q.ListActiveSubscriptionsForOrganization(ctx, orgID)
}

func (r *PlansRepository) GetForOrganization(ctx context.Context, orgID, id uuid.UUID) (dbgen.Subscription, error) {
	return r.q.GetSubscriptionForOrganization(ctx, dbgen.GetSubscriptionForOrganizationParams{ID: id, OrganizationID: orgID})
}

func (r *PlansRepository) Renew(ctx context.Context, id uuid.UUID, currentPeriodEnd pgtype.Date) (dbgen.Subscription, error) {
	return r.q.RenewSubscription(ctx, dbgen.RenewSubscriptionParams{ID: id, CurrentPeriodEnd: currentPeriodEnd})
}

// -- billing sweeps (used by internal/jobs) --

func (r *PlansRepository) ListPastDueCandidates(ctx context.Context, today pgtype.Date) ([]dbgen.Subscription, error) {
	return r.q.ListPastDueCandidateSubscriptions(ctx, today)
}

func (r *PlansRepository) MarkPastDue(ctx context.Context, id uuid.UUID, pastDueSince, graceDeadline pgtype.Date) error {
	return r.q.MarkSubscriptionPastDue(ctx, dbgen.MarkSubscriptionPastDueParams{
		ID:            id,
		PastDueSince:  pastDueSince,
		GraceDeadline: graceDeadline,
	})
}

func (r *PlansRepository) ListGraceExpired(ctx context.Context, today pgtype.Date) ([]dbgen.Subscription, error) {
	return r.q.ListGraceExpiredSubscriptions(ctx, today)
}

func (r *PlansRepository) MarkExpired(ctx context.Context, id uuid.UUID) error {
	return r.q.MarkSubscriptionExpired(ctx, id)
}

func (r *PlansRepository) ListActiveFlash(ctx context.Context) ([]dbgen.Subscription, error) {
	return r.q.ListActiveFlashSubscriptions(ctx)
}
