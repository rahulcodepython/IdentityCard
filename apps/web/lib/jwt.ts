// Client-side peek at a JWT's payload — no signature check, since this
// never makes a trust decision (that's Go's job, verifying against
// better-auth's JWKS — see apps/server/internal/pkg/jwt). Only used to
// read the organizationId/role claims (baked in by definePayload, see
// apps/web/lib/auth.ts) into the session store without a second network
// round trip.
export function decodeJwtPayload(token: string): { organizationId: string | null; role: string | null } {
    try {
        const [, payload] = token.split(".")
        const json = atob(payload.replace(/-/g, "+").replace(/_/g, "/"))
        const claims = JSON.parse(json) as { organizationId?: string | null; role?: string | null }
        return { organizationId: claims.organizationId ?? null, role: claims.role ?? null }
    } catch {
        return { organizationId: null, role: null }
    }
}
