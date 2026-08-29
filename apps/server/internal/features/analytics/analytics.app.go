package analytics

import (
	"identitycard-server/internal/features/attendance"
	"identitycard-server/internal/features/devices"
	"identitycard-server/internal/features/events"
	"identitycard-server/internal/features/subevents"
)

type App struct {
	events     *events.App
	subevents  *subevents.App
	attendance *attendance.App
	devices    *devices.App
}

func New(eventsApp *events.App, subeventsApp *subevents.App, attendanceApp *attendance.App, devicesApp *devices.App) *App {
	return &App{
		events:     eventsApp,
		subevents:  subeventsApp,
		attendance: attendanceApp,
		devices:    devicesApp,
	}
}
