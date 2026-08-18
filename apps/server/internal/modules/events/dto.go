package events

import "github.com/google/uuid"

// EventDayInput is both the create/update wire shape and, unqualified, the
// same shape subevents.CreateSubEventRequest's days use — see
// internal/timeutil for the "2006-01-02" / "15:04" formats.
type EventDayInput struct {
	Date      string `json:"date" validate:"required,datetime=2006-01-02"`
	EntryTime string `json:"entry_time" validate:"required,datetime=15:04"`
	ExitTime  string `json:"exit_time" validate:"required,datetime=15:04"`
}

type CreateEventRequest struct {
	Name  string          `json:"name" validate:"required,min=2,max=200"`
	Kind  string          `json:"kind" validate:"required,oneof=established flash"`
	Venue string          `json:"venue" validate:"omitempty,max=300"`
	Days  []EventDayInput `json:"days" validate:"required,min=1,dive"`
}

// UpdateEventRequest replaces the event's name, venue, and full day list —
// kind can't change after creation (switching flash/established mid-flight
// has no sane semantics) and only draft events are editable at all.
type UpdateEventRequest struct {
	Name  string          `json:"name" validate:"required,min=2,max=200"`
	Venue string          `json:"venue" validate:"omitempty,max=300"`
	Days  []EventDayInput `json:"days" validate:"required,min=1,dive"`
}

type EventDayResponse struct {
	Date      string `json:"date"`
	EntryTime string `json:"entry_time"`
	ExitTime  string `json:"exit_time"`
}

// EventSummary is the list-view shape (no days — fetch the detail
// endpoint for those).
type EventSummary struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Kind        string    `json:"kind"`
	Status      string    `json:"status"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
	Venue       *string   `json:"venue"`
	PublishedAt *string   `json:"published_at"`
}

type EventResponse struct {
	EventSummary
	Days []EventDayResponse `json:"days"`
}
