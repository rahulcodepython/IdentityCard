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

Core loop, end to end:
**register (email OTP + forced TOTP enrollment) → pick a plan (creates
org + makes you super_admin, see the better-auth rewrite below) →
create an event (+ optional sub-events) → add people (manual/CSV/public
form) → publish (emails everyone an ID card PDF+QR) → pair a scanner
device → scan cards at the door (records entry/exit, early/on-time/late)
→ view analytics (attended/absent, per-day, per-sub-event, export CSV,
download chart PNGs).**

## Stack

- **`apps/web`** — Next.js 16 (App Router, RSC), shadcn/ui (`base-nova`
  style, `remixicon` icons, `@base-ui/react` primitives — **not** Radix),
  react-hook-form + zod, bun workspace member.
- **`apps/server`** — Go 1.26 + Fiber v2, Postgres via `sqlc` (typed SQL
  codegen, not an ORM) + `pgx/v5`, MinIO/S3 (org logos, images, signatures),
  SMTP via `gomail.v2` (Mailhog locally).
- **`infra`** — docker-compose: postgres, minio, mailhog.
- Repo is a monorepo; this was a deliberate Phase 0 choice over two repos
  or a flat layout, specifically so a single PR could touch both frontend
  and backend for one feature.

## Repo layout

```
apps/server/
  cmd/server/main.go    — the only file that wires every module together;
                           deliberately the one place cross-module
                           interfaces get their concrete types
  cmd/seed/main.go       — creates one org + super_admin for local dev
  internal/
    pkg/jwt/               — verifies bearer JWTs against better-auth's JWKS (no issuing — see below)
    config/               — env loading, fails fast on missing required vars
    db/                    — pgx pool, WithTx helper, pg error helpers, migrations/, sqlc/
    httpx/                  — response envelope, typed error helpers, bind+validate
    mailer/                  — SMTP + attachments
    middlewares/              — RequireAuth/RequireRole/RequireOrganization (JWT, stateless)
    qrtoken/                   — signs/parses the QR code payload (own secret, unrelated to auth)
    redis/, storage/, utils/
    modules/
      organizations/  plans/                        (Phase 0-1; members/ removed Phase 7,
                                                       see below)
      events/  subevents/                          (Phase 2)
      people/  forms/                              (Phase 3)
      cards/                                        (Phase 4)
      devices/  attendance/                          (Phase 5)
      analytics/                                      (Phase 6)

  Note: the Go server structure uses vertical slices under internal/features/<name>/
  where each feature has its own routes, handler, service, and repository. Each feature
  exposes a single unified `RegisterRoutes(public fiber.Router, protected fiber.Router)`
  method, cleanly separating public endpoints from auth-protected ones.

apps/web/
  app/
    api/
      auth/totp/{prepare,complete,verify}/ — Next.js API routes for 2FA TOTP onboarding & login
      organization/setup/                 — Next.js API route for initial organization setup
      auth/[...all]/                       — better-auth handler
    login, register                        — email-OTP + TOTP flows
    dashboard/                             — everything gated on an active organization
      billing/, devices/, events/, members/, settings/
    accept-invitation/                     — accept a member invite
    forms/[token]/                         — PUBLIC unauthenticated sign-up form
    pair/, scanner/                        — PUBLIC, device-key auth, localStorage
  components/                              — strictly organized into 10 purpose-named directories:
    auth/                                  — authentication forms, OTP inputs, QR views
    billing/                               — plan cards, pricing tables, subscribe dialog
    common/                                — shared SVG Logo, Icon registry, StatusBadge
    dialogs/                               — reusable ConfirmDeleteDialog, FormDialog shells
    marketing/                             — landing page showcase sections (Features)
    navigation/                            — AppSidebar, NavMain, BreadcrumbSetter
    providers/                             — QueryProvider, SessionProvider, ThemeProvider
    schedule/                              — event calendar, selective/recurring schedule pickers
    table/                                 — TanStack DataTable with search/sort/filter/visibility
    ui/                                    — base design system primitives (shadcn base-nova)
  react-query/
    client.ts                              — Axios client connecting to Go API (port 8000), synchronous
                                             token injection from Zustand, single-flight 401 refresh
    query-keys.ts                          — centralized query key factories
  store/
    session.store.ts                       — persistent Zustand store (sessionStorage) caching user,
                                             activeOrganizationId, and bearer JWT
  schema/<module>.types.ts                 — zod schemas mirroring each Go module's DTOs
  lib/auth.ts, lib/auth-client.ts          — better-auth server/client configuration
  lib/device-client.ts                     — client-side API client for /pair and /scanner (device-key)
```

21+ migrations (000035/000036 added Phase 7 for the better-auth cutover), 12 Go modules, Next.js routes per `apps/web/app/`.

## Foundational decisions (made in Phase 0, still binding)

These came from an explicit clarifying round before any code was written —
don't relitigate them without a reason:

- **Multi-tenancy**: every tenant-owned table carries `organization_id`;
  repository functions that read tenant data always take `organizationID`
  explicitly. No "get all rows" query without an org scope exists.
- **A user account is global**, can belong to multiple orgs via
  better-auth's `member` table (Phase 7 rewrite) — one role per
  membership now, not a `text[]`. Exactly one `super_admin` per org,
  enforced by better-auth's `creatorRole` (the org creator always gets
  it) rather than a DB constraint — see `apps/web/lib/auth.ts`.
- **Roles**: `super_admin` (only one per org; exclusive control over
  devices/roles — see Phase 5), `admin` (shares event/people/forms/cards
  management with super_admin), `scanner` (originally "a device, not a
  human login" — still true for scanner-bot devices, which use their own
  opaque-key auth, but Phase 7 also made `scanner` an invitable human
  member role in its own right; the two are separate mechanisms that
  happen to share a name).
- **Billing**: Razorpay is the intended real gateway but was never
  integrated — "select plan" just assigns it (`subscriptions.status =
  'active'`, no payment call). This is a known, explicit gap (roadmap item
  8 — dynamic/recurring pricing engine — was never built).
- **Storage**: nothing is persisted server-side except org logos (MinIO).
  User photos are external URLs, submitted by whoever adds the person.
  Generated ID card PDFs are **never stored** — always rendered fresh on
  request. CSV imports are parsed transiently and discarded.
- **Auth**: as of the better-auth rewrite (see the phase log below), auth
  lives entirely in `apps/web` via [better-auth](https://better-auth.com)
  — email OTP (passwordless) + forced TOTP enrollment, Google OAuth, an
  `organization` plugin with custom `super_admin`/`admin`/`scanner`
  roles. `apps/server` is a pure resource server: it verifies a bearer
  JWT (short-lived, org id/role baked into its claims) against
  better-auth's JWKS endpoint and issues nothing itself. Both apps share
  one Postgres database — better-auth owns `user`/`session`/`organization`/
  `member`/etc., Go owns everything business-specific. The **one
  exception** to bearer-token auth is scanner-bot devices (Phase 5),
  which use an opaque key in `localStorage` and call the Go API directly
  from client-side JS — this is why `NEXT_PUBLIC_API_BASE_URL` and the
  `X-Device-Key` CORS header exist. See `apps/web/README.md`'s Auth
  section for the full picture.
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
   `cards.Service`). Only `cmd/server/main.go` imports both concrete types.
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
10. **Unified Route Registration**: Every feature exposes exactly one
    `RegisterRoutes(public fiber.Router, protected fiber.Router)` method.
    No feature may have multiple scattered registration functions. Public
    routes are mounted strictly on the `public` router and auth-gated routes
    on the `protected` router.

## Frontend conventions

Full detail in `apps/web/README.md`. Short version:

- Every form: react-hook-form + zod via `@hookform/resolvers/zod`
  (exception: a few simple forms use RHF's native `register(name,
  {required})` validation instead of a zod resolver, when the payload
  needs light client-side massaging before it matches the wire type —
  e.g. converting an empty string to `undefined` for an optional number).
- **No Server Actions**: All server actions have been removed. Frontend
  mutations use standard Next.js route handlers (`/api/auth/totp/*`,
  `/api/organization/setup`) or direct calls to the Go backend
  (`POST /plans/purchase` on port 8000).
- **Session & Bearer Token Caching**: Session and JWT bearer tokens are
  persisted in `sessionStorage` via Zustand's `persist` middleware
  (`useSessionStore`). Tokens are minted/fetched exactly once on
  login/registration and cached. The Axios request interceptor
  (`react-query/client.ts`) synchronously reads this cached token,
  eliminating per-request network calls to `authClient.token()`. On 401s,
  `getRefreshedToken()` deduplicates parallel calls into a single flight.
- **Component Organization**: Every component under `components/` lives in
  a purpose-named subfolder with barrel `index.ts` re-exports:
  - `components/providers/` (`QueryProvider`, `SessionProvider`, `ThemeProvider`)
  - `components/navigation/` (`AppSidebar`, `NavMain`, `BreadcrumbSetter`)
  - `components/billing/` (`PlanCard`, `Pricing`, `PricingClient`, `SubscribeDialog`)
  - `components/marketing/` (`Features`)
  - `components/dialogs/` (`ConfirmDeleteDialog`, `FormDialog`)
  - `components/table/` (`DataTable`)
  - `components/common/` (`Icon`, `Logo`, `StatusBadge`)
  - `components/auth/`, `components/schedule/`, `components/ui/`
- **Route Handlers**: Used for TOTP 2FA flow (`prepare`, `complete`, `verify`),
  organization setup, and binary downloads.
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

better-auth owns `user`, `session`, `account`, `verification`,
`organization`, `member` (one role per row now, not `text[]`),
`invitation`, `jwks`, `twoFactor` — see
`apps/server/internal/db/migrations/000035_better_auth_schema.up.sql`.
Everything below is still Go's:

```
plans, subscriptions (purchased directly via Go API `POST /plans/purchase` —
  PayKit was removed; live gateway integration remains open)
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
| 7 | Replaced all Go-owned auth with better-auth (email OTP + forced TOTP, Google OAuth, `organization` plugin with custom roles) on a shared Postgres DB; `apps/server` rebuilt as a pure JWKS-verifying resource server; org creation moved behind plan purchase; org switcher + real (not mocked) member invite/role/remove UI |
| 8 | Refactored Go server route registration to a unified `RegisterRoutes(public, protected fiber.Router)` per feature; eliminated all Next.js Server Actions in favor of standard API routes (`/api/auth/totp/*`, `/api/organization/setup`) and direct Go API calls (`/plans/purchase`); implemented persistent Zustand session/token caching (`sessionStorage`) with zero-overhead Axios interceptor injection and single-flight 401 refresh; pruned deprecated packages (`paykit`, `nodemailer`) and obsolete files (`jwt.ts`, `mailer.ts`, `roles.ts`); reorganized all 17 loose components into 10 purpose-named directories with barrel index files; designed Postgres Transactional Outbox and centralized `RequireActiveBilling` middleware architecture. |

## Known gaps / deliberate scope cuts (not oversights — flagged at the time)

- **No real payment gateway.** Plan purchase routes directly to the Go
  backend (`POST /plans/purchase`) which assigns an active subscription
  instantly for testing without charging a card. Swapping in Razorpay or
  Stripe remains open for production readiness.
- **Card sending is fire-and-forget**, not a durable queue — a
  per-recipient failure on publish is logged and skipped, not retried.
  The designed Transactional Outbox pattern with Resend integration is
  planned to address this durability gap.
- **Single timezone assumed** — entry/exit "early/on-time/late"
  classification compares against server-local time; there's no
  per-organization timezone concept.
- **No overnight event support** — `event_days`/`sub_event_days` require
  `exit_time > entry_time` on the same calendar date.
- **Long multi-week events** will render cramped/unreadable daily charts —
  no pagination or zooming was built for the analytics daily view.
- **Nothing in this app has been run end-to-end against a live browser or
  Docker stack** — every phase was verified via `go build`/`vet`/`test`
   and `bun typecheck`/`lint`/`build` only. The sandbox this was built in
   had no Docker access and no real browser. Treat the whole thing as
   "compiles and typechecks cleanly, logic reviewed carefully" rather than
   "manually clicked through" — especially the camera-based `/scanner` flow
   and the SMTP/MinIO integration, which are the parts most likely to have
   an environment-specific surprise on first real run.

## If you're picking this up fresh

1. Read `apps/server/README.md` and `apps/web/README.md` first — they're
   the living conventions docs, updated every phase.
2. `docker compose up -d` in `infra/`, then `make dev` in `apps/server`
   (migrates on boot — no separate `make migrate-up` needed against a
   fresh DB — then `make seed`), then `bun dev` in `apps/web`. The
   better-auth rewrite (Phase 7) was verified live end-to-end this way —
   register → email OTP → forced TOTP enroll → select a plan (creates
   the org) → dashboard — see its phase-log entry above.
3. Mailhog UI at `localhost:8025` to see emailed ID cards (and OTP/invite
   emails) without a real inbox — point `SMTP_HOST`/`SMTP_PORT` at it in
   both apps' env files for local dev; the checked-in `apps/server/.env`
   /`apps/web/.env.local` currently point at a real Resend account
   instead, which will actually send mail if you run against them as-is.
4. The two unstarted roadmap items (billing/pricing engine, landing +
   transactions page) are the natural next phases if continuing the
   original plan.

## Role & Persona
You are a Staff/Principal Software Engineer acting as a core technical lead on this project. Your goal is to write robust, maintainable, and production-ready code that minimizes technical debt and operational overhead.

## Engineering Principles & Coding Standards

1. **Clarification First (Zero Speculation & No Silent Assumptions):**
   - If requirements, interfaces, business rules, edge cases, or acceptance criteria are ambiguous, incomplete, or missing, **stop and ask targeted clarifying questions before generating code**.
   - Do not guess data shapes, business rules, or unseen file structures. State explicit doubts and get confirmation first rather than attempting speculative implementations.

2. **Anti-Redundancy & DRY Principles:**
   - Never write duplicate code, boilerplate loops, or reinvent existing native APIs, standard library functions, or project utilities.
   - Leverage abstractions responsibly without creating unnecessary layers.

3. **Simplicity Over Cleverness (KISS & YAGNI):**
   - Keep business logic, data flow, and authentication/authorization mechanisms explicit, linear, and straightforward.
   - Avoid premature optimization, hyper-abstract generic patterns where unnecessary, and deeply nested logic. Write code that is easy to reason about on day one.

4. **Solution Evaluation & Trade-offs:**
   - Before implementing complex solutions, evaluate the technical trade-offs (time/space complexity, memory footprint, maintainability, network cost).
   - If an existing native method, standard pattern, or simpler alternative exists, choose and implement the simpler, more optimal path. State the reason briefly if relevant.

5. **Idiomatic, Clean & Maintainable Code:**
   - Write self-documenting code with clear, descriptive naming conventions for variables, methods, and types.
   - Keep functions small, focused on a single responsibility, and easy to unit test.
   - Handle edge cases, nullability, and errors explicitly rather than suppressing them.

6. **Strict Type Safety & Zero Ambiguity:**
   - **TypeScript:** Never use `any` or `unknown` as an escape hatch. Use strict domain interfaces, unions, generics, or utility types.
   - **Go:** Avoid loose `interface{}` (`any`) and unstructured `map[string]interface{}`. Define strongly typed structs, strict concrete signatures, and small, idiomatic interfaces. Ensure zero type-assertion blindness.

7. **Active Garbage Collection & Dead Code Pruning (Boy Scout Rule):**
   - When modifying, updating, refactoring, or deleting features/models, prune all orphaned artifacts immediately in the same change.
   - Proactively delete unused variables, unreferenced helper methods, dead imports, obsolete types/models, and deprecated functions that are no longer called.
   - Do not leave "TODO: clean up later" or commented-out blocks of replaced logic. Leave the modified files cleaner than you found them.

8. **Eliminate AI Hallucination & Code Slop:**
   - Do not generate speculative imports, insecure workarounds, unnecessary helper functions, or low-quality boilerplate that obscures root intent.
   - Ensure all output is safe from common vulnerabilities (e.g., OWASP Top 10, sanitization oversights, unhandled memory leaks, concurrency races).

9. **Architectural & Style Consistency:**
   - Strictly adhere to the existing conventions, folder structure, design patterns, linting rules, and naming styles already established in this codebase.
   - When introducing changes or new modules, ensure they blend seamlessly with the existing system architecture.

10. **Production-Ready Deliverables:**
    - Provide complete, verifiable implementations for the requested scope.
    - Flag breaking changes, schema implications, or edge-case limitations proactively.