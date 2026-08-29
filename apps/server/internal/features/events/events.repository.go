package events

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

func (a *App) CreateEvent(ctx context.Context, tx pgx.Tx, orgID uuid.UUID, eventType string, startDate, endDate pgtype.Date) (dbgen.Event, error) {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.CreateEvent(ctx, dbgen.CreateEventParams{
		OrganizationID: orgID,
		EventType:      eventType,
		StartDate:      startDate,
		EndDate:        endDate,
	})
}

func (a *App) GetEvent(ctx context.Context, orgID, id uuid.UUID) (dbgen.Event, error) {
	return a.queries.GetEvent(ctx, dbgen.GetEventParams{ID: id, OrganizationID: orgID})
}

func (a *App) ListEvents(ctx context.Context, orgID uuid.UUID) ([]dbgen.Event, error) {
	return a.queries.ListEvents(ctx, orgID)
}

func (a *App) UpdateDates(ctx context.Context, tx pgx.Tx, orgID, id uuid.UUID, startDate, endDate pgtype.Date) (dbgen.Event, error) {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.UpdateEventDates(ctx, dbgen.UpdateEventDatesParams{
		ID: id, OrganizationID: orgID, StartDate: startDate, EndDate: endDate,
	})
}

func (a *App) PublishEvent(ctx context.Context, tx pgx.Tx, orgID, id uuid.UUID) (dbgen.Event, error) {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.PublishEvent(ctx, dbgen.PublishEventParams{ID: id, OrganizationID: orgID})
}

func (a *App) DeleteDraft(ctx context.Context, orgID, id uuid.UUID) (bool, error) {
	n, err := a.queries.DeleteDraftEvent(ctx, dbgen.DeleteDraftEventParams{ID: id, OrganizationID: orgID})
	return n > 0, err
}

func (a *App) Delete(ctx context.Context, id uuid.UUID) error {
	return a.queries.DeleteEvent(ctx, id)
}

func (a *App) ListFlashOlderThanRepo(ctx context.Context, cutoff pgtype.Date) ([]dbgen.Event, error) {
	return a.queries.ListFlashEventsOlderThan(ctx, cutoff)
}

func (a *App) MarkUnjoinedPeopleJoinedAt(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, joinedAt pgtype.Timestamptz) error {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.MarkUnjoinedPeopleJoinedAt(ctx, dbgen.MarkUnjoinedPeopleJoinedAtParams{
		EventID:  eventID,
		JoinedAt: joinedAt,
	})
}

func (a *App) CreateDay(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, date pgtype.Date, entryTime, exitTime pgtype.Time) (dbgen.EventDay, error) {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.CreateEventDay(ctx, dbgen.CreateEventDayParams{
		EventID:   eventID,
		Date:      date,
		EntryTime: entryTime,
		ExitTime:  exitTime,
	})
}

func (a *App) ListDays(ctx context.Context, eventID uuid.UUID) ([]dbgen.EventDay, error) {
	return a.queries.ListEventDays(ctx, eventID)
}

func (a *App) DeleteDaysForEvent(ctx context.Context, tx pgx.Tx, eventID uuid.UUID) error {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.DeleteEventDaysForEvent(ctx, eventID)
}

func (a *App) CreateMetadata(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, name string, venue, organizerName pgtype.Text) (dbgen.EventMetadatum, error) {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.CreateEventMetadata(ctx, dbgen.CreateEventMetadataParams{
		EventID: eventID, Name: name, Venue: venue, OrganizerName: organizerName,
	})
}

func (a *App) GetMetadata(ctx context.Context, eventID uuid.UUID) (dbgen.EventMetadatum, error) {
	return a.queries.GetEventMetadata(ctx, eventID)
}

func (a *App) UpdateMetadata(ctx context.Context, tx pgx.Tx, eventID uuid.UUID, name string, venue, organizerName pgtype.Text) (dbgen.EventMetadatum, error) {
	q := a.queries
	if tx != nil {
		q = q.WithTx(tx)
	}
	return q.UpdateEventMetadata(ctx, dbgen.UpdateEventMetadataParams{
		EventID: eventID, Name: name, Venue: venue, OrganizerName: organizerName,
	})
}

func (a *App) UpdateImage(ctx context.Context, eventID uuid.UUID, objectKey string) (dbgen.EventMetadatum, error) {
	return a.queries.UpdateEventImage(ctx, dbgen.UpdateEventImageParams{
		EventID: eventID, ImageObjectKey: pgtype.Text{String: objectKey, Valid: true},
	})
}

func (a *App) UpdateOrganizerSignature(ctx context.Context, eventID uuid.UUID, objectKey string) (dbgen.EventMetadatum, error) {
	return a.queries.UpdateEventOrganizerSignature(ctx, dbgen.UpdateEventOrganizerSignatureParams{
		EventID: eventID, OrganizerSignatureObjectKey: pgtype.Text{String: objectKey, Valid: true},
	})
}
