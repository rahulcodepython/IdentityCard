package forms

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/features/people"
	"identitycard-server/internal/utils"
)

func (a *App) Create(ctx context.Context, orgID, eventID uuid.UUID, req CreateFormRequest) (FormResponse, error) {
	if _, err := a.events.GetContext(ctx, orgID, eventID); err != nil {
		return FormResponse{}, err
	}

	subEventID := pgtype.UUID{}
	if req.SubEventID != nil {
		if _, err := a.subevents.Get(ctx, orgID, eventID, *req.SubEventID); err != nil {
			return FormResponse{}, utils.ErrValidation(map[string]string{"sub_event_id": "unknown sub-event"})
		}
		subEventID = pgtype.UUID{Bytes: [16]byte(*req.SubEventID), Valid: true}
	}

	token, err := formsGenerateToken()
	if err != nil {
		return FormResponse{}, utils.ErrInternal()
	}

	form, err := a.CreateForm(ctx, orgID, eventID, subEventID, token, formsToPgInt4(req.Capacity))
	if err != nil {
		return FormResponse{}, utils.ErrInternal()
	}
	return toFormResponse(form), nil
}

func (a *App) List(ctx context.Context, orgID, eventID uuid.UUID) ([]FormResponse, error) {
	if _, err := a.events.GetContext(ctx, orgID, eventID); err != nil {
		return nil, err
	}
	rows, err := a.ListForms(ctx, orgID, eventID)
	if err != nil {
		return nil, utils.ErrInternal()
	}
	resp := make([]FormResponse, len(rows))
	for i, row := range rows {
		resp[i] = toFormResponse(row)
	}
	return resp, nil
}

func (a *App) Update(ctx context.Context, orgID, eventID, id uuid.UUID, req UpdateFormRequest) (FormResponse, error) {
	form, err := a.UpdateForm(ctx, orgID, eventID, id, formsToPgInt4(req.Capacity), req.IsActive)
	if err != nil {
		return FormResponse{}, utils.ErrNotFound("form")
	}
	return toFormResponse(form), nil
}

func (a *App) Delete(ctx context.Context, orgID, eventID, id uuid.UUID) error {
	deleted, err := a.DeleteForm(ctx, orgID, eventID, id)
	if err != nil {
		return utils.ErrInternal()
	}
	if !deleted {
		return utils.ErrNotFound("form")
	}
	return nil
}

func (a *App) GetPublic(ctx context.Context, token string) (PublicFormResponse, error) {
	row, err := a.GetPublicByToken(ctx, token)
	if err != nil {
		return PublicFormResponse{}, utils.ErrNotFound("form")
	}

	var subEventName *string
	if row.SubEventName.Valid {
		subEventName = &row.SubEventName.String
	}
	isOpen := row.IsActive && (!row.Capacity.Valid || row.SubmissionsCount < row.Capacity.Int32)

	return PublicFormResponse{EventName: row.EventName, SubEventName: subEventName, IsOpen: isOpen}, nil
}

func (a *App) Submit(ctx context.Context, token string, req SubmitFormRequest) error {
	row, err := a.GetPublicByToken(ctx, token)
	if err != nil {
		return utils.ErrNotFound("form")
	}

	if _, err := a.IncrementSubmissions(ctx, row.ID); err != nil {
		return utils.NewError(http.StatusConflict, "form_closed", "this form is no longer accepting submissions")
	}

	var subEventID *uuid.UUID
	if row.SubEventID.Valid {
		id := uuid.UUID(row.SubEventID.Bytes)
		subEventID = &id
	}

	_, inserted, err := a.people.Submit(ctx, row.OrganizationID, row.EventID, people.PeopleSubmitInput{
		Email: req.Email, Mobile: req.Mobile, Name: req.Name,
		ImageURL: req.ImageURL, Age: req.Age, Gender: req.Gender,
		SubEventID: subEventID,
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

func formsToPgInt4(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}

func toFormResponse(f dbgen.EventForm) FormResponse {
	var subEventID *uuid.UUID
	if f.SubEventID.Valid {
		id := uuid.UUID(f.SubEventID.Bytes)
		subEventID = &id
	}
	var capacity *int32
	if f.Capacity.Valid {
		capacity = &f.Capacity.Int32
	}
	return FormResponse{
		ID: f.ID, Token: f.Token, SubEventID: subEventID,
		Capacity: capacity, SubmissionsCount: f.SubmissionsCount, IsActive: f.IsActive,
	}
}
