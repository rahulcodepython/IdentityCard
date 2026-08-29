# apps/web

Next.js (App Router) frontend: shadcn/ui, react-hook-form + zod, Server
Components talking to `apps/server` directly. Owns auth entirely (see
below) — `apps/server` is a pure resource server that trusts this app's
JWTs.

## Setup

```bash
cp .env.example .env.local   # fill in BETTER_AUTH_SECRET, DATABASE_URL
bun install                  # from the repo root, or here — workspaces resolve either way
bun dev                      # http://localhost:3000, expects apps/server on :8080
```

`DATABASE_URL` points at the same Postgres `apps/server` uses — this app
and the Go API share one database (see `apps/server/internal/db/migrations/
000035_better_auth_schema.up.sql`).

## Auth

Auth lives entirely here now, via [better-auth](https://better-auth.com)
(`lib/auth.ts` server config, `lib/auth-client.ts` client) — plugins:
`jwt` (bearer tokens for `apps/server`), `organization` (custom
`super_admin`/`admin`/`scanner` roles, see `lib/auth-access-control.ts`),
`emailOTP` (the one credential — passwordless), `twoFactor` (TOTP,
`allowPasswordless: true`, enrollment forced on first sign-in — see
`components/verification.tsx`). `proxy.ts` does a fast cookie-presence
check at the edge; `components/session-provider.tsx` calls
`getSession()`/`token()` exactly once per tab and writes the result into
`store/session.store.ts` — nothing else should call those again.
`organizationId`/`role` are baked into the JWT at issue time
(`definePayload` in `lib/auth.ts`) and decoded client-side
(`lib/jwt.ts`) rather than re-fetched, so switching the active org (see
the switcher in `components/app-sidebar.tsx`) only needs a fresh
`token()` call, not a second `getSession()`.

A new organization is never created from a standalone form — only as a
side effect of a successful plan purchase (`lib/actions/checkout.ts`'s
`purchasePlan` Server Action; see `app/select-plan`), gated by
`organizationLimit: 1` in `lib/auth.ts`. Plans/checkout go through
[PayKit](https://usepaykit.dev) (`lib/paykit.ts`) — currently a
hand-written manual provider (`lib/paykit-manual-provider.ts`, instant
success, no live gateway yet); swapping in a real one later only touches
that file. Go's `plans`/`subscriptions` tables stay the system of record
for what an org purchased and its event quota.

## Conventions

- Every form: `react-hook-form` + a zod schema via
  `@hookform/resolvers/zod`. Schemas live in `schema/<module>.types.ts`
  and mirror the matching Go module's request/response DTOs 1:1 — when a
  backend field changes, update the schema here by hand (no codegen yet).
- `lib/api/<module>.ts`: thin typed fetch wrappers per backend module,
  server-only (mint a fresh bearer JWT per call via `auth.api.getToken`,
  see `lib/api/client.ts`). `lib/client-api/<module>.ts` is the client-side
  equivalent, reading the token from `store/session.store.ts` (see
  `react-query/client.ts`'s interceptor — it also retries once with a
  freshly minted token on a 401, since the JWT is short-lived by design).
  Every response is parsed through its zod schema before being trusted.
- Route Handlers are only for a true BFF proxy need — e.g.
  `app/dashboard/events/[id]/people/export/route.ts`, which a
  browser-navigated download link can't attach a bearer token to itself
  and which returns CSV, not the JSON `apiFetch` expects. Same
  fresh-token-per-request pattern as `lib/api/client.ts`.
- A file upload (CSV import) posts a `FormData` body through the same
  `apiFetch` — it skips the JSON `Content-Type` automatically when the body
  is `FormData` so the browser can set its own multipart boundary. Native
  `<input type="file">` + `useActionState`, not react-hook-form, since
  there's nothing to client-validate before the file reaches the server.
- shadcn/ui components: `npx shadcn@latest add <component>` (see
  `components.json` for the configured style/aliases).
