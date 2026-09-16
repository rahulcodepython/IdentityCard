package events

import (
    "github.com/jackc/pgx/v5/pgxpool"
)

// App is the composition root for the events feature.
type App struct {
    DB *pgxpool.Pool
}

// NewApp wires together repository, service, and handler for events.
func NewApp(db *pgxpool.Pool) *App {
    return &App{
        DB: db,
    }
}
