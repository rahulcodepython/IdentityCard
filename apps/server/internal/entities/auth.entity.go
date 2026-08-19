package entities

import "github.com/google/uuid"

// RegisterRequest creates a brand-new organization and its first,
// unverified user. No plan is chosen and no subscription created here —
// see the plans domain, done later from the dashboard. The account can't
// sign in until VerifyOTP or VerifyTOTP succeeds.
type RegisterRequest struct {
	Name             string `json:"name" validate:"required,min=2,max=120"`
	Email            string `json:"email" validate:"required,email"`
	OrganizationName string `json:"organization_name" validate:"required,min=2,max=120"`
}

// RegisterResponse carries everything the verification screen needs to
// enroll TOTP right away — this is the only moment the secret is ever
// shown in the clear (see AuthService.Register).
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

// MeResponse's organization fields are nil until HasOrganization is true —
// a Google signup lands with a valid session but no org yet, pending
// /onboarding (see AuthService.Me).
type MeResponse struct {
	UserID           uuid.UUID  `json:"user_id"`
	Email            string     `json:"email"`
	Name             string     `json:"name"`
	HasOrganization  bool       `json:"has_organization"`
	OrganizationID   *uuid.UUID `json:"organization_id,omitempty"`
	OrganizationName *string    `json:"organization_name,omitempty"`
	Roles            []string   `json:"roles,omitempty"`
}

// TokenPairResponse is returned by every session-issuing endpoint. Go is a
// pure bearer-token API — cookie issuance/storage is entirely the Next.js
// auth-bridge routes' job (apps/web/app/api/auth/**).
type TokenPairResponse struct {
	AccessToken            string `json:"access_token"`
	RefreshToken           string `json:"refresh_token"`
	AccessTokenTTLSeconds  int    `json:"access_ttl_seconds"`
	RefreshTokenTTLSeconds int    `json:"refresh_ttl_seconds"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type OAuthExchangeRequest struct {
	Code string `json:"code" validate:"required"`
}
