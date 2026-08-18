package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"identitycard-server/internal/config"
	"identitycard-server/internal/db"
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/httpx"
	"identitycard-server/internal/mailer"
	"identitycard-server/internal/modules/analytics"
	"identitycard-server/internal/modules/attendance"
	authmod "identitycard-server/internal/modules/auth"
	"identitycard-server/internal/modules/cards"
	"identitycard-server/internal/modules/devices"
	"identitycard-server/internal/modules/events"
	"identitycard-server/internal/modules/forms"
	"identitycard-server/internal/modules/members"
	"identitycard-server/internal/modules/organizations"
	"identitycard-server/internal/modules/people"
	"identitycard-server/internal/modules/plans"
	"identitycard-server/internal/modules/subevents"
	"identitycard-server/internal/redis"
	"identitycard-server/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	rdb, err := redis.Connect(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rdb.Close()

	objectStore, err := storage.Connect(ctx, cfg)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}

	queries := dbgen.New(pool)
	mail := mailer.New(cfg)

	orgsRepo := organizations.NewRepository(queries)
	orgsService := organizations.NewService(orgsRepo, objectStore)
	orgsHandler := organizations.NewHandler(orgsService)

	membersRepo := members.NewRepository(queries)

	plansRepo := plans.NewRepository(queries)
	plansService := plans.NewService(plansRepo)
	plansHandler := plans.NewHandler(plansService)

	authRepo := authmod.NewRepository(queries)
	authService := authmod.NewService(cfg, authRepo, membersRepo, orgsService, pool, queries, rdb)
	authHandler := authmod.NewHandler(cfg, authService)

	eventsRepo := events.NewRepository(queries)
	eventsService := events.NewService(eventsRepo, pool)

	subEventsRepo := subevents.NewRepository(queries)
	subEventsService := subevents.NewService(subEventsRepo, eventsService, pool)
	subEventsHandler := subevents.NewHandler(subEventsService)

	peopleRepo := people.NewRepository(queries)
	peopleService := people.NewService(peopleRepo, eventsService, subEventsService, pool)
	peopleHandler := people.NewHandler(peopleService)

	formsRepo := forms.NewRepository(queries)
	formsService := forms.NewService(formsRepo, eventsService, subEventsService, peopleService)
	formsHandler := forms.NewHandler(formsService)

	cardsService := cards.NewService(cfg, eventsService, peopleService, subEventsService, orgsService, mail)
	cardsHandler := cards.NewHandler(cardsService)

	devicesRepo := devices.NewRepository(queries)
	devicesService := devices.NewService(devicesRepo, orgsService)
	devicesHandler := devices.NewHandler(devicesService)

	attendanceRepo := attendance.NewRepository(queries)
	attendanceService := attendance.NewService(cfg, attendanceRepo, eventsService, peopleService, subEventsService)
	attendanceHandler := attendance.NewHandler(attendanceService)

	analyticsService := analytics.NewService(subEventsService, attendanceService)
	analyticsHandler := analytics.NewHandler(analyticsService)

	// events.Handler is built last: it fires cardsService on publish (see
	// events.CardSender) and can only take the concrete service once it
	// exists, since events itself can't import the cards package (cards
	// depends on events — see the cards package doc comment).
	eventsHandler := events.NewHandler(eventsService, cardsService)

	app := fiber.New(fiber.Config{
		ErrorHandler: httpx.ErrorHandler,
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{Format: "${time} ${status} ${method} ${path} (${latency})\n"}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.WebOrigin,
		AllowCredentials: true,
		// X-Device-Key: the scanner UI calls this API directly from the
		// browser (its key lives in localStorage, not an httpOnly cookie —
		// see devices.KeyHeader), so it's a real cross-origin request, not
		// just Server Components/Actions calling in server-side.
		AllowHeaders: "Content-Type," + devices.KeyHeader,
		AllowMethods: "GET,POST,PATCH,DELETE",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return httpx.OK(c, fiber.StatusOK, fiber.Map{"status": "ok"})
	})

	authmod.RegisterRoutes(app, cfg, authHandler)
	organizations.RegisterRoutes(app, cfg, orgsHandler)
	plans.RegisterRoutes(app, plansHandler)
	events.RegisterRoutes(app, cfg, eventsHandler)
	subevents.RegisterRoutes(app, cfg, subEventsHandler)
	people.RegisterRoutes(app, cfg, peopleHandler)
	forms.RegisterRoutes(app, cfg, formsHandler)
	forms.RegisterPublicRoutes(app, formsHandler)
	cards.RegisterRoutes(app, cfg, cardsHandler)
	devices.RegisterRoutes(app, cfg, devicesHandler)
	devices.RegisterPublicRoutes(app, devicesHandler)
	attendance.RegisterRoutes(app, cfg, attendanceHandler)
	analytics.RegisterRoutes(app, cfg, analyticsHandler)

	scannerGroup := devices.RegisterScannerRoutes(app, devicesService, devicesHandler)
	attendance.RegisterScanRoute(scannerGroup, attendanceHandler)

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("server: %v", err)
		}
	}()
	log.Printf("identitycard-server listening on :%s (%s)", cfg.Port, cfg.Env)

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
