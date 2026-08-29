package router

import (
    "github.com/gofiber/fiber/v2"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/redis/go-redis/v9"

    "identitycard-server/internal/config"
    dbgen "identitycard-server/internal/db/sqlc/generated"
    "identitycard-server/internal/features/analytics"
    "identitycard-server/internal/features/attendance"
    "identitycard-server/internal/features/cards"
    "identitycard-server/internal/features/devices"
    "identitycard-server/internal/features/events"
    "identitycard-server/internal/features/forms"
    "identitycard-server/internal/features/organizations"
    "identitycard-server/internal/features/people"
    "identitycard-server/internal/features/plans"
    "identitycard-server/internal/features/subevents"
    "identitycard-server/internal/generic"
    "identitycard-server/internal/pkg/mailer"
    "identitycard-server/internal/pkg/storage"
    "identitycard-server/internal/utils"
)

type Router struct {
    app *fiber.App
    cfg *config.Config

    Organizations *organizations.App
    Plans         *plans.App
    Events        *events.App
    SubEvents     *subevents.App
    People        *people.App
    Forms         *forms.App
    Cards         *cards.App
    Devices       *devices.App
    Attendance    *attendance.App
    Analytics     *analytics.App
}

func NewRouter(
    app *fiber.App,
    cfg *config.Config,
    pool *pgxpool.Pool,
    rdb *redis.Client,
    objectStore *storage.Storage,
    mail *mailer.Mailer,
    queries *dbgen.Queries,
) *Router {
    orgsApp := organizations.New(queries, objectStore)
    plansApp := plans.New(pool, queries)
    eventsApp := events.New(pool, queries, plansApp, objectStore)
    subEventsApp := subevents.New(queries, eventsApp)
    peopleApp := people.New(pool, queries, eventsApp, subEventsApp)
    formsApp := forms.New(queries, eventsApp, subEventsApp, peopleApp)
    cardsApp := cards.New(cfg, eventsApp, peopleApp, subEventsApp, orgsApp, mail)
    devicesApp := devices.New(queries, orgsApp)
    attendanceApp := attendance.New(cfg, queries, eventsApp, peopleApp, subEventsApp)
    analyticsApp := analytics.New(eventsApp, subEventsApp, attendanceApp, devicesApp)

    eventsApp.SetCardSender(cardsApp)

    return &Router{
        app:           app,
        cfg:           cfg,
        Organizations: orgsApp,
        Plans:         plansApp,
        Events:        eventsApp,
        SubEvents:     subEventsApp,
        People:        peopleApp,
        Forms:         formsApp,
        Cards:         cardsApp,
        Devices:       devicesApp,
        Attendance:    attendanceApp,
        Analytics:     analyticsApp,
    }
}

func (r *Router) SetUp() {
    healthHandler := func(c *fiber.Ctx) error {
        return utils.OK(c, fiber.StatusOK, fiber.Map{"status": generic.MsgStatusOk})
    }

    r.app.Get("/health", healthHandler)
    registerDocsRoutes(r.app)

    api := r.app.Group(generic.APIV1Prefix)
    api.Get("/health", healthHandler)

    r.Organizations.RegisterRoutes(api)
    r.Plans.RegisterRoutes(api)
    r.Events.RegisterRoutes(api)
    r.SubEvents.RegisterRoutes(api)
    r.People.RegisterRoutes(api)
    r.Forms.RegisterRoutes(api)
    r.Forms.RegisterPublicRoutes(api)
    r.Cards.RegisterRoutes(api)
    r.Devices.RegisterRoutes(api)
    r.Devices.RegisterPublicRoutes(api)
    r.Attendance.RegisterRoutes(api)
    r.Analytics.RegisterRoutes(api)

    scannerGroup := r.Devices.RegisterScannerRoutes(api)
    r.Attendance.RegisterScanRoute(scannerGroup)
}
