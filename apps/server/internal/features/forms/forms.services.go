package forms

import (
    "context"
    "crypto/rand"
    "encoding/base64"
    "errors"
    "net/http"

    "github.com/google/uuid"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/features/people"
    "identitycard-server/internal/pkg/postgres"
    "identitycard-server/internal/utils"
)

func (a *App) Create(ctx context.Context, orgID, eventID uuid.UUID, req CreateFormRequest) (FormResponse, error) {
    if _, err := a.events.GetContext(ctx, orgID, eventID); err != nil {
        return FormResponse{}, err
    }

    if req.SubEventID != nil {
        if _, err := a.subevents.Get(ctx, orgID, eventID, *req.SubEventID); err != nil {
            return FormResponse{}, utils.ErrValidation(map[string]string{"sub_event_id": "Unknown sub-event."})
        }
    }

    token, err := formsGenerateToken()
    if err != nil {
        return FormResponse{}, utils.ErrInternal("Failed to generate form token.", err)
    }

    form, err := a.CreateForm(ctx, orgID, eventID, req.SubEventID, token, req.Capacity)
    if err != nil {
        return FormResponse{}, utils.ErrInternal("Failed to create form.", err)
    }
    return *form, nil
}

func (a *App) List(ctx context.Context, orgID, eventID uuid.UUID) ([]FormResponse, error) {
    if _, err := a.events.GetContext(ctx, orgID, eventID); err != nil {
        return nil, err
    }
    rows, err := a.ListForms(ctx, orgID, eventID)
    if err != nil {
        return nil, utils.ErrInternal("Failed to list forms.", err)
    }
    return rows, nil
}

func (a *App) Update(ctx context.Context, orgID, eventID, id uuid.UUID, req UpdateFormRequest) (FormResponse, error) {
    form, err := a.UpdateForm(ctx, orgID, eventID, id, req.Capacity, req.IsActive)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) || errors.Is(err, generic.ErrFormsNotFound) {
            return FormResponse{}, utils.ErrNotFound("Form not found.", err)
        }
        return FormResponse{}, utils.ErrInternal("Failed to update form.", err)
    }
    return *form, nil
}

func (a *App) Delete(ctx context.Context, orgID, eventID, id uuid.UUID) error {
    deleted, err := a.DeleteForm(ctx, orgID, eventID, id)
    if err != nil {
        return utils.ErrInternal("Failed to delete form.", err)
    }
    if !deleted {
        return utils.ErrNotFound("Form not found.", generic.ErrFormsNotFound)
    }
    return nil
}

func (a *App) GetPublic(ctx context.Context, token string) (PublicFormResponse, error) {
    row, err := a.GetPublicByToken(ctx, token)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) || errors.Is(err, generic.ErrFormsNotFound) {
            return PublicFormResponse{}, utils.ErrNotFound("Form not found.", err)
        }
        return PublicFormResponse{}, utils.ErrInternal("Failed to fetch public form.", err)
    }

    isOpen := row.IsActive && (row.Capacity == nil || row.SubmissionsCount < *row.Capacity)
    return PublicFormResponse{
        EventName:    row.EventName,
        SubEventName: row.SubEventName,
        IsOpen:       isOpen,
    }, nil
}

func (a *App) Submit(ctx context.Context, token string, req SubmitFormRequest) error {
    row, err := a.GetPublicByToken(ctx, token)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) || errors.Is(err, generic.ErrFormsNotFound) {
            return utils.ErrNotFound("Form not found.", err)
        }
        return utils.ErrInternal("Failed to fetch form.", err)
    }

    if _, err := a.IncrementSubmissions(ctx, row.ID); err != nil {
        return utils.NewError(http.StatusConflict, "This form is no longer accepting submissions.", generic.ErrFormsCapacityExceeded)
    }

    _, inserted, err := a.people.Submit(ctx, row.OrganizationID, row.EventID, people.PeopleSubmitInput{
        Email:      req.Email,
        Mobile:     req.Mobile,
        Name:       req.Name,
        ImageURL:   req.ImageURL,
        Age:        req.Age,
        Gender:     req.Gender,
        SubEventID: row.SubEventID,
    })
    if err != nil {
        _ = a.DecrementSubmissions(ctx, row.ID)
        return err
    }
    if !inserted {
        _ = a.DecrementSubmissions(ctx, row.ID)
    }
    return nil
}

func formsGenerateToken() (string, error) {
    buf := make([]byte, 18)
    if _, err := rand.Read(buf); err != nil {
        return "", err
    }
    return base64.RawURLEncoding.EncodeToString(buf), nil
}
