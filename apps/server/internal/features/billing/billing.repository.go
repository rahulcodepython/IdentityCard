package billing

import (
    "context"
    "fmt"
    "uuid"

    "github.com/jackc/pgx/v5"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/pkg/postgres"
)

func (a *App) EnsureBillingExists(ctx context.Context, orgID uuid.UUID) error {
    _, err := a.pool.Exec(ctx, EnsureOrganizationBillingQuery, orgID, generic.InitialSignupCredits)
    return err
}

func (a *App) GetBilling(ctx context.Context, orgID uuid.UUID) (*OrganizationBillingDB, error) {
    return postgres.QueryJSON[OrganizationBillingDB](ctx, a.pool, GetOrganizationBillingQuery, orgID)
}

func (a *App) CountEvents(ctx context.Context, orgID uuid.UUID) (int, error) {
    var count int
    err := a.pool.QueryRow(ctx, CountEventsForOrgQuery, orgID).Scan(&count)
    return count, err
}

func (a *App) ListTransactions(ctx context.Context, orgID uuid.UUID) ([]BillingTransactionDB, error) {
    rows, err := postgres.QueryJSON[[]BillingTransactionDB](ctx, a.pool, ListBillingTransactionsQuery, orgID)
    if err != nil {
        return nil, err
    }
    if rows == nil {
        return []BillingTransactionDB{}, nil
    }
    return *rows, nil
}

func (a *App) AddCreditsTx(ctx context.Context, tx pgx.Tx, orgID uuid.UUID, quantity int, amount int64, currency, description string) error {
    var newBalance int
    err := tx.QueryRow(ctx, AddCreditsQuery, orgID, quantity).Scan(&newBalance)
    if err != nil {
        return fmt.Errorf("failed to update credit balance: %w", err)
    }

    _, err = tx.Exec(ctx, InsertBillingTransactionQuery, orgID, generic.TxTypeCreditPurchase, quantity, amount, currency, nil, description)
    if err != nil {
        return fmt.Errorf("failed to record credit transaction: %w", err)
    }
    return nil
}

func (a *App) RenewAnnualTx(ctx context.Context, tx pgx.Tx, orgID uuid.UUID, amount int64, currency, description string) error {
    _, err := tx.Exec(ctx, RenewAnnualSubscriptionQuery, orgID)
    if err != nil {
        return fmt.Errorf("failed to renew annual subscription: %w", err)
    }

    _, err = tx.Exec(ctx, InsertBillingTransactionQuery, orgID, generic.TxTypeAnnualRenewal, 0, amount, currency, nil, description)
    if err != nil {
        return fmt.Errorf("failed to record renewal transaction: %w", err)
    }
    return nil
}
