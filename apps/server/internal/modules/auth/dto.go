package auth

import "github.com/google/uuid"

// RegisterRequest creates a brand-new organization and its first,
// unverified user. No plan is chosen and no subscription created here —
// see internal/modules/plans, done later from the dashboard. The account
// can't sign in until VerifyOTP or VerifyTOTP succeeds.
type RegisterRequest struct {
	Name             string `json:"name" validate:"required,min=2,max=120"`
	Email            string `json:"email" validate:"required,email"`
	OrganizationName string `json:"organization_name" validate:"required,min=2,max=120"`
}

// RegisterResponse carries everything the verification screen needs to
// enroll TOTP right away — this is the only moment the secret is ever
// shown in the clear (see Service.Register).
type RegisterResponse struct {
	Email       string `json:"email"`
	TOTPQRImage string `json:"totp_qr_image"` // data:image/png;base64,...
	TOTPSecret  string `json:"totp_secret"`   // manual-entry fallback
}

type SendOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6,numeric"`
}

type VerifyTOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6,numeric"`
}

type CreateOrganizationRequest struct {
	OrganizationName string `json:"organization_name" validate:"required,min=2,max=120"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

// MeResponse's organization fields are nil until HasOrganization is true —
// a Google signup lands with a valid session but no org yet, pending
// /onboarding (see Service.Me).
type MeResponse struct {
	UserID           uuid.UUID  `json:"user_id"`
	Email            string     `json:"email"`
	Name             string     `json:"name"`
	HasOrganization  bool       `json:"has_organization"`
	OrganizationID   *uuid.UUID `json:"organization_id,omitempty"`
	OrganizationName *string    `json:"organization_name,omitempty"`
	Roles            []string   `json:"roles,omitempty"`
}
