package plans

import (
    "context"
    "errors"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"

    "identitycard-server/internal/pkg/postgres"
)

// -- Plans Catalog --

func (a *App) ListPlans(ctx context.Context) ([]PlanDB, error) {
    return postgres.QueryJSONSlice[PlanDB](ctx, a.pool, ListPlansQuery)
}

func (a *App) GetPlanByCode(ctx context.Context, code string) (*PlanDB, error) {
    return postgres.QueryJSON[PlanDB](ctx, a.pool, GetPlanByCodeQuery, code)
}

func (a *App) GetPlanByID(ctx context.Context, id uuid.UUID) (*PlanDB, error) {
    return postgres.QueryJSON[PlanDB](ctx, a.pool, GetPlanByIDQuery, id)
}

// -- Billing Lineages --

func (a *App) CreateBillingLineageRoot(ctx context.Context, tx pgx.Tx, orgID, planID uuid.UUID, periodStart, periodEnd time.Time, status string, amount int64) (*BillingDB, error) {
    if tx != nil {
        return postgres.QueryJSONTx[BillingDB](ctx, tx, CreateBillingLineageRootQuery, orgID, planID, periodStart, periodEnd, status, amount)
    }
    return postgres.QueryJSON[BillingDB](ctx, a.pool, CreateBillingLineageRootQuery, orgID, planID, periodStart, periodEnd, status, amount)
}

func (a *App) CreateBillingRenewal(ctx context.Context, tx pgx.Tx, orgID, planID, lineageRootID uuid.UUID, billingNumber int, periodStart, periodEnd time.Time, status string, amount int64) (*BillingDB, error) {
    if tx != nil {
        return postgres.QueryJSONTx[BillingDB](ctx, tx, CreateBillingRenewalQuery, orgID, planID, lineageRootID, billingNumber, periodStart, periodEnd, status, amount)
    }
    return postgres.QueryJSON[BillingDB](ctx, a.pool, CreateBillingRenewalQuery, orgID, planID, lineageRootID, billingNumber, periodStart, periodEnd, status, amount)
}

func (a *App) GetLatestBillingForLineage(ctx context.Context, lineageRootID uuid.UUID) (*BillingDB, error) {
    return postgres.QueryJSON[BillingDB](ctx, a.pool, GetLatestBillingForLineageQuery, lineageRootID)
}

func (a *App) ListLatestBillingForOrganization(ctx context.Context, orgID uuid.UUID) ([]BillingDB, error) {
    return postgres.QueryJSONSlice[BillingDB](ctx, a.pool, ListLatestBillingForOrganizationQuery, orgID)
}

func (a *App) MarkBillingPaid(ctx context.Context, tx pgx.Tx, id, transactionID uuid.UUID) (*BillingDB, error) {
    if tx != nil {
        return postgres.QueryJSONTx[BillingDB](ctx, tx, MarkBillingPaidQuery, id, transactionID)
    }
    return postgres.QueryJSON[BillingDB](ctx, a.pool, MarkBillingPaidQuery, id, transactionID)
}

func (a *App) CancelBilling(ctx context.Context, lineageRootID uuid.UUID) (*BillingDB, error) {
    return postgres.QueryJSON[BillingDB](ctx, a.pool, CancelBillingLineageQuery, lineageRootID)
}

func (a *App) UpgradeBillingPlan(ctx context.Context, lineageRootID, planID uuid.UUID, amount int64) (*BillingDB, error) {
    return postgres.QueryJSON[BillingDB](ctx, a.pool, UpgradeBillingPlanQuery, lineageRootID, planID, amount)
}

func (a *App) FlipLapsedBillingToPending(ctx context.Context, today time.Time) ([]BillingDB, error) {
    return postgres.QueryJSONSlice[BillingDB](ctx, a.pool, FlipLapsedBillingToPendingQuery, today)
}

// -- Credits --

func (a *App) CreateCredit(ctx context.Context, tx pgx.Tx, orgID, billingID uuid.UUID, creditType string) (*CreditDB, error) {
    if tx != nil {
        return postgres.QueryJSONTx[CreditDB](ctx, tx, CreateCreditQuery, orgID, billingID, creditType)
    }
    return postgres.QueryJSON[CreditDB](ctx, a.pool, CreateCreditQuery, orgID, billingID, creditType)
}

func (a *App) FindAvailableCredit(ctx context.Context, tx pgx.Tx, orgID uuid.UUID, types []string) (*CreditDB, error) {
    if tx != nil {
        return postgres.QueryJSONTx[CreditDB](ctx, tx, FindAvailableCreditQuery, orgID, types)
    }
    return postgres.QueryJSON[CreditDB](ctx, a.pool, FindAvailableCreditQuery, orgID, types)
}

func (a *App) ConsumeCredit(ctx context.Context, tx pgx.Tx, creditID, eventID uuid.UUID) (*CreditDB, error) {
    if tx != nil {
        return postgres.QueryJSONTx[CreditDB](ctx, tx, ConsumeCreditQuery, creditID, eventID)
    }
    return postgres.QueryJSON[CreditDB](ctx, a.pool, ConsumeCreditQuery, creditID, eventID)
}

func (a *App) GetCreditForEvent(ctx context.Context, eventID uuid.UUID) (*CreditDB, error) {
    return postgres.QueryJSON[CreditDB](ctx, a.pool, GetCreditForEventQuery, eventID)
}

func (a *App) GetActiveUnlimitedLineageRoot(ctx context.Context, tx pgx.Tx, orgID uuid.UUID) (uuid.UUID, bool, error) {
    var lineageRootID uuid.UUID
    var err error
    if tx != nil {
        err = tx.QueryRow(ctx, GetActiveUnlimitedLineageRootQuery, orgID).Scan(&lineageRootID)
    } else {
        err = a.pool.QueryRow(ctx, GetActiveUnlimitedLineageRootQuery, orgID).Scan(&lineageRootID)
    }
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, postgres.ErrNotFound) {
            return uuid.Nil, false, nil
        }
        return uuid.Nil, false, postgres.MapPgError(err)
    }
    return lineageRootID, true, nil
}

func (a *App) ListCreditsForOrganization(ctx context.Context, orgID uuid.UUID) ([]CreditDB, error) {
    return postgres.QueryJSONSlice[CreditDB](ctx, a.pool, ListCreditsForOrganizationQuery, orgID)
}

func (a *App) SyncCreditRestriction(ctx context.Context) ([]CreditDB, error) {
    return postgres.QueryJSONSlice[CreditDB](ctx, a.pool, SyncCreditRestrictionQuery)
}

func (a *App) ListRestrictedCreditsOlderThanRepo(ctx context.Context, cutoff time.Time) ([]CreditDB, error) {
    return postgres.QueryJSONSlice[CreditDB](ctx, a.pool, ListRestrictedCreditsOlderThanQuery, cutoff)
}

// -- Transactions --

func (a *App) CreateTransaction(ctx context.Context, tx pgx.Tx, orgID, planID uuid.UUID, amount int64, currency string) (*TransactionDB, error) {
    if tx != nil {
        return postgres.QueryJSONTx[TransactionDB](ctx, tx, CreateTransactionQuery, orgID, planID, amount, currency)
    }
    return postgres.QueryJSON[TransactionDB](ctx, a.pool, CreateTransactionQuery, orgID, planID, amount, currency)
}
