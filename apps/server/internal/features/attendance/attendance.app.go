package attendance

import (
	"identitycard-server/internal/config"
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/features/events"
	"identitycard-server/internal/features/people"
	"identitycard-server/internal/features/subevents"
)

type App struct {
	cfg       *config.Config
	queries   *dbgen.Queries
	events    *events.App
	people    *people.App
	subevents *subevents.App
}

func New(
	cfg *config.Config,
	queries *dbgen.Queries,
	eventsApp *events.App,
	peopleApp *people.App,
	subeventsApp *subevents.App,
) *App {
	return &App{
		cfg:       cfg,
		queries:   queries,
		events:    eventsApp,
		people:    peopleApp,
		subevents: subeventsApp,
	}
}
