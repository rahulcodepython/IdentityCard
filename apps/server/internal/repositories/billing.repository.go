// Package repositories: billing is one row per payment period, chained
// into a lineage via lineage_root_id (see the migration's own doc
// comment) — this is what gates whether an org can keep creating/editing
// the events its credits funded.
package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

type BillingRepository struct {
	q *dbgen.Queries
}

func NewBillingRepository(q *dbgen.Queries) *BillingRepository {
	return &BillingRepository{q: q}
}

func (r *BillingRepository) WithTx(tx pgx.Tx) *BillingRepository {
	return &BillingRepository{q: r.q.WithTx(tx)}
}

func (r *BillingRepository) CreateLineageRoot(ctx context.Context, orgID, planID uuid.UUID, periodStart, periodEnd pgtype.Date, status string, amount int64) (dbgen.Billing, error) {
	return r.q.CreateBillingLineageRoot(ctx, dbgen.CreateBillingLineageRootParams{
		OrganizationID: orgID, PlanID: planID, PeriodStart: periodStart, PeriodEnd: periodEnd, Status: status, Amount: amount,
	})
}

func (r *BillingRepository) CreateRenewal(ctx context.Context, orgID, planID, lineageRootID uuid.UUID, billingNumber int32, periodStart, periodEnd pgtype.Date, status string, amount int64) (dbgen.Billing, error) {
	return r.q.CreateBillingRenewal(ctx, dbgen.CreateBillingRenewalParams{
		OrganizationID: orgID, PlanID: planID, LineageRootID: lineageRootID, BillingNumber: billingNumber,
		PeriodStart: periodStart, PeriodEnd: periodEnd, Status: status, Amount: amount,
	})
}

func (r *BillingRepository) Get(ctx context.Context, orgID, id uuid.UUID) (dbgen.Billing, error) {
	return r.q.GetBilling(ctx, dbgen.GetBillingParams{ID: id, OrganizationID: orgID})
}

func (r *BillingRepository) GetLatestForLineage(ctx context.Context, lineageRootID uuid.UUID) (dbgen.Billing, error) {
	return r.q.GetLatestBillingForLineage(ctx, lineageRootID)
}

func (r *BillingRepository) ListLatestForOrganization(ctx context.Context, orgID uuid.UUID) ([]dbgen.Billing, error) {
	return r.q.ListLatestBillingForOrganization(ctx, orgID)
}

func (r *BillingRepository) MarkPaid(ctx context.Context, id, transactionID uuid.UUID) (dbgen.Billing, error) {
	return r.q.MarkBillingPaid(ctx, dbgen.MarkBillingPaidParams{ID: id, TransactionID: pgtype.UUID{Bytes: transactionID, Valid: true}})
}

func (r *BillingRepository) Cancel(ctx context.Context, lineageRootID uuid.UUID) (dbgen.Billing, error) {
	return r.q.CancelBillingLineage(ctx, lineageRootID)
}

func (r *BillingRepository) UpgradePlan(ctx context.Context, lineageRootID, planID uuid.UUID, amount int64) (dbgen.Billing, error) {
	return r.q.UpgradeBillingPlan(ctx, dbgen.UpgradeBillingPlanParams{LineageRootID: lineageRootID, PlanID: planID, Amount: amount})
}

// FlipLapsedToPending is the daily cron's first step — see internal/jobs.
func (r *BillingRepository) FlipLapsedToPending(ctx context.Context, today pgtype.Date) ([]dbgen.Billing, error) {
	return r.q.FlipLapsedBillingToPending(ctx, today)
}
