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

func (r *Repository) Create(ctx context.Context, orgID, subscriptionID uuid.UUID, name, scheduleMode string, startDate, endDate pgtype.Date, venue pgtype.Text) (dbgen.Event, error) {
	return r.q.CreateEvent(ctx, dbgen.CreateEventParams{
		OrganizationID: orgID,
		SubscriptionID: subscriptionID,
		Name:           name,
		ScheduleMode:   scheduleMode,
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

// CountBySubscription is used by Service.checkEventLimit to enforce a
// subscription's event_quota.
func (r *Repository) CountBySubscription(ctx context.Context, subscriptionID uuid.UUID) (int64, error) {
	return r.q.CountEventsBySubscription(ctx, subscriptionID)
}

// DeleteForSubscription hard-deletes every event a subscription funded —
// used by internal/jobs once that subscription's grace period (or, for
// Flash, its post-event retention window) has expired. Cascades through
// event_days, sub_events, people, attendance_records, and event_forms via
// their existing FKs on events.id.
func (r *Repository) DeleteForSubscription(ctx context.Context, subscriptionID uuid.UUID) error {
	return r.q.DeleteEventsForSubscription(ctx, subscriptionID)
}

// EndDateForSubscription returns the end_date of the (single) event a
// Flash subscription funds, or ok=false if no event has been created
// under it yet.
func (r *Repository) EndDateForSubscription(ctx context.Context, subscriptionID uuid.UUID) (date pgtype.Date, ok bool, err error) {
	d, err := r.q.GetEventEndDateForSubscription(ctx, subscriptionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return pgtype.Date{}, false, nil
		}
		return pgtype.Date{}, false, err
	}
	return d, true, nil
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

// DeleteDay removes a single materialized day — used when an exclusion
// date added post-publish (see Service.AddExcludedDate) falls on an
// already-materialized recurring event day.
func (r *Repository) DeleteDay(ctx context.Context, eventID uuid.UUID, date pgtype.Date) error {
	return r.q.DeleteEventDay(ctx, dbgen.DeleteEventDayParams{EventID: eventID, Date: date})
}

// LatestDay returns the furthest-out materialized date for an event, or
// ok=false if it has none yet — Service.ExtendRecurringHorizon resumes
// materializing from the day after this.
func (r *Repository) LatestDay(ctx context.Context, eventID uuid.UUID) (date pgtype.Date, ok bool, err error) {
	d, err := r.q.LatestEventDay(ctx, eventID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return pgtype.Date{}, false, nil
		}
		return pgtype.Date{}, false, err
	}
	return d.Date, true, nil
}

// UpsertRecurrence writes eventID's recurrence rule, replacing it if one
// already exists — the shared write path for both Create and a draft-only
// Update (see Service.Update).
func (r *Repository) UpsertRecurrence(ctx context.Context, eventID uuid.UUID, startsOn, endsOn pgtype.Date) (dbgen.EventRecurrence, error) {
	return r.q.UpsertEventRecurrence(ctx, dbgen.UpsertEventRecurrenceParams{
		EventID: eventID, StartsOn: startsOn, EndsOn: endsOn,
	})
}

func (r *Repository) GetRecurrence(ctx context.Context, eventID uuid.UUID) (dbgen.EventRecurrence, error) {
	return r.q.GetEventRecurrence(ctx, eventID)
}

// UpdateRecurrenceEndsOn is the only field of a recurring event's rule
// editable post-publish (see Service.Publish).
func (r *Repository) UpdateRecurrenceEndsOn(ctx context.Context, eventID uuid.UUID, endsOn pgtype.Date) (dbgen.EventRecurrence, error) {
	return r.q.UpdateEventRecurrenceEndsOn(ctx, dbgen.UpdateEventRecurrenceEndsOnParams{EventID: eventID, EndsOn: endsOn})
}

func (r *Repository) CreateRecurrenceWeekday(ctx context.Context, eventID uuid.UUID, weekday int16, entryTime, exitTime pgtype.Time) (dbgen.EventRecurrenceWeekday, error) {
	return r.q.CreateRecurrenceWeekday(ctx, dbgen.CreateRecurrenceWeekdayParams{
		EventID: eventID, Weekday: weekday, EntryTime: entryTime, ExitTime: exitTime,
	})
}

func (r *Repository) ListRecurrenceWeekdays(ctx context.Context, eventID uuid.UUID) ([]dbgen.EventRecurrenceWeekday, error) {
	return r.q.ListRecurrenceWeekdays(ctx, eventID)
}

func (r *Repository) DeleteRecurrenceWeekdaysForEvent(ctx context.Context, eventID uuid.UUID) error {
	return r.q.DeleteRecurrenceWeekdaysForEvent(ctx, eventID)
}

// CreateExcludedDate is an upsert (see the query's ON CONFLICT) so it's
// safe to call both from bulk creation and from the single-date
// post-publish AddExcludedDate action without a pre-check.
func (r *Repository) CreateExcludedDate(ctx context.Context, eventID uuid.UUID, date pgtype.Date) (dbgen.EventExcludedDate, error) {
	return r.q.CreateExcludedDate(ctx, dbgen.CreateExcludedDateParams{EventID: eventID, Date: date})
}

func (r *Repository) ListExcludedDates(ctx context.Context, eventID uuid.UUID) ([]dbgen.EventExcludedDate, error) {
	return r.q.ListExcludedDates(ctx, eventID)
}

func (r *Repository) DeleteExcludedDatesForEvent(ctx context.Context, eventID uuid.UUID) error {
	return r.q.DeleteExcludedDatesForEvent(ctx, eventID)
}

// ListRecurringForMaterialization returns every published recurring event
// still running as of today — see Service.ExtendRecurringHorizon.
func (r *Repository) ListRecurringForMaterialization(ctx context.Context, today pgtype.Date) ([]dbgen.Event, error) {
	return r.q.ListRecurringEventsForMaterialization(ctx, today)
}
