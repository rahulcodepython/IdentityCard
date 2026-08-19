package plans

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"identitycard-server/internal/db"
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/httpx"
	"identitycard-server/internal/timeutil"
)

type Service struct {
	repo *Repository

	// Used only by Subscribe/Renew, which write a subscription and its
	// first/next subscription_periods row atomically — List/ListForOrganization
	// work through repo alone.
	pool        *pgxpool.Pool
	baseQueries *dbgen.Queries
}

func NewService(repo *Repository, pool *pgxpool.Pool, baseQueries *dbgen.Queries) *Service {
	return &Service{repo: repo, pool: pool, baseQueries: baseQueries}
}

func (s *Service) List(ctx context.Context) ([]PlanResponse, error) {
	rows, err := s.repo.List(ctx)
	if err != nil {
		return nil, httpx.ErrInternal()
	}
	resp := make([]PlanResponse, len(rows))
	for i, row := range rows {
		resp[i] = toPlanResponse(row)
	}
	return resp, nil
}

// ListForOrganization returns every subscription orgID has ever
// purchased, plus a computed capacity summary. An active 'unlimited'
// subscription makes the whole org unlimited regardless of anything
// else; otherwise total capacity is the sum of event_quota across
// 'active' subscriptions, and "used" is how many events are actually
// attributed to those subscriptions (see events.subscription_id).
func (s *Service) ListForOrganization(ctx context.Context, orgID uuid.UUID) (OrgSubscriptionsResponse, error) {
	subs, err := s.repo.ListForOrganization(ctx, orgID)
	if err != nil {
		return OrgSubscriptionsResponse{}, httpx.ErrInternal()
	}

	resp := OrgSubscriptionsResponse{Subscriptions: make([]SubscriptionResponse, len(subs))}
	for i, sub := range subs {
		plan, err := s.repo.GetPlanByID(ctx, sub.PlanID)
		if err != nil {
			return OrgSubscriptionsResponse{}, httpx.ErrInternal()
		}
		resp.Subscriptions[i] = toSubscriptionResponse(sub, plan)

		if sub.Status != "active" {
			continue
		}
		if sub.Kind == "unlimited" {
			resp.Unlimited = true
			continue
		}
		if sub.EventQuota.Valid {
			q := int(sub.EventQuota.Int32)
			if resp.TotalQuota == nil {
				resp.TotalQuota = &q
			} else {
				*resp.TotalQuota += q
			}
		}
	}
	return resp, nil
}

// SubscriptionCandidate is the pure-subscriptions-table view of one
// active subscription — deliberately carries no events-table data (plans
// doesn't know about events at all, to avoid an import cycle: events
// already imports plans). The caller (events.Service.checkEventLimit)
// does its own per-subscription event counting.
type SubscriptionCandidate struct {
	SubscriptionID uuid.UUID
	Kind           string
	EventQuota     *int // nil = unlimited
}

// ActiveSubscriptionsForOrgOrderedByAge returns orgID's 'active' (not
// past_due/expired) subscriptions oldest-first — see
// events.Service.checkEventLimit, which spends an org's earliest-purchased
// capacity before newer top-ups, so a subscription that later lapses only
// ever "owns" the events it actually funded.
func (s *Service) ActiveSubscriptionsForOrgOrderedByAge(ctx context.Context, orgID uuid.UUID) ([]SubscriptionCandidate, error) {
	subs, err := s.repo.ListActiveForOrganization(ctx, orgID)
	if err != nil {
		return nil, httpx.ErrInternal()
	}
	out := make([]SubscriptionCandidate, len(subs))
	for i, sub := range subs {
		c := SubscriptionCandidate{SubscriptionID: sub.ID, Kind: sub.Kind}
		if sub.EventQuota.Valid {
			q := int(sub.EventQuota.Int32)
			c.EventQuota = &q
		}
		out[i] = c
	}
	return out, nil
}

// Subscribe purchases a new subscription for orgID. This is always
// additive — buying more Custom capacity on top of an existing
// Base/Custom/Unlimited subscription creates a new row rather than
// canceling anything (confirmed: top-ups stack, each with its own
// independent billing cycle and grace/deletion timeline; see
// ActiveSubscriptionsForOrgOrderedByAge for how capacity is pooled
// across them). Still stubbed — no real payment call, same as before.
func (s *Service) Subscribe(ctx context.Context, orgID uuid.UUID, req SubscribeRequest) (SubscriptionResponse, error) {
	plan, err := s.repo.GetPlanByCode(ctx, req.PlanCode)
	if err != nil {
		return SubscriptionResponse{}, httpx.NewError(http.StatusBadRequest, "invalid_plan", "unknown plan")
	}

	quota := plan.EventQuota
	if plan.Kind == "custom" {
		if req.EventQuantity == nil {
			return SubscriptionResponse{}, httpx.ErrValidation(map[string]string{"event_quantity": "required for a custom plan"})
		}
		quota = pgtype.Int4{Int32: int32(*req.EventQuantity), Valid: true}
	}

	now := time.Now().UTC()
	var periodStart, periodEnd pgtype.Date
	if plan.BillingCycle != "one_time" {
		periodStart = pgtype.Date{Time: dateOnly(now), Valid: true}
		periodEnd = pgtype.Date{Time: addCycle(dateOnly(now), plan.BillingCycle), Valid: true}
	}

	var sub dbgen.Subscription
	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := NewRepository(s.baseQueries.WithTx(tx))
		created, err := repo.CreateSubscription(ctx, orgID, plan.ID, plan.Kind, plan.BillingCycle, quota, periodEnd)
		if err != nil {
			return httpx.ErrInternal()
		}
		sub = created

		if plan.BillingCycle != "one_time" {
			if err := repo.CreateSubscriptionPeriod(ctx, sub.ID, periodStart, periodEnd); err != nil {
				return httpx.ErrInternal()
			}
		}
		return nil
	})
	if txErr != nil {
		return SubscriptionResponse{}, txErr
	}

	return toSubscriptionResponse(sub, plan), nil
}

// Renew extends subscriptionID by one billing cycle and clears any
// past_due state — the confirmed grace-period rule ("still fully
// functional until the deadline") means a renewal any time before
// grace_deadline simply restores normal service with no penalty.
func (s *Service) Renew(ctx context.Context, orgID, subscriptionID uuid.UUID) (SubscriptionResponse, error) {
	sub, err := s.repo.GetForOrganization(ctx, orgID, subscriptionID)
	if err != nil {
		return SubscriptionResponse{}, httpx.ErrNotFound("subscription")
	}
	if sub.BillingCycle == "one_time" {
		return SubscriptionResponse{}, httpx.NewError(http.StatusBadRequest, "not_renewable", "this plan does not renew")
	}
	if sub.Status == "expired" {
		return SubscriptionResponse{}, httpx.NewError(http.StatusConflict, "expired", "this subscription has already expired — purchase a new one")
	}

	now := dateOnly(time.Now().UTC())
	newPeriodEnd := pgtype.Date{Time: addCycle(now, sub.BillingCycle), Valid: true}

	var updated dbgen.Subscription
	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := NewRepository(s.baseQueries.WithTx(tx))
		u, err := repo.Renew(ctx, sub.ID, newPeriodEnd)
		if err != nil {
			return httpx.ErrInternal()
		}
		updated = u
		return repo.CreateSubscriptionPeriod(ctx, sub.ID, pgtype.Date{Time: now, Valid: true}, newPeriodEnd)
	})
	if txErr != nil {
		return SubscriptionResponse{}, txErr
	}

	plan, err := s.repo.GetPlanByID(ctx, updated.PlanID)
	if err != nil {
		return SubscriptionResponse{}, httpx.ErrInternal()
	}
	return toSubscriptionResponse(updated, plan), nil
}

// -- billing sweeps (used by internal/jobs; plain errors, not httpx —
// these are never handler-facing) --

// SweepPastDue flips 'active' monthly/yearly subscriptions whose
// current_period_end has passed into 'past_due', computing each one's
// grace_deadline from its billing cycle (1 month for monthly, 6 months
// for yearly). Flash subscriptions are never candidates — see
// ListPastDueCandidateSubscriptions's filter.
func (s *Service) SweepPastDue(ctx context.Context, today time.Time) error {
	todayDate := pgtype.Date{Time: dateOnly(today), Valid: true}
	subs, err := s.repo.ListPastDueCandidates(ctx, todayDate)
	if err != nil {
		return err
	}
	for _, sub := range subs {
		grace := dateOnly(today).AddDate(0, 1, 0)
		if sub.BillingCycle == "yearly" {
			grace = dateOnly(today).AddDate(0, 6, 0)
		}
		if err := s.repo.MarkPastDue(ctx, sub.ID, todayDate, pgtype.Date{Time: grace, Valid: true}); err != nil {
			return err
		}
	}
	return nil
}

// ListGraceExpired returns 'past_due' subscriptions whose grace_deadline
// has passed — internal/jobs deletes their events and calls MarkExpired.
func (s *Service) ListGraceExpired(ctx context.Context, today time.Time) ([]dbgen.Subscription, error) {
	return s.repo.ListGraceExpired(ctx, pgtype.Date{Time: dateOnly(today), Valid: true})
}

func (s *Service) MarkExpired(ctx context.Context, id uuid.UUID) error {
	return s.repo.MarkExpired(ctx, id)
}

// ListActiveFlashSubscriptions returns Flash subscriptions still 'active'
// — internal/jobs cross-references each against its funded event's
// end_date (via events.Service) to decide whether the 1-month retention
// window has passed, since that date lives on the events table, not here.
func (s *Service) ListActiveFlashSubscriptions(ctx context.Context) ([]dbgen.Subscription, error) {
	return s.repo.ListActiveFlash(ctx)
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func addCycle(t time.Time, billingCycle string) time.Time {
	if billingCycle == "yearly" {
		return t.AddDate(1, 0, 0)
	}
	return t.AddDate(0, 1, 0)
}

func toPlanResponse(row dbgen.Plan) PlanResponse {
	return PlanResponse{
		ID: row.ID, Code: row.Code, Kind: row.Kind, BillingCycle: row.BillingCycle,
		Name: row.Name, Amount: int8Ptr(row.Amount), PerEventAmount: int8Ptr(row.PerEventAmount),
		Currency: row.Currency, EventQuota: int4Ptr(row.EventQuota),
	}
}

func toSubscriptionResponse(sub dbgen.Subscription, plan dbgen.Plan) SubscriptionResponse {
	return SubscriptionResponse{
		ID: sub.ID, PlanCode: plan.Code, Kind: sub.Kind, BillingCycle: sub.BillingCycle,
		Status: sub.Status, EventQuota: int4Ptr(sub.EventQuota),
		StartedAt:        sub.StartedAt.Time.Format(time.RFC3339),
		CurrentPeriodEnd: datePtr(sub.CurrentPeriodEnd),
		GraceDeadline:    datePtr(sub.GraceDeadline),
	}
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

func datePtr(v pgtype.Date) *string {
	if !v.Valid {
		return nil
	}
	s := timeutil.FormatDate(v)
	return &s
}
