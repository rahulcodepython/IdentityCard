package devices

import (
    "github.com/jackc/pgx/v5/pgxpool"
    "identitycard-server/internal/features/organizations"
)

type App struct {
    pool *pgxpool.Pool
    orgs *organizations.App
}

func New(pool *pgxpool.Pool, orgsApp *organizations.App) *App {
    return &App{
        pool: pool,
        orgs: orgsApp,
    }
}
