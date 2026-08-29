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
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"identitycard-server/internal/config"
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/generic"
	"identitycard-server/internal/jobs"
	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/pkg/jwt"
	"identitycard-server/internal/pkg/mailer"
	"identitycard-server/internal/pkg/postgres"
	"identitycard-server/internal/pkg/redis"
	"identitycard-server/internal/pkg/storage"
	"identitycard-server/internal/router"
	"identitycard-server/internal/utils"
)

const migrationsPath = "file://internal/db/migrations"

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

	verifier, err := jwt.NewVerifier(ctx, cfg.BetterAuthJWKSURL)
	if err != nil {
		log.Fatalf("jwt: fetch JWKS: %v", err)
	}

	middlewares.InitAuth(verifier)

	queries := dbgen.New(pool)
	mail := mailer.New(cfg)

	app := fiber.New(fiber.Config{
		ErrorHandler:   utils.ErrorHandler,
		ReadBufferSize: 16384,
	})

	app.Use(recover.New())
	app.Use(limiter.New(limiter.Config{Max: 100, Expiration: time.Minute}))
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{Format: "${time} ${status} ${method} ${path} (${latency})\n"}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.WebOrigin,
		AllowHeaders: "Content-Type,Authorization," + generic.HeaderDeviceKey,
		AllowMethods: "GET,POST,PATCH,DELETE,OPTIONS",
	}))

	r := router.NewRouter(app, cfg, pool, rdb, objectStore, mail, queries)
	r.SetUp()

	backgroundTasks := jobs.NewBillingTasks(r.Plans, r.Events)
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
