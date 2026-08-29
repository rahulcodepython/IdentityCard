package entities

import "github.com/google/uuid"

// OrganizationSettingsResponse — named with the domain prefix since
// "SettingsResponse" alone is too generic for a flat, shared entities
// package (every domain that has its own settings screen would want the
// same short name).
type OrganizationSettingsResponse struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Slug    string    `json:"slug"`
	HasLogo bool      `json:"has_logo"` // fetch the image itself from GET /organizations/logo
}

type UpdateOrganizationSettingsRequest struct {
	Name string `json:"name" validate:"required,min=2,max=120"`
}
