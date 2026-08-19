package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

type OrganizationsRepository struct {
	q *dbgen.Queries
}

func NewOrganizationsRepository(q *dbgen.Queries) *OrganizationsRepository {
	return &OrganizationsRepository{q: q}
}

func (r *OrganizationsRepository) Create(ctx context.Context, name, slug string) (dbgen.Organization, error) {
	return r.q.CreateOrganization(ctx, dbgen.CreateOrganizationParams{Name: name, Slug: slug})
}

func (r *OrganizationsRepository) GetByID(ctx context.Context, id uuid.UUID) (dbgen.Organization, error) {
	return r.q.GetOrganizationByID(ctx, id)
}

func (r *OrganizationsRepository) GetBySlug(ctx context.Context, slug string) (dbgen.Organization, error) {
	return r.q.GetOrganizationBySlug(ctx, slug)
}

func (r *OrganizationsRepository) UpdateLogo(ctx context.Context, id uuid.UUID, objectKey string) (dbgen.Organization, error) {
	return r.q.UpdateOrganizationLogo(ctx, dbgen.UpdateOrganizationLogoParams{
		ID:            id,
		LogoObjectKey: pgtype.Text{String: objectKey, Valid: true},
	})
}
