// Package repositories: devices owns scanner-bot pairing: a device is
// created by a super_admin (with a short-lived OTP shown to a human, who
// relays it to the physical device), and once paired authenticates every
// subsequent request with an opaque key — never a session cookie, since a
// device never logs in as an org member.
package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

type DevicesRepository struct {
	q *dbgen.Queries
}

func NewDevicesRepository(q *dbgen.Queries) *DevicesRepository {
	return &DevicesRepository{q: q}
}

func (r *DevicesRepository) Create(ctx context.Context, orgID uuid.UUID, name, otpCode string, otpExpiresAt time.Time) (dbgen.Device, error) {
	return r.q.CreateDevice(ctx, dbgen.CreateDeviceParams{
		OrganizationID: orgID,
		Name:           name,
		OtpCode:        pgtype.Text{String: otpCode, Valid: true},
		OtpExpiresAt:   pgtype.Timestamptz{Time: otpExpiresAt, Valid: true},
	})
}

func (r *DevicesRepository) List(ctx context.Context, orgID uuid.UUID) ([]dbgen.Device, error) {
	return r.q.ListDevices(ctx, orgID)
}

func (r *DevicesRepository) Get(ctx context.Context, orgID, id uuid.UUID) (dbgen.Device, error) {
	return r.q.GetDevice(ctx, dbgen.GetDeviceParams{ID: id, OrganizationID: orgID})
}

func (r *DevicesRepository) GetPendingByOTP(ctx context.Context, otpCode string) (dbgen.Device, error) {
	return r.q.GetPendingDeviceByOTP(ctx, pgtype.Text{String: otpCode, Valid: true})
}

func (r *DevicesRepository) MarkVerified(ctx context.Context, id uuid.UUID, keyHash string) (dbgen.Device, error) {
	return r.q.MarkDeviceVerified(ctx, dbgen.MarkDeviceVerifiedParams{
		ID:      id,
		KeyHash: pgtype.Text{String: keyHash, Valid: true},
	})
}

func (r *DevicesRepository) GetByKeyHash(ctx context.Context, keyHash string) (dbgen.Device, error) {
	return r.q.GetDeviceByKeyHash(ctx, pgtype.Text{String: keyHash, Valid: true})
}

// Revoke reports whether a row was actually changed.
func (r *DevicesRepository) Revoke(ctx context.Context, orgID, id uuid.UUID) (bool, error) {
	n, err := r.q.RevokeDevice(ctx, dbgen.RevokeDeviceParams{ID: id, OrganizationID: orgID})
	return n > 0, err
}
