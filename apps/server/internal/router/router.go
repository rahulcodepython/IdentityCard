package router

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"identitycard-server/internal/config"
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
	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/pkg/cache"
	"identitycard-server/internal/pkg/jwt"
	"identitycard-server/internal/pkg/mailer"
	"identitycard-server/internal/pkg/storage"
	"identitycard-server/internal/utils"
)

type Router struct {
	App      *fiber.App
	CFG      *config.Config
	DB       *pgxpool.Pool
	Cache    *cache.Cache
	Verifier *jwt.Verifier
	RootCtx  context.Context

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
	verifier *jwt.Verifier,
	rootCtx context.Context,
) *Router {
	cch := cache.New(rdb)

	orgsApp := organizations.New(pool, objectStore)
	plansApp := plans.New(pool, cch)
	eventsApp := events.New(pool, plansApp, objectStore)
	subEventsApp := subevents.New(pool, eventsApp)
	peopleApp := people.New(pool, eventsApp, subEventsApp)
	formsApp := forms.New(pool, eventsApp, subEventsApp, peopleApp)
	cardsApp := cards.New(cfg, eventsApp, peopleApp, subEventsApp, orgsApp, mail)
    devicesApp := devices.New(pool, orgsApp)
    attendanceApp := attendance.New(cfg, pool, eventsApp, peopleApp, subEventsApp, devicesApp)
    analyticsApp := analytics.New(eventsApp, subEventsApp, attendanceApp, devicesApp)

    eventsApp.SetCardSender(cardsApp)

    return &Router{
        App:           app,
        CFG:           cfg,
        DB:            pool,
        Cache:         cch,
        Verifier:      verifier,
        RootCtx:       rootCtx,
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

    // 6. CORS must be mounted BEFORE rate limiter so that 429 responses still carry
    // Access-Control-Allow-Origin headers, preventing client-side opaque fetch errors.
    r.App.Use(cors.New(cors.Config{
        AllowOrigins:     r.CFG.WebOrigin,
        AllowHeaders:     "Content-Type,Authorization," + generic.HeaderDeviceKey,
        AllowMethods:     "GET,POST,PATCH,DELETE,OPTIONS",
        AllowCredentials: true,
    }))

    // 7. Rate limiting per client IP
    r.App.Use(middlewares.RateLimiterMiddleware())

    // Health check handler
    healthHandler := func(c *fiber.Ctx) error {
        return utils.OK(c, fiber.StatusOK, fiber.Map{"status": generic.MsgStatusOk})
    }

    r.App.Get("/health", healthHandler)
    registerDocsRoutes(r.App)

    public := r.App.Group(generic.APIV1Prefix)

    // Base authentication middleware instantiated with router's JWKS verifier
    auth := middlewares.BaseAuthMiddleware(r.Verifier)
    billing := middlewares.RequireActiveBilling(r.DB, r.Cache)
    protected := r.App.Group(generic.APIV1Prefix, auth, billing)

    // Register all feature routes
    r.Organizations.RegisterRoutes(protected, public)
    r.Plans.RegisterRoutes(protected, public)
    r.Events.RegisterRoutes(protected, public)
    r.SubEvents.RegisterRoutes(protected, public)
    r.People.RegisterRoutes(protected, public)
    r.Forms.RegisterRoutes(protected, public)
    r.Cards.RegisterRoutes(protected, public)
    r.Devices.RegisterRoutes(protected, public)
    r.Attendance.RegisterRoutes(protected, public)
    r.Analytics.RegisterRoutes(protected, public)

}
