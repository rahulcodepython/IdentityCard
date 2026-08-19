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
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"identitycard-server/internal/config"
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/jobs"
	"identitycard-server/internal/pkg/mailer"
	"identitycard-server/internal/pkg/oauth"
	"identitycard-server/internal/pkg/postgres"
	"identitycard-server/internal/pkg/redis"
	"identitycard-server/internal/pkg/storage"
	"identitycard-server/internal/routes"
	"identitycard-server/internal/services"
	"identitycard-server/internal/utils"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
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
	googleOAuth := oauth.New(cfg)

	app := fiber.New(fiber.Config{
		ErrorHandler: utils.ErrorHandler,
	})

	app.Use(recover.New())
	app.Use(limiter.New(limiter.Config{Max: 100, Expiration: time.Minute}))
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{Format: "${time} ${status} ${method} ${path} (${latency})\n"}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.WebOrigin,
		// Go API now directly issues and verifies session cookies.
		AllowCredentials: true,
		AllowHeaders:     "Content-Type,Authorization," + services.KeyHeader,
		AllowMethods:     "GET,POST,PATCH,DELETE,OPTIONS",
	}))

	router := routes.NewRouter(app, cfg, pool, rdb, objectStore, mail, googleOAuth, queries)
	router.SetUp()

	// Background sweeps — billing (mark lapsed subscriptions past_due,
	// hard-delete events past their grace/retention deadline) and
	// recurring-event horizon extension — run on the same ctx as the
	// server itself, so they stop cleanly on the same shutdown signal. A
	// prototype-simple daily ticker — see internal/jobs — rather than a
	// dedicated cron dependency.
	backgroundTasks := append(
		jobs.NewBillingTasks(router.PlansService, router.EventsService),
		jobs.NewRecurrenceTasks(router.EventsService)...,
	)
	go jobs.Run(ctx, 24*time.Hour, backgroundTasks...)

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
