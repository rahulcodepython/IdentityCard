"use client"

import { useEffect, useState } from "react"
import QRCode from "qrcode"

interface TotpQrViewProps {
    totpUri: string
    secret?: string | null
    description?: string
}

export function TotpQrView({
    totpUri,
    secret,
    description = "Scan this QR code with your authenticator app (Google Authenticator, Authy, or Apple Passwords).",
}: TotpQrViewProps) {
    const [qrDataUrl, setQrDataUrl] = useState<string | null>(null)

    useEffect(() => {
        if (!totpUri) return
        let cancelled = false
        void QRCode.toDataURL(totpUri).then((url) => {
            if (!cancelled) setQrDataUrl(url)
        })
        return () => {
            cancelled = true
        }
    }, [totpUri])

    const manualSecret =
        secret ??
        (totpUri ? new URLSearchParams(totpUri.split("?")[1]).get("secret") : null)

    return (
        <div className="flex flex-col items-center gap-3 text-center">
            {qrDataUrl && (
                // eslint-disable-next-line @next/next/no-img-element -- data: URI from QR generation
                <img
                    src={qrDataUrl}
                    alt="Scan with your authenticator app"
                    className="size-44 rounded-xl border bg-white p-2 shadow-sm"
                />
            )}
            <p className="max-w-xs text-xs text-muted-foreground">
                {description}
                {manualSecret && (
                    <>
                        {" "}
                        Can&apos;t scan? Enter manually:{" "}
                        <code className="select-all rounded bg-muted px-1.5 py-0.5 font-mono font-medium text-foreground">
                            {manualSecret}
                        </code>
                    </>
                )}
            </p>
        </div>
    )
}
