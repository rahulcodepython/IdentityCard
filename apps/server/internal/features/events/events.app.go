package events

import (
    "context"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"

    "identitycard-server/internal/features/plans"
    "identitycard-server/internal/pkg/storage"
)

type CardSender interface {
    SendForEvent(ctx context.Context, orgID, eventID uuid.UUID)
}

type App struct {
    pool       *pgxpool.Pool
    plans      *plans.App
    storage    *storage.Storage
    cardSender CardSender
}

func New(pool *pgxpool.Pool, plansApp *plans.App, storage *storage.Storage) *App {
    return &App{
        pool:    pool,
        plans:   plansApp,
        storage: storage,
    }
}

func (a *App) SetCardSender(sender CardSender) {
    a.cardSender = sender
}
