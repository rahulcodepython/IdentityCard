# apps/server

Go/Fiber API: Postgres (via [sqlc](https://sqlc.dev) + pgx) and Redis
(refresh-token/session storage).

## Setup

```bash
cp .env.example .env          # defaults match infra/docker-compose.yml
cd ../../infra && docker compose up -d
cd ../apps/server
make migrate-up
make seed                     # creates the first org + super_admin login
make dev                      # http://localhost:8080
```

`make seed` prints the admin email/password it created (override with
`SEED_ADMIN_EMAIL` / `SEED_ADMIN_PASSWORD` env vars) — handy for a quick
login, but the normal path is `POST /auth/register` (see `apps/web`'s
`/plans` → `/register` flow), which creates an org + super_admin from
scratch against a selected plan.

## Conventions

- One folder per feature under `internal/modules/<name>`: `routes.go` →
  `handler.go` (bind + validate) → `service.go` (business logic) →
  `repository.go` (sqlc queries). No cross-module repository calls for
  reads — those go through the other module's `Service`. The one exception
  is a write that must be atomic across modules (e.g. `auth.Service.Register`
  creating a user + org + subscription + membership together): it uses
  `db.WithTx` and constructs short-lived, tx-scoped repositories directly
  via `dbgen.Queries.WithTx(tx)` — see `internal/modules/auth/service.go`.
- Every request DTO is bound + validated with `httpx.BindAndValidate`
  (`internal/httpx/bind.go`); every response is a typed struct passed to
  `httpx.OK`. Never `map[string]interface{}` on either side.
- Tenant-owned tables always carry `organization_id`; repository functions
  that read them always take `organizationID` as an explicit argument.
- SQL lives in `internal/db/sqlc/queries/*.sql`; run `make sqlc-generate`
  after adding/changing a query to regenerate `internal/db/sqlc/generated`
  (gitignored, not committed).
- Schema changes are plain SQL migrations: `make migrate-new name=<desc>`,
  then `make migrate-up`.
- A sub-resource that only ever makes sense nested under its parent (e.g.
  `subevents`/`people`/`forms` under `events`) gets its own module and its
  own tx-scoped `Repository.WithTx`, but reads the parent's state through
  the parent's `Service` rather than its repository — see
  `events.Service.GetContext`, used by `subevents`, `people`, and `forms`.
  Its routes are mounted nested too: `/events/:eventId/<resource>/...`.
- A route that must work with no authenticated session at all (the public
  sign-up form) lives in its owning module's own `RegisterPublicRoutes`,
  mounted outside every `RequireAuth` group — see
  `internal/modules/forms/routes.go` (`/public/forms/:token`). Token
  possession is the only access control; there is no user identity to check.
- Three ingestion paths (manual entry, CSV import, public-form submission)
  all write a person through the same `people.Service.upsert`: one row per
  `(event_id, email, mobile)`, `INSERT ... ON CONFLICT DO UPDATE`, telling
  insert from update via Postgres's `xmax = 0` trick in the `RETURNING`
  clause rather than a separate existence check.
- A capacity-limited resource (`event_forms.submissions_count`) is charged
  by an atomic conditional `UPDATE ... WHERE submissions_count < capacity
  RETURNING *` (zero rows back means full) *before* the write it's gating,
  then refunded with a plain decrement if that write turns out to be a
  no-op (an existing registrant resubmitting) — see `forms.Service.Submit`.
  This never lets the gated write happen without a reserved slot, at the
  cost of a brief, self-correcting overcount under concurrent duplicates.
- A response that isn't JSON (CSV export, the org logo, a generated PDF)
  skips `httpx.OK` and sets `Content-Type`/`Content-Disposition` directly —
  see `people.Handler.Export`, `organizations.Handler.GetLogo`,
  `cards.Handler.Download`. A deliberate, narrow exception to the
  `{"data": ...}` envelope convention, not a precedent for skipping it
  elsewhere.
- When module A needs a callback into module B, but B already depends on A
  (so B can't import A back without cycling), A defines a small interface
  for just the method it needs and takes that instead of B's concrete type
  — see `events.CardSender`, satisfied by `cards.Service`. Only `cmd/api`
  imports both packages, so it's the only place that can hand the concrete
  type in; the two packages themselves never reference each other.
- Nothing user-supplied that resolves to a network address is fetched
  without going through a SSRF-safe client: `cards` fetches an attendee's
  external photo URL (submitted, unauthenticated, through the public
  sign-up form) via a custom `DialContext` that re-resolves and validates
  the IP is public *at connect time*, not just when the URL is parsed —
  see `internal/modules/cards/fetchimage.go`. Any failure (blocked
  address, wrong content-type, oversized body) just means the card renders
  without a photo; it never fails generation over a URL we don't control.
- ID card PDFs (with an embedded QR code, see `internal/qrtoken`) are
  generated fresh on every request and never persisted, matching the "no
  server-side photo/PDF storage" decision from the project's Phase 0
  scoping — `cards.Service.generate` is the one place that assembles one.
- A scanner-bot device never logs in as an org member — it authenticates
  every request with an opaque key (paired once via OTP, stored only as a
  hash) sent as `X-Device-Key`, not a session cookie. Because validating it
  means a DB lookup, its middleware lives as a method on `devices.Service`
  (`RequireDevice`) rather than in the generic, stateless
  `internal/middleware` package — see `internal/modules/devices/middleware.go`.
  The scanner UI calls this API directly from the browser (the key lives in
  `localStorage`, so there's no server-side proxy to forward it through),
  which is why `X-Device-Key` is in the CORS `AllowHeaders` list in
  `cmd/api/main.go` alongside `Content-Type`.
- Two modules can each own one route on the same URL group without either
  importing the other: `devices.RegisterScannerRoutes` mounts the
  device-authenticated `/scanner` group and its own `/me` route, returns
  the `fiber.Router` group, and `attendance.RegisterScanRoute` mounts
  `/scan` onto that same returned group. Only `cmd/api/main.go` calls both.
- `attendance.Service.BuildRoster` is the single source of truth for "who
  was expected on which date, and did they show up" — it's not just a scan
  log query, it includes people who were never scanned at all (`Attended:
  false`). `analytics.Service` never touches `attendance_records` directly;
  it calls `BuildRoster` (whole-event, and once per sub-event) and rolls
  the same rows up into summary/daily numbers, so "expected" and
  "attended" can't drift into two different definitions across modules.

## Commands

| command              | what it does                                   |
| -------------------- | ----------------------------------------------- |
| `make dev`            | run the API with live env from `.env`           |
| `make build`           | build `bin/api`                                |
| `make seed`             | create the first org + super_admin user      |
| `make migrate-up/down`   | apply/rollback one migration               |
| `make migrate-new name=` | scaffold a new migration pair             |
| `make sqlc-generate`      | regenerate typed query code               |
| `make lint`                | `go vet` + `gofmt -l`                   |
