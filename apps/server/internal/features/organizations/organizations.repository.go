package organizations

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

func (a *App) GetByID(ctx context.Context, id uuid.UUID) (dbgen.Organization, error) {
	return a.queries.GetOrganizationByID(ctx, id)
}

func (a *App) UpdateName(ctx context.Context, id uuid.UUID, name string) (dbgen.Organization, error) {
	return a.queries.UpdateOrganizationName(ctx, dbgen.UpdateOrganizationNameParams{ID: id, Name: name})
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	return a.queries.DeleteOrganization(ctx, id)
}

func (a *App) UpdateLogo(ctx context.Context, id uuid.UUID, key string) (dbgen.Organization, error) {
	return a.queries.UpdateOrganizationLogo(ctx, dbgen.UpdateOrganizationLogoParams{
		ID:            id,
		LogoObjectKey: pgtype.Text{String: key, Valid: true},
	})
}
