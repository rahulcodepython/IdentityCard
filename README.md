# Identity Card

Identity card generator + attendance tracker for organizations running
events (established or flash) with optional sub-events.

## Layout

- `apps/web` — Next.js app (App Router, shadcn/ui, react-hook-form + zod).
  Owns auth entirely, via [better-auth](https://better-auth.com).
- `apps/server` — Go/Fiber API (Postgres via sqlc + pgx, Redis). Verifies
  bearer JWTs against `apps/web`'s JWKS endpoint; issues none itself.
- `infra` — local dev infrastructure (Postgres, Redis, MinIO, Mailhog via
  Docker Compose).

See `apps/web/README.md` and `apps/server/README.md` for per-app setup.

## Local development

```bash
cd infra && docker compose up -d      # postgres, redis, minio, mailhog
cd apps/server && make dev            # migrates on boot
cd apps/web && bun install && bun dev
```

Publishing an event emails every attendee their ID card (PDF + QR) — locally
that mail goes to Mailhog, not a real inbox: http://localhost:8025.
