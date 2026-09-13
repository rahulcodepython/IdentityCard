import { TabsContent } from "@/components/ui/tabs"
import { OtpInput } from "@/components/auth/otp-input"
import { LoadingSpinner, BackToEmailButton } from "./auth-ui-elements"

interface EmailOtpTabContentProps {
    email: string
    code: string
    pending: boolean
    resendCooldown: number
    onCodeChange: (value: string) => void
    onVerify: (code: string) => void
    onResend: () => void
    onChangeEmail: () => void
}

export function EmailOtpTabContent({
    email,
    code,
    pending,
    resendCooldown,
    onCodeChange,
    onVerify,
    onResend,
    onChangeEmail,
}: EmailOtpTabContentProps) {
    return (
        <TabsContent value="email" className="mt-4 flex flex-col items-center gap-4">
            <div className="space-y-1">
                <p className="text-sm font-medium">Enter 6-digit verification code</p>
                <p className="text-xs text-muted-foreground">
                    Sent to <span className="font-semibold text-foreground">{email}</span>
                </p>
            </div>

            <OtpInput
                value={code}
                disabled={pending}
                onChange={onCodeChange}
                onComplete={onVerify}
            />

            {
                pending ?
                    <LoadingSpinner message="Verifying code and logging in…" />
                    : <div className="flex flex-col items-center gap-2">
                        <button
                            type="button"
                            onClick={onResend}
                            disabled={pending || resendCooldown > 0}
                            className="text-xs text-muted-foreground underline underline-offset-4 transition-colors hover:text-foreground disabled:opacity-50 disabled:no-underline"
                        >
                            {
                                resendCooldown > 0
                                    ? `Resend code in ${resendCooldown}s`
                                    : "Didn't receive code? Resend code"
                            }
                        </button>
                        <BackToEmailButton onClick={onChangeEmail} disabled={pending} />
                    </div>
            }
        </TabsContent>
    )
}