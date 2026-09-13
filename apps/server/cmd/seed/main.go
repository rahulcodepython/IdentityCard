package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"uuid"

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

	email := getOr("SEED_ADMIN_EMAIL", "rahulcodepython@gmail.com")
	name := getOr("SEED_ADMIN_NAME", "Rahul Das")

	now := time.Now().UTC()
	var userID string
	err = pool.QueryRow(ctx, "SELECT id FROM \"user\" WHERE email = $1", email).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		userID = uuid.New().String()
		_, err = pool.Exec(ctx, `
            INSERT INTO "user" ("id", "name", "email", "emailVerified", "createdAt", "updatedAt")
            VALUES ($1, $2, $3, true, $4, $4)
        `, userID, name, email, now)
		if err != nil {
			log.Fatalf("seed: create user: %v", err)
		}
		log.Printf("seed: created user %s (id=%s)", email, userID)
	} else if err != nil {
		log.Fatalf("seed: check existing user: %v", err)
	} else {
		log.Printf("seed: user %s already exists (id=%s), proceeding to organization setup", email, userID)
	}

	type orgSeed struct {
		name string
		slug string
		role generic.Role
	}

	orgsToSeed := []orgSeed{
		{
			name: "Rahul Tech Hub",
			slug: "rahul-tech-hub",
			role: generic.RoleOwner,
		},
		{
			name: "Apex Innovations",
			slug: "apex-innovations",
			role: generic.RoleOwner,
		},
		{
			name: "Global Community",
			slug: "global-community",
			role: generic.RoleMember,
		},
	}

	var primaryOrgID string
	for idx, org := range orgsToSeed {
		var orgID string
		err = pool.QueryRow(ctx, "SELECT id FROM \"organization\" WHERE slug = $1", org.slug).Scan(&orgID)
		if errors.Is(err, pgx.ErrNoRows) {
			orgID = uuid.New().String()
			_, err = pool.Exec(ctx, `
                INSERT INTO "organization" ("id", "name", "slug", "createdAt")
                VALUES ($1, $2, $3, $4)
            `, orgID, org.name, org.slug, now)
			if err != nil {
				log.Fatalf("seed: create organization %q: %v", org.name, err)
			}
			log.Printf("seed: created organization %q (id=%s, slug=%s)", org.name, orgID, org.slug)
		} else if err != nil {
			log.Fatalf("seed: check existing organization %q: %v", org.name, err)
		} else {
			log.Printf("seed: organization %q already exists (id=%s, slug=%s)", org.name, orgID, org.slug)
		}

		if idx == 0 {
			primaryOrgID = orgID
		}

		var memberID string
		var currentRole string
		err = pool.QueryRow(ctx, `
            SELECT id, role FROM "member" WHERE "organizationId" = $1 AND "userId" = $2
        `, orgID, userID).Scan(&memberID, &currentRole)

		if errors.Is(err, pgx.ErrNoRows) {
			memberID = uuid.New().String()
			_, err = pool.Exec(ctx, `
                INSERT INTO "member" ("id", "organizationId", "userId", "role", "createdAt")
                VALUES ($1, $2, $3, $4, $5)
            `, memberID, orgID, userID, string(org.role), now)
			if err != nil {
				log.Fatalf("seed: add membership for %q: %v", org.name, err)
			}
			log.Printf("seed: added %s to %q as %s", email, org.name, org.role)
		} else if err != nil {
			log.Fatalf("seed: check membership for %q: %v", org.name, err)
		} else if currentRole != string(org.role) {
			_, err = pool.Exec(ctx, `
                UPDATE "member" SET "role" = $1 WHERE "id" = $2
            `, string(org.role), memberID)
			if err != nil {
				log.Fatalf("seed: update membership role for %q: %v", org.name, err)
			}
			log.Printf("seed: updated %s in %q to role %s", email, org.name, org.role)
		} else {
			log.Printf("seed: user %s already a %s in %q", email, currentRole, org.name)
		}
	}

	if primaryOrgID != "" {
		_, _ = pool.Exec(ctx, `
            UPDATE "session"
            SET "activeOrganizationId" = $1
            WHERE "userId" = $2 AND "activeOrganizationId" IS NULL
        `, primaryOrgID, userID)
	}

	log.Printf("seed: successfully completed seeding for %s with 3 organizations!", email)
}

func getOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
