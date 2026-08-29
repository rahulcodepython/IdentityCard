package subevents

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/utils"
	"identitycard-server/internal/utils/timeutil"
)

func (a *App) Create(ctx context.Context, orgID, eventID uuid.UUID, req CreateSubEventRequest) (SubEventResponse, error) {
	parent, err := a.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return SubEventResponse{}, err
	}
	if !parent.IsDraft() {
		return SubEventResponse{}, utils.ErrConflict("sub-events can only be added while the event is a draft")
	}
	if parent.EventType != "grouped" {
		return SubEventResponse{}, utils.ErrValidation(map[string]string{"event_type": "sub-events can only be added to a grouped event"})
	}

	date, err := timeutil.ParseDate(req.Date)
	if err != nil {
		return SubEventResponse{}, utils.ErrValidation(map[string]string{"date": "invalid date"})
	}
	if date.Time.Before(parent.StartDate.Time) || date.Time.After(parent.EndDate.Time) {
		return SubEventResponse{}, utils.ErrValidation(map[string]string{"date": "date must fall within the parent event's date range"})
	}
	entry, exit, err := parseTimeWindow(req.EntryTime, req.ExitTime)
	if err != nil {
		return SubEventResponse{}, err
	}

	subEvent, err := a.CreateSubEvent(ctx, orgID, eventID, req.Name, date, entry, exit)
	if err != nil {
		return SubEventResponse{}, utils.ErrInternal()
	}
	return toSubEventResponse(subEvent), nil
}

func (a *App) Get(ctx context.Context, orgID, eventID, id uuid.UUID) (SubEventResponse, error) {
	if _, err := a.events.GetContext(ctx, orgID, eventID); err != nil {
		return SubEventResponse{}, err
	}
	subEvent, err := a.GetSubEvent(ctx, orgID, eventID, id)
	if err != nil {
		return SubEventResponse{}, utils.ErrNotFound("sub-event")
	}
	return toSubEventResponse(subEvent), nil
}

func (a *App) List(ctx context.Context, orgID, eventID uuid.UUID) ([]SubEventResponse, error) {
	if _, err := a.events.GetContext(ctx, orgID, eventID); err != nil {
		return nil, err
	}
	rows, err := a.ListSubEvents(ctx, orgID, eventID)
	if err != nil {
		return nil, utils.ErrInternal()
	}
	resp := make([]SubEventResponse, len(rows))
	for i, row := range rows {
		resp[i] = toSubEventResponse(row)
	}
	return resp, nil
}

func (a *App) Update(ctx context.Context, orgID, eventID, id uuid.UUID, req UpdateSubEventRequest) (SubEventResponse, error) {
	parent, err := a.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return SubEventResponse{}, err
	}
	if !parent.IsDraft() {
		return SubEventResponse{}, utils.ErrConflict("sub-events can only be edited while the event is a draft")
	}

	date, err := timeutil.ParseDate(req.Date)
	if err != nil {
		return SubEventResponse{}, utils.ErrValidation(map[string]string{"date": "invalid date"})
	}
	if date.Time.Before(parent.StartDate.Time) || date.Time.After(parent.EndDate.Time) {
		return SubEventResponse{}, utils.ErrValidation(map[string]string{"date": "date must fall within the parent event's date range"})
	}
	entry, exit, err := parseTimeWindow(req.EntryTime, req.ExitTime)
	if err != nil {
		return SubEventResponse{}, err
	}

	subEvent, err := a.UpdateSubEvent(ctx, orgID, eventID, id, req.Name, date, entry, exit)
	if err != nil {
		return SubEventResponse{}, utils.ErrNotFound("sub-event")
	}
	return toSubEventResponse(subEvent), nil
}

func (a *App) Delete(ctx context.Context, orgID, eventID, id uuid.UUID) error {
	parent, err := a.events.GetContext(ctx, orgID, eventID)
	if err != nil {
		return err
	}
	if !parent.IsDraft() {
		return utils.ErrConflict("sub-events can only be removed while the event is a draft")
	}

	deleted, err := a.DeleteSubEvent(ctx, orgID, eventID, id)
	if err != nil {
		return utils.ErrInternal()
	}
	if !deleted {
		return utils.ErrNotFound("sub-event")
	}
	return nil
}

func (a *App) ValidateIDs(ctx context.Context, orgID, eventID uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]bool, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]bool{}, nil
	}
	valid, err := a.ValidateSubEventIDs(ctx, orgID, eventID, ids)
	if err != nil {
		return nil, utils.ErrInternal()
	}
	result := make(map[uuid.UUID]bool, len(valid))
	for _, id := range valid {
		result[id] = true
	}
	return result, nil
}

func parseTimeWindow(entryStr, exitStr string) (pgtype.Time, pgtype.Time, error) {
	entry, err := timeutil.ParseClock(entryStr)
	if err != nil {
		return pgtype.Time{}, pgtype.Time{}, utils.ErrValidation(map[string]string{"entry_time": "invalid entry_time: " + entryStr})
	}
	exit, err := timeutil.ParseClock(exitStr)
	if err != nil {
		return pgtype.Time{}, pgtype.Time{}, utils.ErrValidation(map[string]string{"exit_time": "invalid exit_time: " + exitStr})
	}
	if exit.Microseconds <= entry.Microseconds {
		return pgtype.Time{}, pgtype.Time{}, utils.ErrValidation(map[string]string{"exit_time": "exit_time must be after entry_time"})
	}
	return entry, exit, nil
}

func toSubEventResponse(se dbgen.SubEvent) SubEventResponse {
	return SubEventResponse{
		ID:        se.ID,
		Name:      se.Name,
		Date:      timeutil.FormatDate(se.Date),
		EntryTime: timeutil.FormatClock(se.EntryTime),
		ExitTime:  timeutil.FormatClock(se.ExitTime),
	}
}
