package organizations

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

type Repository struct {
	q *dbgen.Queries
}

func NewRepository(q *dbgen.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) Create(ctx context.Context, name, slug string) (dbgen.Organization, error) {
	return r.q.CreateOrganization(ctx, dbgen.CreateOrganizationParams{Name: name, Slug: slug})
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (dbgen.Organization, error) {
	return r.q.GetOrganizationByID(ctx, id)
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (dbgen.Organization, error) {
	return r.q.GetOrganizationBySlug(ctx, slug)
}

func (r *Repository) UpdateLogo(ctx context.Context, id uuid.UUID, objectKey string) (dbgen.Organization, error) {
	return r.q.UpdateOrganizationLogo(ctx, dbgen.UpdateOrganizationLogoParams{
		ID:            id,
		LogoObjectKey: pgtype.Text{String: objectKey, Valid: true},
	})
}
