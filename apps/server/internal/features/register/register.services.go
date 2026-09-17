package register

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/google/uuid"

    "identitycard-server/internal/pkg/cache"
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
    ErrDataPayloadTooLarge = errors.New("form data payload exceeds 64KB maximum limit")
)

func (s *App) GetPublicApplyConfigService(ctx context.Context, eventFormID string) (*PublicApplyConfigResponse, error) {
    if _, err := uuid.Parse(eventFormID); err != nil {
        return nil, ErrInvalidEventFormID
    }

    cacheKey := fmt.Sprintf("cache:public_apply:config:%s", eventFormID)
    return cache.RememberWithJitter(ctx, s.Cache, cacheKey, 60*time.Second, cache.DefaultJitterPercentage, func() (*PublicApplyConfigResponse, error) {
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
    })
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

    // Dynamic schema payload safety guard: protect PostgreSQL JSONB column from oversized payloads
    const maxPayloadBytes = 64 * 1024
    if len(dataJSON) > maxPayloadBytes {
        return nil, ErrDataPayloadTooLarge
    }

    statusCode, finalUserID, err := s.SubmitApplicationRepository(ctx, eventFormID, "", trimmedName, trimmedEmail, dataJSON)
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
        if s.Cache != nil {
            _ = s.Cache.Delete(ctx, fmt.Sprintf("cache:public_apply:config:%s", eventFormID))
            _ = s.Cache.DeletePattern(ctx, "cache:event_forms:*")
            _ = s.Cache.DeletePattern(ctx, "cache:assigned_form:*")
        }
        return &SubmitApplicationResponse{
            UserID:    finalUserID,
            Message:   "Application submitted successfully",
            CreatedAt: time.Now().UTC().Format(time.RFC3339),
        }, nil
    default:
        return nil, ErrFormNotFound
    }
}
