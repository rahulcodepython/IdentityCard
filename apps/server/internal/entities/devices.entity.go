package entities

import "github.com/google/uuid"

type CreateDeviceRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type DeviceResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	CreatedAt  string    `json:"created_at"`
	VerifiedAt *string   `json:"verified_at"`
}

// CreateDeviceResponse is the only place the OTP is ever exposed — shown
// once to the admin who created the device, to relay by voice/screen to
// whoever is standing at the physical device.
type CreateDeviceResponse struct {
	DeviceResponse
	OTPCode      string `json:"otp_code"`
	OTPExpiresAt string `json:"otp_expires_at"`
}

type PairDeviceRequest struct {
	OTPCode string `json:"otp_code" validate:"required,len=6,numeric"`
}

// PairDeviceResponse is the only place the device key is ever exposed —
// the pairing page stores it in the browser's localStorage and never
// fetches it again; there is no "view key" endpoint.
type PairDeviceResponse struct {
	DeviceID         uuid.UUID `json:"device_id"`
	OrganizationName string    `json:"organization_name"`
	Key              string    `json:"key"`
}

// DeviceMeResponse lets the scanner UI show "connected to <org>" without
// exposing anything beyond what the device itself already represents.
// Domain-prefixed since MeResponse already exists for the session-auth
// side (auth.entity.go).
type DeviceMeResponse struct {
	DeviceID         uuid.UUID `json:"device_id"`
	DeviceName       string    `json:"device_name"`
	OrganizationName string    `json:"organization_name"`
}
