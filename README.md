# Identity Card

Identity card generator + attendance tracker for organizations running
events (established or flash) with optional sub-events.

## Layout

- `apps/web` — Next.js app (App Router, shadcn/ui, react-hook-form + zod).
- `apps/server` — Go/Fiber API (Postgres via sqlc + pgx, Redis, JWT auth).
- `infra` — local dev infrastructure (Postgres, Redis, MinIO, Mailhog via
  Docker Compose).

See `apps/web/README.md` and `apps/server/README.md` for per-app setup.

## Local development

```bash
cd infra && docker compose up -d      # postgres, redis, minio, mailhog
cd apps/server && make migrate-up && make dev
cd apps/web && pnpm install && pnpm dev
```

Publishing an event emails every attendee their ID card (PDF + QR) — locally
that mail goes to Mailhog, not a real inbox: http://localhost:8025.
