package forms

import (
    "github.com/jackc/pgx/v5/pgxpool"
)

// App is the composition root for the forms feature.
type App struct {
    DB *pgxpool.Pool
}

// NewApp wires together repository, service, and handler for forms.
func NewApp(db *pgxpool.Pool) *App {
    return &App{
        DB: db,
    }
}
