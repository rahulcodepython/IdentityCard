package entities

import "github.com/google/uuid"

// A sub-event is always exactly one day — the parent (grouped) event is
// the date-range container; each sub-event is effectively a flash event
// living inside it. Date must fall within the parent's
// [start_date, end_date] (validated in SubEventsService, since it needs
// the parent's own dates to check against).

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
