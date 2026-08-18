package events

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"identitycard-server/internal/db"
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/httpx"
	"identitycard-server/internal/timeutil"
)

type Service struct {
	repo *Repository
	pool *pgxpool.Pool
}

func NewService(repo *Repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, pool: pool}
}

func (s *Service) Create(ctx context.Context, orgID uuid.UUID, req CreateEventRequest) (EventResponse, error) {
	days, err := parseDays(req.Days)
	if err != nil {
		return EventResponse{}, err
	}
	if req.Kind == "flash" && len(days) != 1 {
		return EventResponse{}, httpx.ErrValidation(map[string]string{"days": "a flash event must have exactly one day"})
	}
	minDate, maxDate := minMaxDates(days)

	var event dbgen.Event
	var createdDays []dbgen.EventDay
	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)

		created, err := repo.Create(ctx, orgID, req.Name, req.Kind, minDate, maxDate, toPgText(req.Venue))
		if err != nil {
			return httpx.ErrInternal()
		}
		event = created

		for _, d := range days {
			day, err := repo.CreateDay(ctx, event.ID, d.date, d.entryTime, d.exitTime)
			if err != nil {
				return httpx.ErrInternal()
			}
			createdDays = append(createdDays, day)
		}
		return nil
	})
	if txErr != nil {
		return EventResponse{}, txErr
	}

	return toEventResponse(event, createdDays), nil
}

func (s *Service) Get(ctx context.Context, orgID, id uuid.UUID) (EventResponse, error) {
	event, err := s.repo.Get(ctx, orgID, id)
	if err != nil {
		return EventResponse{}, httpx.ErrNotFound("event")
	}
	eventDays, err := s.repo.ListDays(ctx, id)
	if err != nil {
		return EventResponse{}, httpx.ErrInternal()
	}
	return toEventResponse(event, eventDays), nil
}

func (s *Service) List(ctx context.Context, orgID uuid.UUID) ([]EventSummary, error) {
	rows, err := s.repo.List(ctx, orgID)
	if err != nil {
		return nil, httpx.ErrInternal()
	}
	summaries := make([]EventSummary, len(rows))
	for i, row := range rows {
		summaries[i] = toEventSummary(row)
	}
	return summaries, nil
}

// Update replaces the event's name and its entire day list. Only allowed
// while the event is still a draft — publishing locks the schedule.
func (s *Service) Update(ctx context.Context, orgID, id uuid.UUID, req UpdateEventRequest) (EventResponse, error) {
	days, err := parseDays(req.Days)
	if err != nil {
		return EventResponse{}, err
	}

	existing, err := s.repo.Get(ctx, orgID, id)
	if err != nil {
		return EventResponse{}, httpx.ErrNotFound("event")
	}
	if existing.Status != "draft" {
		return EventResponse{}, httpx.ErrConflict("only a draft event can be edited")
	}
	if existing.Kind == "flash" && len(days) != 1 {
		return EventResponse{}, httpx.ErrValidation(map[string]string{"days": "a flash event must have exactly one day"})
	}
	minDate, maxDate := minMaxDates(days)

	var event dbgen.Event
	var updatedDays []dbgen.EventDay
	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)

		updated, err := repo.UpdateNameAndDates(ctx, orgID, id, req.Name, minDate, maxDate, toPgText(req.Venue))
		if err != nil {
			// Not found, wrong org, or raced with a publish between the
			// pre-check above and here — all collapse to the same 409.
			return httpx.ErrConflict("only a draft event can be edited")
		}
		event = updated

		if err := repo.DeleteDaysForEvent(ctx, id); err != nil {
			return httpx.ErrInternal()
		}
		for _, d := range days {
			day, err := repo.CreateDay(ctx, id, d.date, d.entryTime, d.exitTime)
			if err != nil {
				return httpx.ErrInternal()
			}
			updatedDays = append(updatedDays, day)
		}
		return nil
	})
	if txErr != nil {
		return EventResponse{}, txErr
	}

	return toEventResponse(event, updatedDays), nil
}

// Publish locks the event's schedule and backfills joined_at (see
// people.joined_at, and 000016_people_joined_at_nullable) for everyone
// added while it was still a draft. Actually emailing ID cards is a later
// phase — this only flips the state and settles join dates.
func (s *Service) Publish(ctx context.Context, orgID, id uuid.UUID) (EventResponse, error) {
	var event dbgen.Event
	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)

		published, err := repo.Publish(ctx, orgID, id)
		if err != nil {
			return httpx.NewError(http.StatusConflict, "not_publishable", "event not found or already published")
		}
		event = published

		return repo.MarkUnjoinedPeopleJoinedAt(ctx, id, event.PublishedAt)
	})
	if txErr != nil {
		return EventResponse{}, txErr
	}

	eventDays, err := s.repo.ListDays(ctx, id)
	if err != nil {
		return EventResponse{}, httpx.ErrInternal()
	}
	return toEventResponse(event, eventDays), nil
}

func (s *Service) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	deleted, err := s.repo.DeleteDraft(ctx, orgID, id)
	if err != nil {
		return httpx.ErrInternal()
	}
	if !deleted {
		return httpx.NewError(http.StatusConflict, "not_deletable", "event not found or no longer a draft")
	}
	return nil
}

// EventContext is the minimal parent-event read shared by every module
// nested under an event (subevents, people, forms): is it still editable,
// and which dates can a sub-event/day-scoped child legally use.
type EventContext struct {
	Status string
	Dates  map[string]bool
}

func (c EventContext) IsDraft() bool { return c.Status == "draft" }

func (s *Service) GetContext(ctx context.Context, orgID, eventID uuid.UUID) (EventContext, error) {
	event, err := s.repo.Get(ctx, orgID, eventID)
	if err != nil {
		return EventContext{}, httpx.ErrNotFound("event")
	}
	eventDays, err := s.repo.ListDays(ctx, eventID)
	if err != nil {
		return EventContext{}, httpx.ErrInternal()
	}

	dates := make(map[string]bool, len(eventDays))
	for _, d := range eventDays {
		dates[timeutil.FormatDate(d.Date)] = true
	}
	return EventContext{Status: event.Status, Dates: dates}, nil
}

type parsedDay struct {
	date      pgtype.Date
	entryTime pgtype.Time
	exitTime  pgtype.Time
}

func parseDays(inputs []EventDayInput) ([]parsedDay, error) {
	seen := make(map[string]bool, len(inputs))
	days := make([]parsedDay, len(inputs))

	for i, in := range inputs {
		if seen[in.Date] {
			return nil, httpx.ErrValidation(map[string]string{"days": "duplicate date: " + in.Date})
		}
		seen[in.Date] = true

		date, err := timeutil.ParseDate(in.Date)
		if err != nil {
			return nil, httpx.ErrValidation(map[string]string{"days": "invalid date: " + in.Date})
		}
		entry, err := timeutil.ParseClock(in.EntryTime)
		if err != nil {
			return nil, httpx.ErrValidation(map[string]string{"days": "invalid entry_time: " + in.EntryTime})
		}
		exit, err := timeutil.ParseClock(in.ExitTime)
		if err != nil {
			return nil, httpx.ErrValidation(map[string]string{"days": "invalid exit_time: " + in.ExitTime})
		}
		if exit.Microseconds <= entry.Microseconds {
			return nil, httpx.ErrValidation(map[string]string{"days": "exit_time must be after entry_time on " + in.Date})
		}

		days[i] = parsedDay{date: date, entryTime: entry, exitTime: exit}
	}
	return days, nil
}

func minMaxDates(days []parsedDay) (pgtype.Date, pgtype.Date) {
	min, max := days[0].date, days[0].date
	for _, d := range days[1:] {
		if d.date.Time.Before(min.Time) {
			min = d.date
		}
		if d.date.Time.After(max.Time) {
			max = d.date
		}
	}
	return min, max
}

func toEventSummary(event dbgen.Event) EventSummary {
	var publishedAt *string
	if event.PublishedAt.Valid {
		s := event.PublishedAt.Time.Format(time.RFC3339)
		publishedAt = &s
	}
	var venue *string
	if event.Venue.Valid {
		venue = &event.Venue.String
	}
	return EventSummary{
		ID:          event.ID,
		Name:        event.Name,
		Kind:        event.Kind,
		Status:      event.Status,
		StartDate:   timeutil.FormatDate(event.StartDate),
		EndDate:     timeutil.FormatDate(event.EndDate),
		Venue:       venue,
		PublishedAt: publishedAt,
	}
}

func toPgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func toEventResponse(event dbgen.Event, eventDays []dbgen.EventDay) EventResponse {
	dayResponses := make([]EventDayResponse, len(eventDays))
	for i, d := range eventDays {
		dayResponses[i] = EventDayResponse{
			Date:      timeutil.FormatDate(d.Date),
			EntryTime: timeutil.FormatClock(d.EntryTime),
			ExitTime:  timeutil.FormatClock(d.ExitTime),
		}
	}
	return EventResponse{EventSummary: toEventSummary(event), Days: dayResponses}
}
