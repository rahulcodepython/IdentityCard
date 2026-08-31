package devices

import (
    "context"
    "time"

    "github.com/google/uuid"

    "identitycard-server/internal/pkg/postgres"
)

func (a *App) CreateDevice(ctx context.Context, orgID uuid.UUID, name, otpCode string, otpExpiresAt time.Time) (*DeviceDB, error) {
    return postgres.QueryJSON[DeviceDB](ctx, a.pool, CreateDeviceQuery, orgID, name, otpCode, otpExpiresAt)
}

func (a *App) ListDevices(ctx context.Context, orgID uuid.UUID) ([]DeviceDB, error) {
    return postgres.QueryJSONSlice[DeviceDB](ctx, a.pool, ListDevicesQuery, orgID)
}

func (a *App) GetDevice(ctx context.Context, orgID, id uuid.UUID) (*DeviceDB, error) {
    return postgres.QueryJSON[DeviceDB](ctx, a.pool, GetDeviceQuery, id, orgID)
}

func (a *App) GetPendingDeviceByOTP(ctx context.Context, otpCode string) (*DeviceDB, error) {
    return postgres.QueryJSON[DeviceDB](ctx, a.pool, GetPendingDeviceByOTPQuery, otpCode)
}

func (a *App) MarkDeviceVerified(ctx context.Context, id uuid.UUID, keyHash string) (*DeviceDB, error) {
    return postgres.QueryJSON[DeviceDB](ctx, a.pool, MarkDeviceVerifiedQuery, id, keyHash)
}

func (a *App) GetDeviceByKeyHash(ctx context.Context, keyHash string) (*DeviceDB, error) {
    return postgres.QueryJSON[DeviceDB](ctx, a.pool, GetDeviceByKeyHashQuery, keyHash)
}

func (a *App) RevokeDevice(ctx context.Context, orgID, id uuid.UUID) (bool, error) {
    res, err := a.pool.Exec(ctx, RevokeDeviceQuery, id, orgID)
    if err != nil {
        return false, postgres.MapPgError(err)
    }
    return res.RowsAffected() > 0, nil
}
