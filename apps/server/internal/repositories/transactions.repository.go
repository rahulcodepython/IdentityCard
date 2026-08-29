package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

type TransactionsRepository struct {
	q *dbgen.Queries
}

func NewTransactionsRepository(q *dbgen.Queries) *TransactionsRepository {
	return &TransactionsRepository{q: q}
}

func (r *TransactionsRepository) WithTx(tx pgx.Tx) *TransactionsRepository {
	return &TransactionsRepository{q: r.q.WithTx(tx)}
}

func (r *TransactionsRepository) Create(ctx context.Context, orgID, planID uuid.UUID, amount int64, currency string) (dbgen.Transaction, error) {
	return r.q.CreateTransaction(ctx, dbgen.CreateTransactionParams{
		OrganizationID: orgID, PlanID: planID, Amount: amount, Currency: currency,
	})
}

func (r *TransactionsRepository) ListForOrganization(ctx context.Context, orgID uuid.UUID) ([]dbgen.Transaction, error) {
	return r.q.ListTransactionsForOrganization(ctx, orgID)
}
