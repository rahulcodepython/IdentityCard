package router

import (
    "context"
    "time"

    "github.com/go-webauthn/webauthn/webauthn"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/compress"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/helmet"
    "github.com/gofiber/fiber/v2/middleware/recover"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/redis/go-redis/v9"

    "identitycard-server/internal/config"
    "identitycard-server/internal/features/applicants"
    "identitycard-server/internal/features/attendance"
    "identitycard-server/internal/features/devices"
    "identitycard-server/internal/features/events/dates"
    "identitycard-server/internal/features/events/events"
    "identitycard-server/internal/features/forms"
    "identitycard-server/internal/features/register"
    "identitycard-server/internal/generic"
    "identitycard-server/internal/middlewares"
    "identitycard-server/internal/pkg/cache"
    "identitycard-server/internal/utils"
)

type Router struct {
    App     *fiber.App
    CFG     *config.Config
    DB      *pgxpool.Pool
    Cache   *cache.Cache
    RootCtx context.Context

    Events     *events.App
    Dates      *dates.App
    Forms      *forms.App
    Register   *register.App
    Applicants *applicants.App
    Devices    *devices.App
    Attendance *attendance.App
}

func NewRouter(
    app *fiber.App,
    cfg *config.Config,
    pool *pgxpool.Pool,
    rdb *redis.Client,
    wa *webauthn.WebAuthn,
    rootCtx context.Context,
) *Router {
    cch := cache.New(rdb)

    eventsApp := events.NewApp(pool)
    datesApp := dates.NewApp(pool)
    formsApp := forms.NewApp(pool)
    registerApp := register.NewApp(pool)
    applicantsApp := applicants.NewApp(pool)
    devicesApp := devices.NewApp(pool, rdb, wa, cfg.RPID, cfg.CanonicalOrigin)
    attendanceApp := attendance.NewApp(pool)

    return &Router{
        App:     app,
        CFG:     cfg,
        DB:      pool,
        Cache:   cch,
        RootCtx: rootCtx,

        Events:     eventsApp,
        Dates:      datesApp,
        Forms:      formsApp,
        Register:   registerApp,
        Applicants: applicantsApp,
        Devices:    devicesApp,
        Attendance: attendanceApp,
    }
}

func (r *Router) SetUp() {
    // 1. Establish per-request context bounded by request timeout and server lifecycle
    r.App.Use(middlewares.RequestContextMiddleware(r.RootCtx, time.Duration(r.CFG.RequestTimeoutSec)*time.Second))

    // 2. Structured request logging
    r.App.Use(middlewares.LoggerMiddleware())

    // 3. Outermost panic recovery
    r.App.Use(recover.New())

    // 4. Standard security headers
    r.App.Use(helmet.New())

    // 5. Response compression
    r.App.Use(compress.New())

    // 6. CORS configuration
    r.App.Use(cors.New(cors.Config{
        AllowOrigins:     r.CFG.WebOrigin,
        AllowHeaders:     "Content-Type,Authorization,X-Device-Token",
        AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
        AllowCredentials: true,
    }))

    // 7. Rate limiting per client IP
    r.App.Use(middlewares.RateLimiterMiddleware())

    // Health check handler
    healthHandler := func(c *fiber.Ctx) error {
        return utils.OK(c, generic.MsgStatusOk, fiber.Map{"status": generic.MsgStatusOk})
    }

    r.App.Get("/health", healthHandler)
    registerDocsRoutes(r.App)

    // Base API v1 group
    api := r.App.Group(generic.APIV1Prefix)

    r.Events.RegisterRoutes(api)
    r.Dates.RegisterRoutes(api)
    r.Forms.RegisterRoutes(api)
    r.Register.RegisterRoutes(api)
    r.Applicants.RegisterRoutes(api)
    r.Devices.RegisterRoutes(api)
    r.Attendance.RegisterRoutes(api)
}
