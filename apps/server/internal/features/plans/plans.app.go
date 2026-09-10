package plans

import (
    "github.com/jackc/pgx/v5/pgxpool"

    "identitycard-server/internal/pkg/cache"
)

type App struct {
    pool  *pgxpool.Pool
    cache *cache.Cache
}

func New(pool *pgxpool.Pool, cch *cache.Cache) *App {
    return &App{
        pool:  pool,
        cache: cch,
    }
}
