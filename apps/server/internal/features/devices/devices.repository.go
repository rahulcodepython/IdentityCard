package devices

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

func (a *App) CreateDevice(ctx context.Context, orgID uuid.UUID, name, otpCode string, otpExpiresAt time.Time) (dbgen.Device, error) {
	return a.queries.CreateDevice(ctx, dbgen.CreateDeviceParams{
		OrganizationID: orgID,
		Name:           name,
		OtpCode:        pgtype.Text{String: otpCode, Valid: true},
		OtpExpiresAt:   pgtype.Timestamptz{Time: otpExpiresAt, Valid: true},
	})
}

func (a *App) ListDevices(ctx context.Context, orgID uuid.UUID) ([]dbgen.Device, error) {
	return a.queries.ListDevices(ctx, orgID)
}

func (a *App) GetDevice(ctx context.Context, orgID, id uuid.UUID) (dbgen.Device, error) {
	return a.queries.GetDevice(ctx, dbgen.GetDeviceParams{ID: id, OrganizationID: orgID})
}

func (a *App) GetPendingDeviceByOTP(ctx context.Context, otpCode string) (dbgen.Device, error) {
	return a.queries.GetPendingDeviceByOTP(ctx, pgtype.Text{String: otpCode, Valid: true})
}

func (a *App) MarkDeviceVerified(ctx context.Context, id uuid.UUID, keyHash string) (dbgen.Device, error) {
	return a.queries.MarkDeviceVerified(ctx, dbgen.MarkDeviceVerifiedParams{
		ID:      id,
		KeyHash: pgtype.Text{String: keyHash, Valid: true},
	})
}

func (a *App) GetDeviceByKeyHash(ctx context.Context, keyHash string) (dbgen.Device, error) {
	return a.queries.GetDeviceByKeyHash(ctx, pgtype.Text{String: keyHash, Valid: true})
}

func (a *App) RevokeDevice(ctx context.Context, orgID, id uuid.UUID) (bool, error) {
	n, err := a.queries.RevokeDevice(ctx, dbgen.RevokeDeviceParams{ID: id, OrganizationID: orgID})
	return n > 0, err
}
