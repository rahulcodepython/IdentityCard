// Package repositories: events owns events, their per-day schedule
// (event_days, flash/standard only), and their display metadata
// (event_metadata). Sub-events (subevents.*) are a separate domain that
// reads this one's Service to validate their own date against the
// parent's.
package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

type EventsRepository struct {
	q *dbgen.Queries
}

func NewEventsRepository(q *dbgen.Queries) *EventsRepository {
	return &EventsRepository{q: q}
}

// WithTx returns an EventsRepository whose queries run inside tx — used by
// service methods that write the event, its event_days, and its metadata
// atomically.
func (r *EventsRepository) WithTx(tx pgx.Tx) *EventsRepository {
	return &EventsRepository{q: r.q.WithTx(tx)}
}

func (r *EventsRepository) Create(ctx context.Context, orgID uuid.UUID, eventType string, startDate, endDate pgtype.Date) (dbgen.Event, error) {
	return r.q.CreateEvent(ctx, dbgen.CreateEventParams{
		OrganizationID: orgID,
		EventType:      eventType,
		StartDate:      startDate,
		EndDate:        endDate,
	})
}

func (r *EventsRepository) Get(ctx context.Context, orgID, id uuid.UUID) (dbgen.Event, error) {
	return r.q.GetEvent(ctx, dbgen.GetEventParams{ID: id, OrganizationID: orgID})
}

func (r *EventsRepository) List(ctx context.Context, orgID uuid.UUID) ([]dbgen.Event, error) {
	return r.q.ListEvents(ctx, orgID)
}

func (r *EventsRepository) UpdateDates(ctx context.Context, orgID, id uuid.UUID, startDate, endDate pgtype.Date) (dbgen.Event, error) {
	return r.q.UpdateEventDates(ctx, dbgen.UpdateEventDatesParams{
		ID: id, OrganizationID: orgID, StartDate: startDate, EndDate: endDate,
	})
}

func (r *EventsRepository) Publish(ctx context.Context, orgID, id uuid.UUID) (dbgen.Event, error) {
	return r.q.PublishEvent(ctx, dbgen.PublishEventParams{ID: id, OrganizationID: orgID})
}

// DeleteDraft reports whether a row was actually deleted — false means
// either the event doesn't exist in this org, or it's no longer a draft.
func (r *EventsRepository) DeleteDraft(ctx context.Context, orgID, id uuid.UUID) (bool, error) {
	n, err := r.q.DeleteDraftEvent(ctx, dbgen.DeleteDraftEventParams{ID: id, OrganizationID: orgID})
	return n > 0, err
}

// Delete unconditionally hard-deletes — used by the credits-restriction/
// flash cleanup cron (internal/jobs), not gated on draft status or
// org-scoped, since it's a system sweep, not a user request.
func (r *EventsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteEvent(ctx, id)
}

// ListFlashOlderThan is the cleanup cron's first deletion criterion —
// flash events whose (single) day is older than the cutoff.
func (r *EventsRepository) ListFlashOlderThan(ctx context.Context, cutoff pgtype.Date) ([]dbgen.Event, error) {
	return r.q.ListFlashEventsOlderThan(ctx, cutoff)
}

func (r *EventsRepository) MarkUnjoinedPeopleJoinedAt(ctx context.Context, eventID uuid.UUID, joinedAt pgtype.Timestamptz) error {
	return r.q.MarkUnjoinedPeopleJoinedAt(ctx, dbgen.MarkUnjoinedPeopleJoinedAtParams{
		EventID:  eventID,
		JoinedAt: joinedAt,
	})
}

func (r *EventsRepository) CreateDay(ctx context.Context, eventID uuid.UUID, date pgtype.Date, entryTime, exitTime pgtype.Time) (dbgen.EventDay, error) {
	return r.q.CreateEventDay(ctx, dbgen.CreateEventDayParams{
		EventID:   eventID,
		Date:      date,
		EntryTime: entryTime,
		ExitTime:  exitTime,
	})
}

func (r *EventsRepository) ListDays(ctx context.Context, eventID uuid.UUID) ([]dbgen.EventDay, error) {
	return r.q.ListEventDays(ctx, eventID)
}

func (r *EventsRepository) DeleteDaysForEvent(ctx context.Context, eventID uuid.UUID) error {
	return r.q.DeleteEventDaysForEvent(ctx, eventID)
}

// -- event_metadata (name/venue/organizer/image) --

func (r *EventsRepository) CreateMetadata(ctx context.Context, eventID uuid.UUID, name string, venue, organizerName pgtype.Text) (dbgen.EventMetadatum, error) {
	return r.q.CreateEventMetadata(ctx, dbgen.CreateEventMetadataParams{
		EventID: eventID, Name: name, Venue: venue, OrganizerName: organizerName,
	})
}

func (r *EventsRepository) GetMetadata(ctx context.Context, eventID uuid.UUID) (dbgen.EventMetadatum, error) {
	return r.q.GetEventMetadata(ctx, eventID)
}

func (r *EventsRepository) UpdateMetadata(ctx context.Context, eventID uuid.UUID, name string, venue, organizerName pgtype.Text) (dbgen.EventMetadatum, error) {
	return r.q.UpdateEventMetadata(ctx, dbgen.UpdateEventMetadataParams{
		EventID: eventID, Name: name, Venue: venue, OrganizerName: organizerName,
	})
}

func (r *EventsRepository) UpdateImage(ctx context.Context, eventID uuid.UUID, objectKey string) (dbgen.EventMetadatum, error) {
	return r.q.UpdateEventImage(ctx, dbgen.UpdateEventImageParams{
		EventID: eventID, ImageObjectKey: pgtype.Text{String: objectKey, Valid: true},
	})
}

func (r *EventsRepository) UpdateOrganizerSignature(ctx context.Context, eventID uuid.UUID, objectKey string) (dbgen.EventMetadatum, error) {
	return r.q.UpdateEventOrganizerSignature(ctx, dbgen.UpdateEventOrganizerSignatureParams{
		EventID: eventID, OrganizerSignatureObjectKey: pgtype.Text{String: objectKey, Valid: true},
	})
}
