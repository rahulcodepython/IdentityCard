package subevents

import "github.com/google/uuid"

type CreateSubEventRequest struct {
	Name      string `json:"name" validate:"required,min=2,max=200"`
	Date      string `json:"date" validate:"required,datetime=2006-01-02"`
	EntryTime string `json:"entry_time" validate:"required,datetime=15:04"`
	ExitTime  string `json:"exit_time" validate:"required,datetime=15:04"`
}

type UpdateSubEventRequest struct {
	Name      string `json:"name" validate:"required,min=2,max=200"`
	Date      string `json:"date" validate:"required,datetime=2006-01-02"`
	EntryTime string `json:"entry_time" validate:"required,datetime=15:04"`
	ExitTime  string `json:"exit_time" validate:"required,datetime=15:04"`
}

type SubEventResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Date      string    `json:"date"`
	EntryTime string    `json:"entry_time"`
	ExitTime  string    `json:"exit_time"`
}
