// Command seed creates the first organization, its super_admin user, and
// the membership linking them — a fast way to get a usable account
// locally without going through better-auth's real email-OTP/TOTP
// sign-up flow first. It writes directly into better-auth's own tables
// (see internal/db/migrations/000035_better_auth_schema.up.sql) — no
// password to fake, since the app is passwordless; the seeded user still
// signs in for real, in the Next.js app, via the email-OTP tab against
// local Mailhog (http://localhost:8025) once this row exists.
package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/jackc/pgx/v5"

	"identitycard-server/internal/config"
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/generic"
	"identitycard-server/internal/pkg/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()
	queries := dbgen.New(pool)

	email := getOr("SEED_ADMIN_EMAIL", "admin@example.com")
	name := getOr("SEED_ADMIN_NAME", "Super Admin")
	orgName := getOr("SEED_ORG_NAME", "Acme Events")
	orgSlug := getOr("SEED_ORG_SLUG", "acme-events")

	if existing, err := queries.GetUserByEmail(ctx, email); err == nil {
		log.Printf("seed: user %s already exists (id=%s), nothing to do", email, existing.ID)
		return
	} else if !errors.Is(err, pgx.ErrNoRows) {
		log.Fatalf("seed: check existing user: %v", err)
	}

	user, err := queries.CreateUser(ctx, dbgen.CreateUserParams{Email: email, Name: name})
	if err != nil {
		log.Fatalf("seed: create user: %v", err)
	}

	org, err := queries.CreateOrganization(ctx, dbgen.CreateOrganizationParams{Name: orgName, Slug: orgSlug})
	if err != nil {
		log.Fatalf("seed: create organization: %v", err)
	}

	if _, err := queries.CreateMember(ctx, dbgen.CreateMemberParams{
		OrganizationId: org.ID,
		UserId:         user.ID,
		Role:           string(generic.RoleAdmin),
	}); err != nil {
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
