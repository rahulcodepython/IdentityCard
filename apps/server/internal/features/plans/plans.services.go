package plans

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/pkg/postgres"
	"identitycard-server/internal/utils"
	"identitycard-server/internal/utils/timeutil"
)

func (a *App) List(ctx context.Context) ([]PlanResponse, error) {
	rows, err := a.ListPlans(ctx)
	if err != nil {
		return nil, utils.ErrInternal()
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
		return OrgBillingResponse{}, utils.ErrInternal()
	}
	creditRows, err := a.ListCreditsForOrganization(ctx, orgID)
	if err != nil {
		return OrgBillingResponse{}, utils.ErrInternal()
	}

	resp := OrgBillingResponse{
		Billings:        make([]BillingResponse, len(lineages)),
		Credits:         make([]CreditResponse, len(creditRows)),
		AvailableByType: map[string]int{},
	}
	for i, b := range lineages {
		plan, err := a.GetPlanByID(ctx, b.PlanID)
		if err != nil {
			return OrgBillingResponse{}, utils.ErrInternal()
		}
		resp.Billings[i] = toBillingResponse(b, plan)
	}
	for i, c := range creditRows {
		resp.Credits[i] = toCreditResponse(c)
		if !c.EventID.Valid && !c.IsRestricted {
			resp.AvailableByType[c.Type]++
		}
	}
	return resp, nil
}

func (a *App) Purchase(ctx context.Context, orgID uuid.UUID, req PurchaseRequest) (BillingResponse, error) {
	plan, err := a.GetPlanByCode(ctx, req.PlanCode)
	if err != nil {
		return BillingResponse{}, utils.NewError(http.StatusBadRequest, "invalid_plan", "unknown plan")
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
	amount := computeBillingAmount(plan, 1, quantity)

	var lineage dbgen.Billing
	txErr := postgres.WithTx(ctx, a.pool, func(tx pgx.Tx) error {
		created, err := a.CreateBillingLineageRoot(ctx, tx, orgID, plan.ID,
			pgtype.Date{Time: now, Valid: true}, pgtype.Date{Time: periodEnd, Valid: true}, "active", amount)
		if err != nil {
			return utils.ErrInternal()
		}
		lineage = created

		txn, err := a.CreateTransaction(ctx, tx, orgID, plan.ID, amount, plan.Currency)
		if err != nil {
			return utils.ErrInternal()
		}
		if _, err := a.MarkBillingPaid(ctx, tx, lineage.ID, txn.ID); err != nil {
			return utils.ErrInternal()
		}
		lineage.Status = "paid"
		lineage.TransactionID = pgtype.UUID{Bytes: txn.ID, Valid: true}

		for i := 0; i < creditTypesQuantity(plan.Kind, quantity); i++ {
			if _, err := a.CreateCredit(ctx, tx, orgID, lineage.ID, creditTypeForPlanKind(plan.Kind)); err != nil {
				return utils.ErrInternal()
			}
		}
		return nil
	})
	if txErr != nil {
		return BillingResponse{}, txErr
	}

	return toBillingResponse(lineage, plan), nil
}

func (a *App) Renew(ctx context.Context, orgID, lineageRootID uuid.UUID) (BillingResponse, error) {
	current, err := a.GetLatestBillingForLineage(ctx, lineageRootID)
	if err != nil || current.OrganizationID != orgID {
		return BillingResponse{}, utils.ErrNotFound("billing lineage")
	}
	if current.Status == "cancel" {
		return BillingResponse{}, utils.NewError(http.StatusConflict, "cancelled", "this billing lineage has been cancelled — purchase a new one")
	}
	if current.Status == "paid" {
		return BillingResponse{}, utils.NewError(http.StatusConflict, "already_paid", "the current period is already paid")
	}

	plan, err := a.GetPlanByID(ctx, current.PlanID)
	if err != nil {
		return BillingResponse{}, utils.ErrInternal()
	}

	now := dateOnly(time.Now().UTC())
	periodEnd := addCycle(now, plan.BillingCycle)
	nextNumber := int(current.BillingNumber) + 1
	amount := computeBillingAmount(plan, nextNumber, 0)

	var next dbgen.Billing
	txErr := postgres.WithTx(ctx, a.pool, func(tx pgx.Tx) error {
		txn, err := a.CreateTransaction(ctx, tx, orgID, plan.ID, amount, plan.Currency)
		if err != nil {
			return utils.ErrInternal()
		}
		if _, err := a.MarkBillingPaid(ctx, tx, current.ID, txn.ID); err != nil {
			return utils.ErrInternal()
		}

		created, err := a.CreateBillingRenewal(ctx, tx, orgID, plan.ID, lineageRootID, int32(nextNumber),
			pgtype.Date{Time: now, Valid: true}, pgtype.Date{Time: periodEnd, Valid: true}, "active", amount)
		if err != nil {
			return utils.ErrInternal()
		}
		next = created

		if plan.Kind == "flash" {
			if _, err := a.CreateCredit(ctx, tx, orgID, lineageRootID, "flash"); err != nil {
				return utils.ErrInternal()
			}
		}
		return nil
	})
	if txErr != nil {
		return BillingResponse{}, txErr
	}

	return toBillingResponse(next, plan), nil
}

func (a *App) Upgrade(ctx context.Context, orgID, lineageRootID uuid.UUID, req UpgradeRequest) (BillingResponse, error) {
	current, err := a.GetLatestBillingForLineage(ctx, lineageRootID)
	if err != nil || current.OrganizationID != orgID {
		return BillingResponse{}, utils.ErrNotFound("billing lineage")
	}
	newPlan, err := a.GetPlanByCode(ctx, req.PlanCode)
	if err != nil {
		return BillingResponse{}, utils.NewError(http.StatusBadRequest, "invalid_plan", "unknown plan")
	}
	oldPlan, err := a.GetPlanByID(ctx, current.PlanID)
	if err != nil {
		return BillingResponse{}, utils.ErrInternal()
	}
	if newPlan.Kind != oldPlan.Kind {
		return BillingResponse{}, utils.NewError(http.StatusBadRequest, "kind_mismatch", "can only upgrade within the same plan kind")
	}

	amount := computeBillingAmount(newPlan, int(current.BillingNumber), 0)
	updated, err := a.UpgradeBillingPlan(ctx, lineageRootID, newPlan.ID, amount)
	if err != nil {
		return BillingResponse{}, utils.ErrInternal()
	}
	return toBillingResponse(updated, newPlan), nil
}

func (a *App) Cancel(ctx context.Context, orgID, lineageRootID uuid.UUID) (BillingResponse, error) {
	current, err := a.GetLatestBillingForLineage(ctx, lineageRootID)
	if err != nil || current.OrganizationID != orgID {
		return BillingResponse{}, utils.ErrNotFound("billing lineage")
	}
	plan, err := a.GetPlanByID(ctx, current.PlanID)
	if err != nil {
		return BillingResponse{}, utils.ErrInternal()
	}
	updated, err := a.CancelBilling(ctx, lineageRootID)
	if err != nil {
		return BillingResponse{}, utils.ErrInternal()
	}
	return toBillingResponse(updated, plan), nil
}

// -- Credit consumption for events --

func creditTypesForEventType(eventType string) []string {
	if eventType == "flash" {
		return []string{"flash", "all"}
	}
	return []string{"events", "all"}
}

func (a *App) FindCreditTx(ctx context.Context, tx pgx.Tx, orgID uuid.UUID, eventType string) (dbgen.Credit, error) {
	credit, err := a.FindAvailableCredit(ctx, tx, orgID, creditTypesForEventType(eventType))
	if err != nil {
		if err == pgx.ErrNoRows {
			return dbgen.Credit{}, utils.NewError(http.StatusForbidden, "no_credit_available", "no available credit for this event type — purchase a plan or wait for renewal")
		}
		return dbgen.Credit{}, utils.ErrInternal()
	}
	return credit, nil
}

func (a *App) LinkCreditTx(ctx context.Context, tx pgx.Tx, creditID, eventID uuid.UUID) error {
	if _, err := a.ConsumeCredit(ctx, tx, creditID, eventID); err != nil {
		return utils.ErrInternal()
	}
	return nil
}

func (a *App) MaybeReplenishUnlimitedTx(ctx context.Context, tx pgx.Tx, orgID uuid.UUID, consumedCreditType string) error {
	if consumedCreditType != "all" {
		return nil
	}
	lineageRootID, ok, err := a.GetActiveUnlimitedLineageRoot(ctx, tx, orgID)
	if err != nil {
		return utils.ErrInternal()
	}
	if !ok {
		return nil
	}
	if _, err := a.CreateCredit(ctx, tx, orgID, lineageRootID, "all"); err != nil {
		return utils.ErrInternal()
	}
	return nil
}

func (a *App) GetCreditRestriction(ctx context.Context, eventID uuid.UUID) (bool, error) {
	credit, err := a.GetCreditForEvent(ctx, eventID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, utils.ErrInternal()
	}
	return credit.IsRestricted, nil
}

// -- Daily sweeps for internal/jobs --

func (a *App) SyncBillingStatus(ctx context.Context, today time.Time) error {
	todayDate := pgtype.Date{Time: dateOnly(today), Valid: true}
	if _, err := a.FlipLapsedBillingToPending(ctx, todayDate); err != nil {
		return err
	}
	_, err := a.SyncCreditRestriction(ctx)
	return err
}

func (a *App) ListRestrictedCreditsOlderThan(ctx context.Context, cutoff time.Time) ([]dbgen.Credit, error) {
	return a.ListRestrictedCreditsOlderThanRepo(ctx, pgtype.Timestamptz{Time: cutoff, Valid: true})
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

func computeBillingAmount(plan dbgen.Plan, billingNumber, eventCount int) int64 {
	var amount int64
	if plan.Amount.Valid {
		amount = plan.Amount.Int64
	}
	if plan.Kind == "custom" && plan.PerEventAmount.Valid {
		amount += plan.PerEventAmount.Int64 * int64(eventCount)
	}
	amount += int64(billingNumber-1) * plan.NominalIncrement
	return amount
}

func toPlanResponse(row dbgen.Plan) PlanResponse {
	return PlanResponse{
		ID: row.ID, Code: row.Code, Kind: row.Kind, BillingCycle: row.BillingCycle,
		Name: row.Name, Amount: int8Ptr(row.Amount), PerEventAmount: int8Ptr(row.PerEventAmount),
		Currency: row.Currency, EventQuota: int4Ptr(row.EventQuota), NominalIncrement: row.NominalIncrement,
	}
}

func toBillingResponse(b dbgen.Billing, plan dbgen.Plan) BillingResponse {
	return BillingResponse{
		ID: b.ID, LineageRootID: b.LineageRootID, PlanCode: plan.Code, Kind: plan.Kind, BillingCycle: plan.BillingCycle,
		BillingNumber: int(b.BillingNumber), PeriodStart: timeutil.FormatDate(b.PeriodStart), PeriodEnd: timeutil.FormatDate(b.PeriodEnd),
		Status: b.Status, Amount: b.Amount, Currency: plan.Currency, PaidAt: timestamptzPtr(b.PaidAt),
	}
}

func toCreditResponse(c dbgen.Credit) CreditResponse {
	resp := CreditResponse{
		ID: c.ID, Type: c.Type, IsRestricted: c.IsRestricted, CreatedAt: c.CreatedAt.Time.Format(time.RFC3339),
	}
	if c.EventID.Valid {
		id := uuid.UUID(c.EventID.Bytes)
		resp.EventID = &id
	}
	if c.RestrictedSince.Valid {
		s := c.RestrictedSince.Time.Format(time.RFC3339)
		resp.RestrictedSince = &s
	}
	return resp
}

func int8Ptr(v pgtype.Int8) *int64 {
	if !v.Valid {
		return nil
	}
	val := v.Int64
	return &val
}

func int4Ptr(v pgtype.Int4) *int {
	if !v.Valid {
		return nil
	}
	val := int(v.Int32)
	return &val
}

func timestamptzPtr(v pgtype.Timestamptz) *string {
	if !v.Valid {
		return nil
	}
	s := v.Time.Format(time.RFC3339)
	return &s
}
