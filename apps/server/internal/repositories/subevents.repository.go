// Package repositories: subevents lets an organizer split an event's
// people across optional sub-groups, each optionally restricted to a
// subset of the parent event's days (see services.EventsService.GetContext).
// An event with no sub-events treats everyone assigned to it as one
// whole-event group.
package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

type SubEventsRepository struct {
	q *dbgen.Queries
}

func NewSubEventsRepository(q *dbgen.Queries) *SubEventsRepository {
	return &SubEventsRepository{q: q}
}

func (r *SubEventsRepository) WithTx(tx pgx.Tx) *SubEventsRepository {
	return &SubEventsRepository{q: r.q.WithTx(tx)}
}

func (r *SubEventsRepository) Create(ctx context.Context, orgID, eventID uuid.UUID, name, scheduleMode string) (dbgen.SubEvent, error) {
	return r.q.CreateSubEvent(ctx, dbgen.CreateSubEventParams{
		OrganizationID: orgID,
		EventID:        eventID,
		Name:           name,
		ScheduleMode:   scheduleMode,
	})
}

func (r *SubEventsRepository) Get(ctx context.Context, orgID, eventID, id uuid.UUID) (dbgen.SubEvent, error) {
	return r.q.GetSubEvent(ctx, dbgen.GetSubEventParams{ID: id, EventID: eventID, OrganizationID: orgID})
}

func (r *SubEventsRepository) List(ctx context.Context, orgID, eventID uuid.UUID) ([]dbgen.SubEvent, error) {
	return r.q.ListSubEvents(ctx, dbgen.ListSubEventsParams{EventID: eventID, OrganizationID: orgID})
}

func (r *SubEventsRepository) UpdateName(ctx context.Context, orgID, eventID, id uuid.UUID, name string) (dbgen.SubEvent, error) {
	return r.q.UpdateSubEventName(ctx, dbgen.UpdateSubEventNameParams{
		ID: id, EventID: eventID, OrganizationID: orgID, Name: name,
	})
}

// Delete reports whether a row was actually deleted.
func (r *SubEventsRepository) Delete(ctx context.Context, orgID, eventID, id uuid.UUID) (bool, error) {
	n, err := r.q.DeleteSubEvent(ctx, dbgen.DeleteSubEventParams{ID: id, EventID: eventID, OrganizationID: orgID})
	return n > 0, err
}

// ValidateIDs returns the subset of ids that are real sub-events of this
// event/org — the caller diffs the count to reject any bogus id.
func (r *SubEventsRepository) ValidateIDs(ctx context.Context, orgID, eventID uuid.UUID, ids []uuid.UUID) ([]uuid.UUID, error) {
	return r.q.ValidateSubEventIDs(ctx, dbgen.ValidateSubEventIDsParams{EventID: eventID, OrganizationID: orgID, Ids: ids})
}

func (r *SubEventsRepository) CreateDay(ctx context.Context, subEventID uuid.UUID, date pgtype.Date, entryTime, exitTime pgtype.Time) (dbgen.SubEventDay, error) {
	return r.q.CreateSubEventDay(ctx, dbgen.CreateSubEventDayParams{
		SubEventID: subEventID,
		Date:       date,
		EntryTime:  entryTime,
		ExitTime:   exitTime,
	})
}

func (r *SubEventsRepository) ListDays(ctx context.Context, subEventID uuid.UUID) ([]dbgen.SubEventDay, error) {
	return r.q.ListSubEventDays(ctx, subEventID)
}

func (r *SubEventsRepository) DeleteDaysForSubEvent(ctx context.Context, subEventID uuid.UUID) error {
	return r.q.DeleteSubEventDaysForSubEvent(ctx, subEventID)
}
