package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

// OrganizationsRepository reads/writes better-auth's own "organization"
// table (see internal/db/migrations/000035_better_auth_schema.up.sql) —
// org creation and membership are better-auth's job now; this is just
// the settings slice (name, logo) Go still owns.
type OrganizationsRepository struct {
	q *dbgen.Queries
}

func NewOrganizationsRepository(q *dbgen.Queries) *OrganizationsRepository {
	return &OrganizationsRepository{q: q}
}

func (r *OrganizationsRepository) GetByID(ctx context.Context, id uuid.UUID) (dbgen.Organization, error) {
	return r.q.GetOrganizationByID(ctx, id)
}

func (r *OrganizationsRepository) UpdateName(ctx context.Context, id uuid.UUID, name string) (dbgen.Organization, error) {
	return r.q.UpdateOrganizationName(ctx, dbgen.UpdateOrganizationNameParams{ID: id, Name: name})
}

func (r *OrganizationsRepository) UpdateLogo(ctx context.Context, id uuid.UUID, objectKey string) (dbgen.Organization, error) {
	return r.q.UpdateOrganizationLogo(ctx, dbgen.UpdateOrganizationLogoParams{
		ID:            id,
		LogoObjectKey: pgtype.Text{String: objectKey, Valid: true},
	})
}

func (r *OrganizationsRepository) DeleteLogo(ctx context.Context, id uuid.UUID) (dbgen.Organization, error) {
	return r.q.DeleteOrganizationLogo(ctx, id)
}

func (r *OrganizationsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteOrganization(ctx, id)
}
