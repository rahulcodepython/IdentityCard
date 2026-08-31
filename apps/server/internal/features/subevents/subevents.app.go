package subevents

import (
    "github.com/jackc/pgx/v5/pgxpool"
    "identitycard-server/internal/features/events"
)

type App struct {
    pool   *pgxpool.Pool
    events *events.App
}

func New(pool *pgxpool.Pool, eventsApp *events.App) *App {
    return &App{
        pool:   pool,
        events: eventsApp,
    }
}
