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
	"identitycard-server/internal/features/eventdates"
	"identitycard-server/internal/features/events"
	"identitycard-server/internal/features/forms"
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
	EventDates *eventdates.App
	Forms      *forms.App
}

func NewRouter(
	app *fiber.App,
	cfg *config.Config,
	pool *pgxpool.Pool,
	rdb *redis.Client,
	rootCtx context.Context,
) *Router {
	cch := cache.New(rdb)

	events := events.NewApp(pool)
	eventDates := eventdates.NewApp(pool)
	forms := forms.NewApp(pool)

	return &Router{
		App:     app,
		CFG:     cfg,
		DB:      pool,
		Cache:   cch,
		RootCtx: rootCtx,

		Events:     events,
		EventDates: eventDates,
		Forms:      forms,
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
		AllowHeaders:     "Content-Type,Authorization",
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
	r.EventDates.RegisterRoutes(api)
	r.Forms.RegisterRoutes(api)
}
