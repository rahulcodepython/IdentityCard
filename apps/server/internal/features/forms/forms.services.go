package forms

import (
    "context"
    "encoding/json"
    "errors"
    "strings"

    "github.com/google/uuid"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/pkg/postgres"
)

var (
    ErrInvalidFormID             = errors.New("invalid form id")
    ErrFormNotFound              = errors.New("form not found")
    ErrInvalidFormName           = errors.New("form name cannot be empty")
    ErrMissingMandatoryFields    = errors.New("mandatory fields (name and email) cannot be removed")
    ErrInvalidFieldType          = errors.New("unsupported field type")
    ErrFormAlreadyPublished      = errors.New("form is already published")
    ErrFormPublishedCannotModify = errors.New("published forms cannot be modified")
)

// DefaultFormFields generates mandatory system fields that every form must have.
func DefaultFormFields() []FormField {
    return []FormField{
        {
            ID:          "field_default_name",
            Key:         "name",
            Label:       "Full Name",
            Type:        "text",
            Required:    true,
            Placeholder: "Enter your full name",
            IsSystem:    true,
            Options:     []FieldOption{},
        },
        {
            ID:          "field_default_email",
            Key:         "email",
            Label:       "Email Address",
            Type:        "email",
            Required:    true,
            Placeholder: "name@example.com",
            IsSystem:    true,
            Options:     []FieldOption{},
        },
    }
}

// ListService retrieves paginated forms.
func (s *App) ListService(ctx context.Context, search string, isPublished *bool, page, limit int) (*generic.PaginatedResponse[[]Form], error) {
    search = strings.TrimSpace(search)
    offset := (page - 1) * limit

    return s.ListRepository(ctx, search, isPublished, page, limit, offset)
}

// GetService finds a form by UUID.
func (s *App) GetService(ctx context.Context, id string) (*Form, error) {
    if _, err := uuid.Parse(id); err != nil {
        return nil, ErrInvalidFormID
    }

    f, err := s.GetRepository(ctx, id)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) {
            return nil, ErrFormNotFound
        }
        return nil, err
    }

    return f, nil
}

// CreateService validates name and initializes form with default system fields.
func (s *App) CreateService(ctx context.Context, req CreateFormRequest) (*Form, error) {
    trimmedName := strings.TrimSpace(req.Name)
    if trimmedName == "" {
        return nil, ErrInvalidFormName
    }

    defaultFields := DefaultFormFields()
    fieldsJSON, err := json.Marshal(defaultFields)
    if err != nil {
        return nil, err
    }

    return s.CreateRepository(ctx, trimmedName, fieldsJSON)
}

// UpdateService updates form metadata (name).
func (s *App) UpdateService(ctx context.Context, id string, req UpdateFormRequest) (*Form, error) {
    if _, err := uuid.Parse(id); err != nil {
        return nil, ErrInvalidFormID
    }

    existing, err := s.GetService(ctx, id)
    if err != nil {
        return nil, err
    }
    if existing.IsPublished {
        return nil, ErrFormPublishedCannotModify
    }

    if req.Name != nil {
        trimmed := strings.TrimSpace(*req.Name)
        if trimmed == "" {
            return nil, ErrInvalidFormName
        }
        req.Name = &trimmed
    }

    f, err := s.UpdateRepository(ctx, id, req)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) {
            return nil, ErrFormNotFound
        }
        return nil, err
    }

    return f, nil
}

// PublishService permanently publishes a form.
func (s *App) PublishService(ctx context.Context, id string) (*Form, error) {
    if _, err := uuid.Parse(id); err != nil {
        return nil, ErrInvalidFormID
    }

    existing, err := s.GetService(ctx, id)
    if err != nil {
        return nil, err
    }
    if existing.IsPublished {
        return nil, ErrFormAlreadyPublished
    }

    f, err := s.PublishRepository(ctx, id)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) {
            return nil, ErrFormNotFound
        }
        return nil, err
    }

    return f, nil
}

// UpdateFieldsService validates all field types and ensures mandatory fields remain intact.
func (s *App) UpdateFieldsService(ctx context.Context, id string, req UpdateFormFieldsRequest) (*Form, error) {
    if _, err := uuid.Parse(id); err != nil {
        return nil, ErrInvalidFormID
    }

    existing, err := s.GetService(ctx, id)
    if err != nil {
        return nil, err
    }
    if existing.IsPublished {
        return nil, ErrFormPublishedCannotModify
    }

    hasName := false
    hasEmail := false

    normalizedFields := make([]FormField, len(req.Fields))
    for i, field := range req.Fields {
        if !SupportedFieldTypes[field.Type] {
            return nil, ErrInvalidFieldType
        }

        if field.ID == "" {
            field.ID = uuid.NewString()
        }

        if field.Key == "name" || (field.IsSystem && field.Type == "text") {
            hasName = true
            field.IsSystem = true
            field.Required = true
            field.Key = "name"
        }

        if field.Key == "email" || (field.IsSystem && field.Type == "email") {
            hasEmail = true
            field.IsSystem = true
            field.Required = true
            field.Key = "email"
        }

        if field.Options == nil {
            field.Options = []FieldOption{}
        }

        normalizedFields[i] = field
    }

    if !hasName || !hasEmail {
        return nil, ErrMissingMandatoryFields
    }

    fieldsJSON, err := json.Marshal(normalizedFields)
    if err != nil {
        return nil, err
    }

    f, err := s.UpdateFieldsRepository(ctx, id, fieldsJSON)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) {
            return nil, ErrFormNotFound
        }
        return nil, err
    }

    return f, nil
}

// DeleteService removes a form by UUID.
func (s *App) DeleteService(ctx context.Context, id string) error {
    if _, err := uuid.Parse(id); err != nil {
        return ErrInvalidFormID
    }

    err := s.DeleteRepository(ctx, id)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) {
            return ErrFormNotFound
        }
        return err
    }

    return nil
}
