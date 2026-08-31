package people

import (
    "github.com/jackc/pgx/v5/pgxpool"
    "identitycard-server/internal/features/events"
    "identitycard-server/internal/features/subevents"
)

type App struct {
    pool      *pgxpool.Pool
    events    *events.App
    subevents *subevents.App
}

func New(pool *pgxpool.Pool, eventsApp *events.App, subEventsApp *subevents.App) *App {
    return &App{
        pool:      pool,
        events:    eventsApp,
        subevents: subEventsApp,
    }
}
