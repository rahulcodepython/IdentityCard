package devices

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
	"identitycard-server/internal/httpx"
	"identitycard-server/internal/modules/organizations"
)

const otpTTL = 10 * time.Minute

type Service struct {
	repo *Repository
	orgs *organizations.Service
}

func NewService(repo *Repository, orgsService *organizations.Service) *Service {
	return &Service{repo: repo, orgs: orgsService}
}

func (s *Service) Create(ctx context.Context, orgID uuid.UUID, name string) (CreateDeviceResponse, error) {
	otp, err := generateOTP()
	if err != nil {
		return CreateDeviceResponse{}, httpx.ErrInternal()
	}
	expiresAt := time.Now().Add(otpTTL)

	device, err := s.repo.Create(ctx, orgID, name, otp, expiresAt)
	if err != nil {
		return CreateDeviceResponse{}, httpx.ErrInternal()
	}

	return CreateDeviceResponse{
		DeviceResponse: toDeviceResponse(device),
		OTPCode:        otp,
		OTPExpiresAt:   expiresAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) List(ctx context.Context, orgID uuid.UUID) ([]DeviceResponse, error) {
	rows, err := s.repo.List(ctx, orgID)
	if err != nil {
		return nil, httpx.ErrInternal()
	}
	resp := make([]DeviceResponse, len(rows))
	for i, row := range rows {
		resp[i] = toDeviceResponse(row)
	}
	return resp, nil
}

func (s *Service) Revoke(ctx context.Context, orgID, id uuid.UUID) error {
	revoked, err := s.repo.Revoke(ctx, orgID, id)
	if err != nil {
		return httpx.ErrInternal()
	}
	if !revoked {
		return httpx.ErrNotFound("device")
	}
	return nil
}

// Pair consumes the OTP (it's only valid once — see the 'pending' status
// check baked into GetPendingByOTP) and mints a new device key, returned
// exactly once. There's no "reveal key" endpoint after this; a lost key
// means re-pairing.
func (s *Service) Pair(ctx context.Context, otpCode string) (PairDeviceResponse, error) {
	device, err := s.repo.GetPendingByOTP(ctx, otpCode)
	if err != nil {
		return PairDeviceResponse{}, httpx.NewError(http.StatusBadRequest, "invalid_otp", "code is invalid or expired")
	}

	rawKey, keyHash, err := generateDeviceKey()
	if err != nil {
		return PairDeviceResponse{}, httpx.ErrInternal()
	}
	if _, err := s.repo.MarkVerified(ctx, device.ID, keyHash); err != nil {
		return PairDeviceResponse{}, httpx.ErrInternal()
	}

	org, err := s.orgs.GetSettings(ctx, device.OrganizationID)
	if err != nil {
		return PairDeviceResponse{}, httpx.ErrInternal()
	}

	return PairDeviceResponse{DeviceID: device.ID, OrganizationName: org.Name, Key: rawKey}, nil
}

func (s *Service) Me(ctx context.Context, orgID, deviceID uuid.UUID) (MeResponse, error) {
	device, err := s.repo.Get(ctx, orgID, deviceID)
	if err != nil {
		return MeResponse{}, httpx.ErrNotFound("device")
	}
	org, err := s.orgs.GetSettings(ctx, orgID)
	if err != nil {
		return MeResponse{}, httpx.ErrInternal()
	}
	return MeResponse{DeviceID: device.ID, DeviceName: device.Name, OrganizationName: org.Name}, nil
}

// Claims is what RequireDevice attaches to the request context — the
// device-auth equivalent of auth.AccessClaims.
type Claims struct {
	DeviceID       uuid.UUID
	OrganizationID uuid.UUID
}

func (s *Service) Authenticate(ctx context.Context, rawKey string) (Claims, error) {
	if rawKey == "" {
		return Claims{}, httpx.ErrUnauthorized("")
	}
	device, err := s.repo.GetByKeyHash(ctx, hashKey(rawKey))
	if err != nil {
		return Claims{}, httpx.ErrUnauthorized("")
	}
	return Claims{DeviceID: device.ID, OrganizationID: device.OrganizationID}, nil
}

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
	return raw, hashKey(raw), nil
}

func hashKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func toDeviceResponse(d dbgen.Device) DeviceResponse {
	var verifiedAt *string
	if d.VerifiedAt.Valid {
		s := d.VerifiedAt.Time.Format(time.RFC3339)
		verifiedAt = &s
	}
	return DeviceResponse{
		ID: d.ID, Name: d.Name, Status: d.Status,
		CreatedAt: d.CreatedAt.Time.Format(time.RFC3339), VerifiedAt: verifiedAt,
	}
}
