package forms

import (
    "context"
    "encoding/json"
    "errors"
    "strings"
    "time"

    "github.com/google/uuid"

    "identitycard-server/internal/pkg/postgres"
)

var (
    ErrInvalidEventID             = errors.New("invalid event id")
    ErrEventFormNotFound          = errors.New("event form not found")
    ErrEventFormAlreadyExists     = errors.New("an event form has already been created for this event")
    ErrFormIsLocked               = errors.New("form is locked and cannot be modified")
    ErrCannotDeleteWithApplicants = errors.New("cannot delete form: applicants have already registered")
    ErrInvalidTemplateID          = errors.New("invalid template id")
    ErrInvalidExpiresAt           = errors.New("expiration date must be in the future")
    ErrInvalidMaxApplicants       = errors.New("max applicants must be -1 or at least 1")
    ErrInvalidSource              = errors.New("source must be either 'template' or 'scratch'")
)

type rawFieldCheck struct {
    Key      string `json:"key"`
    Type     string `json:"type"`
    IsSystem bool   `json:"is_system"`
}

func defaultEventFormFields() []byte {
    return []byte(`[
        {"id":"field_default_name","key":"name","type":"text","label":"Full Name","options":[],"required":true,"is_system":true,"placeholder":"Enter your full name"},
        {"id":"field_default_email","key":"email","type":"email","label":"Email Address","options":[],"required":true,"is_system":true,"placeholder":"name@example.com"}
    ]`)
}

func validateEventFieldsJSON(raw []byte) error {
    if len(raw) == 0 {
        return ErrMissingMandatoryFields
    }
    var fields []rawFieldCheck
    if err := json.Unmarshal(raw, &fields); err != nil {
        return err
    }
    var hasName, hasEmail bool
    for _, f := range fields {
        if f.Key == "name" || (f.IsSystem && f.Type == "text") {
            hasName = true
        }
        if f.Key == "email" || (f.IsSystem && f.Type == "email") {
            hasEmail = true
        }
    }
    if !hasName || !hasEmail {
        return ErrMissingMandatoryFields
    }
    return nil
}

func (s *App) GetEventFormService(ctx context.Context, eventID string) (*EventFormDetails, error) {
    if _, err := uuid.Parse(eventID); err != nil {
        return nil, ErrInvalidEventID
    }

    ef, err := s.GetEventFormRepository(ctx, eventID)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) {
            return nil, ErrEventFormNotFound
        }
        return nil, err
    }

    return ef, nil
}

func (s *App) CreateEventFormService(ctx context.Context, eventID string, req CreateEventFormRequest) (*EventFormDetails, error) {
    if _, err := uuid.Parse(eventID); err != nil {
        return nil, ErrInvalidEventID
    }

    // Guard: ensure event does not already have a form
    existing, err := s.GetEventFormRepository(ctx, eventID)
    if err != nil && !errors.Is(err, postgres.ErrNotFound) {
        return nil, err
    }
    if existing != nil {
        return nil, ErrEventFormAlreadyExists
    }

    name := strings.TrimSpace(req.Name)
    if name == "" {
        name = "Event Registration Form"
    }

    if req.MaxApplicants != -1 && req.MaxApplicants < 1 {
        return nil, ErrInvalidMaxApplicants
    }

    if time.Now().After(req.ExpiresAt) {
        return nil, ErrInvalidExpiresAt
    }

    source := strings.ToLower(strings.TrimSpace(req.Source))
    switch source {
    case "template":
        if req.TemplateID == nil || *req.TemplateID == "" {
            return nil, ErrInvalidTemplateID
        }
        if _, err := uuid.Parse(*req.TemplateID); err != nil {
            return nil, ErrInvalidTemplateID
        }
        _, err = s.CreateFromTemplateRepository(ctx, eventID, *req.TemplateID, name, req.MaxApplicants, req.ExpiresAt)
        if err != nil {
            return nil, err
        }

    case "scratch":
        fieldsBytes := req.Fields
        if len(fieldsBytes) == 0 || string(fieldsBytes) == "null" || string(fieldsBytes) == "[]" {
            fieldsBytes = defaultEventFormFields()
        } else {
            if err := validateEventFieldsJSON(fieldsBytes); err != nil {
                return nil, err
            }
        }
        _, err = s.CreateFromScratchRepository(ctx, eventID, name, fieldsBytes, req.MaxApplicants, req.ExpiresAt)
        if err != nil {
            return nil, err
        }

    default:
        return nil, ErrInvalidSource
    }

    return s.GetEventFormService(ctx, eventID)
}

func (s *App) UpdateEventFormService(ctx context.Context, eventID string, req UpdateEventFormRequest) (*EventFormDetails, error) {
    if _, err := uuid.Parse(eventID); err != nil {
        return nil, ErrInvalidEventID
    }

    existing, err := s.GetEventFormService(ctx, eventID)
    if err != nil {
        return nil, err
    }

    if existing.IsLocked {
        return nil, ErrFormIsLocked
    }

    if req.MaxApplicants != nil {
        if *req.MaxApplicants != -1 && *req.MaxApplicants < 1 {
            return nil, ErrInvalidMaxApplicants
        }
    }

    if req.ExpiresAt != nil {
        if time.Now().After(*req.ExpiresAt) {
            return nil, ErrInvalidExpiresAt
        }
    }

    var fieldsJSON []byte
    if req.Fields != nil {
        fieldsJSON = *req.Fields
        if err := validateEventFieldsJSON(fieldsJSON); err != nil {
            return nil, err
        }
    }

    var trimmedName *string
    if req.Name != nil {
        n := strings.TrimSpace(*req.Name)
        if n == "" {
            return nil, ErrInvalidFormName
        }
        trimmedName = &n
    }

    err = s.UpdateEventFormRepository(ctx, eventID, trimmedName, fieldsJSON, req.MaxApplicants, req.ExpiresAt)
    if err != nil {
        return nil, err
    }

    return s.GetEventFormService(ctx, eventID)
}

func (s *App) LockEventFormService(ctx context.Context, eventID string) (*EventFormDetails, error) {
    if _, err := uuid.Parse(eventID); err != nil {
        return nil, ErrInvalidEventID
    }

    existing, err := s.GetEventFormService(ctx, eventID)
    if err != nil {
        return nil, err
    }

    if existing.IsLocked {
        return existing, nil
    }

    if time.Now().After(existing.ExpiresAt) {
        return nil, ErrInvalidExpiresAt
    }

    if err := validateEventFieldsJSON(existing.Fields); err != nil {
        return nil, err
    }

    err = s.LockEventFormRepository(ctx, eventID)
    if err != nil {
        return nil, err
    }

    return s.GetEventFormService(ctx, eventID)
}

func (s *App) DeleteEventFormService(ctx context.Context, eventID string) error {
    if _, err := uuid.Parse(eventID); err != nil {
        return ErrInvalidEventID
    }

    existing, err := s.GetEventFormService(ctx, eventID)
    if err != nil {
        return err
    }

    count, err := s.CountApplicantsRepository(ctx, eventID)
    if err != nil {
        return err
    }
    if count > 0 || existing.TotalApplicants > 0 {
        return ErrCannotDeleteWithApplicants
    }

    return s.DeleteEventFormRepository(ctx, eventID)
}
