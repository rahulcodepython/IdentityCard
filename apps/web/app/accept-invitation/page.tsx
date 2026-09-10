"use client"

import { Suspense, useState } from "react"
import Link from "next/link"
import { useRouter, useSearchParams } from "next/navigation"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { authClient } from "@/lib/auth-client"
import { useSessionStore } from "@/store/session.store"

function AcceptInvitationCard() {
    const router = useRouter()
    const searchParams = useSearchParams()
    const invitationId = searchParams.get("id")
    const status = useSessionStore((s) => s.status)
    const setToken = useSessionStore((s) => s.setToken)
    const [error, setError] = useState<string | null>(null)
    const [pending, setPending] = useState(false)

    async function accept() {
        if (!invitationId) return
        setPending(true)
        setError(null)
        const { error: acceptError } = await authClient.organization.acceptInvitation({ invitationId })
        if (acceptError) {
            setPending(false)
            setError(acceptError.message ?? "Couldn't accept this invitation.")
            return
        }
        // The org this invite just joined isn't necessarily the active
        // one yet — switch to it explicitly before minting a token.
        const { data: invitation } = await authClient.organization.getInvitation({ query: { id: invitationId } })
        if (invitation?.organizationId) {
            await authClient.organization.setActive({ organizationId: invitation.organizationId })
        }
        const { data } = await authClient.token()
        if (data?.token) {
            setToken(data.token, invitation?.organizationId ?? null, null)
        }
        router.push("/dashboard")
    }

    if (!invitationId) {
        return (
            <CardContent>
                <p className="text-sm text-destructive">This invitation link is missing its id.</p>
            </CardContent>
        )
    }

    if (status === "idle" || status === "loading") {
        return (
            <CardContent>
                <p className="text-sm text-muted-foreground">Loading…</p>
            </CardContent>
        )
    }

    if (status === "unauthenticated") {
        return (
            <CardContent className="flex flex-col gap-4">
                <p className="text-sm text-muted-foreground">
                    Sign in with the email this invitation was sent to, then come back to this link.
                </p>
                <Button render={<Link href="/login" />}>Sign in</Button>
            </CardContent>
        )
    }

    return (
        <CardContent className="flex flex-col gap-4">
            {error && <p className="text-sm text-destructive">{error}</p>}
            <Button onClick={() => void accept()} disabled={pending}>
                {pending ? "Joining…" : "Accept invitation"}
            </Button>
        </CardContent>
    )
}

export default function AcceptInvitationPage() {
    return (
        <div className="flex min-h-svh items-center justify-center p-6">
            <Card className="w-full max-w-sm">
                <CardHeader>
                    <CardTitle>Join organization</CardTitle>
                    <CardDescription>Accept your invitation to start collaborating.</CardDescription>
                </CardHeader>
                <Suspense fallback={null}>
                    <AcceptInvitationCard />
                </Suspense>
            </Card>
        </div>
    )
}
