package billing

import (
    "context"
    "fmt"
    "uuid"

    "github.com/jackc/pgx/v5"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/pkg/postgres"
    "identitycard-server/internal/utils"
)

func (a *App) GetOverview(ctx context.Context, orgID uuid.UUID) (BillingOverviewResponse, error) {
    if err := a.EnsureBillingExists(ctx, orgID); err != nil {
        return BillingOverviewResponse{}, utils.ErrInternal("Failed to initialize billing", err)
    }

    billing, err := a.GetBilling(ctx, orgID)
    if err != nil {
        return BillingOverviewResponse{}, utils.ErrInternal("Failed to fetch organization billing", err)
    }

    eventCount, err := a.CountEvents(ctx, orgID)
    if err != nil {
        return BillingOverviewResponse{}, utils.ErrInternal("Failed to count organization events", err)
    }

    txs, err := a.ListTransactions(ctx, orgID)
    if err != nil {
        return BillingOverviewResponse{}, utils.ErrInternal("Failed to list billing transactions", err)
    }

    var startStr, endStr *string
    if billing.CurrentPeriodStart != nil {
        s := billing.CurrentPeriodStart.Format("2006-01-02")
        startStr = &s
    }
    if billing.CurrentPeriodEnd != nil {
        e := billing.CurrentPeriodEnd.Format("2006-01-02")
        endStr = &e
    }

    return BillingOverviewResponse{
        CreditBalance:      billing.CreditBalance,
        AnnualFeeStatus:    billing.AnnualFeeStatus,
        CurrentPeriodStart: startStr,
        CurrentPeriodEnd:   endStr,
        EventCount:         eventCount,
        Pricing: PricingConfig{
            CreditUnitPrice:     generic.CreditUnitPrice,
            Currency:            generic.CreditCurrency,
            AnnualRenewalAmount: generic.FlatAnnualRenewalAmount,
        },
        Transactions: txs,
    }, nil
}

func (a *App) PurchaseCredits(ctx context.Context, orgID uuid.UUID, req PurchaseCreditsRequest) (BillingOverviewResponse, error) {
    if req.Quantity <= 0 {
        return BillingOverviewResponse{}, utils.ErrValidation(map[string]string{"quantity": "Quantity must be greater than 0"})
    }

    totalAmount := int64(req.Quantity) * generic.CreditUnitPrice
    desc := fmt.Sprintf("Purchased %d event creation credit(s)", req.Quantity)

    txErr := postgres.WithTx(ctx, a.pool, func(tx pgx.Tx) error {
        if err := a.EnsureBillingExists(ctx, orgID); err != nil {
            return err
        }
        return a.AddCreditsTx(ctx, tx, orgID, req.Quantity, totalAmount, generic.CreditCurrency, desc)
    })
    if txErr != nil {
        return BillingOverviewResponse{}, utils.ErrInternal("Failed to complete credit purchase", txErr)
    }

    return a.GetOverview(ctx, orgID)
}

func (a *App) RenewAnnual(ctx context.Context, orgID uuid.UUID) (BillingOverviewResponse, error) {
    eventCount, err := a.CountEvents(ctx, orgID)
    if err != nil {
        return BillingOverviewResponse{}, utils.ErrInternal("Failed to check organization events", err)
    }
    if eventCount == 0 {
        return BillingOverviewResponse{}, utils.ErrValidation(map[string]string{"organization": "Your organization has 0 events and is on the permanent free tier. Annual maintenance renewal is not required."})
    }

    desc := "Annual platform maintenance fee renewal"

    txErr := postgres.WithTx(ctx, a.pool, func(tx pgx.Tx) error {
        if err := a.EnsureBillingExists(ctx, orgID); err != nil {
            return err
        }
        return a.RenewAnnualTx(ctx, tx, orgID, generic.FlatAnnualRenewalAmount, generic.CreditCurrency, desc)
    })
    if txErr != nil {
        return BillingOverviewResponse{}, utils.ErrInternal("Failed to renew annual platform subscription", txErr)
    }

    return a.GetOverview(ctx, orgID)
}
