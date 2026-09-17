package main

import (
    "context"
    "log"
    "log/slog"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-webauthn/webauthn/webauthn"
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/config"
    "identitycard-server/internal/pkg/postgres"
    "identitycard-server/internal/pkg/redis"
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

    // Initialize WebAuthn Relying Party from configuration
    wa, err := webauthn.New(&webauthn.Config{
        RPDisplayName: cfg.RPDisplayName,
        RPID:          cfg.RPID,
        RPOrigins:     []string{cfg.CanonicalOrigin},
    })
    if err != nil {
        log.Fatalf("[main] webauthn init failed: %v", err)
    }

    // Create root context for server lifecycle
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Create the Fiber app with production-oriented defaults.
    app := fiber.New(fiber.Config{
        AppName:                 "IdentityCard API v1.0",
        ErrorHandler:            utils.ErrorHandler,
        BodyLimit:               1 * 1024 * 1024,
        DisableStartupMessage:   false,
        ReadBufferSize:          16 * 1024,
        EnableTrustedProxyCheck: true,
        TrustedProxies:          cfg.TrustedProxies,
        ProxyHeader:             fiber.HeaderXForwardedFor,
    })

    // Setup router composition root
    r := router.NewRouter(app, cfg, pool, rdb, wa, ctx)
    r.SetUp()

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
