package attendance

import (
    "github.com/jackc/pgx/v5/pgxpool"
)

// App is the composition root for the attendance feature.
type App struct {
    DB *pgxpool.Pool
}

// NewApp creates a new App instance for attendance.
func NewApp(db *pgxpool.Pool) *App {
    return &App{
        DB: db,
    }
}
