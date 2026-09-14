package events

import (
    "time"
)

// Event represents an event record with display metadata.
type Event struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    StartDate string    `json:"start_date"`
    EndDate   string    `json:"end_date"`
    Venue     *string   `json:"venue,omitempty"`
    Logo      *string   `json:"logo,omitempty"`
    Organizer *string   `json:"organizer,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// CreateEventRequest payload for creating a new event.
type CreateEventRequest struct {
    Name      string `json:"name" validate:"required"`
    StartDate string `json:"start_date" validate:"required"`
    EndDate   string `json:"end_date" validate:"required"`
}

// UpdateEventRequest payload for modifying an existing event.
type UpdateEventRequest struct {
    Name      *string `json:"name"`
    StartDate *string `json:"start_date"`
    EndDate   *string `json:"end_date"`
    Venue     *string `json:"venue"`
    Logo      *string `json:"logo"`
    Organizer *string `json:"organizer"`
}
