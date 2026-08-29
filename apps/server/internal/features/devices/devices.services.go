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
	"identitycard-server/internal/utils"
)

const deviceOTPTTL = 10 * time.Minute

func (a *App) Create(ctx context.Context, orgID uuid.UUID, name string) (CreateDeviceResponse, error) {
	otp, err := generateOTP()
	if err != nil {
		return CreateDeviceResponse{}, utils.ErrInternal()
	}
	expiresAt := time.Now().Add(deviceOTPTTL)

	device, err := a.CreateDevice(ctx, orgID, name, otp, expiresAt)
	if err != nil {
		return CreateDeviceResponse{}, utils.ErrInternal()
	}

	return CreateDeviceResponse{
		DeviceResponse: toDeviceResponse(device),
		OTPCode:        otp,
		OTPExpiresAt:   expiresAt.Format(time.RFC3339),
	}, nil
}

func (a *App) List(ctx context.Context, orgID uuid.UUID) ([]DeviceResponse, error) {
	rows, err := a.ListDevices(ctx, orgID)
	if err != nil {
		return nil, utils.ErrInternal()
	}
	resp := make([]DeviceResponse, len(rows))
	for i, row := range rows {
		resp[i] = toDeviceResponse(row)
	}
	return resp, nil
}

func (a *App) Revoke(ctx context.Context, orgID, id uuid.UUID) error {
	revoked, err := a.RevokeDevice(ctx, orgID, id)
	if err != nil {
		return utils.ErrInternal()
	}
	if !revoked {
		return utils.ErrNotFound("device")
	}
	return nil
}

func (a *App) Pair(ctx context.Context, otpCode string) (PairDeviceResponse, error) {
	device, err := a.GetPendingDeviceByOTP(ctx, otpCode)
	if err != nil {
		return PairDeviceResponse{}, utils.NewError(http.StatusBadRequest, "invalid_otp", "code is invalid or expired")
	}

	rawKey, keyHash, err := generateDeviceKey()
	if err != nil {
		return PairDeviceResponse{}, utils.ErrInternal()
	}
	if _, err := a.MarkDeviceVerified(ctx, device.ID, keyHash); err != nil {
		return PairDeviceResponse{}, utils.ErrInternal()
	}

	org, err := a.orgs.GetSettings(ctx, device.OrganizationID)
	if err != nil {
		return PairDeviceResponse{}, utils.ErrInternal()
	}

	return PairDeviceResponse{DeviceID: device.ID, OrganizationName: org.Name, Key: rawKey}, nil
}

func (a *App) Me(ctx context.Context, orgID, deviceID uuid.UUID) (DeviceMeResponse, error) {
	device, err := a.GetDevice(ctx, orgID, deviceID)
	if err != nil {
		return DeviceMeResponse{}, utils.ErrNotFound("device")
	}
	org, err := a.orgs.GetSettings(ctx, orgID)
	if err != nil {
		return DeviceMeResponse{}, utils.ErrInternal()
	}
	return DeviceMeResponse{DeviceID: device.ID, DeviceName: device.Name, OrganizationName: org.Name}, nil
}

func (a *App) Authenticate(ctx context.Context, rawKey string) (DeviceClaims, error) {
	if rawKey == "" {
		return DeviceClaims{}, utils.ErrUnauthorized("")
	}
	device, err := a.GetDeviceByKeyHash(ctx, hashDeviceKey(rawKey))
	if err != nil {
		return DeviceClaims{}, utils.ErrUnauthorized("")
	}
	return DeviceClaims{DeviceID: device.ID, OrganizationID: device.OrganizationID}, nil
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

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
