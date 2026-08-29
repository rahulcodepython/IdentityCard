# apps/server

Go/Fiber API: Postgres (via [sqlc](https://sqlc.dev) + pgx) and Redis.
Auth lives entirely in `apps/web` now (better-auth) — this API is a pure
resource server that verifies bearer JWTs against better-auth's JWKS
endpoint (`internal/middlewares/auth.go`, `internal/pkg/jwt`). It never
issues a session and has no login/register/OTP/TOTP/OAuth endpoints of
its own; `internal/db/migrations/000035_better_auth_schema.up.sql` is the
schema better-auth's CLI generated, pasted in so one `migrate` pipeline
sets up both apps' tables.

## Setup

```bash
cp .env.example .env          # defaults match infra/docker-compose.yml
cd ../../infra && docker compose up -d
cd ../apps/server
make dev                      # http://localhost:8080 — migrates on boot
make seed                     # creates the first org + super_admin login
```

Migrations (this app's and better-auth's) apply automatically on boot —
no separate `make migrate-up` step needed against a fresh database. `make
seed` writes directly into better-auth's tables (no password, the app is
passwordless) so there's an account to sign in with immediately; the
normal path is registering for real through `apps/web`'s `/register`
(email-OTP + forced TOTP enrollment), which is gated behind choosing a
plan before an organization exists at all (see `apps/web/lib/auth.ts`).

## Conventions

- One folder per feature under `internal/modules/<name>`: `routes.go` →
  `handler.go` (bind + validate) → `service.go` (business logic) →
  `repository.go` (sqlc queries). No cross-module repository calls for
  reads — those go through the other module's `Service`. The one exception
  is a write that must be atomic across modules: it uses `db.WithTx` and
  constructs short-lived, tx-scoped repositories directly via
  `dbgen.Queries.WithTx(tx)`.
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
  — see `events.CardSender`, satisfied by `cards.Service`. Only `cmd/server`
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
  `cmd/server/main.go` alongside `Content-Type`.
- Two modules can each own one route on the same URL group without either
  importing the other: `devices.RegisterScannerRoutes` mounts the
  device-authenticated `/scanner` group and its own `/me` route, returns
  the `fiber.Router` group, and `attendance.RegisterScanRoute` mounts
  `/scan` onto that same returned group. Only `cmd/server/main.go` calls both.
- `attendance.Service.BuildRoster` is the single source of truth for "who
  was expected on which date, and did they show up" — it's not just a scan
  log query, it includes people who were never scanned at all (`Attended:
  false`). `analytics.Service` never touches `attendance_records` directly;
  it calls `BuildRoster` (whole-event, and once per sub-event) and rolls
  the same rows up into summary/daily numbers, so "expected" and
  "attended" can't drift into two different definitions across modules.
- **Auth is entirely better-auth's** (`apps/web/lib/auth.ts` — jwt,
  organization, emailOTP, twoFactor plugins). This API never issues a
  session; `internal/middlewares/auth.go`'s `RequireAuth` only verifies
  the bearer JWT against better-auth's JWKS endpoint
  (`internal/pkg/jwt.Verifier`, fetched once and refreshed in the
  background, not per request). `organizationId`/`role` are read straight
  off the JWT's claims (baked in at issue time by better-auth's
  `definePayload`) — no DB round trip needed for RBAC.
- **Add-ons** (`internal/modules/addons`) are purchased on top of an
  existing subscription without changing plans — a separate concept from
  `plans.kind = 'flash'`, which is a standalone plan chosen at
  registration. Stubbed the same way `plans.CreateSubscription` is: a
  purchase just inserts an `active` `subscription_addons` row, no payment
  gateway call.
- **Plan upgrade** (`plans.Service.Upgrade`, `POST
  /plans/subscription/upgrade`) cancels the org's current active
  subscription and inserts a new one atomically (`db.WithTx`), keeping
  subscription history instead of mutating `plan_id` in place — mirrors
  how Stripe-style plan changes are usually modeled. Also stubbed, no
  payment call.
- `GET /analytics/overview` (dashboard home page's stats) is built the
  same way `analytics.Service.Summary`/`Daily` are: it calls
  `attendance.Service.BuildRoster` once per event and rolls the results
  up, rather than querying `attendance_records` or `people` directly, so
  org-wide numbers can't drift from the per-event definitions of
  "expected"/"attended". The daily trend is capped to the last 30 days
  (`overviewTrendDays`) for the same reason the per-event daily chart has
  no pagination yet — an unbounded trend line for a long-running org would
  be unreadable.

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
