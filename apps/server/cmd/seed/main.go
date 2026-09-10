package main

import (
    "context"
    "errors"
    "log"
    "os"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"

    "identitycard-server/internal/config"
    "identitycard-server/internal/generic"
    "identitycard-server/internal/pkg/postgres"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("config: %v", err)
    }

    ctx := context.Background()
    pool, err := postgres.Connect(ctx, cfg)
    if err != nil {
        log.Fatalf("database: %v", err)
    }
    defer pool.Close()

    email := getOr("SEED_ADMIN_EMAIL", "admin@example.com")
    name := getOr("SEED_ADMIN_NAME", "Super Admin")
    orgName := getOr("SEED_ORG_NAME", "Acme Events")
    orgSlug := getOr("SEED_ORG_SLUG", "acme-events")

    var existingID string
    err = pool.QueryRow(ctx, "SELECT id FROM \"user\" WHERE email = $1", email).Scan(&existingID)
    if err == nil {
        log.Printf("seed: user %s already exists (id=%s), nothing to do", email, existingID)
        return
    } else if !errors.Is(err, pgx.ErrNoRows) {
        log.Fatalf("seed: check existing user: %v", err)
    }

    userID := uuid.New().String()
    orgID := uuid.New().String()
    memberID := uuid.New().String()
    now := time.Now().UTC()

    _, err = pool.Exec(ctx, `
        INSERT INTO "user" (id, name, email, email_verified, created_at, updated_at)
        VALUES ($1, $2, $3, true, $4, $4)
    `, userID, name, email, now)
    if err != nil {
        log.Fatalf("seed: create user: %v", err)
    }

    _, err = pool.Exec(ctx, `
        INSERT INTO organization (id, name, slug, created_at)
        VALUES ($1, $2, $3, $4)
    `, orgID, orgName, orgSlug, now)
    if err != nil {
        log.Fatalf("seed: create organization: %v", err)
    }

    _, err = pool.Exec(ctx, `
        INSERT INTO member (id, organization_id, user_id, role, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `, memberID, orgID, userID, string(generic.RoleAdmin), now)
    if err != nil {
        log.Fatalf("seed: create membership: %v", err)
    }

    log.Printf("seed: created organization %q and super_admin %s — sign in with the email-code tab", orgName, email)
}

func getOr(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
