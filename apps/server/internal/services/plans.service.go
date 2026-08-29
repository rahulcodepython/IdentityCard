// Package services: plans owns the plan catalog plus the whole
// credits/billing/transactions purchase model — a purchase mints one or
// more single-use event-creation credits (credits.repository.go), and a
// separate billing lineage (billing.repository.go) tracks the recurring
// payment obligation that gates whether those credits' events stay
// usable. See the migration's doc comments for the schema-level design.
package services

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/entities"
	"identitycard-server/internal/pkg/postgres"
	"identitycard-server/internal/repositories"
	"identitycard-server/internal/utils"
	"identitycard-server/internal/utils/timeutil"
)

type PlansService struct {
	plans        *repositories.PlansRepository
	billing      *repositories.BillingRepository
	credits      *repositories.CreditsRepository
	transactions *repositories.TransactionsRepository

	// Used by every method that writes across transactions/billing/
	// credits atomically — List/ListBilling read through the repos alone.
	pool        *pgxpool.Pool
	baseQueries *dbgen.Queries
}

func NewPlansService(
	plans *repositories.PlansRepository,
	billing *repositories.BillingRepository,
	credits *repositories.CreditsRepository,
	transactions *repositories.TransactionsRepository,
	pool *pgxpool.Pool,
	baseQueries *dbgen.Queries,
) *PlansService {
	return &PlansService{plans: plans, billing: billing, credits: credits, transactions: transactions, pool: pool, baseQueries: baseQueries}
}

func (s *PlansService) List(ctx context.Context) ([]entities.PlanResponse, error) {
	rows, err := s.plans.List(ctx)
	if err != nil {
		return nil, utils.ErrInternal()
	}
	resp := make([]entities.PlanResponse, len(rows))
	for i, row := range rows {
		resp[i] = toPlanResponse(row)
	}
	return resp, nil
}

// ListBilling returns orgID's whole billing picture: the latest row of
// every lineage it has ever started, its full credit ledger, and a
// computed available-by-type count.
func (s *PlansService) ListBilling(ctx context.Context, orgID uuid.UUID) (entities.OrgBillingResponse, error) {
	lineages, err := s.billing.ListLatestForOrganization(ctx, orgID)
	if err != nil {
		return entities.OrgBillingResponse{}, utils.ErrInternal()
	}
	creditRows, err := s.credits.ListForOrganization(ctx, orgID)
	if err != nil {
		return entities.OrgBillingResponse{}, utils.ErrInternal()
	}

	resp := entities.OrgBillingResponse{
		Billings:        make([]entities.BillingResponse, len(lineages)),
		Credits:         make([]entities.CreditResponse, len(creditRows)),
		AvailableByType: map[string]int{},
	}
	for i, b := range lineages {
		plan, err := s.plans.GetPlanByID(ctx, b.PlanID)
		if err != nil {
			return entities.OrgBillingResponse{}, utils.ErrInternal()
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

// Purchase starts a new billing lineage for orgID and mints its initial
// credits. Always additive — buying more Custom capacity on top of an
// existing Base/Custom/Unlimited lineage starts a new independent
// lineage rather than replacing anything (confirmed: top-ups stack, each
// with its own billing cycle and restriction timeline).
func (s *PlansService) Purchase(ctx context.Context, orgID uuid.UUID, req entities.PurchaseRequest) (entities.BillingResponse, error) {
	plan, err := s.plans.GetPlanByCode(ctx, req.PlanCode)
	if err != nil {
		return entities.BillingResponse{}, utils.NewError(http.StatusBadRequest, "invalid_plan", "unknown plan")
	}

	quantity := 1
	if plan.Kind == "custom" {
		if req.EventQuantity == nil {
			return entities.BillingResponse{}, utils.ErrValidation(map[string]string{"event_quantity": "required for a custom plan"})
		}
		quantity = *req.EventQuantity
	}

	now := dateOnly(time.Now().UTC())
	periodEnd := addCycle(now, plan.BillingCycle)
	amount := computeBillingAmount(plan, 1, quantity)

	var lineage dbgen.Billing
	txErr := postgres.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		billingRepo := s.billing.WithTx(tx)
		creditsRepo := s.credits.WithTx(tx)
		txRepo := s.transactions.WithTx(tx)

		created, err := billingRepo.CreateLineageRoot(ctx, orgID, plan.ID,
			pgtype.Date{Time: now, Valid: true}, pgtype.Date{Time: periodEnd, Valid: true}, "active", amount)
		if err != nil {
			return utils.ErrInternal()
		}
		lineage = created

		txn, err := txRepo.Create(ctx, orgID, plan.ID, amount, plan.Currency)
		if err != nil {
			return utils.ErrInternal()
		}
		if _, err := billingRepo.MarkPaid(ctx, lineage.ID, txn.ID); err != nil {
			return utils.ErrInternal()
		}
		lineage.Status = "paid"
		lineage.TransactionID = pgtype.UUID{Bytes: txn.ID, Valid: true}

		for i := 0; i < creditTypesQuantity(plan.Kind, quantity); i++ {
			if _, err := creditsRepo.Create(ctx, orgID, lineage.ID, creditTypeForPlanKind(plan.Kind)); err != nil {
				return utils.ErrInternal()
			}
		}
		return nil
	})
	if txErr != nil {
		return entities.BillingResponse{}, txErr
	}

	return toBillingResponse(lineage, plan), nil
}

// Renew records payment for lineageRootID's current period and spawns
// the next one. For a flash lineage this also grants one fresh flash
// credit as a side effect — flash billing recurs daily, and paying it is
// literally how an org gets "one more" flash credit; non-flash renewals
// grant no new credits, they purely extend restriction-free access to
// the events already funded by that lineage.
func (s *PlansService) Renew(ctx context.Context, orgID, lineageRootID uuid.UUID) (entities.BillingResponse, error) {
	current, err := s.billing.GetLatestForLineage(ctx, lineageRootID)
	if err != nil || current.OrganizationID != orgID {
		return entities.BillingResponse{}, utils.ErrNotFound("billing lineage")
	}
	if current.Status == "cancel" {
		return entities.BillingResponse{}, utils.NewError(http.StatusConflict, "cancelled", "this billing lineage has been cancelled — purchase a new one")
	}
	if current.Status == "paid" {
		return entities.BillingResponse{}, utils.NewError(http.StatusConflict, "already_paid", "the current period is already paid")
	}

	plan, err := s.plans.GetPlanByID(ctx, current.PlanID)
	if err != nil {
		return entities.BillingResponse{}, utils.ErrInternal()
	}

	now := dateOnly(time.Now().UTC())
	periodEnd := addCycle(now, plan.BillingCycle)
	nextNumber := int(current.BillingNumber) + 1
	amount := computeBillingAmount(plan, nextNumber, 0)

	var next dbgen.Billing
	txErr := postgres.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		billingRepo := s.billing.WithTx(tx)
		creditsRepo := s.credits.WithTx(tx)
		txRepo := s.transactions.WithTx(tx)

		txn, err := txRepo.Create(ctx, orgID, plan.ID, amount, plan.Currency)
		if err != nil {
			return utils.ErrInternal()
		}
		if _, err := billingRepo.MarkPaid(ctx, current.ID, txn.ID); err != nil {
			return utils.ErrInternal()
		}

		created, err := billingRepo.CreateRenewal(ctx, orgID, plan.ID, lineageRootID, int32(nextNumber),
			pgtype.Date{Time: now, Valid: true}, pgtype.Date{Time: periodEnd, Valid: true}, "active", amount)
		if err != nil {
			return utils.ErrInternal()
		}
		next = created

		if plan.Kind == "flash" {
			if _, err := creditsRepo.Create(ctx, orgID, lineageRootID, "flash"); err != nil {
				return utils.ErrInternal()
			}
		}
		return nil
	})
	if txErr != nil {
		return entities.BillingResponse{}, txErr
	}

	return toBillingResponse(next, plan), nil
}

// Upgrade switches lineageRootID's current period to a new plan in
// place — no new transaction/billing/credits rows spawned immediately;
// the new plan's terms (and credit grants) apply starting the next
// Renew. Only plans of the same kind can be swapped to (switching kind
// would change what a credit from this lineage even means).
func (s *PlansService) Upgrade(ctx context.Context, orgID, lineageRootID uuid.UUID, req entities.UpgradeRequest) (entities.BillingResponse, error) {
	current, err := s.billing.GetLatestForLineage(ctx, lineageRootID)
	if err != nil || current.OrganizationID != orgID {
		return entities.BillingResponse{}, utils.ErrNotFound("billing lineage")
	}
	newPlan, err := s.plans.GetPlanByCode(ctx, req.PlanCode)
	if err != nil {
		return entities.BillingResponse{}, utils.NewError(http.StatusBadRequest, "invalid_plan", "unknown plan")
	}
	oldPlan, err := s.plans.GetPlanByID(ctx, current.PlanID)
	if err != nil {
		return entities.BillingResponse{}, utils.ErrInternal()
	}
	if newPlan.Kind != oldPlan.Kind {
		return entities.BillingResponse{}, utils.NewError(http.StatusBadRequest, "kind_mismatch", "can only upgrade within the same plan kind")
	}

	amount := computeBillingAmount(newPlan, int(current.BillingNumber), 0)
	updated, err := s.billing.UpgradePlan(ctx, lineageRootID, newPlan.ID, amount)
	if err != nil {
		return entities.BillingResponse{}, utils.ErrInternal()
	}
	return toBillingResponse(updated, newPlan), nil
}

// Cancel marks lineageRootID's current period 'cancel' — the next daily
// sync restricts every credit funded by it (see SyncBillingStatus).
func (s *PlansService) Cancel(ctx context.Context, orgID, lineageRootID uuid.UUID) (entities.BillingResponse, error) {
	current, err := s.billing.GetLatestForLineage(ctx, lineageRootID)
	if err != nil || current.OrganizationID != orgID {
		return entities.BillingResponse{}, utils.ErrNotFound("billing lineage")
	}
	plan, err := s.plans.GetPlanByID(ctx, current.PlanID)
	if err != nil {
		return entities.BillingResponse{}, utils.ErrInternal()
	}
	updated, err := s.billing.Cancel(ctx, lineageRootID)
	if err != nil {
		return entities.BillingResponse{}, utils.ErrInternal()
	}
	return toBillingResponse(updated, plan), nil
}

// -- credit consumption (transaction-scoped; called by events' service
// from inside its own WithTx, alongside the event insert) --

// creditTypesForEventType returns the acceptable credit types for
// creating an event of eventType, required-type first — an 'all'
// (unlimited-plan) credit always covers any event type.
func creditTypesForEventType(eventType string) []string {
	if eventType == "flash" {
		return []string{"flash", "all"}
	}
	return []string{"events", "all"}
}

// FindCreditTx locks and returns the oldest available credit that can
// fund an event of eventType, or a 403 no_credit_available error if none
// exists. Must run inside the same tx as the event insert that will
// consume it.
func (s *PlansService) FindCreditTx(ctx context.Context, tx pgx.Tx, orgID uuid.UUID, eventType string) (dbgen.Credit, error) {
	credit, err := s.credits.WithTx(tx).FindAvailable(ctx, orgID, creditTypesForEventType(eventType))
	if err != nil {
		if err == pgx.ErrNoRows {
			return dbgen.Credit{}, utils.NewError(http.StatusForbidden, "no_credit_available", "no available credit for this event type — purchase a plan or wait for renewal")
		}
		return dbgen.Credit{}, utils.ErrInternal()
	}
	return credit, nil
}

// LinkCreditTx marks creditID consumed by eventID.
func (s *PlansService) LinkCreditTx(ctx context.Context, tx pgx.Tx, creditID, eventID uuid.UUID) error {
	if _, err := s.credits.WithTx(tx).Consume(ctx, creditID, eventID); err != nil {
		return utils.ErrInternal()
	}
	return nil
}

// MaybeReplenishUnlimitedTx re-grants an 'all' credit immediately after
// one is consumed, if orgID still has an active/paid unlimited lineage —
// a no-op for any other consumedCreditType.
func (s *PlansService) MaybeReplenishUnlimitedTx(ctx context.Context, tx pgx.Tx, orgID uuid.UUID, consumedCreditType string) error {
	if consumedCreditType != "all" {
		return nil
	}
	creditsRepo := s.credits.WithTx(tx)
	lineageRootID, ok, err := creditsRepo.GetActiveUnlimitedLineageRoot(ctx, orgID)
	if err != nil {
		return utils.ErrInternal()
	}
	if !ok {
		return nil
	}
	if _, err := creditsRepo.Create(ctx, orgID, lineageRootID, "all"); err != nil {
		return utils.ErrInternal()
	}
	return nil
}

// GetCreditRestriction reports whether eventID's funding credit is
// currently restricted — events' service gates event/sub-event updates
// on this.
func (s *PlansService) GetCreditRestriction(ctx context.Context, eventID uuid.UUID) (bool, error) {
	credit, err := s.credits.GetForEvent(ctx, eventID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, utils.ErrInternal()
	}
	return credit.IsRestricted, nil
}

// -- daily cron sweeps (used by internal/jobs; plain errors, not utils) --

// SyncBillingStatus flips every lineage's lapsed 'active' period to
// 'pending', then recomputes every credit's is_restricted from its
// lineage's current status.
func (s *PlansService) SyncBillingStatus(ctx context.Context, today time.Time) error {
	todayDate := pgtype.Date{Time: dateOnly(today), Valid: true}
	if _, err := s.billing.FlipLapsedToPending(ctx, todayDate); err != nil {
		return err
	}
	_, err := s.credits.SyncRestriction(ctx)
	return err
}

// ListRestrictedCreditsOlderThan returns credits that have been
// restricted for 30+ days and still fund a live event — the cleanup
// cron's second deletion criterion (see EventsService.ListFlashOlderThan
// for the first). The caller deletes each credit's EventID.
func (s *PlansService) ListRestrictedCreditsOlderThan(ctx context.Context, cutoff time.Time) ([]dbgen.Credit, error) {
	return s.credits.ListRestrictedOlderThan(ctx, pgtype.Timestamptz{Time: cutoff, Valid: true})
}

// -- helpers --

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func addCycle(t time.Time, billingCycle string) time.Time {
	switch billingCycle {
	case "daily":
		return t.AddDate(0, 0, 1)
	case "yearly":
		return t.AddDate(1, 0, 0)
	default: // monthly
		return t.AddDate(0, 1, 0)
	}
}

// creditTypeForPlanKind maps a plan kind to the type of credit it mints.
func creditTypeForPlanKind(kind string) string {
	switch kind {
	case "flash":
		return "flash"
	case "unlimited":
		return "all"
	default: // base, custom
		return "events"
	}
}

// creditTypesQuantity is how many credits a purchase of this kind mints
// — 1 for everything except custom, which mints one per purchased event.
func creditTypesQuantity(kind string, requestedQuantity int) int {
	if kind == "custom" {
		return requestedQuantity
	}
	return 1
}

// computeBillingAmount implements the confirmed pricing formula:
// plan.amount + (plan.per_event_amount × eventCount, custom only) +
// (billingNumber-1) × plan.nominal_increment.
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

func toPlanResponse(row dbgen.Plan) entities.PlanResponse {
	return entities.PlanResponse{
		ID: row.ID, Code: row.Code, Kind: row.Kind, BillingCycle: row.BillingCycle,
		Name: row.Name, Amount: int8Ptr(row.Amount), PerEventAmount: int8Ptr(row.PerEventAmount),
		Currency: row.Currency, EventQuota: int4Ptr(row.EventQuota), NominalIncrement: row.NominalIncrement,
	}
}

func toBillingResponse(b dbgen.Billing, plan dbgen.Plan) entities.BillingResponse {
	return entities.BillingResponse{
		ID: b.ID, LineageRootID: b.LineageRootID, PlanCode: plan.Code, Kind: plan.Kind, BillingCycle: plan.BillingCycle,
		BillingNumber: int(b.BillingNumber), PeriodStart: timeutil.FormatDate(b.PeriodStart), PeriodEnd: timeutil.FormatDate(b.PeriodEnd),
		Status: b.Status, Amount: b.Amount, Currency: plan.Currency, PaidAt: timestamptzPtr(b.PaidAt),
	}
}

func toCreditResponse(c dbgen.Credit) entities.CreditResponse {
	resp := entities.CreditResponse{
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
