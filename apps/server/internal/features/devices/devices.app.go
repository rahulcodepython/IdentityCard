package devices

import (
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/features/organizations"
)

type App struct {
	queries *dbgen.Queries
	orgs    *organizations.App
}

func New(queries *dbgen.Queries, orgsApp *organizations.App) *App {
	return &App{
		queries: queries,
		orgs:    orgsApp,
	}
}
