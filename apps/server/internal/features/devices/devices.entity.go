package devices

import (
    "time"

    "github.com/google/uuid"
)

type DeviceDB struct {
    ID             uuid.UUID  `json:"id"`
    OrganizationID uuid.UUID  `json:"organization_id"`
    Name           string     `json:"name"`
    KeyHash        *string    `json:"key_hash"`
    OtpCode        *string    `json:"otp_code"`
    OtpExpiresAt   *time.Time `json:"otp_expires_at"`
    Status         string     `json:"status"`
    VerifiedAt     *time.Time `json:"verified_at"`
    CreatedAt      time.Time  `json:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at"`
}

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

type CreateDeviceResponse struct {
    DeviceResponse
    OTPCode      string `json:"otp_code"`
    OTPExpiresAt string `json:"otp_expires_at"`
}

type PairDeviceRequest struct {
    OTPCode string `json:"otp_code" validate:"required,len=6,numeric"`
}

type PairDeviceResponse struct {
    DeviceID         uuid.UUID `json:"device_id"`
    OrganizationName string    `json:"organization_name"`
    Key              string    `json:"key"`
}

type DeviceMeResponse struct {
    DeviceID         uuid.UUID `json:"device_id"`
    DeviceName       string    `json:"device_name"`
    OrganizationName string    `json:"organization_name"`
}

type DeviceClaims struct {
    DeviceID       uuid.UUID
    OrganizationID uuid.UUID
}
