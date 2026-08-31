package plans

import (
    "context"
    "errors"
    "net/http"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/pkg/postgres"
    "identitycard-server/internal/utils"
    "identitycard-server/internal/utils/timeutil"
)

func (a *App) List(ctx context.Context) ([]PlanResponse, error) {
    rows, err := a.ListPlans(ctx)
    if err != nil {
        return nil, utils.ErrInternal("Failed to list plans.", err)
    }
    resp := make([]PlanResponse, len(rows))
    for i, row := range rows {
        resp[i] = toPlanResponse(row)
    }
    return resp, nil
}

func (a *App) ListBilling(ctx context.Context, orgID uuid.UUID) (OrgBillingResponse, error) {
    lineages, err := a.ListLatestBillingForOrganization(ctx, orgID)
    if err != nil {
        return OrgBillingResponse{}, utils.ErrInternal("Failed to list billing records.", err)
    }
    creditRows, err := a.ListCreditsForOrganization(ctx, orgID)
    if err != nil {
        return OrgBillingResponse{}, utils.ErrInternal("Failed to list credits.", err)
    }

    resp := OrgBillingResponse{
        Billings:        make([]BillingResponse, len(lineages)),
        Credits:         make([]CreditResponse, len(creditRows)),
        AvailableByType: map[string]int{},
    }
    for i, b := range lineages {
        plan, err := a.GetPlanByID(ctx, b.PlanID)
        if err != nil {
            return OrgBillingResponse{}, utils.ErrInternal("Failed to fetch plan for billing.", err)
        }
        resp.Billings[i] = toBillingResponse(b, *plan)
    }
    for i, c := range creditRows {
        resp.Credits[i] = toCreditResponse(c)
        if c.EventID == nil && !c.IsRestricted {
            resp.AvailableByType[c.Type]++
        }
    }
    return resp, nil
}

func (a *App) Purchase(ctx context.Context, orgID uuid.UUID, req PurchaseRequest) (BillingResponse, error) {
    plan, err := a.GetPlanByCode(ctx, req.PlanCode)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) || errors.Is(err, generic.ErrPlansNotFound) {
            return BillingResponse{}, utils.NewError(http.StatusBadRequest, "Unknown plan code.", err)
        }
        return BillingResponse{}, utils.ErrInternal("Failed to fetch plan.", err)
    }

    quantity := 1
    if plan.Kind == "custom" {
        if req.EventQuantity == nil {
            return BillingResponse{}, utils.ErrValidation(map[string]string{"event_quantity": "required for a custom plan"})
        }
        quantity = *req.EventQuantity
    }

    now := dateOnly(time.Now().UTC())
    periodEnd := addCycle(now, plan.BillingCycle)
    amount := computeBillingAmount(*plan, 1, quantity)

    var lineage *BillingDB
    txErr := postgres.WithTx(ctx, a.pool, func(tx pgx.Tx) error {
        created, err := a.CreateBillingLineageRoot(ctx, tx, orgID, plan.ID, now, periodEnd, "active", amount)
        if err != nil {
            return err
        }
        lineage = created

        txn, err := a.CreateTransaction(ctx, tx, orgID, plan.ID, amount, plan.Currency)
        if err != nil {
            return err
        }
        if _, err := a.MarkBillingPaid(ctx, tx, lineage.ID, txn.ID); err != nil {
            return err
        }
        lineage.Status = "paid"
        lineage.TransactionID = &txn.ID

        for i := 0; i < creditTypesQuantity(plan.Kind, quantity); i++ {
            if _, err := a.CreateCredit(ctx, tx, orgID, lineage.ID, creditTypeForPlanKind(plan.Kind)); err != nil {
                return err
            }
        }
        return nil
    })
    if txErr != nil {
        return BillingResponse{}, utils.ErrInternal("Failed to complete purchase.", txErr)
    }

    return toBillingResponse(*lineage, *plan), nil
}

func (a *App) Renew(ctx context.Context, orgID, lineageRootID uuid.UUID) (BillingResponse, error) {
    current, err := a.GetLatestBillingForLineage(ctx, lineageRootID)
    if err != nil || current.OrganizationID != orgID {
        return BillingResponse{}, utils.ErrNotFound("Billing lineage not found.", err)
    }
    if current.Status == "cancel" {
        return BillingResponse{}, utils.NewError(http.StatusConflict, "This billing lineage has been cancelled — purchase a new one.", nil)
    }
    if current.Status == "paid" {
        return BillingResponse{}, utils.NewError(http.StatusConflict, "The current period is already paid.", nil)
    }

    plan, err := a.GetPlanByID(ctx, current.PlanID)
    if err != nil {
        return BillingResponse{}, utils.ErrInternal("Failed to fetch plan.", err)
    }

    now := dateOnly(time.Now().UTC())
    periodEnd := addCycle(now, plan.BillingCycle)
    nextNumber := current.BillingNumber + 1
    amount := computeBillingAmount(*plan, nextNumber, 0)

    var next *BillingDB
    txErr := postgres.WithTx(ctx, a.pool, func(tx pgx.Tx) error {
        txn, err := a.CreateTransaction(ctx, tx, orgID, plan.ID, amount, plan.Currency)
        if err != nil {
            return err
        }
        if _, err := a.MarkBillingPaid(ctx, tx, current.ID, txn.ID); err != nil {
            return err
        }

        created, err := a.CreateBillingRenewal(ctx, tx, orgID, plan.ID, lineageRootID, nextNumber, now, periodEnd, "active", amount)
        if err != nil {
            return err
        }
        next = created

        if plan.Kind == "flash" {
            if _, err := a.CreateCredit(ctx, tx, orgID, lineageRootID, "flash"); err != nil {
                return err
            }
        }
        return nil
    })
    if txErr != nil {
        return BillingResponse{}, utils.ErrInternal("Failed to renew billing.", txErr)
    }

    return toBillingResponse(*next, *plan), nil
}

func (a *App) Upgrade(ctx context.Context, orgID, lineageRootID uuid.UUID, req UpgradeRequest) (BillingResponse, error) {
    current, err := a.GetLatestBillingForLineage(ctx, lineageRootID)
    if err != nil || current.OrganizationID != orgID {
        return BillingResponse{}, utils.ErrNotFound("Billing lineage not found.", err)
    }
    newPlan, err := a.GetPlanByCode(ctx, req.PlanCode)
    if err != nil {
        return BillingResponse{}, utils.NewError(http.StatusBadRequest, "Unknown plan.", err)
    }
    oldPlan, err := a.GetPlanByID(ctx, current.PlanID)
    if err != nil {
        return BillingResponse{}, utils.ErrInternal("Failed to fetch current plan.", err)
    }
    if newPlan.Kind != oldPlan.Kind {
        return BillingResponse{}, utils.NewError(http.StatusBadRequest, "Can only upgrade within the same plan kind.", nil)
    }

    amount := computeBillingAmount(*newPlan, current.BillingNumber, 0)
    updated, err := a.UpgradeBillingPlan(ctx, lineageRootID, newPlan.ID, amount)
    if err != nil {
        return BillingResponse{}, utils.ErrInternal("Failed to upgrade billing plan.", err)
    }
    return toBillingResponse(*updated, *newPlan), nil
}

func (a *App) Cancel(ctx context.Context, orgID, lineageRootID uuid.UUID) (BillingResponse, error) {
    current, err := a.GetLatestBillingForLineage(ctx, lineageRootID)
    if err != nil || current.OrganizationID != orgID {
        return BillingResponse{}, utils.ErrNotFound("Billing lineage not found.", err)
    }
    plan, err := a.GetPlanByID(ctx, current.PlanID)
    if err != nil {
        return BillingResponse{}, utils.ErrInternal("Failed to fetch plan.", err)
    }
    updated, err := a.CancelBilling(ctx, lineageRootID)
    if err != nil {
        return BillingResponse{}, utils.ErrInternal("Failed to cancel billing lineage.", err)
    }
    return toBillingResponse(*updated, *plan), nil
}

// -- Credit consumption for events --

func creditTypesForEventType(eventType string) []string {
    if eventType == "flash" {
        return []string{"flash", "all"}
    }
    return []string{"events", "all"}
}

func (a *App) FindCreditTx(ctx context.Context, tx pgx.Tx, orgID uuid.UUID, eventType string) (*CreditDB, error) {
    credit, err := a.FindAvailableCredit(ctx, tx, orgID, creditTypesForEventType(eventType))
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) || errors.Is(err, generic.ErrPlansNoCredits) {
            return nil, utils.NewError(http.StatusForbidden, "No available credit for this event type — purchase a plan or wait for renewal.", err)
        }
        return nil, utils.ErrInternal("Failed to find available credit.", err)
    }
    return credit, nil
}

func (a *App) LinkCreditTx(ctx context.Context, tx pgx.Tx, creditID, eventID uuid.UUID) error {
    if _, err := a.ConsumeCredit(ctx, tx, creditID, eventID); err != nil {
        return utils.ErrInternal("Failed to link credit to event.", err)
    }
    return nil
}

func (a *App) MaybeReplenishUnlimitedTx(ctx context.Context, tx pgx.Tx, orgID uuid.UUID, consumedCreditType string) error {
    if consumedCreditType != "all" {
        return nil
    }
    lineageRootID, ok, err := a.GetActiveUnlimitedLineageRoot(ctx, tx, orgID)
    if err != nil {
        return utils.ErrInternal("Failed to check unlimited lineage.", err)
    }
    if !ok {
        return nil
    }
    if _, err := a.CreateCredit(ctx, tx, orgID, lineageRootID, "all"); err != nil {
        return utils.ErrInternal("Failed to replenish unlimited credit.", err)
    }
    return nil
}

func (a *App) GetCreditRestriction(ctx context.Context, eventID uuid.UUID) (bool, error) {
    credit, err := a.GetCreditForEvent(ctx, eventID)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) {
            return false, nil
        }
        return false, utils.ErrInternal("Failed to get credit restriction.", err)
    }
    return credit.IsRestricted, nil
}

// -- Daily sweeps for internal/jobs --

func (a *App) SyncBillingStatus(ctx context.Context, today time.Time) error {
    if _, err := a.FlipLapsedBillingToPending(ctx, dateOnly(today)); err != nil {
        return err
    }
    _, err := a.SyncCreditRestriction(ctx)
    return err
}

func (a *App) ListRestrictedCreditsOlderThan(ctx context.Context, cutoff time.Time) ([]CreditDB, error) {
    return a.ListRestrictedCreditsOlderThanRepo(ctx, cutoff)
}

// -- Helpers --

func dateOnly(t time.Time) time.Time {
    return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func addCycle(t time.Time, billingCycle string) time.Time {
    switch billingCycle {
    case "daily":
        return t.AddDate(0, 0, 1)
    case "yearly":
        return t.AddDate(1, 0, 0)
    default:
        return t.AddDate(0, 1, 0)
    }
}

func creditTypeForPlanKind(kind string) string {
    switch kind {
    case "flash":
        return "flash"
    case "unlimited":
        return "all"
    default:
        return "events"
    }
}

func creditTypesQuantity(kind string, requestedQuantity int) int {
    if kind == "custom" {
        return requestedQuantity
    }
    return 1
}

func computeBillingAmount(plan PlanDB, billingNumber, eventCount int) int64 {
    var amount int64
    if plan.Amount != nil {
        amount = *plan.Amount
    }
    if plan.Kind == "custom" && plan.PerEventAmount != nil {
        amount += *plan.PerEventAmount * int64(eventCount)
    }
    amount += int64(billingNumber-1) * plan.NominalIncrement
    return amount
}

func toPlanResponse(row PlanDB) PlanResponse {
    return PlanResponse{
        ID:               row.ID,
        Code:             row.Code,
        Kind:             row.Kind,
        BillingCycle:     row.BillingCycle,
        Name:             row.Name,
        Amount:           row.Amount,
        PerEventAmount:   row.PerEventAmount,
        Currency:         row.Currency,
        EventQuota:       row.EventQuota,
        NominalIncrement: row.NominalIncrement,
    }
}

func toBillingResponse(b BillingDB, plan PlanDB) BillingResponse {
    var lineageRoot uuid.UUID
    if b.LineageRootID != nil {
        lineageRoot = *b.LineageRootID
    }
    var paidAt *string
    if b.PaidAt != nil {
        s := b.PaidAt.Format(time.RFC3339)
        paidAt = &s
    }
    return BillingResponse{
        ID:            b.ID,
        LineageRootID: lineageRoot,
        PlanCode:      plan.Code,
        Kind:          plan.Kind,
        BillingCycle:  plan.BillingCycle,
        BillingNumber: b.BillingNumber,
        PeriodStart:   timeutil.FormatDateDirect(b.PeriodStart),
        PeriodEnd:     timeutil.FormatDateDirect(b.PeriodEnd),
        Status:        b.Status,
        Amount:        b.Amount,
        Currency:      plan.Currency,
        PaidAt:        paidAt,
    }
}

func toCreditResponse(c CreditDB) CreditResponse {
    resp := CreditResponse{
        ID:           c.ID,
        Type:         c.Type,
        IsRestricted: c.IsRestricted,
        CreatedAt:    c.CreatedAt.Format(time.RFC3339),
    }
    if c.EventID != nil {
        resp.EventID = c.EventID
    }
    if c.RestrictedSince != nil {
        s := c.RestrictedSince.Format(time.RFC3339)
        resp.RestrictedSince = &s
    }
    return resp
}
