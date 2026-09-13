"use client"

import { Suspense, useState } from "react"
import Link from "next/link"
import { Logo } from "@/components/common/logo"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import {
    Field,
    FieldDescription,
    FieldGroup,
    FieldSeparator,
} from "@/components/ui/field"
import { authClient } from "@/lib/auth-client"
import { RiGoogleFill } from "@remixicon/react"

interface AuthCardProps {
    title: string
    subtitle: string
    footerText: string
    footerLinkText: string
    footerLinkHref: string
    children: React.ReactNode
    showSocial?: boolean
}

export function AuthCard(props: AuthCardProps) {
    const [googlePending, setGooglePending] = useState(false)

    const continueWithGoogle = async () => {
        setGooglePending(true)
        await authClient.signIn.social({
            provider: "google",
            callbackURL: "/dashboard",
            errorCallbackURL: `${props.footerLinkHref === "/auth/register" ? "/auth/login" : "/auth/register"}?error=oauth_failed`,
        })
    }

    return (
        <div className="flex min-h-svh items-center justify-center p-6">
            <div className="w-full max-w-md">
                <Card className="py-0">
                    <CardContent className="p-6 md:p-8">
                        <FieldGroup>
                            <div className="flex flex-col items-center gap-2 text-center">
                                <Logo className="h-9 w-9" />
                                <h1 className="text-2xl font-bold">{props.title}</h1>
                                <p className="text-sm text-balance text-muted-foreground">
                                    {props.subtitle}
                                </p>
                            </div>

                            {props.children}

                            {
                                props.showSocial && <FieldGroup>
                                    <FieldSeparator className="*:data-[slot=field-separator-content]:bg-card">
                                        Or continue with
                                    </FieldSeparator>

                                    <Field>
                                        <Button
                                            variant="outline"
                                            type="button"
                                            disabled={googlePending}
                                            onClick={() => void continueWithGoogle()}
                                        >
                                            <RiGoogleFill />
                                            {googlePending ? "Redirecting…" : "Continue with Google"}
                                        </Button>
                                    </Field>

                                    <FieldDescription className="text-center">
                                        {props.footerText}{" "}
                                        <Link
                                            href={props.footerLinkHref}
                                            className="underline underline-offset-4 hover:text-foreground"
                                        >
                                            {props.footerLinkText}
                                        </Link>
                                    </FieldDescription>
                                </FieldGroup>
                            }
                        </FieldGroup>
                    </CardContent>
                </Card>
            </div>
        </div>
    )
}
