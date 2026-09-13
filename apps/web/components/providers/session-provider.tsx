"use client"

import { useEffect, useState } from "react"

import { authClient } from "@/lib/auth-client"
import { useSessionStore } from "@/store/session.store"
import z from "zod"

const SessionDataSchema = z.object({
    token: z.string(),
    user: z.object({
        id: z.string(),
        name: z.string(),
        email: z.string(),
        image: z.string().nullable().optional(),
    }),
    session: z.object({
        activeOrganizationId: z.string().nullable().optional(),
    }),
})

export function SessionProvider({ children }: { children: React.ReactNode }) {
    const isAuthenticated = useSessionStore((state) => state.isAuthenticated)
    const setSession = useSessionStore((state) => state.setSession)

    const [isLoading, setIsLoading] = useState(true)

    const getSession = async () => {
        try {
            const { data } = await authClient.getSession()
            const sessionData = SessionDataSchema.parse(data)

            setSession({
                token: sessionData.token,
                user: sessionData.user,
                activeOrgId: sessionData.session.activeOrganizationId ?? null,
            })
        } catch (err) {
            console.error("Failed to retrieve session:", err)
            return null
        } finally {
            setIsLoading(false)
        }
    }

    useEffect(() => {
        if (isAuthenticated) {
            setIsLoading(false)
            return
        }

        getSession()
    }, [])

    if (isLoading) {
        return <>
            Loading...
        </>
    }

    return children
}
