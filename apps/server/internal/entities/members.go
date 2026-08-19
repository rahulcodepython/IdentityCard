package entities

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationMemberResponse struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"userId"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Roles          []string  `json:"roles"`
	CreatedAt      time.Time `json:"createdAt"`
	Status         string    `json:"status"`         // MOCKED for now
	AssignedEvents []string  `json:"assignedEvents"` // MOCKED for now
}
