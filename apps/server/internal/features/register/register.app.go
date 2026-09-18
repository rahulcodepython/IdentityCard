package register

import (
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/redis/go-redis/v9"

    "identitycard-server/internal/pkg/cache"
)

type App struct {
    DB    *pgxpool.Pool
    RDB   *redis.Client
    Cache *cache.Cache
}

func NewApp(db *pgxpool.Pool, rdb *redis.Client, cache *cache.Cache) *App {
    return &App{
        DB:    db,
        RDB:   rdb,
        Cache: cache,
    }
}
