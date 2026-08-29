package services

import (
    "context"

    "github.com/google/uuid"

    dbgen "identitycard-server/internal/db/sqlc/generated"
    "identitycard-server/internal/entities"
    "identitycard-server/internal/repositories"
    "identitycard-server/internal/utils"
    "identitycard-server/internal/utils/timeutil"
)


type SubEventsService struct {
    repo   *repositories.SubEventsRepository
    events *EventsService
}

func NewSubEventsService(repo *repositories.SubEventsRepository, eventsService *EventsService) *SubEventsService {
    return &SubEventsService{repo: repo, events: eventsService}
}

// Create validates that the parent event is a grouped draft, that req.Date
// falls within the parent's [start_date, end_date], then writes a single
// sub_events row (date/entry_time/exit_time live on the row directly — no
// sub_event_days table). No credit consumption: one credit funds the grouped
// parent at creation time; sub-events are free.
func (s *SubEventsService) Create(ctx context.Context, orgID, eventID uuid.UUID, req entities.CreateSubEventRequest) (entities.SubEventResponse, error) {
    parent, err := s.events.GetContext(ctx, orgID, eventID)
    if err != nil {
        return entities.SubEventResponse{}, err
    }
    if !parent.IsDraft() {
        return entities.SubEventResponse{}, utils.ErrConflict("sub-events can only be added while the event is a draft")
    }
    if parent.EventType != "grouped" {
        return entities.SubEventResponse{}, utils.ErrValidation(map[string]string{"event_type": "sub-events can only be added to a grouped event"})
    }

    date, err := timeutil.ParseDate(req.Date)
    if err != nil {
        return entities.SubEventResponse{}, utils.ErrValidation(map[string]string{"date": "invalid date"})
    }
    if date.Time.Before(parent.StartDate.Time) || date.Time.After(parent.EndDate.Time) {
        return entities.SubEventResponse{}, utils.ErrValidation(map[string]string{"date": "date must fall within the parent event's date range"})
    }
    entry, exit, err := parseTimeWindow(req.EntryTime, req.ExitTime)
    if err != nil {
        return entities.SubEventResponse{}, err
    }

    subEvent, err := s.repo.Create(ctx, orgID, eventID, req.Name, date, entry, exit)
    if err != nil {
        return entities.SubEventResponse{}, utils.ErrInternal()
    }
    return toSubEventResponse(subEvent), nil
}

func (s *SubEventsService) Get(ctx context.Context, orgID, eventID, id uuid.UUID) (entities.SubEventResponse, error) {
    if _, err := s.events.GetContext(ctx, orgID, eventID); err != nil {
        return entities.SubEventResponse{}, err
    }
    subEvent, err := s.repo.Get(ctx, orgID, eventID, id)
    if err != nil {
        return entities.SubEventResponse{}, utils.ErrNotFound("sub-event")
    }
    return toSubEventResponse(subEvent), nil
}

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
        resp[i] = toSubEventResponse(row)
    }
    return resp, nil
}

// Update validates the same constraints as Create (parent grouped + draft,
// date in range) then replaces name/date/entry_time/exit_time atomically.
// Also gated on the parent event's credit restriction, since a restricted
// event's sub-events should be equally frozen.
func (s *SubEventsService) Update(ctx context.Context, orgID, eventID, id uuid.UUID, req entities.UpdateSubEventRequest) (entities.SubEventResponse, error) {
    parent, err := s.events.GetContext(ctx, orgID, eventID)
    if err != nil {
        return entities.SubEventResponse{}, err
    }
    if !parent.IsDraft() {
        return entities.SubEventResponse{}, utils.ErrConflict("sub-events can only be edited while the event is a draft")
    }

    restricted, err := s.events.plans.GetCreditRestriction(ctx, eventID)
    if err != nil {
        return entities.SubEventResponse{}, err
    }
    if restricted {
        return entities.SubEventResponse{}, utils.NewError(403, "credit_restricted", "this event's billing is past due — renew your plan to edit")
    }

    date, err := timeutil.ParseDate(req.Date)
    if err != nil {
        return entities.SubEventResponse{}, utils.ErrValidation(map[string]string{"date": "invalid date"})
    }
    if date.Time.Before(parent.StartDate.Time) || date.Time.After(parent.EndDate.Time) {
        return entities.SubEventResponse{}, utils.ErrValidation(map[string]string{"date": "date must fall within the parent event's date range"})
    }
    entry, exit, err := parseTimeWindow(req.EntryTime, req.ExitTime)
    if err != nil {
        return entities.SubEventResponse{}, err
    }

    subEvent, err := s.repo.Update(ctx, orgID, eventID, id, req.Name, date, entry, exit)
    if err != nil {
        return entities.SubEventResponse{}, utils.ErrNotFound("sub-event")
    }
    return toSubEventResponse(subEvent), nil
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

func toSubEventResponse(se dbgen.SubEvent) entities.SubEventResponse {
    return entities.SubEventResponse{
        ID:        se.ID,
        Name:      se.Name,
        Date:      timeutil.FormatDate(se.Date),
        EntryTime: timeutil.FormatClock(se.EntryTime),
        ExitTime:  timeutil.FormatClock(se.ExitTime),
    }
}
