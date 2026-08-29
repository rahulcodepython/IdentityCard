// Package repositories: plans is just the catalog now — purchasing,
// credits, and billing live in their own repositories (credits.
// repository.go, billing.repository.go, transactions.repository.go).
package repositories

import (
	"context"

	"github.com/google/uuid"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

type PlansRepository struct {
	q *dbgen.Queries
}

func NewPlansRepository(q *dbgen.Queries) *PlansRepository {
	return &PlansRepository{q: q}
}

func (r *PlansRepository) List(ctx context.Context) ([]dbgen.Plan, error) {
	return r.q.ListPlans(ctx)
}

func (r *PlansRepository) GetPlanByCode(ctx context.Context, code string) (dbgen.Plan, error) {
	return r.q.GetPlanByCode(ctx, code)
}

func (r *PlansRepository) GetPlanByID(ctx context.Context, id uuid.UUID) (dbgen.Plan, error) {
	return r.q.GetPlanByID(ctx, id)
}
