package forms

import "github.com/google/uuid"

type CreateFormRequest struct {
	SubEventID *uuid.UUID `json:"sub_event_id" validate:"omitempty"`
	Capacity   *int32     `json:"capacity" validate:"omitempty,min=1"`
}

type UpdateFormRequest struct {
	Capacity *int32 `json:"capacity" validate:"omitempty,min=1"`
	IsActive bool   `json:"is_active"`
}

type FormResponse struct {
	ID               uuid.UUID  `json:"id"`
	Token            string     `json:"token"`
	SubEventID       *uuid.UUID `json:"sub_event_id"`
	Capacity         *int32     `json:"capacity"`
	SubmissionsCount int32      `json:"submissions_count"`
	IsActive         bool       `json:"is_active"`
}

// PublicFormResponse and SubmitFormRequest are served with no
// authentication — token possession is the only access control, so
// neither carries anything beyond what a prospective registrant needs to
// see the form and submit it.

type PublicFormResponse struct {
	EventName    string  `json:"event_name"`
	SubEventName *string `json:"sub_event_name"`
	IsOpen       bool    `json:"is_open"`
}

type SubmitResponse struct {
	Message string `json:"message"`
}

type SubmitFormRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Mobile   string `json:"mobile" validate:"required,min=3,max=32"`
	Name     string `json:"name" validate:"required,min=1,max=200"`
	ImageURL string `json:"image_url" validate:"omitempty,url,max=2048"`
	Age      *int16 `json:"age" validate:"omitempty,min=0,max=150"`
	Gender   string `json:"gender" validate:"omitempty,max=50"`
}
