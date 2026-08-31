package organizations

import (
    "github.com/jackc/pgx/v5/pgxpool"
    "identitycard-server/internal/pkg/storage"
)

type App struct {
    pool    *pgxpool.Pool
    storage *storage.Storage
}

func New(pool *pgxpool.Pool, storage *storage.Storage) *App {
    return &App{
        pool:    pool,
        storage: storage,
    }
}
