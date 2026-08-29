package subevents

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

func (a *App) CreateSubEvent(ctx context.Context, orgID, eventID uuid.UUID, name string, date pgtype.Date, entryTime, exitTime pgtype.Time) (dbgen.SubEvent, error) {
	return a.queries.CreateSubEvent(ctx, dbgen.CreateSubEventParams{
		OrganizationID: orgID,
		EventID:        eventID,
		Name:           name,
		Date:           date,
		EntryTime:      entryTime,
		ExitTime:       exitTime,
	})
}

func (a *App) GetSubEvent(ctx context.Context, orgID, eventID, id uuid.UUID) (dbgen.SubEvent, error) {
	return a.queries.GetSubEvent(ctx, dbgen.GetSubEventParams{ID: id, EventID: eventID, OrganizationID: orgID})
}

func (a *App) ListSubEvents(ctx context.Context, orgID, eventID uuid.UUID) ([]dbgen.SubEvent, error) {
	return a.queries.ListSubEvents(ctx, dbgen.ListSubEventsParams{EventID: eventID, OrganizationID: orgID})
}

func (a *App) UpdateSubEvent(ctx context.Context, orgID, eventID, id uuid.UUID, name string, date pgtype.Date, entryTime, exitTime pgtype.Time) (dbgen.SubEvent, error) {
	return a.queries.UpdateSubEvent(ctx, dbgen.UpdateSubEventParams{
		ID: id, EventID: eventID, OrganizationID: orgID, Name: name, Date: date, EntryTime: entryTime, ExitTime: exitTime,
	})
}

func (a *App) DeleteSubEvent(ctx context.Context, orgID, eventID, id uuid.UUID) (bool, error) {
	n, err := a.queries.DeleteSubEvent(ctx, dbgen.DeleteSubEventParams{ID: id, EventID: eventID, OrganizationID: orgID})
	return n > 0, err
}

func (a *App) ValidateSubEventIDs(ctx context.Context, orgID, eventID uuid.UUID, ids []uuid.UUID) ([]uuid.UUID, error) {
	return a.queries.ValidateSubEventIDs(ctx, dbgen.ValidateSubEventIDsParams{EventID: eventID, OrganizationID: orgID, Ids: ids})
}
