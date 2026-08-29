// better-auth's social sign-in redirects back to errorCallbackURL with an
// `?error=<code>` query param on failure (e.g. "signup_disabled",
// "account_not_linked") — humanize whatever code comes back rather than
// maintaining a hardcoded lookup, since the exact set of codes is
// better-auth's internal detail, not a public contract.
export function oauthErrorMessage(code?: string): string | null {
    if (!code) return null
    return `Google sign-in failed: ${code.replace(/_/g, " ")}. Please try again.`
}
