package plans

import (
    "github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
    pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *App {
    return &App{
        pool: pool,
    }
}
