package entities

import "github.com/google/uuid"

// email/mobile are the identity that keys the upsert (see
// repositories.PeopleRepository.Upsert) and can't be changed after the
// fact — fixing a typo means deleting and re-adding the person, an
// acceptable scope cut for now.

type CreatePersonRequest struct {
	Email       string      `json:"email" validate:"required,email"`
	Mobile      string      `json:"mobile" validate:"required,min=3,max=32"`
	Name        string      `json:"name" validate:"required,min=1,max=200"`
	ImageURL    string      `json:"image_url" validate:"omitempty,url,max=2048"`
	Age         *int16      `json:"age" validate:"omitempty,min=0,max=150"`
	Gender      string      `json:"gender" validate:"omitempty,max=50"`
	SubEventIDs []uuid.UUID `json:"sub_event_ids" validate:"omitempty,dive"`
}

type UpdatePersonRequest struct {
	Name        string      `json:"name" validate:"required,min=1,max=200"`
	ImageURL    string      `json:"image_url" validate:"omitempty,url,max=2048"`
	Age         *int16      `json:"age" validate:"omitempty,min=0,max=150"`
	Gender      string      `json:"gender" validate:"omitempty,max=50"`
	SubEventIDs []uuid.UUID `json:"sub_event_ids" validate:"omitempty,dive"`
}

type PersonResponse struct {
	ID          uuid.UUID   `json:"id"`
	Email       string      `json:"email"`
	Mobile      string      `json:"mobile"`
	Name        string      `json:"name"`
	ImageURL    *string     `json:"image_url"`
	Age         *int16      `json:"age"`
	Gender      *string     `json:"gender"`
	JoinedAt    *string     `json:"joined_at"`
	CardSentAt  *string     `json:"card_sent_at"`
	SubEventIDs []uuid.UUID `json:"sub_event_ids"`
}

// PeopleImportRowError/PeopleImportSummary are domain-prefixed since
// "import row error/summary" is a shape multiple domains (events, forms)
// each need their own copy of in this flat entities package.
type PeopleImportRowError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

type PeopleImportSummary struct {
	Inserted int                    `json:"inserted"`
	Updated  int                    `json:"updated"`
	Skipped  int                    `json:"skipped"`
	Errors   []PeopleImportRowError `json:"errors"`
}
