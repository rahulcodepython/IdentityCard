// Package repositories: credits are single-use event-creation rights —
// a purchase mints one or more, each optionally consumed by exactly one
// event (credits.event_id). See PlansService for the consumption/
// replenishment rules.
package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

type CreditsRepository struct {
	q *dbgen.Queries
}

func NewCreditsRepository(q *dbgen.Queries) *CreditsRepository {
	return &CreditsRepository{q: q}
}

func (r *CreditsRepository) WithTx(tx pgx.Tx) *CreditsRepository {
	return &CreditsRepository{q: r.q.WithTx(tx)}
}

func (r *CreditsRepository) Create(ctx context.Context, orgID, billingID uuid.UUID, creditType string) (dbgen.Credit, error) {
	return r.q.CreateCredit(ctx, dbgen.CreateCreditParams{OrganizationID: orgID, BillingID: billingID, Type: creditType})
}

// FindAvailable locks (FOR UPDATE) and returns the oldest unconsumed,
// non-restricted credit of one of the given types — see the query's own
// doc comment for the flash same-day rule. Returns pgx.ErrNoRows if
// none.
func (r *CreditsRepository) FindAvailable(ctx context.Context, orgID uuid.UUID, types []string) (dbgen.Credit, error) {
	return r.q.FindAvailableCredit(ctx, dbgen.FindAvailableCreditParams{OrganizationID: orgID, Types: types})
}

func (r *CreditsRepository) Consume(ctx context.Context, creditID, eventID uuid.UUID) (dbgen.Credit, error) {
	return r.q.ConsumeCredit(ctx, dbgen.ConsumeCreditParams{ID: creditID, EventID: pgtype.UUID{Bytes: eventID, Valid: true}})
}

func (r *CreditsRepository) GetForEvent(ctx context.Context, eventID uuid.UUID) (dbgen.Credit, error) {
	return r.q.GetCreditForEvent(ctx, pgtype.UUID{Bytes: eventID, Valid: true})
}

// GetActiveUnlimitedLineageRoot returns the org's active/paid
// unlimited-plan billing lineage id, or ok=false if it has none — used
// to decide whether to replenish an 'all' credit after one is consumed.
func (r *CreditsRepository) GetActiveUnlimitedLineageRoot(ctx context.Context, orgID uuid.UUID) (id uuid.UUID, ok bool, err error) {
	lineageID, err := r.q.GetActiveUnlimitedLineageRoot(ctx, orgID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, err
	}
	return lineageID, true, nil
}

// ListForOrganization returns orgID's whole credit ledger, oldest first
// — used to build the available/consumed/restricted summary in
// PlansService.ListBilling.
func (r *CreditsRepository) ListForOrganization(ctx context.Context, orgID uuid.UUID) ([]dbgen.Credit, error) {
	return r.q.ListCreditsForOrganization(ctx, orgID)
}

// -- daily cron sweeps (used by internal/jobs; plain errors, not utils) --

func (r *CreditsRepository) SyncRestriction(ctx context.Context) ([]dbgen.Credit, error) {
	return r.q.SyncCreditRestriction(ctx)
}

func (r *CreditsRepository) ListRestrictedOlderThan(ctx context.Context, cutoff pgtype.Timestamptz) ([]dbgen.Credit, error) {
	return r.q.ListRestrictedCreditsOlderThan(ctx, cutoff)
}
