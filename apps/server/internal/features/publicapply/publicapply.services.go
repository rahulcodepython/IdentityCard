package publicapply

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/google/uuid"
    "identitycard-server/internal/pkg/postgres"
)

var (
    ErrInvalidEventFormID = errors.New("invalid event form id")
    ErrFormNotFound       = errors.New("registration form not found")
    ErrFormNotLive        = errors.New("registration form is not live")
    ErrFormExpired        = errors.New("registration deadline has passed")
    ErrFormLimitReached   = errors.New("registration capacity is full")
    ErrAlreadyRegistered  = errors.New("an application has already been submitted with this email for this event")
    ErrInvalidName        = errors.New("name cannot be empty")
    ErrInvalidEmail       = errors.New("valid email address is required")
)

// GenerateApplicantUserID creates a custom user_id format: (event_uuid + random(uuid) + date).
// Format: {event_uuid_first_8}-{random_uuid_first_8}-{YYYYMMDD}
func GenerateApplicantUserID(eventID string) string {
    eventClean := strings.ReplaceAll(eventID, "-", "")
    if len(eventClean) > 8 {
        eventClean = eventClean[:8]
    }

    randomClean := strings.ReplaceAll(uuid.NewString(), "-", "")
    if len(randomClean) > 8 {
        randomClean = randomClean[:8]
    }

    datePart := time.Now().UTC().Format("20060102")
    return fmt.Sprintf("%s-%s-%s", eventClean, randomClean, datePart)
}

func (s *App) GetPublicApplyConfigService(ctx context.Context, eventFormID string) (*PublicApplyConfigResponse, error) {
    if _, err := uuid.Parse(eventFormID); err != nil {
        return nil, ErrInvalidEventFormID
    }

    res, err := s.GetPublicApplyConfigRepository(ctx, eventFormID)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) {
            return nil, ErrFormNotFound
        }
        return nil, err
    }
    if res == nil || res.EventFormID == "" {
        return nil, ErrFormNotFound
    }

    // If form is not live, expired, or full, nullify the form schema
    if res.Status != "live" || res.IsExpired || res.IsFull {
        res.Form = nil
    }

    return res, nil
}

func (s *App) SubmitApplicationService(ctx context.Context, eventFormID string, req SubmitApplicationRequest) (*SubmitApplicationResponse, error) {
    if _, err := uuid.Parse(eventFormID); err != nil {
        return nil, ErrInvalidEventFormID
    }

    trimmedName := strings.TrimSpace(req.Name)
    if trimmedName == "" {
        return nil, ErrInvalidName
    }

    trimmedEmail := strings.ToLower(strings.TrimSpace(req.Email))
    if trimmedEmail == "" || !strings.Contains(trimmedEmail, "@") {
        return nil, ErrInvalidEmail
    }

    if req.Data == nil {
        req.Data = make(map[string]interface{})
    }

    dataJSON, err := json.Marshal(req.Data)
    if err != nil {
        return nil, err
    }

    userID := GenerateApplicantUserID(eventFormID)

    statusCode, err := s.SubmitApplicationRepository(ctx, eventFormID, userID, trimmedName, trimmedEmail, dataJSON)
    if err != nil {
        return nil, err
    }

    switch statusCode {
    case "not_found":
        return nil, ErrFormNotFound
    case "not_live":
        return nil, ErrFormNotLive
    case "expired":
        return nil, ErrFormExpired
    case "limit_reached":
        return nil, ErrFormLimitReached
    case "already_registered":
        return nil, ErrAlreadyRegistered
    case "ok":
        return &SubmitApplicationResponse{
            UserID:    userID,
            Message:   "Application submitted successfully",
            CreatedAt: time.Now().UTC().Format(time.RFC3339),
        }, nil
    default:
        return nil, ErrFormNotFound
    }
}
