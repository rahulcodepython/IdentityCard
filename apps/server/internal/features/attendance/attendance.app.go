package attendance

import (
    "github.com/jackc/pgx/v5/pgxpool"
    "identitycard-server/internal/config"
    "identitycard-server/internal/features/devices"
    "identitycard-server/internal/features/events"
    "identitycard-server/internal/features/people"
    "identitycard-server/internal/features/subevents"
)

type App struct {
    cfg       *config.Config
    pool      *pgxpool.Pool
    events    *events.App
    people    *people.App
    subevents *subevents.App
    devices   *devices.App
}

func New(
    cfg *config.Config,
    pool *pgxpool.Pool,
    eventsApp *events.App,
    peopleApp *people.App,
    subEventsApp *subevents.App,
    devicesApp *devices.App,
) *App {
    return &App{
        cfg:       cfg,
        pool:      pool,
        events:    eventsApp,
        people:    peopleApp,
        subevents: subEventsApp,
        devices:   devicesApp,
    }
}
