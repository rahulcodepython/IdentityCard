package dates

import (
    "github.com/jackc/pgx/v5/pgxpool"
)

// App is the composition root for the event dates feature.
type App struct {
    DB *pgxpool.Pool
}

// NewApp wires together dependencies for the event dates feature.
func NewApp(db *pgxpool.Pool) *App {
    return &App{
        DB: db,
    }
}
