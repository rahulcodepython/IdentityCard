package subevents

import (
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/features/events"
)

type App struct {
	queries *dbgen.Queries
	events  *events.App
}

func New(queries *dbgen.Queries, eventsApp *events.App) *App {
	return &App{
		queries: queries,
		events:  eventsApp,
	}
}
