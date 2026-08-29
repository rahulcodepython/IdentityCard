package entities

import "github.com/google/uuid"

// EventDayInput is the wire shape for a single explicit day — used by
// flash (exactly one) and standard (one or more, manually picked) event
// types. See internal/utils/timeutil for the "2006-01-02" / "15:04"
// formats.
type EventDayInput struct {
	Date      string `json:"date" validate:"required,datetime=2006-01-02"`
	EntryTime string `json:"entry_time" validate:"required,datetime=15:04"`
	ExitTime  string `json:"exit_time" validate:"required,datetime=15:04"`
}

type EventDayResponse struct {
	Date      string `json:"date"`
	EntryTime string `json:"entry_time"`
	ExitTime  string `json:"exit_time"`
}

// CreateEventRequest is discriminated by EventType — which of the other
// fields are read depends on it (validated further in
// EventsService.Create, since a cross-field rule like "flash requires
// exactly one Days entry" isn't cleanly expressible as a struct tag
// alone):
//   - flash:    Days (exactly one entry) — start_date/end_date are the
//     same, derived from that one day.
//   - standard: Days (one or more entries, each its own time) — the
//     organizer's manually picked dates. start_date/end_date are the
//     min/max of them.
//   - grouped:  RangeStart/RangeEnd only — no Days; the parent event is a
//     date-range container, its own schedule comes entirely from the
//     sub-events created under it afterward.
type CreateEventRequest struct {
	EventType     string          `json:"event_type" validate:"required,oneof=flash standard grouped"`
	Name          string          `json:"name" validate:"required,min=2,max=200"`
	Venue         string          `json:"venue" validate:"omitempty,max=300"`
	OrganizerName string          `json:"organizer_name" validate:"omitempty,max=200"`
	Days          []EventDayInput `json:"days" validate:"omitempty,dive"`
	RangeStart    string          `json:"range_start" validate:"omitempty,datetime=2006-01-02"`
	RangeEnd      string          `json:"range_end" validate:"omitempty,datetime=2006-01-02"`
}

// UpdateEventRequest carries the same fields as CreateEventRequest minus
// EventType — like every other immutable attribute in this app, the
// event type can't change after creation (only a draft event is
// editable at all; see EventsService.Update).
type UpdateEventRequest struct {
	Name          string          `json:"name" validate:"required,min=2,max=200"`
	Venue         string          `json:"venue" validate:"omitempty,max=300"`
	OrganizerName string          `json:"organizer_name" validate:"omitempty,max=200"`
	Days          []EventDayInput `json:"days" validate:"omitempty,dive"`
	RangeStart    string          `json:"range_start" validate:"omitempty,datetime=2006-01-02"`
	RangeEnd      string          `json:"range_end" validate:"omitempty,datetime=2006-01-02"`
}

// EventSummary is the list-view shape (no days — fetch the detail
// endpoint for those). HasImage/HasOrganizerSignature mirror
// organizations' HasLogo pattern — fetch the bytes from
// GET /events/:id/image or /organizer-signature.
type EventSummary struct {
	ID                    uuid.UUID `json:"id"`
	Name                  string    `json:"name"`
	EventType             string    `json:"event_type"`
	Status                string    `json:"status"`
	StartDate             string    `json:"start_date"`
	EndDate               string    `json:"end_date"`
	Venue                 *string   `json:"venue"`
	OrganizerName         *string   `json:"organizer_name"`
	HasImage              bool      `json:"has_image"`
	HasOrganizerSignature bool      `json:"has_organizer_signature"`
	PublishedAt           *string   `json:"published_at"`
}

type EventResponse struct {
	EventSummary
	Days []EventDayResponse `json:"days"` // empty for a grouped event
}

// DayImportRowError and DayImportSummary mirror the people/forms per-row
// CSV import shape, applied here to a standard event's bulk day upload.
type DayImportRowError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

type DayImportSummary struct {
	Imported int                 `json:"imported"`
	Skipped  int                 `json:"skipped"`
	Errors   []DayImportRowError `json:"errors"`
}
