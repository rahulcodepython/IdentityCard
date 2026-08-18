import "server-only"
import { cookies } from "next/headers"

/**
 * Re-applies the `Set-Cookie` headers from a Go API response onto the
 * Next.js response, so the httpOnly session cookies the API issues
 * (see apps/server/internal/auth/cookies.go) reach the browser without a
 * token ever passing through client-side JS.
 *
 * Only callable from a Server Action or Route Handler — Next.js forbids
 * mutating cookies while rendering, so this is a no-op there.
 */
export async function forwardSetCookies(res: Response) {
  const setCookieHeaders = res.headers.getSetCookie()
  if (setCookieHeaders.length === 0) return

  const store = await cookies()
  for (const header of setCookieHeaders) {
    const [pair, ...attrParts] = header.split(";").map((part) => part.trim())
    const eqIndex = pair.indexOf("=")
    const name = pair.slice(0, eqIndex)
    const value = pair.slice(eqIndex + 1)

    const attrs: Record<string, string> = {}
    for (const attr of attrParts) {
      const [key, val] = attr.split("=")
      attrs[key.toLowerCase()] = val ?? "true"
    }

    try {
      store.set(name, value, {
        path: attrs.path,
        expires: attrs.expires ? new Date(attrs.expires) : undefined,
        httpOnly: "httponly" in attrs,
        secure: "secure" in attrs,
        sameSite:
          (attrs.samesite?.toLowerCase() as
            "lax" | "strict" | "none" | undefined) ?? "lax",
      })
    } catch {
      // Called during a render pass (e.g. a GET that happens to clear an
      // expired cookie) where cookie mutation isn't allowed — safe to skip.
    }
  }
}
