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
    ErrInvalidPhone       = errors.New("mobile number is required")
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

    trimmedPhone := strings.TrimSpace(req.Phone)
    if trimmedPhone == "" {
        if p, ok := req.Data["phone"].(string); ok && strings.TrimSpace(p) != "" {
            trimmedPhone = strings.TrimSpace(p)
        } else if m, ok := req.Data["mobile"].(string); ok && strings.TrimSpace(m) != "" {
            trimmedPhone = strings.TrimSpace(m)
        }
    }
    if trimmedPhone == "" {
        return nil, ErrInvalidPhone
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

    // 1. Fetch cached form config to validate live status & limits
    cfg, err := s.GetPublicApplyConfigService(ctx, eventFormID)
    if err != nil {
        return nil, err
    }
    if cfg == nil || cfg.Event == nil {
        return nil, ErrFormNotFound
    }
    if cfg.Status != "live" {
        return nil, ErrFormNotLive
    }
    if cfg.IsExpired || time.Now().After(cfg.ExpiresAt) {
        return nil, ErrFormExpired
    }
    if cfg.MaxApplicants != -1 && cfg.IsFull {
        return nil, ErrFormLimitReached
    }

    // 2. Pre-generate unique UserID (Edge ID Generation)
    eventPrefix := "app"
    if len(cfg.Event.ID) >= 8 {
        eventPrefix = strings.ReplaceAll(cfg.Event.ID[:8], "-", "")
    }
    randomSuffix := strings.ReplaceAll(uuid.NewString()[:8], "-", "")
    datePart := time.Now().Format("20060102")
    finalUserID := fmt.Sprintf("%s-%s-%s", eventPrefix, randomSuffix, datePart)
    nowISO := time.Now().UTC().Format(time.RFC3339)

    // 3. Fast Ingestion via Redis Stream & atomic reservation
    if s.RDB != nil {
        sub := &QueuedSubmission{
            EventID:     cfg.Event.ID,
            EventFormID: eventFormID,
            UserID:      finalUserID,
            Name:        trimmedName,
            Email:       trimmedEmail,
            Phone:       trimmedPhone,
            Data:        req.Data,
            CreatedAt:   nowISO,
        }

        queueStatus, qErr := s.EnqueueSubmission(ctx, sub, cfg.MaxApplicants, cfg.CurrentApplicants)
        if qErr == nil {
            switch queueStatus {
            case "already_registered":
                return nil, ErrAlreadyRegistered
            case "limit_reached":
                return nil, ErrFormLimitReached
            case "ok":
                return &SubmitApplicationResponse{
                    UserID:    finalUserID,
                    Status:    "submitted",
                    Message:   "Application submitted successfully",
                    CreatedAt: nowISO,
                }, nil
            }
        }
    }

    // 4. Fallback to direct transactional database persistence if Redis is unreachable
    statusCode, generatedID, err := s.SubmitApplicationRepository(ctx, eventFormID, finalUserID, trimmedName, trimmedEmail, trimmedPhone, dataJSON)
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
        if generatedID != "" {
            finalUserID = generatedID
        }
        return &SubmitApplicationResponse{
            UserID:    finalUserID,
            Status:    "submitted",
            Message:   "Application submitted successfully",
            CreatedAt: nowISO,
        }, nil
    default:
        return nil, ErrFormNotFound
    }
}
