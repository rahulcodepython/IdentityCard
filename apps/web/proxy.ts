import { NextResponse, type NextRequest } from "next/server"
import { getSessionCookie } from "better-auth/cookies"

const AUTH_URL_PREFIX = "/auth"
const DASHBOARD_URL_PREFIX = "/dashboard"

export async function proxy(request: NextRequest) {
    const sessionCookie = getSessionCookie(request)

    if (request.nextUrl.pathname.startsWith(AUTH_URL_PREFIX) && sessionCookie) {
        return NextResponse.redirect(new URL("/dashboard", request.url))
    }

    if (request.nextUrl.pathname.startsWith(DASHBOARD_URL_PREFIX) && !sessionCookie) {
        return NextResponse.redirect(new URL("/auth/login", request.url))
    }

    return NextResponse.next()
}

export const config = {
    matcher: ["/dashboard/:path*", "/auth/:path*"],
}
