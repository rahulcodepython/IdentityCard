package organizations

import (
    "context"
    "uuid"

    "identitycard-server/internal/pkg/postgres"
)

func (a *App) List(ctx context.Context, userID uuid.UUID) ([]ListOrganizationsResponse, error) {
    res, err := postgres.QueryJSON[[]ListOrganizationsResponse](ctx, a.pool, ListOrganizationsQuery, userID)
    if err != nil {
        return nil, err
    }
    if res == nil {
        return []ListOrganizationsResponse{}, nil
    }
    return *res, nil
}

func (a *App) GetByID(ctx context.Context, id uuid.UUID) (*OrganizationDB, error) {
    return postgres.QueryJSON[OrganizationDB](ctx, a.pool, GetOrganizationByIDQuery, id)
}

func (a *App) UpdateName(ctx context.Context, id uuid.UUID, name string) (*OrganizationDB, error) {
    return postgres.QueryJSON[OrganizationDB](ctx, a.pool, UpdateOrganizationNameQuery, id, name)
}

func (a *App) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*DeleteOrganizationResult, error) {
    return postgres.QueryJSON[DeleteOrganizationResult](ctx, a.pool, DeleteOrganizationQuery, id, userID)
}
