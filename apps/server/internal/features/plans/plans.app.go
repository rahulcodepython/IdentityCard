package plans

import (
	"github.com/jackc/pgx/v5/pgxpool"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

type App struct {
	pool    *pgxpool.Pool
	queries *dbgen.Queries
}

func New(pool *pgxpool.Pool, queries *dbgen.Queries) *App {
	return &App{
		pool:    pool,
		queries: queries,
	}
}
