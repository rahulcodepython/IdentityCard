package organizations

import (
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/pkg/storage"
)

type App struct {
	queries *dbgen.Queries
	storage *storage.Storage
}

func New(queries *dbgen.Queries, store *storage.Storage) *App {
	return &App{
		queries: queries,
		storage: store,
	}
}
