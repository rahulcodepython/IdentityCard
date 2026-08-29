package events

import "github.com/google/uuid"

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

type CreateEventRequest struct {
	EventType     string          `json:"event_type" validate:"required,oneof=flash standard grouped"`
	Name          string          `json:"name" validate:"required,min=2,max=200"`
	Venue         string          `json:"venue" validate:"omitempty,max=300"`
	OrganizerName string          `json:"organizer_name" validate:"omitempty,max=200"`
	Days          []EventDayInput `json:"days" validate:"omitempty,dive"`
	RangeStart    string          `json:"range_start" validate:"omitempty,datetime=2006-01-02"`
	RangeEnd      string          `json:"range_end" validate:"omitempty,datetime=2006-01-02"`
}

type UpdateEventRequest struct {
	Name          string          `json:"name" validate:"required,min=2,max=200"`
	Venue         string          `json:"venue" validate:"omitempty,max=300"`
	OrganizerName string          `json:"organizer_name" validate:"omitempty,max=200"`
	Days          []EventDayInput `json:"days" validate:"omitempty,dive"`
	RangeStart    string          `json:"range_start" validate:"omitempty,datetime=2006-01-02"`
	RangeEnd      string          `json:"range_end" validate:"omitempty,datetime=2006-01-02"`
}

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
	Days []EventDayResponse `json:"days"`
}

type DayImportRowError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

type DayImportSummary struct {
	Imported int                 `json:"imported"`
	Skipped  int                 `json:"skipped"`
	Errors   []DayImportRowError `json:"errors"`
}
