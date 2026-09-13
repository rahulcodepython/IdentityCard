import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import type { AuthTab } from "./auth.types"
import { ErrorAlert } from "./auth-ui-elements"
import { EmailOtpTabContent } from "./email-otp-tab"
import { TotpTabContent } from "./totp-tab"

interface VerificationPanelProps {
    mode: "login" | "register"
    email: string
    activeTab: AuthTab
    code: string
    pending: boolean
    errorMessage: string | null
    resendCooldown: number
    totpLoading: boolean
    totpUri: string | null
    totpSecret: string | null
    onTabChange: (tab: AuthTab) => void
    onCodeChange: (value: string) => void
    onVerify: (code: string) => void
    onResendEmailOtp: () => void
    onChangeEmail: () => void
}

export function VerificationPanel({
    mode,
    email,
    activeTab,
    code,
    pending,
    errorMessage,
    resendCooldown,
    totpLoading,
    totpUri,
    totpSecret,
    onTabChange,
    onCodeChange,
    onVerify,
    onResendEmailOtp,
    onChangeEmail,
}: VerificationPanelProps) {
    return (
        <div className="flex flex-col items-center gap-4 text-center">
            <ErrorAlert message={errorMessage} />

            <Tabs
                value={activeTab}
                onValueChange={(val) => onTabChange(val as AuthTab)}
                className="w-full"
            >
                <TabsList className="grid w-full grid-cols-2">
                    <TabsTrigger value="email">Email OTP</TabsTrigger>
                    <TabsTrigger value="totp">Authenticator (TOTP)</TabsTrigger>
                </TabsList>

                <EmailOtpTabContent
                    email={email}
                    code={code}
                    pending={pending}
                    resendCooldown={resendCooldown}
                    onCodeChange={onCodeChange}
                    onVerify={onVerify}
                    onResend={onResendEmailOtp}
                    onChangeEmail={onChangeEmail}
                />

                <TotpTabContent
                    mode={mode}
                    code={code}
                    pending={pending}
                    totpLoading={totpLoading}
                    totpUri={totpUri}
                    totpSecret={totpSecret}
                    onCodeChange={onCodeChange}
                    onVerify={onVerify}
                    onChangeEmail={onChangeEmail}
                />
            </Tabs>
        </div>
    )
}