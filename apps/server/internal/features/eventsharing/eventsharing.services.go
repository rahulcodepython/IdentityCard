package eventsharing

import (
    "context"
    "errors"
    "time"

    "github.com/google/uuid"
)

var (
    ErrInvalidEventID       = errors.New("invalid event id")
    ErrInvalidFormID        = errors.New("invalid form id")
    ErrFormNotPublished     = errors.New("only published forms can be assigned to an event")
    ErrCannotChangeForm     = errors.New("cannot change form because applicants have already submitted data")
    ErrInvalidMaxApplicants = errors.New("max applicants must be -1 (unlimited) or at least 1")
    ErrInvalidExpiryDate    = errors.New("expiry date must be in the future")
    ErrInvalidStatus        = errors.New("status must be either 'waiting' or 'live'")
)

func (s *App) GetSharingService(ctx context.Context, eventID string) (*EventSharingResponse, error) {
    if _, err := uuid.Parse(eventID); err != nil {
        return nil, ErrInvalidEventID
    }
    return s.GetSharingRepository(ctx, eventID)
}

func (s *App) UpdateSharingService(ctx context.Context, eventID string, req UpdateEventSharingRequest) (*EventSharingResponse, error) {
    if _, err := uuid.Parse(eventID); err != nil {
        return nil, ErrInvalidEventID
    }
    if _, err := uuid.Parse(req.FormID); err != nil {
        return nil, ErrInvalidFormID
    }

    if req.MaxApplicants != -1 && req.MaxApplicants < 1 {
        return nil, ErrInvalidMaxApplicants
    }

    if req.ExpiresAt.Before(time.Now()) {
        return nil, ErrInvalidExpiryDate
    }

    status := req.Status
    if status == "" {
        status = StatusWaiting
    }
    if status != StatusWaiting && status != StatusLive {
        return nil, ErrInvalidStatus
    }

    res, err := s.UpsertSharingRepository(ctx, eventID, req.FormID, req.MaxApplicants, req.ExpiresAt, status)
    if err != nil {
        return nil, err
    }
    if res == nil {
        return nil, errors.New("failed to update sharing")
    }

    switch res.StatusCode {
    case "form_not_found":
        return nil, ErrInvalidFormID
    case "form_not_published":
        return nil, ErrFormNotPublished
    case "cannot_change_form":
        return nil, ErrCannotChangeForm
    case "ok":
        return res.Response, nil
    default:
        return nil, errors.New("unexpected error updating sharing")
    }
}
