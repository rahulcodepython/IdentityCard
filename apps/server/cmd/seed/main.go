// Command seed creates the first organization, its super_admin user, and
// the membership linking them — a fast way to get a usable account
// locally without going through the OTP/TOTP registration flow. The
// seeded user is inserted already email-verified with no TOTP secret, so
// it signs in via the email-OTP tab against local Mailhog
// (http://localhost:8025).
package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/jackc/pgx/v5"

	"identitycard-server/internal/auth"
	"identitycard-server/internal/config"
	"identitycard-server/internal/db"
	dbgen "identitycard-server/internal/db/sqlc/generated"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
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
	if _, err := queries.MarkEmailVerified(ctx, user.ID); err != nil {
		log.Fatalf("seed: mark user verified: %v", err)
	}

	org, err := queries.CreateOrganization(ctx, dbgen.CreateOrganizationParams{Name: orgName, Slug: orgSlug})
	if err != nil {
		log.Fatalf("seed: create organization: %v", err)
	}

	if _, err := queries.CreateOrganizationMember(ctx, dbgen.CreateOrganizationMemberParams{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Roles:          []string{string(auth.RoleSuperAdmin)},
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
