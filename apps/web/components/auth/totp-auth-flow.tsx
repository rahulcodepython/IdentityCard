"use client"

import { useState } from "react"
import { useForm, Controller } from "react-hook-form"
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
import { TotpQrView } from "@/components/auth/totp-qr-view"
import {
    prepareTotpRegistration,
    completeTotpRegistration,
    signInWithTotp,
} from "@/lib/actions/totp-auth"
import { useSessionStore } from "@/store/session.store"
import {
    type RegisterInput,
    type TotpLoginInput,
    registerSchema,
    totpLoginSchema,
} from "@/schema/auth.types"

interface TotpAuthFlowProps {
    mode: "login" | "register"
    onSuccess?: () => void
}

export function TotpAuthFlow({ mode, onSuccess }: TotpAuthFlowProps) {
    const [step, setStep] = useState<"form" | "verify">("form")
    const [email, setEmail] = useState("")
    const [totpUri, setTotpUri] = useState<string | null>(null)
    const [secret, setSecret] = useState<string | null>(null)
    const [code, setCode] = useState("")
    const [pending, setPending] = useState(false)
    const [errorMessage, setErrorMessage] = useState<string | null>(null)

    const registerForm = useForm<RegisterInput>({
        resolver: zodResolver(registerSchema),
        defaultValues: { name: "", organizationName: "", email: "" },
    })

    const loginForm = useForm<TotpLoginInput>({
        resolver: zodResolver(totpLoginSchema),
        defaultValues: { email: "", code: "" },
    })

    const onPrepareRegister = registerForm.handleSubmit(async (values) => {
        setPending(true)
        setErrorMessage(null)

        const result = await prepareTotpRegistration({
            email: values.email,
            name: values.name,
            organizationName: values.organizationName,
        })
        setPending(false)

        if (!result.success) {
            setErrorMessage(result.error)
            toast.error(result.error)
            return
        }

        setEmail(values.email)
        setTotpUri(result.totpURI)
        setSecret(result.secret)
        setStep("verify")
        setCode("")
        toast.success("Scan the QR code with your authenticator app.")
    })

    const handleCompleteRegistration = async (verifyCode: string) => {
        if (pending) return
        setPending(true)
        setErrorMessage(null)

        const result = await completeTotpRegistration({
            email,
            code: verifyCode,
        })
        setPending(false)

        if (!result.success) {
            setErrorMessage(result.error)
            toast.error(result.error)
            setCode("")
            return
        }

        toast.success("Account created successfully with Authenticator!")
        useSessionStore.getState().reset()

        if (onSuccess) {
            onSuccess()
        } else {
            window.location.assign("/dashboard")
        }
    }

    const handleLoginWithCode = async (targetEmail: string, verifyCode: string) => {
        if (pending) return
        setPending(true)
        setErrorMessage(null)

        const result = await signInWithTotp({
            email: targetEmail,
            code: verifyCode,
        })
        setPending(false)

        if (!result.success) {
            setErrorMessage(result.error)
            toast.error(result.error)
            loginForm.setValue("code", "")
            return
        }

        toast.success("Signed in successfully!")
        useSessionStore.getState().reset()

        if (onSuccess) {
            onSuccess()
        } else {
            window.location.assign("/dashboard")
        }
    }

    const onLoginSubmit = loginForm.handleSubmit(async (values) => {
        await handleLoginWithCode(values.email, values.code)
    })

    if (mode === "register" && step === "verify" && totpUri) {
        return (
            <div className="flex flex-col items-center gap-4 text-center">
                {errorMessage && (
                    <p className="w-full rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-center text-xs text-destructive">
                        {errorMessage}
                    </p>
                )}

                <TotpQrView totpUri={totpUri} secret={secret} />

                <div className="flex w-full flex-col items-center gap-2">
                    <p className="text-sm font-medium">Enter 6-digit code from your app:</p>
                    <OtpInput
                        value={code}
                        disabled={pending}
                        onChange={setCode}
                        onComplete={handleCompleteRegistration}
                    />
                </div>

                {pending ? (
                    <div className="flex animate-pulse items-center gap-2 text-xs font-medium text-primary">
                        <RiLoader4Line className="size-4 animate-spin" />
                        <span>Verifying and creating account…</span>
                    </div>
                ) : (
                    <button
                        type="button"
                        onClick={() => {
                            setStep("form")
                            setCode("")
                            setErrorMessage(null)
                        }}
                        disabled={pending}
                        className="text-center text-xs text-muted-foreground underline underline-offset-4 transition-colors hover:text-foreground"
                    >
                        ← Back to registration details
                    </button>
                )}
            </div>
        )
    }

    if (mode === "register") {
        const { errors } = registerForm.formState
        return (
            <form onSubmit={onPrepareRegister} noValidate>
                <FieldGroup>
                    <div className="rounded-lg border border-primary/20 bg-primary/5 p-3 text-xs text-muted-foreground">
                        As this is registration and no authenticator is linked yet, you will be shown a QR code to set up your authenticator app.
                    </div>

                    {errorMessage && (
                        <p className="rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-center text-xs text-destructive">
                            {errorMessage}
                        </p>
                    )}

                    <Field data-invalid={!!errors.name}>
                        <FieldLabel htmlFor="totp-reg-name">Your name</FieldLabel>
                        <Input
                            id="totp-reg-name"
                            autoComplete="name"
                            placeholder="e.g. Alex Smith"
                            aria-invalid={!!errors.name}
                            {...registerForm.register("name")}
                        />
                        <FieldError errors={[errors.name]} />
                    </Field>

                    <Field data-invalid={!!errors.organizationName}>
                        <FieldLabel htmlFor="totp-reg-org">Organization name</FieldLabel>
                        <Input
                            id="totp-reg-org"
                            placeholder="e.g. Acme Corp"
                            aria-invalid={!!errors.organizationName}
                            {...registerForm.register("organizationName")}
                        />
                        <FieldError errors={[errors.organizationName]} />
                    </Field>

                    <Field data-invalid={!!errors.email}>
                        <FieldLabel htmlFor="totp-reg-email">Email</FieldLabel>
                        <Input
                            id="totp-reg-email"
                            type="email"
                            autoComplete="email"
                            placeholder="you@example.com"
                            aria-invalid={!!errors.email}
                            {...registerForm.register("email")}
                        />
                        <FieldError errors={[errors.email]} />
                    </Field>

                    <Field>
                        <Button type="submit" disabled={pending}>
                            {pending ? (
                                <div className="flex items-center gap-2">
                                    <RiLoader4Line className="size-4 animate-spin" />
                                    <span>Generating setup…</span>
                                </div>
                            ) : (
                                "Set Up Authenticator"
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
                    <FieldLabel htmlFor="totp-login-email">Email</FieldLabel>
                    <Input
                        id="totp-login-email"
                        type="email"
                        autoComplete="email"
                        placeholder="you@example.com"
                        aria-invalid={!!errors.email}
                        {...loginForm.register("email")}
                    />
                    <FieldDescription>Enter your account email.</FieldDescription>
                    <FieldError errors={[errors.email]} />
                </Field>

                <Field data-invalid={!!errors.code}>
                    <FieldLabel htmlFor="totp-login-code">Authenticator Code</FieldLabel>
                    <Controller
                        name="code"
                        control={loginForm.control}
                        render={({ field }) => (
                            <OtpInput
                                id="totp-login-code"
                                value={field.value}
                                disabled={pending}
                                onChange={(val) => field.onChange(val)}
                                onComplete={(val) => {
                                    const currentEmail = loginForm.getValues("email")
                                    if (currentEmail) {
                                        void handleLoginWithCode(currentEmail, val)
                                    } else {
                                        void loginForm.trigger("email")
                                    }
                                }}
                            />
                        )}
                    />
                    <FieldDescription className="text-center">
                        Enter the 6-digit code from Google Authenticator, Authy, or Apple Passwords.
                    </FieldDescription>
                    <FieldError errors={[errors.code]} />
                </Field>

                <Field>
                    <Button type="submit" disabled={pending}>
                        {pending ? (
                            <div className="flex items-center gap-2">
                                <RiLoader4Line className="size-4 animate-spin" />
                                <span>Verifying code…</span>
                            </div>
                        ) : (
                            "Sign In with Authenticator"
                        )}
                    </Button>
                </Field>
            </FieldGroup>
        </form>
    )
}
