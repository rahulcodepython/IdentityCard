package attendance

import (
	"context"
	"time"

	"github.com/google/uuid"

	"identitycard-server/internal/pkg/postgres"
	"identitycard-server/internal/utils"
)

func (a *App) GetAttendanceRecord(ctx context.Context, personID uuid.UUID, date string) (*AttendanceRecordDB, error) {
	d, err := utils.ParseDate(date)
	if err != nil {
		return nil, err
	}
	return postgres.QueryJSON[AttendanceRecordDB](ctx, a.pool, GetAttendanceRecordQuery, personID, d.Time)
}

func (a *App) CreateAttendanceEntry(ctx context.Context, orgID, eventID, personID uuid.UUID, date string, at time.Time, status string, deviceID uuid.UUID) (*AttendanceRecordDB, error) {
	d, err := utils.ParseDate(date)
	if err != nil {
		return nil, err
	}
	return postgres.QueryJSON[AttendanceRecordDB](ctx, a.pool, CreateAttendanceEntryQuery, orgID, eventID, personID, d.Time, at, status, deviceID)
}

func (a *App) RecordAttendanceExit(ctx context.Context, personID uuid.UUID, date string, at time.Time, status string, deviceID uuid.UUID) (*AttendanceRecordDB, error) {
	d, err := utils.ParseDate(date)
	if err != nil {
		return nil, err
	}
	return postgres.QueryJSON[AttendanceRecordDB](ctx, a.pool, RecordAttendanceExitQuery, personID, d.Time, at, status, deviceID)
}

func (a *App) ListAttendanceForEvent(ctx context.Context, orgID, eventID uuid.UUID) ([]AttendanceRowDB, error) {
	return postgres.QueryJSONSlice[AttendanceRowDB](ctx, a.pool, ListAttendanceForEventQuery, eventID, orgID)
}
