"use client"

import { useEffect, useState } from "react"
import { toast } from "sonner"
import { RiLoader4Line } from "@remixicon/react"
import QRCode from "qrcode"

import { authClient } from "@/lib/auth-client"
import { Button } from "@/components/ui/button"
import { InputOTP, InputOTPGroup, InputOTPSlot } from "@/components/ui/input-otp"

const SLOTS = [0, 1, 2, 3, 4, 5]

type Step = "request" | "otp" | "enroll-totp" | "verify-totp"

// Email OTP is the one credential — better-auth auto-creates the account
// on first sign-in (see signIn.emailOtp's `name` param). TOTP is a true
// second factor layered on top of it, not an independent alternative
// like the old app's tabs: the first-ever successful OTP sign-in forces
// enrollment (an authenticator app is now required going forward), and
// every sign-in after that requires a TOTP code too, since
// authClient.twoFactor.verifyTotp works the same way whether it's
// enrolling or re-verifying an already-trusted secret (see
// better-auth's totp2fa plugin — it only branches on the two-factor
// row's own `verified` flag, not on how the caller got its session).
export function Verification({
    email,
    name,
    onVerified,
}: {
    email: string
    name?: string
    onVerified: () => void
}) {
    const [step, setStep] = useState<Step>("request")
    const [code, setCode] = useState("")
    const [pending, setPending] = useState(false)
    const [totpUri, setTotpUri] = useState<string | null>(null)
    const [qrDataUrl, setQrDataUrl] = useState<string | null>(null)

    useEffect(() => {
        if (!totpUri) {
            setQrDataUrl(null)
            return
        }
        let cancelled = false
        void QRCode.toDataURL(totpUri).then((url) => {
            if (!cancelled) setQrDataUrl(url)
        })
        return () => {
            cancelled = true
        }
    }, [totpUri])

    async function sendCode() {
        setPending(true)
        const { error } = await authClient.emailOtp.sendVerificationOtp({ email, type: "sign-in" })
        setPending(false)
        if (error) {
            toast.error(error.message ?? "Something went wrong.")
            return
        }
        setStep("otp")
        toast.success(`Verification code sent to ${email}`)
    }

    async function submitOtp(value: string) {
        setPending(true)
        const { data, error } = await authClient.signIn.emailOtp({ email, otp: value, name })
        if (error || !data) {
            setPending(false)
            toast.error(error?.message ?? "Invalid or expired code.")
            setCode("")
            return
        }

        if (data.user.twoFactorEnabled) {
            setPending(false)
            setCode("")
            setStep("verify-totp")
            return
        }

        const { data: enableData, error: enableError } = await authClient.twoFactor.enable({
            method: "totp",
            issuer: "IdentityCard",
        })
        setPending(false)
        if (enableError || enableData?.method !== "totp") {
            toast.error(enableError?.message ?? "Couldn't start authenticator setup.")
            return
        }
        setTotpUri(enableData.totpURI)
        setCode("")
        setStep("enroll-totp")
    }

    async function submitTotp(value: string) {
        setPending(true)
        const { error } = await authClient.twoFactor.verifyTotp({ code: value, trustDevice: true })
        setPending(false)
        if (error) {
            toast.error(error.message ?? "Invalid code.")
            setCode("")
            return
        }
        onVerified()
    }

    if (step === "request") {
        return (
            <div className="flex flex-col items-center gap-4 text-center">
                <p className="text-sm text-muted-foreground">
                    We&apos;ll email a 6-digit code to{" "}
                    <span className="font-medium text-foreground">{email}</span>.
                </p>
                <Button className="w-full" onClick={() => void sendCode()} disabled={pending}>
                    {pending ? "Sending…" : "Send code"}
                </Button>
            </div>
        )
    }

    if (step === "otp") {
        return (
            <div className="flex flex-col items-center gap-4 text-center">
                <p className="text-sm text-muted-foreground">
                    Enter the code sent to <span className="font-medium text-foreground">{email}</span>.
                </p>
                <CodeInput
                    value={code}
                    disabled={pending}
                    onChange={(value) => {
                        setCode(value)
                        if (value.length === 6) void submitOtp(value)
                    }}
                />
                {pending && <Spinner label="Verifying code…" />}
                <button
                    type="button"
                    onClick={() => void sendCode()}
                    disabled={pending}
                    className="mt-2 text-xs text-muted-foreground underline underline-offset-4 disabled:opacity-50 hover:text-foreground transition-colors"
                >
                    Resend code
                </button>
            </div>
        )
    }

    if (step === "enroll-totp") {
        const secret = totpUri ? new URLSearchParams(totpUri.split("?")[1]).get("secret") : null
        return (
            <div className="flex flex-col items-center gap-4 text-center">
                <div className="flex flex-col items-center gap-2">
                    {qrDataUrl && (
                        // eslint-disable-next-line @next/next/no-img-element -- data: URI, not a Next-optimizable asset
                        <img
                            src={qrDataUrl}
                            alt="Scan with your authenticator app"
                            className="size-40 rounded-lg border p-2"
                        />
                    )}
                    <p className="max-w-64 text-xs text-muted-foreground">
                        Set up an authenticator app (Google Authenticator or similar) — required once, the first
                        time you sign in.
                        {secret && (
                            <>
                                {" "}
                                Can&apos;t scan? Enter this code manually:{" "}
                                <code className="rounded bg-muted px-1 py-0.5">{secret}</code>
                            </>
                        )}
                    </p>
                </div>
                <CodeInput
                    value={code}
                    disabled={pending}
                    onChange={(value) => {
                        setCode(value)
                        if (value.length === 6) void submitTotp(value)
                    }}
                />
                {pending && <Spinner label="Verifying…" />}
            </div>
        )
    }

    return (
        <div className="flex flex-col items-center gap-4 text-center">
            <p className="text-sm text-muted-foreground">Enter the 6-digit code from your authenticator app.</p>
            <CodeInput
                value={code}
                disabled={pending}
                onChange={(value) => {
                    setCode(value)
                    if (value.length === 6) void submitTotp(value)
                }}
            />
            {pending && <Spinner label="Verifying…" />}
        </div>
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

function Spinner({ label }: { label: string }) {
    return (
        <div className="flex items-center gap-2 text-xs text-primary animate-pulse">
            <RiLoader4Line className="size-4 animate-spin" />
            <span>{label}</span>
        </div>
    )
}
