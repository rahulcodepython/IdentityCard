package attendance

import (
    "github.com/jackc/pgx/v5/pgxpool"

    "identitycard-server/internal/pkg/cache"
)

// App is the composition root for the attendance feature.
type App struct {
    DB    *pgxpool.Pool
    Cache *cache.Cache
}

// NewApp creates a new App instance for attendance.
func NewApp(db *pgxpool.Pool, cache *cache.Cache) *App {
    return &App{
        DB:    db,
        Cache: cache,
    }
}
