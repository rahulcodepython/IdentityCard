package subevents

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"identitycard-server/internal/generic"
	"identitycard-server/internal/pkg/postgres"
	"identitycard-server/internal/utils"
)

func (a *App) Create(ctx context.Context, orgID, eventID uuid.UUID, req CreateSubEventRequest) (SubEventResponse, error) {
	parent, err := a.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return SubEventResponse{}, err
	}
	if !parent.IsDraft() {
		return SubEventResponse{}, utils.ErrConflict("Sub-events can only be added while the event is a draft.", generic.ErrEventsNotDraft)
	}
	if parent.EventType != "grouped" {
		return SubEventResponse{}, utils.ErrValidation(map[string]string{"event_type": "Sub-events can only be added to a grouped event."})
	}

	pgDate, err := utils.ParseDate(req.Date)
	if err != nil {
		return SubEventResponse{}, utils.ErrValidation(map[string]string{"date": "Invalid date."})
	}
	date := pgDate.Time
	parentStart, _ := utils.ParseDate(parent.StartDate)
	parentEnd, _ := utils.ParseDate(parent.EndDate)
	if date.Before(parentStart.Time) || date.After(parentEnd.Time) {
		return SubEventResponse{}, utils.ErrValidation(map[string]string{"date": "Date must fall within the parent event's date range."})
	}
	entry, exit, err := parseTimeWindow(req.EntryTime, req.ExitTime)
	if err != nil {
		return SubEventResponse{}, err
	}

	subEvent, err := a.CreateSubEvent(ctx, orgID, eventID, req.Name, date, entry, exit)
	if err != nil {
		return SubEventResponse{}, utils.ErrInternal("Failed to create sub-event.", err)
	}
	return *subEvent, nil
}

func (a *App) Get(ctx context.Context, orgID, eventID, id uuid.UUID) (SubEventResponse, error) {
	if _, err := a.events.GetContext(ctx, orgID, eventID); err != nil {
		return SubEventResponse{}, err
	}
	subEvent, err := a.GetSubEvent(ctx, orgID, eventID, id)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) || errors.Is(err, generic.ErrSubEventsNotFound) {
			return SubEventResponse{}, utils.ErrNotFound("Sub-event not found.", err)
		}
		return SubEventResponse{}, utils.ErrInternal("Failed to fetch sub-event.", err)
	}
	return *subEvent, nil
}

func (a *App) List(ctx context.Context, orgID, eventID uuid.UUID) ([]SubEventResponse, error) {
	if _, err := a.events.GetContext(ctx, orgID, eventID); err != nil {
		return nil, err
	}
	rows, err := a.ListSubEvents(ctx, orgID, eventID)
	if err != nil {
		return nil, utils.ErrInternal("Failed to list sub-events.", err)
	}
	return rows, nil
}

func (a *App) Update(ctx context.Context, orgID, eventID, id uuid.UUID, req UpdateSubEventRequest) (SubEventResponse, error) {
	parent, err := a.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return SubEventResponse{}, err
	}
	if !parent.IsDraft() {
		return SubEventResponse{}, utils.ErrConflict("Sub-events can only be edited while the event is a draft.", generic.ErrEventsNotDraft)
	}

	pgDate, err := utils.ParseDate(req.Date)
	if err != nil {
		return SubEventResponse{}, utils.ErrValidation(map[string]string{"date": "Invalid date."})
	}
	date := pgDate.Time
	parentStart, _ := utils.ParseDate(parent.StartDate)
	parentEnd, _ := utils.ParseDate(parent.EndDate)
	if date.Before(parentStart.Time) || date.After(parentEnd.Time) {
		return SubEventResponse{}, utils.ErrValidation(map[string]string{"date": "Date must fall within the parent event's date range."})
	}
	entry, exit, err := parseTimeWindow(req.EntryTime, req.ExitTime)
	if err != nil {
		return SubEventResponse{}, err
	}

	subEvent, err := a.UpdateSubEvent(ctx, orgID, eventID, id, req.Name, date, entry, exit)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) || errors.Is(err, generic.ErrSubEventsNotFound) {
			return SubEventResponse{}, utils.ErrNotFound("Sub-event not found.", err)
		}
		return SubEventResponse{}, utils.ErrInternal("Failed to update sub-event.", err)
	}
	return *subEvent, nil
}

func (a *App) Delete(ctx context.Context, orgID, eventID, id uuid.UUID) error {
	parent, err := a.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return err
	}
	if !parent.IsDraft() {
		return utils.ErrConflict("Sub-events can only be removed while the event is a draft.", generic.ErrEventsNotDraft)
	}

	deleted, err := a.DeleteSubEvent(ctx, orgID, eventID, id)
	if err != nil {
		return utils.ErrInternal("Failed to delete sub-event.", err)
	}
	if !deleted {
		return utils.ErrNotFound("Sub-event not found.", generic.ErrSubEventsNotFound)
	}
	return nil
}

func (a *App) ValidateIDs(ctx context.Context, orgID, eventID uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]bool, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]bool{}, nil
	}
	valid, err := a.ValidateSubEventIDs(ctx, orgID, eventID, ids)
	if err != nil {
		return nil, utils.ErrInternal("Failed to validate sub-event IDs.", err)
	}
	result := make(map[uuid.UUID]bool, len(valid))
	for _, id := range valid {
		result[id] = true
	}
	return result, nil
}

func parseTimeWindow(entryStr, exitStr string) (string, string, error) {
	entry, err := utils.ParseClock(entryStr)
	if err != nil {
		return "", "", utils.ErrValidation(map[string]string{"entry_time": "Invalid entry_time: " + entryStr})
	}
	exit, err := utils.ParseClock(exitStr)
	if err != nil {
		return "", "", utils.ErrValidation(map[string]string{"exit_time": "Invalid exit_time: " + exitStr})
	}
	if exit.Microseconds <= entry.Microseconds {
		return "", "", utils.ErrValidation(map[string]string{"exit_time": "exit_time must be after entry_time"})
	}
	return entryStr, exitStr, nil
}
