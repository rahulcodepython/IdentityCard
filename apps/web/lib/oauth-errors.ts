// Maps the `?error=` codes apps/server's /auth/google/callback redirects
// with (see internal/modules/auth/handler.go's oauthErrorRedirect) to a
// message for the login/register pages' inline alert.
const OAUTH_ERROR_MESSAGES: Record<string, string> = {
  oauth_not_configured:
    "Google sign-in isn't set up yet — use email and password instead.",
  google_no_account:
    "No account found for that Google email. Create one first.",
  account_exists:
    "An account with that email already exists — sign in instead.",
  oauth_failed: "Something went wrong with Google sign-in. Please try again.",
  invalid_state: "Your sign-in attempt expired. Please try again.",
}

export function oauthErrorMessage(code?: string): string | null {
  if (!code) return null
  return OAUTH_ERROR_MESSAGES[code] ?? "Something went wrong. Please try again."
}
