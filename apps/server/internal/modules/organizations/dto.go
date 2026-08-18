package organizations

import "github.com/google/uuid"

type SettingsResponse struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Slug    string    `json:"slug"`
	HasLogo bool      `json:"has_logo"` // fetch the image itself from GET /organizations/logo
}
