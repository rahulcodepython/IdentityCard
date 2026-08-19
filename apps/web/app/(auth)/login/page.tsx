"use client"

import { Suspense, useState } from "react"
import { zodResolver } from "@hookform/resolvers/zod"
import Link from "next/link"
import { useSearchParams } from "next/navigation"
import { useForm } from "react-hook-form"

import { Logo } from "@/components/logo"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import {
    Field,
    FieldDescription,
    FieldError,
    FieldGroup,
    FieldLabel,
    FieldSeparator,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Verification } from "@/components/verification"
import { googleAuthUrl } from "@/lib/google-auth-url"
import { oauthErrorMessage } from "@/lib/oauth-errors"
import { type SendOtpInput, sendOtpSchema } from "@/lib/validation/auth"

function OAuthError() {
    const searchParams = useSearchParams()
    const message = oauthErrorMessage(searchParams.get("error") ?? undefined)
    if (!message) return null
    return (
        <p className="rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-center text-sm text-destructive">
            {message}
        </p>
    )
}

export default function LoginPage() {
    const [email, setEmail] = useState<string | null>(null)
    const {
        register,
        handleSubmit,
        formState: { errors },
    } = useForm<SendOtpInput>({ resolver: zodResolver(sendOtpSchema) })

    const onSubmit = handleSubmit((values) => setEmail(values.email))

    return (
        <div className="flex min-h-svh items-center justify-center p-6">
            <div className="w-full max-w-md">
                <Card className="py-0">
                    <CardContent className="p-6 md:p-8">
                        <FieldGroup>
                            <div className="flex flex-col items-center gap-2 text-center">
                                <Logo className="h-9 w-9" />
                                <h1 className="text-2xl font-bold">
                                    {email ? "Verify it's you" : "Welcome back"}
                                </h1>
                                <p className="text-balance text-sm text-muted-foreground">
                                    {email
                                        ? "Choose how you'd like to verify."
                                        : "Sign in to manage your organization's events."}
                                </p>
                            </div>

                            <Suspense fallback={null}>
                                <OAuthError />
                            </Suspense>

                            {email ? (
                                <>
                                    <Verification email={email} mode="login" />
                                    <button
                                        type="button"
                                        onClick={() => setEmail(null)}
                                        className="text-center text-sm text-muted-foreground underline underline-offset-4"
                                    >
                                        Use a different email
                                    </button>
                                </>
                            ) : (
                                <form onSubmit={onSubmit} noValidate>
                                    <FieldGroup>
                                        <Field data-invalid={!!errors.email}>
                                            <FieldLabel htmlFor="email">Email</FieldLabel>
                                            <Input
                                                id="email"
                                                type="email"
                                                autoComplete="email"
                                                placeholder="you@example.com"
                                                aria-invalid={!!errors.email}
                                                {...register("email")}
                                            />
                                            <FieldDescription>Enter your registered account email.</FieldDescription>
                                            <FieldError errors={[errors.email]} />
                                        </Field>
                                        <Field>
                                            <Button type="submit">Continue with Email</Button>
                                        </Field>

                                        <FieldSeparator className="*:data-[slot=field-separator-content]:bg-card">
                                            Or continue with
                                        </FieldSeparator>

                                        <Field>
                                            <Button
                                                variant="outline"
                                                type="button"
                                                render={<a href={googleAuthUrl("login")} />}
                                            >
                                                <GoogleLogo />
                                                Continue with Google
                                            </Button>
                                        </Field>

                                        <FieldDescription className="text-center">
                                            Don&apos;t have an account?{" "}
                                            <Link href="/register" className="underline underline-offset-4">
                                                Sign up
                                            </Link>
                                        </FieldDescription>
                                    </FieldGroup>
                                </form>
                            )}
                        </FieldGroup>
                    </CardContent>
                </Card>
            </div>
        </div>
    )
}

const GoogleLogo = () => (
    <svg
        className="inline-block size-lg shrink-0 align-sub text-inherit"
        fill="none"
        height="1.2em"
        viewBox="0 0 16 16"
        width="1.2em"
        xmlns="http://www.w3.org/2000/svg"
    >
        <g clipPath="url(#login-google-clip)">
            <path
                d="M15.6823 8.18368C15.6823 7.63986 15.6382 7.0931 15.5442 6.55811H7.99829V9.63876H12.3194C12.1401 10.6323 11.564 11.5113 10.7203 12.0698V14.0687H13.2983C14.8122 12.6753 15.6823 10.6176 15.6823 8.18368Z"
                fill="#4285F4"
            />
            <path
                d="M7.99812 16C10.1558 16 11.9753 15.2915 13.3011 14.0687L10.7231 12.0698C10.0058 12.5578 9.07988 12.8341 8.00106 12.8341C5.91398 12.8341 4.14436 11.426 3.50942 9.53296H0.849121V11.5936C2.2072 14.295 4.97332 16 7.99812 16Z"
                fill="#34A853"
            />
            <path
                d="M3.50665 9.53295C3.17154 8.53938 3.17154 7.4635 3.50665 6.46993V4.4093H0.849292C-0.285376 6.66982 -0.285376 9.33306 0.849292 11.5936L3.50665 9.53295Z"
                fill="#FBBC04"
            />
            <path
                d="M7.99812 3.16589C9.13867 3.14825 10.241 3.57743 11.067 4.36523L13.3511 2.0812C11.9048 0.723121 9.98526 -0.0235266 7.99812 -1.02057e-05C4.97332 -1.02057e-05 2.2072 1.70493 0.849121 4.40932L3.50648 6.46995C4.13848 4.57394 5.91104 3.16589 7.99812 3.16589Z"
                fill="#EA4335"
            />
        </g>
        <defs>
            <clipPath id="login-google-clip">
                <rect fill="white" height="16" width="15.6825" />
            </clipPath>
        </defs>
    </svg>
)
