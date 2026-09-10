package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	"identitycard-server/internal/config"
	"identitycard-server/internal/jobs"
	"identitycard-server/internal/pkg/jwt"
	"identitycard-server/internal/pkg/mailer"
	"identitycard-server/internal/pkg/postgres"
	"identitycard-server/internal/pkg/redis"
	"identitycard-server/internal/pkg/storage"
	"identitycard-server/internal/router"
	"identitycard-server/internal/utils"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// Load application configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[main] config load failed: %v", err)
	}

	// Connect to the database
	pool, err := postgres.Connect(context.Background(), cfg)
	if err != nil {
		log.Fatalf("[main] database connect failed: %v", err)
	}
	defer postgres.Close(pool)

	// Connect to Redis
	rdb := redis.Connect(context.Background(), cfg)
	defer rdb.Close()

	// Connect to MinIO/S3 storage, continuing without file storage if initialization fails.
	objectStore, err := storage.Connect(context.Background(), cfg)
	if err != nil {
		slog.Warn("storage connect failed, continuing without file storage", "error", err)
	}

	// Create root context for background services like JWKS keyfunc refresh
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize JWKS JWT verifier. Auth is not optional — every protected
	// route depends on it, so a bad JWKS URL must fail the deploy here
	// rather than let the server boot and 500 on every authenticated
	// request at runtime.
	verifier, err := jwt.NewVerifier(ctx, cfg.BetterAuthJWKSURL)
	if err != nil {
		log.Fatalf("[main] jwks verifier init failed: %v", err)
	}

	mail := mailer.New(cfg)

	// Create the Fiber app with production-oriented defaults.
	app := fiber.New(fiber.Config{
		AppName:               "IdentityCard API v1.0",
		ErrorHandler:          utils.ErrorHandler,
		BodyLimit:             1 * 1024 * 1024,
		DisableStartupMessage: false,
		// fasthttp's default 4096-byte read buffer covers the request line
		// plus ALL headers combined. Our Authorization header alone carries
		// a JWT with the caller's full roles/permissions array embedded —
		// an admin holding every admin:* permission already produces a
		// ~1.3KB token, and Chrome's own default headers (sec-ch-ua-*,
		// Accept, cookies, etc.) add several hundred more on top. Once that
		// combined total exceeded the buffer, fasthttp didn't return a
		// clean 431 — it just closed the connection, which every browser
		// surfaces as an opaque "Failed to fetch"/network error with zero
		// indication of the real cause. More permissions == a bigger JWT ==
		// this getting worse over time, so headroom needs to be generous,
		// not just bumped to whatever happens to work today.
		ReadBufferSize: 16 * 1024,
	})

	// Setup router composition root
	r := router.NewRouter(app, cfg, pool, rdb, objectStore, mail, verifier, ctx)
	r.SetUp()

	// Background jobs started after routing
	maintenanceTask := jobs.NewDailyMaintenanceTask(pool, objectStore, mail)
	go jobs.Run(ctx, 24*time.Hour, maintenanceTask)

	// Transactional outbox worker for reliable asynchronous operations
	outboxWorker := jobs.NewOutboxWorker(pool, r.Cards)
	go outboxWorker.Start(ctx)

	// Gracefully stop the server on SIGINT or SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		slog.Info("shutting down server")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer shutdownCancel()
		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			slog.Error("shutdown error", "error", err)
		}
	}()

	addr := ":" + cfg.Port
	slog.Info("IdentityCard API listening", "addr", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("[main] listen: %v", err)
	}
}
