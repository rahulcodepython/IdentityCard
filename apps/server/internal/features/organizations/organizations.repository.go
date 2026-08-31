package organizations

import (
    "context"

    "github.com/google/uuid"

    "identitycard-server/internal/pkg/postgres"
)

func (a *App) GetByID(ctx context.Context, id uuid.UUID) (*OrganizationDB, error) {
    return postgres.QueryJSON[OrganizationDB](ctx, a.pool, GetOrganizationByIDQuery, id)
}

func (a *App) UpdateName(ctx context.Context, id uuid.UUID, name string) (*OrganizationDB, error) {
    return postgres.QueryJSON[OrganizationDB](ctx, a.pool, UpdateOrganizationNameQuery, id, name)
}

func (a *App) UpdateLogo(ctx context.Context, id uuid.UUID, key string) (*OrganizationDB, error) {
    return postgres.QueryJSON[OrganizationDB](ctx, a.pool, UpdateOrganizationLogoQuery, id, key)
}

func (a *App) DeleteLogoRepo(ctx context.Context, id uuid.UUID) error {
    return postgres.Exec(ctx, a.pool, DeleteOrganizationLogoQuery, id)
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
    return postgres.Exec(ctx, a.pool, DeleteOrganizationQuery, id)
}
