package auth

import (
	"context"

	"github.com/google/uuid"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

// Repository wraps the user-credential queries. Membership/role lookups
// live in internal/modules/members since they're a separate concern
// (an org's roster) reused outside login too.
type Repository struct {
	q *dbgen.Queries
}

func NewRepository(q *dbgen.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) CreateUser(ctx context.Context, email, passwordHash, name string) (dbgen.User, error) {
	return r.q.CreateUser(ctx, dbgen.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
	})
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (dbgen.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (dbgen.User, error) {
	return r.q.GetUserByID(ctx, id)
}
