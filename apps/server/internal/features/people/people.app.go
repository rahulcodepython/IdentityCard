package people

import (
	"github.com/jackc/pgx/v5/pgxpool"

	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/features/events"
	"identitycard-server/internal/features/subevents"
)

type App struct {
	pool      *pgxpool.Pool
	queries   *dbgen.Queries
	events    *events.App
	subevents *subevents.App
}

func New(pool *pgxpool.Pool, queries *dbgen.Queries, eventsApp *events.App, subeventsApp *subevents.App) *App {
	return &App{
		pool:      pool,
		queries:   queries,
		events:    eventsApp,
		subevents: subeventsApp,
	}
}
