package events

import (
    "context"

    "uuid"

    "github.com/jackc/pgx/v5/pgxpool"

    "identitycard-server/internal/pkg/cache"
    "identitycard-server/internal/pkg/storage"
)

type CardSender interface {
    SendForEvent(ctx context.Context, orgID, eventID uuid.UUID) error
}

type App struct {
    pool    *pgxpool.Pool
    storage *storage.Storage
    cache   *cache.Cache
    sender  CardSender
}

func New(pool *pgxpool.Pool, s *storage.Storage, c *cache.Cache, sender CardSender) *App {
    return &App{
        pool:    pool,
        storage: s,
        cache:   c,
        sender:  sender,
    }
}

func (a *App) SetCardSender(sender CardSender) {
    a.sender = sender
}
