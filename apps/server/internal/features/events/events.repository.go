package events

import (
    "context"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"

    "identitycard-server/internal/pkg/postgres"
)

func (a *App) CreateEvent(ctx context.Context, tx pgx.Tx, orgID uuid.UUID, eventType string, startDate, endDate time.Time) (*EventDB, error) {
    if tx != nil {
        return postgres.QueryJSONTx[EventDB](ctx, tx, CreateEventQuery, orgID, eventType, startDate, endDate)
    }
    return postgres.QueryJSON[EventDB](ctx, a.pool, CreateEventQuery, orgID, eventType, startDate, endDate)
}

func (a *App) GetEvent(ctx context.Context, orgID, id uuid.UUID) (*EventResponse, error) {
    return postgres.QueryJSON[EventResponse](ctx, a.pool, GetEventQuery, id, orgID)
}

func (a *App) GetRawEvent(ctx context.Context, orgID, id uuid.UUID) (*EventDB, error) {
    return postgres.QueryJSON[EventDB](ctx, a.pool, GetRawEventQuery, id, orgID)
}

func (a *App) ListEvents(ctx context.Context, orgID uuid.UUID) ([]EventSummary, error) {
    return postgres.QueryJSONSlice[EventSummary](ctx, a.pool, ListEventsQuery, orgID)
}

func (a *App) UpdateDates(ctx context.Context, tx pgx.Tx, orgID, id uuid.UUID, startDate, endDate time.Time) (*EventDB, error) {
    if tx != nil {
        return postgres.QueryJSONTx[EventDB](ctx, tx, UpdateEventDatesQuery, id, orgID, startDate, endDate)
    }
    return postgres.QueryJSON[EventDB](ctx, a.pool, UpdateEventDatesQuery, id, orgID, startDate, endDate)
}

func (a *App) PublishEvent(ctx context.Context, tx pgx.Tx, orgID, id uuid.UUID) (*EventDB, error) {
    if tx != nil {
        return postgres.QueryJSONTx[EventDB](ctx, tx, PublishEventQuery, id, orgID)
    }
    return postgres.QueryJSON[EventDB](ctx, a.pool, PublishEventQuery, id, orgID)
}

func (a *App) DeleteDraft(ctx context.Context, orgID, id uuid.UUID) (bool, error) {
    res, err := a.pool.Exec(ctx, DeleteDraftEventQuery, id, orgID)
    if err != nil {
        return false, postgres.MapPgError(err)
    }
    return res.RowsAffected() > 0, nil
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
    return postgres.Exec(ctx, a.pool, DeleteEventQuery, id)
}

func (a *App) ListFlashOlderThanRepo(ctx context.Context, cutoff time.Time) ([]EventDB, error) {
    return postgres.QueryJSONSlice[EventDB](ctx, a.pool, ListFlashEventsOlderThanQuery, cutoff)
}

func (a *App) MarkUnjoinedPeopleJoinedAt(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, joinedAt time.Time) error {
    if tx != nil {
        return postgres.ExecTx(ctx, tx, MarkUnjoinedPeopleJoinedAtQuery, eventID, joinedAt)
    }
    return postgres.Exec(ctx, a.pool, MarkUnjoinedPeopleJoinedAtQuery, eventID, joinedAt)
}

func (a *App) CreateDay(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, date time.Time, entryTime, exitTime string) (*EventDayDB, error) {
    if tx != nil {
        return postgres.QueryJSONTx[EventDayDB](ctx, tx, CreateEventDayQuery, eventID, date, entryTime, exitTime)
    }
    return postgres.QueryJSON[EventDayDB](ctx, a.pool, CreateEventDayQuery, eventID, date, entryTime, exitTime)
}

func (a *App) ListDays(ctx context.Context, eventID uuid.UUID) ([]EventDayResponse, error) {
    return postgres.QueryJSONSlice[EventDayResponse](ctx, a.pool, ListEventDaysQuery, eventID)
}

func (a *App) DeleteDaysForEvent(ctx context.Context, tx pgx.Tx, eventID uuid.UUID) error {
    if tx != nil {
        return postgres.ExecTx(ctx, tx, DeleteEventDaysForEventQuery, eventID)
    }
    return postgres.Exec(ctx, a.pool, DeleteEventDaysForEventQuery, eventID)
}

func (a *App) CreateMetadata(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, name string, venue, organizerName *string) (*EventMetadataDB, error) {
    if tx != nil {
        return postgres.QueryJSONTx[EventMetadataDB](ctx, tx, CreateEventMetadataQuery, eventID, name, venue, organizerName)
    }
    return postgres.QueryJSON[EventMetadataDB](ctx, a.pool, CreateEventMetadataQuery, eventID, name, venue, organizerName)
}

func (a *App) GetMetadata(ctx context.Context, eventID uuid.UUID) (*EventMetadataDB, error) {
    return postgres.QueryJSON[EventMetadataDB](ctx, a.pool, GetEventMetadataQuery, eventID)
}

func (a *App) UpdateMetadata(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, name string, venue, organizerName *string) (*EventMetadataDB, error) {
    if tx != nil {
        return postgres.QueryJSONTx[EventMetadataDB](ctx, tx, UpdateEventMetadataQuery, eventID, name, venue, organizerName)
    }
    return postgres.QueryJSON[EventMetadataDB](ctx, a.pool, UpdateEventMetadataQuery, eventID, name, venue, organizerName)
}

func (a *App) UpdateImage(ctx context.Context, eventID uuid.UUID, objectKey string) (*EventMetadataDB, error) {
    return postgres.QueryJSON[EventMetadataDB](ctx, a.pool, UpdateEventImageQuery, eventID, objectKey)
}

func (a *App) UpdateOrganizerSignature(ctx context.Context, eventID uuid.UUID, objectKey string) (*EventMetadataDB, error) {
    return postgres.QueryJSON[EventMetadataDB](ctx, a.pool, UpdateEventOrganizerSignatureQuery, eventID, objectKey)
}
