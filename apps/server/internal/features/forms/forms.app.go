package forms

import (
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/features/events"
	"identitycard-server/internal/features/people"
	"identitycard-server/internal/features/subevents"
)

type App struct {
	queries   *dbgen.Queries
	events    *events.App
	subevents *subevents.App
	people    *people.App
}

func New(queries *dbgen.Queries, eventsApp *events.App, subeventsApp *subevents.App, peopleApp *people.App) *App {
	return &App{
		queries:   queries,
		events:    eventsApp,
		subevents: subeventsApp,
		people:    peopleApp,
	}
}
