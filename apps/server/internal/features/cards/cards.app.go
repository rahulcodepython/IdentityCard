package cards

import (
	"identitycard-server/internal/config"
	"identitycard-server/internal/features/events"
	"identitycard-server/internal/features/organizations"
	"identitycard-server/internal/features/people"
	"identitycard-server/internal/features/subevents"
	"identitycard-server/internal/pkg/mailer"
)

type App struct {
	cfg       *config.Config
	events    *events.App
	people    *people.App
	subevents *subevents.App
	orgs      *organizations.App
	mailer    *mailer.Mailer
}

func New(
	cfg *config.Config,
	eventsApp *events.App,
	peopleApp *people.App,
	subeventsApp *subevents.App,
	orgsApp *organizations.App,
	m *mailer.Mailer,
) *App {
	return &App{
		cfg:       cfg,
		events:    eventsApp,
		people:    peopleApp,
		subevents: subeventsApp,
		orgs:      orgsApp,
		mailer:    m,
	}
}
