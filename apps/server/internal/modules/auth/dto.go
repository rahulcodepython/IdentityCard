package auth

import "github.com/google/uuid"

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest creates the first account for a brand-new organization:
// a user, the organization itself, a (stubbed — no real payment yet)
// subscription to the chosen plan, and a super_admin membership tying
// them together. See Service.Register.
type RegisterRequest struct {
	Name             string `json:"name" validate:"required,min=2,max=120"`
	Email            string `json:"email" validate:"required,email"`
	Password         string `json:"password" validate:"required,min=8,max=72"`
	OrganizationName string `json:"organization_name" validate:"required,min=2,max=120"`
	PlanCode         string `json:"plan_code" validate:"required"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type MeResponse struct {
	UserID           uuid.UUID `json:"user_id"`
	Email            string    `json:"email"`
	Name             string    `json:"name"`
	OrganizationID   uuid.UUID `json:"organization_id"`
	OrganizationName string    `json:"organization_name"`
	Roles            []string  `json:"roles"`
}
