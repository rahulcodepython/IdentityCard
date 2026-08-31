package subevents

import (
    "context"
    "time"

    "github.com/google/uuid"

    "identitycard-server/internal/pkg/postgres"
)

func (a *App) CreateSubEvent(ctx context.Context, orgID, eventID uuid.UUID, name string, date time.Time, entryTime, exitTime string) (*SubEventResponse, error) {
    return postgres.QueryJSON[SubEventResponse](ctx, a.pool, CreateSubEventQuery, orgID, eventID, name, date, entryTime, exitTime)
}

func (a *App) GetSubEvent(ctx context.Context, orgID, eventID, id uuid.UUID) (*SubEventResponse, error) {
    return postgres.QueryJSON[SubEventResponse](ctx, a.pool, GetSubEventQuery, id, eventID, orgID)
}

func (a *App) ListSubEvents(ctx context.Context, orgID, eventID uuid.UUID) ([]SubEventResponse, error) {
    return postgres.QueryJSONSlice[SubEventResponse](ctx, a.pool, ListSubEventsQuery, eventID, orgID)
}

func (a *App) UpdateSubEvent(ctx context.Context, orgID, eventID, id uuid.UUID, name string, date time.Time, entryTime, exitTime string) (*SubEventResponse, error) {
    return postgres.QueryJSON[SubEventResponse](ctx, a.pool, UpdateSubEventQuery, id, eventID, orgID, name, date, entryTime, exitTime)
}

func (a *App) DeleteSubEvent(ctx context.Context, orgID, eventID, id uuid.UUID) (bool, error) {
    res, err := a.pool.Exec(ctx, DeleteSubEventQuery, id, eventID, orgID)
    if err != nil {
        return false, postgres.MapPgError(err)
    }
    return res.RowsAffected() > 0, nil
}

func (a *App) ValidateSubEventIDs(ctx context.Context, orgID, eventID uuid.UUID, ids []uuid.UUID) ([]uuid.UUID, error) {
    return postgres.QueryJSONSlice[uuid.UUID](ctx, a.pool, ValidateSubEventIDsQuery, eventID, orgID, ids)
}
