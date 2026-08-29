package plans

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

// -- Plans Catalog --

func (a *App) ListPlans(ctx context.Context) ([]dbgen.Plan, error) {
	return a.queries.ListPlans(ctx)
}

func (a *App) GetPlanByCode(ctx context.Context, code string) (dbgen.Plan, error) {
	return a.queries.GetPlanByCode(ctx, code)
}

func (a *App) GetPlanByID(ctx context.Context, id uuid.UUID) (dbgen.Plan, error) {
	return a.queries.GetPlanByID(ctx, id)
}

// -- Billing Lineages --

func (a *App) CreateBillingLineageRoot(ctx context.Context, tx pgx.Tx, orgID, planID uuid.UUID, periodStart, periodEnd pgtype.Date, status string, amount int64) (dbgen.Billing, error) {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.CreateBillingLineageRoot(ctx, dbgen.CreateBillingLineageRootParams{
		OrganizationID: orgID, PlanID: planID, PeriodStart: periodStart, PeriodEnd: periodEnd, Status: status, Amount: amount,
	})
}

func (a *App) CreateBillingRenewal(ctx context.Context, tx pgx.Tx, orgID, planID, lineageRootID uuid.UUID, billingNumber int32, periodStart, periodEnd pgtype.Date, status string, amount int64) (dbgen.Billing, error) {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.CreateBillingRenewal(ctx, dbgen.CreateBillingRenewalParams{
		OrganizationID: orgID, PlanID: planID, LineageRootID: lineageRootID, BillingNumber: billingNumber,
		PeriodStart: periodStart, PeriodEnd: periodEnd, Status: status, Amount: amount,
	})
}

func (a *App) GetLatestBillingForLineage(ctx context.Context, lineageRootID uuid.UUID) (dbgen.Billing, error) {
	return a.queries.GetLatestBillingForLineage(ctx, lineageRootID)
}

func (a *App) ListLatestBillingForOrganization(ctx context.Context, orgID uuid.UUID) ([]dbgen.Billing, error) {
	return a.queries.ListLatestBillingForOrganization(ctx, orgID)
}

func (a *App) MarkBillingPaid(ctx context.Context, tx pgx.Tx, id, transactionID uuid.UUID) (dbgen.Billing, error) {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.MarkBillingPaid(ctx, dbgen.MarkBillingPaidParams{ID: id, TransactionID: pgtype.UUID{Bytes: transactionID, Valid: true}})
}

func (a *App) CancelBilling(ctx context.Context, lineageRootID uuid.UUID) (dbgen.Billing, error) {
	return a.queries.CancelBillingLineage(ctx, lineageRootID)
}

func (a *App) UpgradeBillingPlan(ctx context.Context, lineageRootID, planID uuid.UUID, amount int64) (dbgen.Billing, error) {
	return a.queries.UpgradeBillingPlan(ctx, dbgen.UpgradeBillingPlanParams{LineageRootID: lineageRootID, PlanID: planID, Amount: amount})
}

func (a *App) FlipLapsedBillingToPending(ctx context.Context, today pgtype.Date) ([]dbgen.Billing, error) {
	return a.queries.FlipLapsedBillingToPending(ctx, today)
}

// -- Credits --

func (a *App) CreateCredit(ctx context.Context, tx pgx.Tx, orgID, billingID uuid.UUID, creditType string) (dbgen.Credit, error) {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.CreateCredit(ctx, dbgen.CreateCreditParams{OrganizationID: orgID, BillingID: billingID, Type: creditType})
}

func (a *App) FindAvailableCredit(ctx context.Context, tx pgx.Tx, orgID uuid.UUID, types []string) (dbgen.Credit, error) {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.FindAvailableCredit(ctx, dbgen.FindAvailableCreditParams{OrganizationID: orgID, Types: types})
}

func (a *App) ConsumeCredit(ctx context.Context, tx pgx.Tx, creditID, eventID uuid.UUID) (dbgen.Credit, error) {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.ConsumeCredit(ctx, dbgen.ConsumeCreditParams{ID: creditID, EventID: pgtype.UUID{Bytes: eventID, Valid: true}})
}

func (a *App) GetCreditForEvent(ctx context.Context, eventID uuid.UUID) (dbgen.Credit, error) {
	return a.queries.GetCreditForEvent(ctx, pgtype.UUID{Bytes: eventID, Valid: true})
}

func (a *App) GetActiveUnlimitedLineageRoot(ctx context.Context, tx pgx.Tx, orgID uuid.UUID) (id uuid.UUID, ok bool, err error) {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	lineageID, err := q.GetActiveUnlimitedLineageRoot(ctx, orgID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, err
	}
	return lineageID, true, nil
}

func (a *App) ListCreditsForOrganization(ctx context.Context, orgID uuid.UUID) ([]dbgen.Credit, error) {
	return a.queries.ListCreditsForOrganization(ctx, orgID)
}

func (a *App) SyncCreditRestriction(ctx context.Context) ([]dbgen.Credit, error) {
	return a.queries.SyncCreditRestriction(ctx)
}

func (a *App) ListRestrictedCreditsOlderThanRepo(ctx context.Context, cutoff pgtype.Timestamptz) ([]dbgen.Credit, error) {
	return a.queries.ListRestrictedCreditsOlderThan(ctx, cutoff)
}

// -- Transactions --

func (a *App) CreateTransaction(ctx context.Context, tx pgx.Tx, orgID, planID uuid.UUID, amount int64, currency string) (dbgen.Transaction, error) {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.CreateTransaction(ctx, dbgen.CreateTransactionParams{
		OrganizationID: orgID, PlanID: planID, Amount: amount, Currency: currency,
	})
}
