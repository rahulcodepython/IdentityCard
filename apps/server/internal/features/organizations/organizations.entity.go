package organizations

import "github.com/google/uuid"

type UpdateOrganizationSettingsRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type OrganizationSettingsResponse struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Slug    string    `json:"slug"`
	HasLogo bool      `json:"has_logo"`
}
