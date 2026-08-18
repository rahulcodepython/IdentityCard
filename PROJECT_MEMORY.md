# Project Memory — Identity Card + Attendance Platform

This file is a consolidated knowledge dump of the project as of the end of
**Phase 6**. It exists so a future session (human or AI) can get full
context fast without re-reading every file or re-deriving decisions that
were already made deliberately. It is not auto-generated documentation —
treat it as the durable record of *why* things are the way they are.

For narrower, always-current documentation see:
- `README.md` — quick start
- `apps/server/README.md` — backend conventions (kept up to date every phase)
- `apps/web/README.md` — frontend conventions

If those and this file ever disagree on a fact about the *current* code,
trust the code and the per-app READMEs — this file's job is decisions and
history, not a live mirror of the source.

## What this product is

A SaaS platform for organizations to run events (conferences, campuses,
offices, one-day meetups) with digital ID cards and QR-based attendance
tracking, replacing physical badge printing. Full spec came from the user
in the very first message of the project; nothing about scope was
invented — every module below maps to an explicit paragraph in that spec.

Core loop, end to end, all working as of Phase 6:
**pick a plan → register (creates org + super_admin) → create an event
(+ optional sub-events) → add people (manual/CSV/public form) → publish
(emails everyone an ID card PDF+QR) → pair a scanner device → scan cards
at the door (records entry/exit, early/on-time/late) → view analytics
(attended/absent, per-day, per-sub-event, export CSV, download chart
PNGs).**

## Stack

- **`apps/web`** — Next.js 16 (App Router, RSC), shadcn/ui (`base-nova`
  style, `remixicon` icons, `@base-ui/react` primitives — **not** Radix),
  react-hook-form + zod, pnpm workspace member.
- **`apps/server`** — Go 1.26 + Fiber v2, Postgres via `sqlc` (typed SQL
  codegen, not an ORM) + `pgx/v5`, Redis (refresh tokens), MinIO/S3 (org
  logos only), SMTP via `gomail.v2` (Mailhog locally).
- **`infra`** — docker-compose: postgres, redis, minio, mailhog.
- Repo is a monorepo; this was a deliberate Phase 0 choice over two repos
  or a flat layout, specifically so a single PR could touch both frontend
  and backend for one feature.

## Repo layout

```
apps/server/
  cmd/api/main.go       — the only file that wires every module together;
                           deliberately the one place cross-module
                           interfaces get their concrete types
  cmd/seed/main.go       — creates one org + super_admin for local dev
  internal/
    auth/                — JWT issue/parse, password hashing, cookie helpers
    config/               — env loading, fails fast on missing required vars
    db/                    — pgx pool, WithTx helper, pg error helpers, migrations/, sqlc/
    httpx/                  — response envelope, typed error helpers, bind+validate
    mailer/                  — SMTP + attachments
    middleware/               — RequireAuth/RequireRole (JWT, stateless)
    qrtoken/                   — signs/parses the QR code payload (own secret, not JWTSecret)
    redis/, storage/, timeutil/
    modules/
      organizations/  members/  plans/            (Phase 0-1)
      events/  subevents/                          (Phase 2)
      people/  forms/                              (Phase 3)
      cards/                                        (Phase 4)
      devices/  attendance/                          (Phase 5)
      analytics/                                      (Phase 6)
      auth/                                           (Phase 0, register added Phase 1)

apps/web/
  app/
    (auth)/login, (auth)/register      — cookie-auth flows
    plans/                              — public plan catalog
    forms/[token]/                      — PUBLIC unauthenticated sign-up form
    pair/, scanner/                     — PUBLIC, device-key auth (not cookie), localStorage
    dashboard/                          — everything behind the httpOnly session cookie
      events/[id]/{edit,people,forms,subevents,analytics}/...
      devices/, settings/
  lib/
    api/<module>.ts     — server-only fetch wrappers (cookie-forwarding)
    validation/<module>.ts — zod schemas mirroring each Go module's DTOs by hand
    device-client.ts     — the ONE client-side (non-"server-only") API client;
                            used only by /pair and /scanner
```

21 migrations, 12 Go modules, 26 Next.js routes as of Phase 6.

## Foundational decisions (made in Phase 0, still binding)

These came from an explicit clarifying round before any code was written —
don't relitigate them without a reason:

- **Multi-tenancy**: every tenant-owned table carries `organization_id`;
  repository functions that read tenant data always take `organizationID`
  explicitly. No "get all rows" query without an org scope exists.
- **A user account is global**, can belong to multiple orgs via
  `organization_members` (roles as a `text[]`, e.g. `super_admin` can also
  hold `scanner`). Exactly one `super_admin` per org, enforced by a
  partial unique index.
- **Roles**: `super_admin` (only one; exclusive control over
  devices/roles — see Phase 5), `admin` (shares event/people/forms/cards
  management with super_admin), `scanner` (a *device*, not a human login —
  see below).
- **Billing**: Razorpay is the intended real gateway but was never
  integrated — "select plan" just assigns it (`subscriptions.status =
  'active'`, no payment call). This is a known, explicit gap (roadmap item
  8 — dynamic/recurring pricing engine — was never built).
- **Storage**: nothing is persisted server-side except org logos (MinIO).
  User photos are external URLs, submitted by whoever adds the person.
  Generated ID card PDFs are **never stored** — always rendered fresh on
  request. CSV imports are parsed transiently and discarded.
- **Auth**: Go-issued JWT (access ~15min + rotating refresh), httpOnly
  secure cookie set by the Go API, forwarded by Next.js Server
  Components/Actions. The **one exception** is scanner-bot devices (Phase
  5), which use an opaque key in `localStorage` and call the Go API
  directly from client-side JS — this is why `NEXT_PUBLIC_API_BASE_URL`
  and the `X-Device-Key` CORS header exist.
- **DB access**: `sqlc` (not GORM) — raw SQL in `.sql` files, typed Go
  generated. Migrations via `golang-migrate`, plain numbered up/down SQL.
- **Contract sync**: zod schemas in `apps/web/lib/validation/*` are
  **hand-maintained** to mirror each Go module's DTOs — no OpenAPI
  codegen. If you change a Go DTO, you must remember to update the
  matching zod file; nothing enforces this automatically.

## Go backend conventions (apply these to any new module)

Full detail lives in `apps/server/README.md`; the short version:

1. **Per-module layering**: `routes.go` → `handler.go` (bind+validate) →
   `service.go` (business logic) → `repository.go` (sqlc queries). No
   cross-module repository calls for reads — go through the other
   module's `Service`. The one exception is an atomic multi-table write
   (e.g. `auth.Service.Register` creating user+org+subscription+membership
   together) — uses `db.WithTx` + short-lived tx-scoped repos via
   `dbgen.Queries.WithTx(tx)`.
2. **Every response is a typed struct** through `httpx.OK`; every request
   DTO goes through `httpx.BindAndValidate`. Never
   `map[string]interface{}`. The narrow, deliberate exceptions are binary
   responses (CSV exports, the org logo, generated PDFs) which set
   `Content-Type`/`Content-Disposition` directly.
3. **Breaking import cycles**: when module A needs a callback into module
   B, but B already depends on A, A defines a tiny interface for just the
   method it needs (e.g. `events.CardSender`, satisfied by
   `cards.Service`). Only `cmd/api/main.go` imports both concrete types.
4. **Nested sub-resources** (`subevents`/`people`/`forms` under `events`)
   read the parent's state through the parent's `Service`
   (`events.Service.GetContext`), and mount nested routes:
   `/events/:eventId/<resource>/...`.
5. **Public (no session) routes** live in the owning module's own
   `RegisterPublicRoutes`, entirely outside any `RequireAuth` group.
6. **SSRF discipline**: anything fetched from a user-supplied URL
   (attendee photo URLs, submitted through the *unauthenticated* public
   form) goes through a custom `DialContext` that re-resolves and
   validates the IP is public **at connect time** — see
   `cards/fetchimage.go`. This closes the DNS-rebinding TOCTOU gap a
   simple upfront URL check would leave open.
7. **The `people.Service.upsert` pattern**: three different ingestion
   paths (manual entry, CSV import, public-form submission) all funnel
   through one upsert keyed on `(event_id, email, mobile)` —
   `INSERT ... ON CONFLICT DO UPDATE ... RETURNING *, (xmax = 0) AS
   inserted`. That `xmax = 0` trick is how the code tells "new person" from
   "resubmission" without a separate existence check.
8. **Capacity-gated writes** (`event_forms.submissions_count`): charge
   atomically *before* the gated write (`UPDATE ... WHERE count <
   capacity RETURNING *`), refund with a plain decrement if the write
   turns out to be a no-op. Never lets the gated action happen without a
   reserved slot; tolerates a brief, self-correcting overcount instead.
9. **`attendance.Service.BuildRoster`** is the single source of truth for
   "who was expected on which date, and did they show up" — including
   people who were never scanned (`attended: false`). `analytics.Service`
   never queries `attendance_records` directly; it always goes through
   `BuildRoster` so "expected"/"attended" can't drift into two
   definitions.

## Frontend conventions

Full detail in `apps/web/README.md`. Short version:

- Every form: react-hook-form + zod via `@hookform/resolvers/zod`
  (exception: a few simple forms use RHF's native `register(name,
  {required})` validation instead of a zod resolver, when the payload
  needs light client-side massaging before it matches the wire type —
  e.g. converting an empty string to `undefined` for an optional number).
- `lib/api/<module>.ts` files are `"server-only"` — they only run in
  Server Components/Actions and forward the httpOnly cookie automatically
  via `apiFetch` (`lib/api/client.ts`).
- **Route Handlers exist only as narrow proxies** for things a Server
  Component/Action can't do: binary downloads a browser navigates to
  directly (CSV exports, the org logo, generated PDF cards) that need the
  cookie attached server-side, or non-JSON responses. Never used as a
  general BFF layer.
- `qr-scanner` npm package (not hand-rolled camera code) for the scanner
  UI's camera decoding.
- Charts (`dataviz` skill) are hand-rolled inline SVG, not a charting
  library — the app only needed one grouped-bar-chart shape used four
  times. Core non-negotiables were kept (fixed categorical color order,
  one axis, status colors reserved + always paired with a text legend,
  data table alongside every chart) using the app's existing shadcn CSS
  tokens (`--primary`, `--destructive`, `--muted-foreground`) rather than
  inventing a new palette.
- shadcn uses `@base-ui/react` primitives, not Radix — the polymorphic
  prop is `render={<Link .../>}`, not `asChild`.

## Data model (high level)

```
organizations, users, organization_members (roles: text[])
plans, subscriptions (stubbed billing)
events (venue, kind: established|flash, status: draft|published)
  event_days (one row per active date — handles flash/multi-day/selective dates uniformly)
sub_events
  sub_event_days (empty = inherits all of the parent event's days)
people (org_id, event_id; unique on event_id+email+mobile; joined_at nullable
        until publish backfills it for anyone added while still draft)
  people_sub_events (join table)
event_forms (public sign-up links: optional sub_event scope, optional capacity)
devices (scanner bots: OTP-paired, key stored only as a SHA-256 hash)
attendance_records (one row per person per calendar date scanned;
                     entry_at/status, exit_at/status; unique on person+date)
```

No table was added in Phase 6 (analytics) — it's a pure read/aggregation
layer over `attendance_records` + the roster-building logic.

## Phase-by-phase build log

| Phase | What it added |
|---|---|
| 0 | Monorepo scaffold, conventions, auth (login/me), orgs, RBAC middleware |
| 1 | Plan catalog, `POST /auth/register` (atomic org+user+subscription+membership create), `/plans` → `/register` flow |
| 2 | Events + sub-events CRUD, day-schedule model (`event_days`), publish flow, join-date backfill |
| 3 | People ingestion (manual/CSV/public form) via shared upsert core, public sign-up forms with capacity, CSV export |
| 4 | ID card PDF+QR generation (`go-pdf/fpdf` + `skip2/go-qrcode`), org logo upload (MinIO), SMTP email on publish, SSRF-safe photo fetch |
| 5 | Scanner-bot device pairing (OTP+key), `/pair` + `/scanner` public UI (`qr-scanner`), scan/verify/entry/exit logic, early/on-time/late classification |
| 6 | Attendance roster with absentees, analytics summary/daily rollups, filterable roster + CSV export, SVG charts with PNG "screenshot" export |

## Known gaps / deliberate scope cuts (not oversights — flagged at the time)

- **No real payment integration.** Plan selection is a stub. Roadmap item
  8 (dynamic/recurring pricing engine with storage-based cost escalation)
  was never started.
- **No landing page** beyond the bare Next.js default at `/` (which just
  redirects into the authed dashboard). Roadmap item 7.
- **No transactions/invoices page** for orgs to view billing history —
  doesn't exist yet since there's no real billing.
- **Card sending is fire-and-forget**, not a durable queue — a
  per-recipient failure on publish is logged and skipped, not retried.
  Redis is already in the stack, so a real queue is the natural next step
  if this becomes a problem at scale.
- **Single timezone assumed** — entry/exit "early/on-time/late"
  classification compares against server-local time; there's no
  per-organization timezone concept.
- **No overnight event support** — `event_days`/`sub_event_days` require
  `exit_time > entry_time` on the same calendar date.
- **Long multi-week events** will render cramped/unreadable daily charts —
  no pagination or zooming was built for the analytics daily view.
- **Nothing in this app has been run end-to-end against a live browser or
  Docker stack** — every phase was verified via `go build`/`vet`/`test`
  and `pnpm typecheck`/`lint`/`build` only. The sandbox this was built in
  had no Docker access and no real browser. Treat the whole thing as
  "compiles and typechecks cleanly, logic reviewed carefully" rather than
  "manually clicked through" — especially the camera-based `/scanner` flow
  and the SMTP/MinIO integration, which are the parts most likely to have
  an environment-specific surprise on first real run.

## If you're picking this up fresh

1. Read `apps/server/README.md` and `apps/web/README.md` first — they're
   the living conventions docs, updated every phase.
2. `docker compose up -d` in `infra/`, then `make migrate-up && make seed
   && make dev` in `apps/server`, then `pnpm dev` in `apps/web` — this is
   the first time the full stack would actually run live end-to-end.
3. Mailhog UI at `localhost:8025` to see emailed ID cards without a real
   inbox.
4. The two unstarted roadmap items (billing/pricing engine, landing +
   transactions page) are the natural next phases if continuing the
   original plan.
