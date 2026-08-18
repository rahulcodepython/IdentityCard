package subevents

import (
	"github.com/google/uuid"

	"identitycard-server/internal/modules/events"
)

// Days reuses events.EventDayInput/EventDayResponse — same wire shape,
// same date/time formats. An empty Days list means the sub-event runs on
// every day of the parent event, using the parent's own times.

type CreateSubEventRequest struct {
	Name string                 `json:"name" validate:"required,min=2,max=200"`
	Days []events.EventDayInput `json:"days" validate:"omitempty,dive"`
}

type UpdateSubEventRequest struct {
	Name string                 `json:"name" validate:"required,min=2,max=200"`
	Days []events.EventDayInput `json:"days" validate:"omitempty,dive"`
}

type SubEventResponse struct {
	ID   uuid.UUID                 `json:"id"`
	Name string                    `json:"name"`
	Days []events.EventDayResponse `json:"days"`
}
