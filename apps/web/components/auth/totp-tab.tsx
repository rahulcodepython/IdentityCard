import type { ReactNode } from "react"
import { RiLoader4Line } from "@remixicon/react"
import { TabsContent } from "@/components/ui/tabs"
import { OtpInput } from "@/components/auth/otp-input"
import { TotpQrView } from "@/components/auth/totp-qr-view"
import { LoadingSpinner, BackToEmailButton } from "./auth-ui-elements"

interface TotpTabContentProps {
    mode: "login" | "register"
    code: string
    pending: boolean
    totpLoading: boolean
    totpUri: string | null
    totpSecret: string | null
    onCodeChange: (value: string) => void
    onVerify: (code: string) => void
    onChangeEmail: () => void
}

export function TotpTabContent({
    mode,
    code,
    pending,
    totpLoading,
    totpUri,
    totpSecret,
    onCodeChange,
    onVerify,
    onChangeEmail,
}: TotpTabContentProps) {
    const renderTotpContent = (): ReactNode => {
        if (mode === "register") {
            if (totpLoading) {
                return (
                    <div className="flex flex-col items-center gap-2 py-8">
                        <RiLoader4Line className="size-6 animate-spin text-muted-foreground" />
                        <span className="text-xs text-muted-foreground">Setting up authenticator…</span>
                    </div>
                )
            }

            if (!totpUri) {
                return (
                    <div className="rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-xs text-destructive">
                        Could not initialize authenticator. Please use Email OTP instead.
                    </div>
                )
            }

            return (
                <div className="flex flex-col items-center gap-4">
                    <TotpQrView totpUri={totpUri} secret={totpSecret} />
                    <p className="text-sm font-medium">Enter 6-digit code from your app:</p>
                    <OtpInput
                        value={code}
                        disabled={pending}
                        onChange={onCodeChange}
                        onComplete={onVerify}
                    />
                </div>
            )
        }

        return (
            <div className="flex flex-col items-center gap-3">
                <div className="space-y-1">
                    <p className="text-sm font-medium">Enter 6-digit authenticator code</p>
                    <p className="text-xs text-muted-foreground">
                        From Google Authenticator, Authy, or your password manager
                    </p>
                </div>

                <OtpInput
                    value={code}
                    disabled={pending}
                    onChange={onCodeChange}
                    onComplete={onVerify}
                />
            </div>
        )
    }

    return (
        <TabsContent value="totp" className="mt-4 flex flex-col items-center gap-4">
            {renderTotpContent()}

            {
                pending ?
                    <LoadingSpinner message="Verifying code and logging in…" />
                    : <div className="flex flex-col items-center gap-2">
                        <span className="text-xs text-muted-foreground">
                            Codes refresh automatically in your app
                        </span>
                        <BackToEmailButton onClick={onChangeEmail} disabled={pending} />
                    </div>
            }
        </TabsContent>
    )
}