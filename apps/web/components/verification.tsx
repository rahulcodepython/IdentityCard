"use client"

import { useEffect, useRef, useState, useTransition } from "react"
import { toast } from "sonner"
import { RiLoader4Line } from "@remixicon/react"

import {
    sendOtpAction,
    verifyOtpAction,
    verifyTotpAction,
} from "@/app/(auth)/actions"
import { Button } from "@/components/ui/button"
import { InputOTP, InputOTPGroup, InputOTPSlot } from "@/components/ui/input-otp"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

const SLOTS = [0, 1, 2, 3, 4, 5]

// Shared by the login and register flows. Login always shows both tabs,
// whether or not this account ever enrolled TOTP — an un-enrolled attempt
// just fails and the user falls back to the email-code tab. Register's
// TOTP tab additionally shows the QR/secret for first-time enrollment.
export function Verification({
    email,
    mode,
    totpQrImage,
    totpSecret,
}: {
    email: string
    mode: "login" | "register"
    totpQrImage?: string
    totpSecret?: string
}) {
    return (
        <Tabs defaultValue="otp" className="w-full">
            <TabsList className="w-full">
                <TabsTrigger value="otp" className="flex-1">
                    Email code
                </TabsTrigger>
                <TabsTrigger value="totp" className="flex-1">
                    Authenticator app
                </TabsTrigger>
            </TabsList>
            <TabsContent value="otp" className="mt-5">
                <OtpTab email={email} />
            </TabsContent>
            <TabsContent value="totp" className="mt-5">
                <TotpTab
                    email={email}
                    mode={mode}
                    qrImage={totpQrImage}
                    secret={totpSecret}
                />
            </TabsContent>
        </Tabs>
    )
}

function CodeInput({
    value,
    onChange,
    disabled,
}: {
    value: string
    onChange: (value: string) => void
    disabled?: boolean
}) {
    return (
        <InputOTP maxLength={6} value={value} onChange={onChange} disabled={disabled}>
            <InputOTPGroup>
                {SLOTS.map((i) => (
                    <InputOTPSlot key={i} index={i} />
                ))}
            </InputOTPGroup>
        </InputOTP>
    )
}

function OtpTab({ email }: { email: string }) {
    const [sent, setSent] = useState(false)
    const [code, setCode] = useState("")
    const [isPending, startTransition] = useTransition()
    const lastAttemptedRef = useRef("")

    const send = () => {
        startTransition(async () => {
            const result = await sendOtpAction({ email })
            if (result?.error) {
                toast.error(result.error)
                return
            }
            setSent(true)
            toast.success(`Verification code sent to ${email}`)
        })
    }

    // Auto-submit API call when 6th digit is entered
    useEffect(() => {
        if (code.length === 6) {
            if (code !== lastAttemptedRef.current) {
                lastAttemptedRef.current = code
                startTransition(async () => {
                    const result = await verifyOtpAction({ email, code })
                    if (result?.error) {
                        toast.error(result.error)
                    }
                })
            }
        } else {
            // Reset last attempted code so backspacing & re-typing 6th digit fires API call again
            lastAttemptedRef.current = ""
        }
    }, [code, email])

    if (!sent) {
        return (
            <div className="flex flex-col items-center gap-4 text-center">
                <p className="text-sm text-muted-foreground">
                    We&apos;ll email a 6-digit code to{" "}
                    <span className="font-medium text-foreground">{email}</span>.
                </p>
                <Button className="w-full" onClick={send} disabled={isPending}>
                    {isPending ? "Sending…" : "Send code"}
                </Button>
            </div>
        )
    }

    return (
        <div className="flex flex-col items-center gap-4 text-center">
            <p className="text-sm text-muted-foreground">
                Enter the code sent to{" "}
                <span className="font-medium text-foreground">{email}</span>.
            </p>

            <CodeInput value={code} onChange={setCode} disabled={isPending} />

            {isPending && (
                <div className="flex items-center gap-2 text-xs text-primary animate-pulse">
                    <RiLoader4Line className="size-4 animate-spin" />
                    <span>Verifying code…</span>
                </div>
            )}

            <button
                type="button"
                onClick={send}
                disabled={isPending}
                className="mt-2 text-xs text-muted-foreground underline underline-offset-4 disabled:opacity-50 hover:text-foreground transition-colors"
            >
                Resend code
            </button>
        </div>
    )
}

function TotpTab({
    email,
    mode,
    qrImage,
    secret,
}: {
    email: string
    mode: "login" | "register"
    qrImage?: string
    secret?: string
}) {
    const [code, setCode] = useState("")
    const [isPending, startTransition] = useTransition()
    const lastAttemptedRef = useRef("")

    // Auto-submit API call when 6th digit is entered
    useEffect(() => {
        if (code.length === 6) {
            if (code !== lastAttemptedRef.current) {
                lastAttemptedRef.current = code
                startTransition(async () => {
                    const result = await verifyTotpAction({ email, code })
                    if (result?.error) {
                        toast.error(result.error)
                    }
                })
            }
        } else {
            // Reset last attempted code so backspacing & re-typing 6th digit fires API call again
            lastAttemptedRef.current = ""
        }
    }, [code, email])

    return (
        <div className="flex flex-col items-center gap-4 text-center">
            {mode === "register" && qrImage ? (
                <div className="flex flex-col items-center gap-2">
                    {/* eslint-disable-next-line @next/next/no-img-element -- data: URI, not a Next-optimizable asset */}
                    <img
                        src={qrImage}
                        alt="Scan with your authenticator app"
                        className="size-40 rounded-lg border p-2"
                    />
                    <p className="max-w-64 text-xs text-muted-foreground">
                        Scan with Google Authenticator (or any TOTP app).
                        {secret && (
                            <>
                                {" "}
                                Can&apos;t scan? Enter this code manually:{" "}
                                <code className="rounded bg-muted px-1 py-0.5">{secret}</code>
                            </>
                        )}
                    </p>
                </div>
            ) : (
                <p className="text-sm text-muted-foreground">
                    Enter the 6-digit code from your authenticator app.
                </p>
            )}

            <CodeInput value={code} onChange={setCode} disabled={isPending} />

            {isPending && (
                <div className="flex items-center gap-2 text-xs text-primary animate-pulse">
                    <RiLoader4Line className="size-4 animate-spin" />
                    <span>Verifying authenticator code…</span>
                </div>
            )}
        </div>
    )
}
