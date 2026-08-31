package forms

import (
    "context"

    "github.com/google/uuid"

    "identitycard-server/internal/pkg/postgres"
)

func (a *App) CreateForm(ctx context.Context, orgID, eventID uuid.UUID, subEventID *uuid.UUID, token string, capacity *int32) (*FormResponse, error) {
    return postgres.QueryJSON[FormResponse](ctx, a.pool, CreateEventFormQuery, orgID, eventID, subEventID, token, capacity)
}

func (a *App) ListForms(ctx context.Context, orgID, eventID uuid.UUID) ([]FormResponse, error) {
    return postgres.QueryJSONSlice[FormResponse](ctx, a.pool, ListEventFormsQuery, eventID, orgID)
}

func (a *App) GetForm(ctx context.Context, orgID, eventID, id uuid.UUID) (*FormResponse, error) {
    return postgres.QueryJSON[FormResponse](ctx, a.pool, GetEventFormQuery, id, eventID, orgID)
}

func (a *App) UpdateForm(ctx context.Context, orgID, eventID, id uuid.UUID, capacity *int32, isActive bool) (*FormResponse, error) {
    return postgres.QueryJSON[FormResponse](ctx, a.pool, UpdateEventFormQuery, id, eventID, orgID, capacity, isActive)
}

func (a *App) DeleteForm(ctx context.Context, orgID, eventID, id uuid.UUID) (bool, error) {
    res, err := a.pool.Exec(ctx, DeleteEventFormQuery, id, eventID, orgID)
    if err != nil {
        return false, postgres.MapPgError(err)
    }
    return res.RowsAffected() > 0, nil
}

func (a *App) GetPublicByToken(ctx context.Context, token string) (*PublicFormDB, error) {
    return postgres.QueryJSON[PublicFormDB](ctx, a.pool, GetPublicFormByTokenQuery, token)
}

func (a *App) IncrementSubmissions(ctx context.Context, id uuid.UUID) (*FormResponse, error) {
    return postgres.QueryJSON[FormResponse](ctx, a.pool, IncrementEventFormSubmissionsQuery, id)
}

func (a *App) DecrementSubmissions(ctx context.Context, id uuid.UUID) error {
    return postgres.Exec(ctx, a.pool, DecrementEventFormSubmissionsQuery, id)
}
