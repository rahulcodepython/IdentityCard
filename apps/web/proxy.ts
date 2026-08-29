import { NextResponse, type NextRequest } from "next/server"
import { getSessionCookie } from "better-auth/cookies"

// Fast, DB-free cookie-presence check — bounces an obviously-unauthed
// visitor away from /dashboard at the edge, before the heavier
// getSession()/token() round trip in components/session-provider.tsx
// ever runs. Not the only gate: a present-but-expired/invalid cookie
// still passes this check and gets caught client-side once
// SessionProvider resolves to "unauthenticated".
export async function proxy(request: NextRequest) {
    const sessionCookie = getSessionCookie(request)
    if (!sessionCookie) {
        return NextResponse.redirect(new URL("/login", request.url))
    }
    return NextResponse.next()
}

export const config = {
    matcher: ["/dashboard/:path*"],
}
