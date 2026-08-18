# apps/web

Next.js (App Router) frontend: shadcn/ui, react-hook-form + zod, Server
Components talking to `apps/server` directly.

## Setup

```bash
pnpm install   # from the repo root, or here — pnpm workspaces resolve either way
pnpm dev       # http://localhost:3000, expects apps/server on :8080
```

## Conventions

- Every form: `react-hook-form` + a zod schema via
  `@hookform/resolvers/zod`. Schemas live in `lib/validation/<module>.ts`
  and mirror the matching Go module's request/response DTOs 1:1 — when a
  backend field changes, update the schema here by hand (no codegen yet).
- `lib/api/<module>.ts`: thin typed fetch wrappers per backend module.
  Every response is parsed through its zod schema before being trusted.
- Auth is an httpOnly cookie set by the Go API — Server Components and
  Server Actions call the API directly and forward the cookie; no token
  ever reaches client-side JS. Route Handlers are only for a true BFF proxy
  need — e.g. `app/dashboard/events/[id]/people/export/route.ts`, which a
  browser-navigated download link can't attach the cookie to itself and
  which returns CSV, not the JSON `apiFetch` expects.
- A file upload (CSV import) posts a `FormData` body through the same
  `apiFetch` — it skips the JSON `Content-Type` automatically when the body
  is `FormData` so the browser can set its own multipart boundary. Native
  `<input type="file">` + `useActionState`, not react-hook-form, since
  there's nothing to client-validate before the file reaches the server.
- shadcn/ui components: `npx shadcn@latest add <component>` (see
  `components.json` for the configured style/aliases).
