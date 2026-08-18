package subevents

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"identitycard-server/internal/db"
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/httpx"
	"identitycard-server/internal/modules/events"
	"identitycard-server/internal/timeutil"
)

type Service struct {
	repo   *Repository
	events *events.Service
	pool   *pgxpool.Pool
}

func NewService(repo *Repository, eventsService *events.Service, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, events: eventsService, pool: pool}
}

func (s *Service) Create(ctx context.Context, orgID, eventID uuid.UUID, req CreateSubEventRequest) (SubEventResponse, error) {
	parent, err := s.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return SubEventResponse{}, err
	}
	if !parent.IsDraft() {
		return SubEventResponse{}, httpx.ErrConflict("sub-events can only be added while the event is a draft")
	}
	days, err := parseSubEventDays(req.Days, parent.Dates)
	if err != nil {
		return SubEventResponse{}, err
	}

	var subEvent dbgen.SubEvent
	var createdDays []dbgen.SubEventDay
	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)

		created, err := repo.Create(ctx, orgID, eventID, req.Name)
		if err != nil {
			return httpx.ErrInternal()
		}
		subEvent = created

		for _, d := range days {
			day, err := repo.CreateDay(ctx, subEvent.ID, d.date, d.entryTime, d.exitTime)
			if err != nil {
				return httpx.ErrInternal()
			}
			createdDays = append(createdDays, day)
		}
		return nil
	})
	if txErr != nil {
		return SubEventResponse{}, txErr
	}

	return toSubEventResponse(subEvent, createdDays), nil
}

func (s *Service) Get(ctx context.Context, orgID, eventID, id uuid.UUID) (SubEventResponse, error) {
	if _, err := s.events.GetContext(ctx, orgID, eventID); err != nil {
		return SubEventResponse{}, err
	}
	subEvent, err := s.repo.Get(ctx, orgID, eventID, id)
	if err != nil {
		return SubEventResponse{}, httpx.ErrNotFound("sub-event")
	}
	days, err := s.repo.ListDays(ctx, id)
	if err != nil {
		return SubEventResponse{}, httpx.ErrInternal()
	}
	return toSubEventResponse(subEvent, days), nil
}

// List is O(n) sub-event-day queries rather than one join — fine at the
// handful-of-sub-events-per-event scale this is built for; worth a join
// query if that assumption stops holding.
func (s *Service) List(ctx context.Context, orgID, eventID uuid.UUID) ([]SubEventResponse, error) {
	if _, err := s.events.GetContext(ctx, orgID, eventID); err != nil {
		return nil, err
	}
	rows, err := s.repo.List(ctx, orgID, eventID)
	if err != nil {
		return nil, httpx.ErrInternal()
	}

	resp := make([]SubEventResponse, len(rows))
	for i, row := range rows {
		days, err := s.repo.ListDays(ctx, row.ID)
		if err != nil {
			return nil, httpx.ErrInternal()
		}
		resp[i] = toSubEventResponse(row, days)
	}
	return resp, nil
}

func (s *Service) Update(ctx context.Context, orgID, eventID, id uuid.UUID, req UpdateSubEventRequest) (SubEventResponse, error) {
	parent, err := s.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return SubEventResponse{}, err
	}
	if !parent.IsDraft() {
		return SubEventResponse{}, httpx.ErrConflict("sub-events can only be edited while the event is a draft")
	}
	days, err := parseSubEventDays(req.Days, parent.Dates)
	if err != nil {
		return SubEventResponse{}, err
	}

	var subEvent dbgen.SubEvent
	var updatedDays []dbgen.SubEventDay
	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)

		updated, err := repo.UpdateName(ctx, orgID, eventID, id, req.Name)
		if err != nil {
			return httpx.ErrNotFound("sub-event")
		}
		subEvent = updated

		if err := repo.DeleteDaysForSubEvent(ctx, id); err != nil {
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
		return SubEventResponse{}, txErr
	}

	return toSubEventResponse(subEvent, updatedDays), nil
}

func (s *Service) Delete(ctx context.Context, orgID, eventID, id uuid.UUID) error {
	parent, err := s.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return err
	}
	if !parent.IsDraft() {
		return httpx.ErrConflict("sub-events can only be removed while the event is a draft")
	}

	deleted, err := s.repo.Delete(ctx, orgID, eventID, id)
	if err != nil {
		return httpx.ErrInternal()
	}
	if !deleted {
		return httpx.ErrNotFound("sub-event")
	}
	return nil
}

// ValidateIDs reports which of the given ids are real sub-events of this
// event/org, so a caller assigning a person to sub-events (people.Service)
// can reject any id that isn't.
func (s *Service) ValidateIDs(ctx context.Context, orgID, eventID uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]bool, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]bool{}, nil
	}
	valid, err := s.repo.ValidateIDs(ctx, orgID, eventID, ids)
	if err != nil {
		return nil, httpx.ErrInternal()
	}
	result := make(map[uuid.UUID]bool, len(valid))
	for _, id := range valid {
		result[id] = true
	}
	return result, nil
}

type parsedDay struct {
	date      pgtype.Date
	entryTime pgtype.Time
	exitTime  pgtype.Time
}

func parseSubEventDays(inputs []events.EventDayInput, allowedDates map[string]bool) ([]parsedDay, error) {
	seen := make(map[string]bool, len(inputs))
	days := make([]parsedDay, 0, len(inputs))

	for _, in := range inputs {
		if !allowedDates[in.Date] {
			return nil, httpx.ErrValidation(map[string]string{"days": "date is not part of the parent event: " + in.Date})
		}
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

		days = append(days, parsedDay{date: date, entryTime: entry, exitTime: exit})
	}
	return days, nil
}

func toSubEventResponse(subEvent dbgen.SubEvent, days []dbgen.SubEventDay) SubEventResponse {
	dayResponses := make([]events.EventDayResponse, len(days))
	for i, d := range days {
		dayResponses[i] = events.EventDayResponse{
			Date:      timeutil.FormatDate(d.Date),
			EntryTime: timeutil.FormatClock(d.EntryTime),
			ExitTime:  timeutil.FormatClock(d.ExitTime),
		}
	}
	return SubEventResponse{ID: subEvent.ID, Name: subEvent.Name, Days: dayResponses}
}
