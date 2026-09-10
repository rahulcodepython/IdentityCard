"use client"

import { useEffect } from "react"

import { authClient } from "@/lib/auth-client"
import { decodeJwtPayload } from "@/lib/jwt"
import { useSessionStore } from "@/store/session.store"

// Calls better-auth's getSession()/token() exactly once per browser tab
// (guarded by the store's status, which starts "idle" and only this
// effect ever moves it out of "idle") and writes the result into the
// zustand session store — user data and the JWT bearer token used for
// every Go API call (see react-query/client.ts). Nothing else in the app
// should call getSession()/token() again; re-render/navigation within
// the same tab reads from the store instead.
export function SessionProvider({ children }: { children: React.ReactNode }) {
    const status = useSessionStore((s) => s.status)
    const setLoading = useSessionStore((s) => s.setLoading)
    const setSession = useSessionStore((s) => s.setSession)
    const setUnauthenticated = useSessionStore((s) => s.setUnauthenticated)

    useEffect(() => {
        if (status !== "idle") return
        setLoading()

        void (async () => {
            try {
                const { data: sessionData } = await authClient.getSession()
                if (!sessionData || !sessionData.user) {
                    setUnauthenticated()
                    return
                }

                let activeToken = ""
                let organizationId: string | null =
                    (sessionData.session as { activeOrganizationId?: string })?.activeOrganizationId ?? null
                let role: string | null = null

                try {
                    const { data: tokenData } = await authClient.token()
                    if (tokenData?.token) {
                        activeToken = tokenData.token
                        const decoded = decodeJwtPayload(activeToken)
                        organizationId = decoded.organizationId || organizationId
                        role = decoded.role || role
                    }
                } catch (tokenErr) {
                    console.warn("Failed to fetch JWT bearer token, proceeding with session:", tokenErr)
                }

                if (!organizationId) {
                    try {
                        const { data: orgs } = await authClient.organization.list()
                        if (orgs && orgs.length > 0) {
                            organizationId = orgs[0].id
                            await authClient.organization.setActive({ organizationId: orgs[0].id })
                            try {
                                const refetched = await authClient.token()
                                if (refetched.data?.token) {
                                    activeToken = refetched.data.token
                                    const decoded = decodeJwtPayload(activeToken)
                                    role = decoded.role || role
                                }
                            } catch {
                                // proceed with existing activeToken
                            }
                        }
                    } catch (orgErr) {
                        console.warn("Failed to list/set active organization:", orgErr)
                    }
                }

                setSession({
                    token: activeToken,
                    user: {
                        id: sessionData.user.id,
                        name: sessionData.user.name,
                        email: sessionData.user.email,
                        image: sessionData.user.image,
                    },
                    activeOrganizationId: organizationId,
                    role,
                })
            } catch (err) {
                console.error("SessionProvider unexpected failure:", err)
                setUnauthenticated()
            }
        })()
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [status])

    return children
}
