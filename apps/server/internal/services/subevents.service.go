package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/entities"
	"identitycard-server/internal/pkg/postgres"
	"identitycard-server/internal/repositories"
	"identitycard-server/internal/utils"
	"identitycard-server/internal/utils/timeutil"
)

type SubEventsService struct {
	repo   *repositories.SubEventsRepository
	events *EventsService
	pool   *pgxpool.Pool
}

func NewSubEventsService(repo *repositories.SubEventsRepository, eventsService *EventsService, pool *pgxpool.Pool) *SubEventsService {
	return &SubEventsService{repo: repo, events: eventsService, pool: pool}
}

func (s *SubEventsService) Create(ctx context.Context, orgID, eventID uuid.UUID, req entities.CreateSubEventRequest) (entities.SubEventResponse, error) {
	parent, err := s.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return entities.SubEventResponse{}, err
	}
	if !parent.IsDraft() {
		return entities.SubEventResponse{}, utils.ErrConflict("sub-events can only be added while the event is a draft")
	}
	days, err := parseSubEventScheduleMode(req.ScheduleMode, req.Days, req.RangeStart, req.RangeEnd, req.RangeEntryTime, req.RangeExitTime, parent.Dates)
	if err != nil {
		return entities.SubEventResponse{}, err
	}

	var subEvent dbgen.SubEvent
	var createdDays []dbgen.SubEventDay
	txErr := postgres.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)

		created, err := repo.Create(ctx, orgID, eventID, req.Name, req.ScheduleMode)
		if err != nil {
			return utils.ErrInternal()
		}
		subEvent = created

		for _, d := range days {
			day, err := repo.CreateDay(ctx, subEvent.ID, d.date, d.entryTime, d.exitTime)
			if err != nil {
				return utils.ErrInternal()
			}
			createdDays = append(createdDays, day)
		}
		return nil
	})
	if txErr != nil {
		return entities.SubEventResponse{}, txErr
	}

	return toSubEventResponse(subEvent, createdDays), nil
}

func (s *SubEventsService) Get(ctx context.Context, orgID, eventID, id uuid.UUID) (entities.SubEventResponse, error) {
	if _, err := s.events.GetContext(ctx, orgID, eventID); err != nil {
		return entities.SubEventResponse{}, err
	}
	subEvent, err := s.repo.Get(ctx, orgID, eventID, id)
	if err != nil {
		return entities.SubEventResponse{}, utils.ErrNotFound("sub-event")
	}
	days, err := s.repo.ListDays(ctx, id)
	if err != nil {
		return entities.SubEventResponse{}, utils.ErrInternal()
	}
	return toSubEventResponse(subEvent, days), nil
}

// List is O(n) sub-event-day queries rather than one join — fine at the
// handful-of-sub-events-per-event scale this is built for; worth a join
// query if that assumption stops holding.
func (s *SubEventsService) List(ctx context.Context, orgID, eventID uuid.UUID) ([]entities.SubEventResponse, error) {
	if _, err := s.events.GetContext(ctx, orgID, eventID); err != nil {
		return nil, err
	}
	rows, err := s.repo.List(ctx, orgID, eventID)
	if err != nil {
		return nil, utils.ErrInternal()
	}

	resp := make([]entities.SubEventResponse, len(rows))
	for i, row := range rows {
		days, err := s.repo.ListDays(ctx, row.ID)
		if err != nil {
			return nil, utils.ErrInternal()
		}
		resp[i] = toSubEventResponse(row, days)
	}
	return resp, nil
}

func (s *SubEventsService) Update(ctx context.Context, orgID, eventID, id uuid.UUID, req entities.UpdateSubEventRequest) (entities.SubEventResponse, error) {
	parent, err := s.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return entities.SubEventResponse{}, err
	}
	if !parent.IsDraft() {
		return entities.SubEventResponse{}, utils.ErrConflict("sub-events can only be edited while the event is a draft")
	}

	existing, err := s.repo.Get(ctx, orgID, eventID, id)
	if err != nil {
		return entities.SubEventResponse{}, utils.ErrNotFound("sub-event")
	}
	days, err := parseSubEventScheduleMode(existing.ScheduleMode, req.Days, req.RangeStart, req.RangeEnd, req.RangeEntryTime, req.RangeExitTime, parent.Dates)
	if err != nil {
		return entities.SubEventResponse{}, err
	}

	var subEvent dbgen.SubEvent
	var updatedDays []dbgen.SubEventDay
	txErr := postgres.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		repo := s.repo.WithTx(tx)

		updated, err := repo.UpdateName(ctx, orgID, eventID, id, req.Name)
		if err != nil {
			return utils.ErrNotFound("sub-event")
		}
		subEvent = updated

		if err := repo.DeleteDaysForSubEvent(ctx, id); err != nil {
			return utils.ErrInternal()
		}
		for _, d := range days {
			day, err := repo.CreateDay(ctx, id, d.date, d.entryTime, d.exitTime)
			if err != nil {
				return utils.ErrInternal()
			}
			updatedDays = append(updatedDays, day)
		}
		return nil
	})
	if txErr != nil {
		return entities.SubEventResponse{}, txErr
	}

	return toSubEventResponse(subEvent, updatedDays), nil
}

func (s *SubEventsService) Delete(ctx context.Context, orgID, eventID, id uuid.UUID) error {
	parent, err := s.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return err
	}
	if !parent.IsDraft() {
		return utils.ErrConflict("sub-events can only be removed while the event is a draft")
	}

	deleted, err := s.repo.Delete(ctx, orgID, eventID, id)
	if err != nil {
		return utils.ErrInternal()
	}
	if !deleted {
		return utils.ErrNotFound("sub-event")
	}
	return nil
}

// ValidateIDs reports which of the given ids are real sub-events of this
// event/org, so a caller assigning a person to sub-events (PeopleService)
// can reject any id that isn't.
func (s *SubEventsService) ValidateIDs(ctx context.Context, orgID, eventID uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]bool, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]bool{}, nil
	}
	valid, err := s.repo.ValidateIDs(ctx, orgID, eventID, ids)
	if err != nil {
		return nil, utils.ErrInternal()
	}
	result := make(map[uuid.UUID]bool, len(valid))
	for _, id := range valid {
		result[id] = true
	}
	return result, nil
}

// subEventParsedDay is subevents' own day-parsing intermediate — named
// distinctly from events' parsedDay since both types now live in the same
// flat services package.
type subEventParsedDay struct {
	date      pgtype.Date
	entryTime pgtype.Time
	exitTime  pgtype.Time
}

// parseSubEventScheduleMode dispatches to the two sub-event authoring
// modes — selective (explicit dates, each validated against the parent's
// own materialized days) or fixed_range (a date span expanded day-by-day,
// every resulting date also validated against the parent — a range that
// reaches outside what the parent actually runs is a user error, not
// silently trimmed).
func parseSubEventScheduleMode(mode string, dayInputs []entities.EventDayInput, rangeStart, rangeEnd, rangeEntryTime, rangeExitTime string, allowedDates map[string]bool) ([]subEventParsedDay, error) {
	switch mode {
	case "selective":
		return parseSubEventDays(dayInputs, allowedDates)
	case "fixed_range":
		return expandSubEventFixedRange(rangeStart, rangeEnd, rangeEntryTime, rangeExitTime, allowedDates)
	default:
		return nil, utils.ErrValidation(map[string]string{"schedule_mode": "unknown schedule mode"})
	}
}

func parseSubEventDays(inputs []entities.EventDayInput, allowedDates map[string]bool) ([]subEventParsedDay, error) {
	seen := make(map[string]bool, len(inputs))
	days := make([]subEventParsedDay, 0, len(inputs))

	for _, in := range inputs {
		if !allowedDates[in.Date] {
			return nil, utils.ErrValidation(map[string]string{"days": "date is not part of the parent event: " + in.Date})
		}
		if seen[in.Date] {
			return nil, utils.ErrValidation(map[string]string{"days": "duplicate date: " + in.Date})
		}
		seen[in.Date] = true

		date, err := timeutil.ParseDate(in.Date)
		if err != nil {
			return nil, utils.ErrValidation(map[string]string{"days": "invalid date: " + in.Date})
		}
		entry, err := timeutil.ParseClock(in.EntryTime)
		if err != nil {
			return nil, utils.ErrValidation(map[string]string{"days": "invalid entry_time: " + in.EntryTime})
		}
		exit, err := timeutil.ParseClock(in.ExitTime)
		if err != nil {
			return nil, utils.ErrValidation(map[string]string{"days": "invalid exit_time: " + in.ExitTime})
		}
		if exit.Microseconds <= entry.Microseconds {
			return nil, utils.ErrValidation(map[string]string{"days": "exit_time must be after entry_time on " + in.Date})
		}

		days = append(days, subEventParsedDay{date: date, entryTime: entry, exitTime: exit})
	}
	if len(days) == 0 {
		return nil, utils.ErrValidation(map[string]string{"days": "at least one day is required"})
	}
	return days, nil
}

func expandSubEventFixedRange(rangeStart, rangeEnd, entryTimeStr, exitTimeStr string, allowedDates map[string]bool) ([]subEventParsedDay, error) {
	start, err := timeutil.ParseDate(rangeStart)
	if err != nil {
		return nil, utils.ErrValidation(map[string]string{"range_start": "invalid date"})
	}
	end, err := timeutil.ParseDate(rangeEnd)
	if err != nil {
		return nil, utils.ErrValidation(map[string]string{"range_end": "invalid date"})
	}
	if end.Time.Before(start.Time) {
		return nil, utils.ErrValidation(map[string]string{"range_end": "must be on or after range_start"})
	}
	entry, err := timeutil.ParseClock(entryTimeStr)
	if err != nil {
		return nil, utils.ErrValidation(map[string]string{"range_entry_time": "invalid entry_time"})
	}
	exit, err := timeutil.ParseClock(exitTimeStr)
	if err != nil {
		return nil, utils.ErrValidation(map[string]string{"range_exit_time": "invalid exit_time"})
	}
	if exit.Microseconds <= entry.Microseconds {
		return nil, utils.ErrValidation(map[string]string{"range_exit_time": "exit_time must be after entry_time"})
	}

	var days []subEventParsedDay
	for cur := start.Time; !cur.After(end.Time); cur = cur.AddDate(0, 0, 1) {
		dateStr := cur.Format(timeutil.DateLayout)
		if !allowedDates[dateStr] {
			return nil, utils.ErrValidation(map[string]string{"range_end": "date is not part of the parent event: " + dateStr})
		}
		days = append(days, subEventParsedDay{date: pgtype.Date{Time: cur, Valid: true}, entryTime: entry, exitTime: exit})
	}
	if len(days) == 0 {
		return nil, utils.ErrValidation(map[string]string{"days": "at least one day is required"})
	}
	return days, nil
}

func toSubEventResponse(subEvent dbgen.SubEvent, days []dbgen.SubEventDay) entities.SubEventResponse {
	dayResponses := make([]entities.EventDayResponse, len(days))
	for i, d := range days {
		dayResponses[i] = entities.EventDayResponse{
			Date:      timeutil.FormatDate(d.Date),
			EntryTime: timeutil.FormatClock(d.EntryTime),
			ExitTime:  timeutil.FormatClock(d.ExitTime),
		}
	}
	return entities.SubEventResponse{ID: subEvent.ID, Name: subEvent.Name, ScheduleMode: subEvent.ScheduleMode, Days: dayResponses}
}
