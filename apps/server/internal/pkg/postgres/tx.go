package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WithTx runs fn inside a transaction, committing on success and rolling
// back otherwise (including on panic). Used for writes that span more than
// one module's repository — e.g. registration creates a user, an
// organization, a subscription, and a membership atomically.
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op once committed

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
