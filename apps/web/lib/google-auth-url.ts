// Not server-only: used by client-side Google sign-in buttons, which do a
// plain browser navigation straight to apps/server (same pattern as
// lib/device-client.ts's NEXT_PUBLIC_API_BASE_URL use) rather than a
// fetch through a Server Action, since the whole point is to leave this
// origin and come back via a redirect. Register intent needs nothing else
// up front — a Google signup collects its organization name afterward, on
// /onboarding.
const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080"

export function googleAuthUrl(intent: "login" | "register") {
  const params = new URLSearchParams({ intent })
  return `${API_BASE_URL}/auth/google/login?${params.toString()}`
}
