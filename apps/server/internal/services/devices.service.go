package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/google/uuid"

	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/entities"
	"identitycard-server/internal/repositories"
	"identitycard-server/internal/utils"
)

const deviceOTPTTL = 10 * time.Minute

type DevicesService struct {
	repo *repositories.DevicesRepository
	orgs *OrganizationsService
}

func NewDevicesService(repo *repositories.DevicesRepository, orgsService *OrganizationsService) *DevicesService {
	return &DevicesService{repo: repo, orgs: orgsService}
}

func (s *DevicesService) Create(ctx context.Context, orgID uuid.UUID, name string) (entities.CreateDeviceResponse, error) {
	otp, err := generateOTP()
	if err != nil {
		return entities.CreateDeviceResponse{}, utils.ErrInternal()
	}
	expiresAt := time.Now().Add(deviceOTPTTL)

	device, err := s.repo.Create(ctx, orgID, name, otp, expiresAt)
	if err != nil {
		return entities.CreateDeviceResponse{}, utils.ErrInternal()
	}

	return entities.CreateDeviceResponse{
		DeviceResponse: toDeviceResponse(device),
		OTPCode:        otp,
		OTPExpiresAt:   expiresAt.Format(time.RFC3339),
	}, nil
}

func (s *DevicesService) List(ctx context.Context, orgID uuid.UUID) ([]entities.DeviceResponse, error) {
	rows, err := s.repo.List(ctx, orgID)
	if err != nil {
		return nil, utils.ErrInternal()
	}
	resp := make([]entities.DeviceResponse, len(rows))
	for i, row := range rows {
		resp[i] = toDeviceResponse(row)
	}
	return resp, nil
}

func (s *DevicesService) Revoke(ctx context.Context, orgID, id uuid.UUID) error {
	revoked, err := s.repo.Revoke(ctx, orgID, id)
	if err != nil {
		return utils.ErrInternal()
	}
	if !revoked {
		return utils.ErrNotFound("device")
	}
	return nil
}

// Pair consumes the OTP (it's only valid once — see the 'pending' status
// check baked into GetPendingByOTP) and mints a new device key, returned
// exactly once. There's no "reveal key" endpoint after this; a lost key
// means re-pairing.
func (s *DevicesService) Pair(ctx context.Context, otpCode string) (entities.PairDeviceResponse, error) {
	device, err := s.repo.GetPendingByOTP(ctx, otpCode)
	if err != nil {
		return entities.PairDeviceResponse{}, utils.NewError(http.StatusBadRequest, "invalid_otp", "code is invalid or expired")
	}

	rawKey, keyHash, err := generateDeviceKey()
	if err != nil {
		return entities.PairDeviceResponse{}, utils.ErrInternal()
	}
	if _, err := s.repo.MarkVerified(ctx, device.ID, keyHash); err != nil {
		return entities.PairDeviceResponse{}, utils.ErrInternal()
	}

	org, err := s.orgs.GetSettings(ctx, device.OrganizationID)
	if err != nil {
		return entities.PairDeviceResponse{}, utils.ErrInternal()
	}

	return entities.PairDeviceResponse{DeviceID: device.ID, OrganizationName: org.Name, Key: rawKey}, nil
}

func (s *DevicesService) Me(ctx context.Context, orgID, deviceID uuid.UUID) (entities.DeviceMeResponse, error) {
	device, err := s.repo.Get(ctx, orgID, deviceID)
	if err != nil {
		return entities.DeviceMeResponse{}, utils.ErrNotFound("device")
	}
	org, err := s.orgs.GetSettings(ctx, orgID)
	if err != nil {
		return entities.DeviceMeResponse{}, utils.ErrInternal()
	}
	return entities.DeviceMeResponse{DeviceID: device.ID, DeviceName: device.Name, OrganizationName: org.Name}, nil
}

// DeviceClaims is what RequireDevice attaches to the request context — the
// device-auth equivalent of pkg/jwt.AccessClaims.
type DeviceClaims struct {
	DeviceID       uuid.UUID
	OrganizationID uuid.UUID
}

func (s *DevicesService) Authenticate(ctx context.Context, rawKey string) (DeviceClaims, error) {
	if rawKey == "" {
		return DeviceClaims{}, utils.ErrUnauthorized("")
	}
	device, err := s.repo.GetByKeyHash(ctx, hashDeviceKey(rawKey))
	if err != nil {
		return DeviceClaims{}, utils.ErrUnauthorized("")
	}
	return DeviceClaims{DeviceID: device.ID, OrganizationID: device.OrganizationID}, nil
}

// generateOTP returns a zero-padded 6-digit numeric code — same shape as
// the auth-side email OTP (see auth.otp.go's generateAuthOTP, renamed
// specifically to avoid colliding with this one in the shared package).
func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// generateDeviceKey returns a high-entropy random secret alongside its
// SHA-256 hash — a fast hash is the right tool here (unlike a human
// password) because the key already has 256 bits of entropy; there's no
// brute-force risk bcrypt's slowness would meaningfully defend against.
func generateDeviceKey() (raw, hash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(buf)
	return raw, hashDeviceKey(raw), nil
}

func hashDeviceKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func toDeviceResponse(d dbgen.Device) entities.DeviceResponse {
	var verifiedAt *string
	if d.VerifiedAt.Valid {
		s := d.VerifiedAt.Time.Format(time.RFC3339)
		verifiedAt = &s
	}
	return entities.DeviceResponse{
		ID: d.ID, Name: d.Name, Status: d.Status,
		CreatedAt: d.CreatedAt.Time.Format(time.RFC3339), VerifiedAt: verifiedAt,
	}
}
