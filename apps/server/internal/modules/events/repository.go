// Package events owns events and their per-day schedule (event_days).
// Sub-events (internal/modules/subevents) are a separate module that reads
// this one's Service to validate their own days against the parent's.
package events

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

type Repository struct {
	q *dbgen.Queries
}

func NewRepository(q *dbgen.Queries) *Repository {
	return &Repository{q: q}
}

// WithTx returns a Repository whose queries run inside tx — used by
// Service methods that write the event and its event_days atomically.
func (r *Repository) WithTx(tx pgx.Tx) *Repository {
	return &Repository{q: r.q.WithTx(tx)}
}

func (r *Repository) Create(ctx context.Context, orgID uuid.UUID, name, kind string, startDate, endDate pgtype.Date, venue pgtype.Text) (dbgen.Event, error) {
	return r.q.CreateEvent(ctx, dbgen.CreateEventParams{
		OrganizationID: orgID,
		Name:           name,
		Kind:           kind,
		StartDate:      startDate,
		EndDate:        endDate,
		Venue:          venue,
	})
}

func (r *Repository) Get(ctx context.Context, orgID, id uuid.UUID) (dbgen.Event, error) {
	return r.q.GetEvent(ctx, dbgen.GetEventParams{ID: id, OrganizationID: orgID})
}

func (r *Repository) List(ctx context.Context, orgID uuid.UUID) ([]dbgen.Event, error) {
	return r.q.ListEvents(ctx, orgID)
}

func (r *Repository) UpdateNameAndDates(ctx context.Context, orgID, id uuid.UUID, name string, startDate, endDate pgtype.Date, venue pgtype.Text) (dbgen.Event, error) {
	return r.q.UpdateEventNameAndDates(ctx, dbgen.UpdateEventNameAndDatesParams{
		ID:             id,
		OrganizationID: orgID,
		Name:           name,
		StartDate:      startDate,
		EndDate:        endDate,
		Venue:          venue,
	})
}

func (r *Repository) Publish(ctx context.Context, orgID, id uuid.UUID) (dbgen.Event, error) {
	return r.q.PublishEvent(ctx, dbgen.PublishEventParams{ID: id, OrganizationID: orgID})
}

// DeleteDraft reports whether a row was actually deleted — false means
// either the event doesn't exist in this org, or it's no longer a draft.
func (r *Repository) DeleteDraft(ctx context.Context, orgID, id uuid.UUID) (bool, error) {
	n, err := r.q.DeleteDraftEvent(ctx, dbgen.DeleteDraftEventParams{ID: id, OrganizationID: orgID})
	return n > 0, err
}

func (r *Repository) MarkUnjoinedPeopleJoinedAt(ctx context.Context, eventID uuid.UUID, joinedAt pgtype.Timestamptz) error {
	return r.q.MarkUnjoinedPeopleJoinedAt(ctx, dbgen.MarkUnjoinedPeopleJoinedAtParams{
		EventID:  eventID,
		JoinedAt: joinedAt,
	})
}

func (r *Repository) CreateDay(ctx context.Context, eventID uuid.UUID, date pgtype.Date, entryTime, exitTime pgtype.Time) (dbgen.EventDay, error) {
	return r.q.CreateEventDay(ctx, dbgen.CreateEventDayParams{
		EventID:   eventID,
		Date:      date,
		EntryTime: entryTime,
		ExitTime:  exitTime,
	})
}

func (r *Repository) ListDays(ctx context.Context, eventID uuid.UUID) ([]dbgen.EventDay, error) {
	return r.q.ListEventDays(ctx, eventID)
}

func (r *Repository) DeleteDaysForEvent(ctx context.Context, eventID uuid.UUID) error {
	return r.q.DeleteEventDaysForEvent(ctx, eventID)
}
