"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { toast } from "sonner"
import { RiLoader4Line } from "@remixicon/react"

import { Button } from "@/components/ui/button"
import {
    Field,
    FieldDescription,
    FieldError,
    FieldGroup,
    FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { OtpInput } from "@/components/auth/otp-input"
import { authClient } from "@/lib/auth-client"
import { useSessionStore } from "@/store/session.store"
import {
    type SendOtpInput,
    type RegisterInput,
    sendOtpSchema,
    registerSchema,
} from "@/schema/auth.types"

interface EmailAuthFlowProps {
    mode: "login" | "register"
    onSuccess?: () => void
}

export function EmailAuthFlow({ mode, onSuccess }: EmailAuthFlowProps) {
    const router = useRouter()
    const [step, setStep] = useState<"form" | "otp">("form")
    const [email, setEmail] = useState("")
    const [name, setName] = useState("")
    const [organizationName, setOrganizationName] = useState("")
    const [otpCode, setOtpCode] = useState("")
    const [pending, setPending] = useState(false)
    const [errorMessage, setErrorMessage] = useState<string | null>(null)

    const loginForm = useForm<SendOtpInput>({
        resolver: zodResolver(sendOtpSchema),
        defaultValues: { email: "" },
    })

    const registerForm = useForm<RegisterInput>({
        resolver: zodResolver(registerSchema),
        defaultValues: { name: "", organizationName: "", email: "" },
    })

    const handleInitialSubmit = async (targetEmail: string, targetName?: string, targetOrg?: string) => {
        setPending(true)
        setErrorMessage(null)

        const { error } = await authClient.emailOtp.sendVerificationOtp({
            email: targetEmail,
            type: "sign-in",
        })

        setPending(false)

        if (error) {
            setErrorMessage(error.message ?? "Failed to send verification code.")
            toast.error(error.message ?? "Failed to send verification code.")
            return
        }

        setEmail(targetEmail)
        if (targetName) setName(targetName)
        if (targetOrg) setOrganizationName(targetOrg)
        setStep("otp")
        setOtpCode("")
        toast.success(`Verification code sent to ${targetEmail}`)
    }

    const onLoginSubmit = loginForm.handleSubmit(async (values) => {
        await handleInitialSubmit(values.email)
    })

    const onRegisterSubmit = registerForm.handleSubmit(async (values) => {
        await handleInitialSubmit(values.email, values.name, values.organizationName)
    })

    const handleVerifyOtp = async (code: string) => {
        if (pending) return
        setPending(true)
        setErrorMessage(null)

        try {
            const { data, error } = await authClient.signIn.emailOtp({
                email,
                otp: code,
                name: name || undefined,
            })

            if (error || !data) {
                setPending(false)
                setErrorMessage(error?.message ?? "Invalid or expired code.")
                toast.error(error?.message ?? "Invalid or expired code.")
                setOtpCode("")
                return
            }

            if (mode === "register" && organizationName) {
                try {
                    await fetch("/api/organization/setup", {
                        method: "POST",
                        headers: { "Content-Type": "application/json" },
                        body: JSON.stringify({ organizationName, userEmail: email }),
                    })
                } catch (err) {
                    console.error("Failed to associate organization:", err)
                }
            }

            try {
                const { data: tokenData } = await authClient.token()
                const { data: sessionData } = await authClient.getSession()
                if (sessionData?.user && tokenData?.token) {
                    useSessionStore.getState().setSession({
                        token: tokenData.token,
                        user: {
                            id: sessionData.user.id,
                            name: sessionData.user.name,
                            email: sessionData.user.email,
                            image: sessionData.user.image,
                        },
                        activeOrganizationId:
                            (sessionData.session as { activeOrganizationId?: string })?.activeOrganizationId ?? null,
                        role: null,
                    })
                }
            } catch (sessErr) {
                console.warn("Failed to cache session after sign-in:", sessErr)
            }

            toast.success(mode === "register" ? "Account created successfully!" : "Signed in successfully!")

            if (onSuccess) {
                onSuccess()
            } else {
                router.push("/dashboard")
            }
        } catch (err) {
            setPending(false)
            const msg = err instanceof Error ? err.message : "Verification failed."
            setErrorMessage(msg)
            toast.error(msg)
            setOtpCode("")
        }
    }

    if (step === "otp") {
        return (
            <div className="flex flex-col items-center gap-4 text-center">
                {errorMessage && (
                    <p className="w-full rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-center text-xs text-destructive">
                        {errorMessage}
                    </p>
                )}

                <div className="space-y-1">
                    <p className="text-sm font-medium">Enter 6-digit verification code</p>
                    <p className="text-xs text-muted-foreground">
                        Sent to <span className="font-semibold text-foreground">{email}</span>
                    </p>
                </div>

                <OtpInput
                    value={otpCode}
                    disabled={pending}
                    onChange={setOtpCode}
                    onComplete={handleVerifyOtp}
                />

                {pending ? (
                    <div className="flex animate-pulse items-center gap-2 text-xs font-medium text-primary">
                        <RiLoader4Line className="size-4 animate-spin" />
                        <span>Verifying code and logging in…</span>
                    </div>
                ) : (
                    <div className="flex flex-col items-center gap-2">
                        <button
                            type="button"
                            onClick={() => void handleInitialSubmit(email, name, organizationName)}
                            disabled={pending}
                            className="text-xs text-muted-foreground underline underline-offset-4 transition-colors hover:text-foreground disabled:opacity-50"
                        >
                            Resend code
                        </button>
                        <button
                            type="button"
                            onClick={() => {
                                setStep("form")
                                setOtpCode("")
                                setErrorMessage(null)
                            }}
                            disabled={pending}
                            className="text-xs text-muted-foreground transition-colors hover:text-foreground"
                        >
                            ← Use different {mode === "register" ? "details" : "email"}
                        </button>
                    </div>
                )}
            </div>
        )
    }

    if (mode === "register") {
        const { errors } = registerForm.formState
        return (
            <form onSubmit={onRegisterSubmit} noValidate>
                <FieldGroup>
                    {errorMessage && (
                        <p className="rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-center text-xs text-destructive">
                            {errorMessage}
                        </p>
                    )}

                    <Field data-invalid={!!errors.name}>
                        <FieldLabel htmlFor="email-reg-name">Your name</FieldLabel>
                        <Input
                            id="email-reg-name"
                            autoComplete="name"
                            placeholder="e.g. Alex Smith"
                            aria-invalid={!!errors.name}
                            {...registerForm.register("name")}
                        />
                        <FieldError errors={[errors.name]} />
                    </Field>

                    <Field data-invalid={!!errors.organizationName}>
                        <FieldLabel htmlFor="email-reg-org">Organization name</FieldLabel>
                        <Input
                            id="email-reg-org"
                            placeholder="e.g. Acme Corp"
                            aria-invalid={!!errors.organizationName}
                            {...registerForm.register("organizationName")}
                        />
                        <FieldError errors={[errors.organizationName]} />
                    </Field>

                    <Field data-invalid={!!errors.email}>
                        <FieldLabel htmlFor="email-reg-email">Email</FieldLabel>
                        <Input
                            id="email-reg-email"
                            type="email"
                            autoComplete="email"
                            placeholder="you@example.com"
                            aria-invalid={!!errors.email}
                            {...registerForm.register("email")}
                        />
                        <FieldDescription>We&apos;ll send your login code here.</FieldDescription>
                        <FieldError errors={[errors.email]} />
                    </Field>

                    <Field>
                        <Button type="submit" disabled={pending}>
                            {pending ? (
                                <div className="flex items-center gap-2">
                                    <RiLoader4Line className="size-4 animate-spin" />
                                    <span>Sending code…</span>
                                </div>
                            ) : (
                                "Continue with Email OTP"
                            )}
                        </Button>
                    </Field>
                </FieldGroup>
            </form>
        )
    }

    const { errors } = loginForm.formState
    return (
        <form onSubmit={onLoginSubmit} noValidate>
            <FieldGroup>
                {errorMessage && (
                    <p className="rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-center text-xs text-destructive">
                        {errorMessage}
                    </p>
                )}

                <Field data-invalid={!!errors.email}>
                    <FieldLabel htmlFor="email-login-email">Email</FieldLabel>
                    <Input
                        id="email-login-email"
                        type="email"
                        autoComplete="email"
                        placeholder="you@example.com"
                        aria-invalid={!!errors.email}
                        {...loginForm.register("email")}
                    />
                    <FieldDescription>Enter your account email to sign in.</FieldDescription>
                    <FieldError errors={[errors.email]} />
                </Field>

                <Field>
                    <Button type="submit" disabled={pending}>
                        {pending ? (
                            <div className="flex items-center gap-2">
                                <RiLoader4Line className="size-4 animate-spin" />
                                <span>Sending code…</span>
                            </div>
                        ) : (
                            "Continue with Email"
                        )}
                    </Button>
                </Field>
            </FieldGroup>
        </form>
    )
}
