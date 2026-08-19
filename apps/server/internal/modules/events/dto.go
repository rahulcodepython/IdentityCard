package events

import "github.com/google/uuid"

// EventDayInput is both the wire shape for a single explicit day (flash and
// selective modes) and, unqualified, the same shape subevents' selective
// mode uses — see internal/timeutil for the "2006-01-02" / "15:04" formats.
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

// RecurrenceWeekdayInput is one active weekday of a recurring event's
// rule, with its own time window — a Friday can close earlier than
// Mon-Thu, so there's deliberately no single shared time for the whole
// week (see the "all the possible edge cases" requirement).
type RecurrenceWeekdayInput struct {
	Weekday   int    `json:"weekday" validate:"min=0,max=6"` // 0=Monday..6=Sunday
	EntryTime string `json:"entry_time" validate:"required,datetime=15:04"`
	ExitTime  string `json:"exit_time" validate:"required,datetime=15:04"`
}

// RecurrenceInput authors a schedule_mode='recurring' event: active on
// StartsOn and every date after it that matches one of Weekdays, up to
// EndsOn — or indefinitely if EndsOn is nil ("open-ended", confirmed).
type RecurrenceInput struct {
	StartsOn string                   `json:"starts_on" validate:"required,datetime=2006-01-02"`
	EndsOn   *string                  `json:"ends_on" validate:"omitempty,datetime=2006-01-02"`
	Weekdays []RecurrenceWeekdayInput `json:"weekdays" validate:"required,min=1,dive"`
}

// CreateEventRequest is discriminated by ScheduleMode — which of the other
// fields are required depends on it (validated further in Service.Create,
// since a cross-field rule like "flash requires exactly one Days entry"
// isn't cleanly expressible as a struct tag alone):
//   - flash:       Days (exactly one entry)
//   - fixed_range: RangeStart/RangeEnd/RangeEntryTime/RangeExitTime, optional ExcludedDates
//   - selective:   Days (one or more entries, each its own time)
//   - recurring:   Recurrence, optional ExcludedDates
type CreateEventRequest struct {
	Name           string           `json:"name" validate:"required,min=2,max=200"`
	ScheduleMode   string           `json:"schedule_mode" validate:"required,oneof=flash fixed_range selective recurring"`
	Venue          string           `json:"venue" validate:"omitempty,max=300"`
	Days           []EventDayInput  `json:"days" validate:"omitempty,dive"`
	RangeStart     string           `json:"range_start" validate:"omitempty,datetime=2006-01-02"`
	RangeEnd       string           `json:"range_end" validate:"omitempty,datetime=2006-01-02"`
	RangeEntryTime string           `json:"range_entry_time" validate:"omitempty,datetime=15:04"`
	RangeExitTime  string           `json:"range_exit_time" validate:"omitempty,datetime=15:04"`
	ExcludedDates  []string         `json:"excluded_dates" validate:"omitempty,dive,datetime=2006-01-02"`
	Recurrence     *RecurrenceInput `json:"recurrence" validate:"omitempty"`
}

// UpdateEventRequest carries the same schedule_mode-discriminated fields as
// CreateEventRequest, except ScheduleMode itself — like kind before it,
// the schedule mode can't change after creation (only draft events are
// editable at all; see Service.Update).
type UpdateEventRequest struct {
	Name           string           `json:"name" validate:"required,min=2,max=200"`
	Venue          string           `json:"venue" validate:"omitempty,max=300"`
	Days           []EventDayInput  `json:"days" validate:"omitempty,dive"`
	RangeStart     string           `json:"range_start" validate:"omitempty,datetime=2006-01-02"`
	RangeEnd       string           `json:"range_end" validate:"omitempty,datetime=2006-01-02"`
	RangeEntryTime string           `json:"range_entry_time" validate:"omitempty,datetime=15:04"`
	RangeExitTime  string           `json:"range_exit_time" validate:"omitempty,datetime=15:04"`
	ExcludedDates  []string         `json:"excluded_dates" validate:"omitempty,dive,datetime=2006-01-02"`
	Recurrence     *RecurrenceInput `json:"recurrence" validate:"omitempty"`
}

// AddExcludedDateRequest is the one action allowed on a recurring event
// after publish (see Service.Publish's doc comment) — adding an
// unplanned holiday, never removing one.
type AddExcludedDateRequest struct {
	Date string `json:"date" validate:"required,datetime=2006-01-02"`
}

type RecurrenceWeekdayResponse struct {
	Weekday   int    `json:"weekday"`
	EntryTime string `json:"entry_time"`
	ExitTime  string `json:"exit_time"`
}

type RecurrenceResponse struct {
	StartsOn string                      `json:"starts_on"`
	EndsOn   *string                     `json:"ends_on"`
	Weekdays []RecurrenceWeekdayResponse `json:"weekdays"`
}

// EventSummary is the list-view shape (no days — fetch the detail
// endpoint for those).
type EventSummary struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	ScheduleMode string    `json:"schedule_mode"`
	Status       string    `json:"status"`
	StartDate    string    `json:"start_date"`
	EndDate      *string   `json:"end_date"` // nil for an open-ended recurring event
	Venue        *string   `json:"venue"`
	PublishedAt  *string   `json:"published_at"`
}

type EventResponse struct {
	EventSummary
	Days          []EventDayResponse  `json:"days"`
	Recurrence    *RecurrenceResponse `json:"recurrence,omitempty"`
	ExcludedDates []string            `json:"excluded_dates,omitempty"`
}

// DayImportRowError and DayImportSummary mirror people.ImportRowError /
// people.ImportSummary — same per-row-error CSV import shape, applied
// here to day/exclusion-date rows instead of people rows.
type DayImportRowError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

type DayImportSummary struct {
	Imported int                 `json:"imported"`
	Skipped  int                 `json:"skipped"`
	Errors   []DayImportRowError `json:"errors"`
}
